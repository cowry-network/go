package horizon

import (
	"errors"
	"fmt"

	"github.com/cowry-network/go/services/horizon/internal/actions"
	"github.com/cowry-network/go/services/horizon/internal/db2"
	"github.com/cowry-network/go/services/horizon/internal/db2/history"
	"github.com/cowry-network/go/services/horizon/internal/render/sse"
	"github.com/cowry-network/go/services/horizon/internal/resourceadapter"
	"github.com/cowry-network/go/support/render/hal"
	supportProblem "github.com/cowry-network/go/support/render/problem"
)

// Interface verifications
var _ actions.JSONer = (*PaymentsIndexAction)(nil)
var _ actions.EventStreamer = (*PaymentsIndexAction)(nil)

// PaymentsIndexAction returns a paged slice of payments based upon the provided
// filters
type PaymentsIndexAction struct {
	Action
	LedgerFilter      int32
	AccountFilter     string
	TransactionFilter string
	PagingParams      db2.PageQuery
	Records           []history.Operation
	Ledgers           *history.LedgerCache
	Page              hal.Page
	IncludeFailed     bool
}

// JSON is a method for actions.JSON
func (action *PaymentsIndexAction) JSON() error {
	action.Do(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.ValidateCursorWithinHistory,
		action.loadRecords,
		action.loadLedgers,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
	return action.Err
}

// SSE is a method for actions.SSE
func (action *PaymentsIndexAction) SSE(stream *sse.Stream) error {
	action.Setup(
		action.EnsureHistoryFreshness,
		action.loadParams,
		action.ValidateCursorWithinHistory,
	)
	action.Do(
		action.loadRecords,
		action.loadLedgers,
		func() {
			stream.SetLimit(int(action.PagingParams.Limit))
			records := action.Records[stream.SentCount():]

			for _, record := range records {
				ledger, found := action.Ledgers.Records[record.LedgerSequence()]
				if !found {
					action.Err = errors.New(fmt.Sprintf("could not find ledger data for sequence %d", record.LedgerSequence()))
					return
				}

				res, err := resourceadapter.NewOperation(action.R.Context(), record, ledger)
				if err != nil {
					action.Err = err
					return
				}

				stream.Send(sse.Event{
					ID:   res.PagingToken(),
					Data: res,
				})
			}
		},
	)

	return action.Err
}

func (action *PaymentsIndexAction) loadParams() {
	action.ValidateCursorAsDefault()
	action.AccountFilter = action.GetAddress("account_id")
	action.LedgerFilter = action.GetInt32("ledger_id")
	action.TransactionFilter = action.GetString("tx_id")
	action.PagingParams = action.GetPageQuery()
	action.IncludeFailed = action.GetBool("include_failed")

	if action.IncludeFailed == true && !action.App.config.IngestFailedTransactions {
		err := errors.New("`include_failed` parameter is unavailable when Horizon is not ingesting failed " +
			"transactions. Set `INGEST_FAILED_TRANSACTIONS=true` to start ingesting them.")
		action.Err = supportProblem.MakeInvalidFieldProblem("include_failed", err)
		return
	}
}

func (action *PaymentsIndexAction) loadRecords() {
	q := action.HistoryQ()
	ops := q.Operations().OnlyPayments()

	switch {
	case action.AccountFilter != "":
		ops.ForAccount(action.AccountFilter)
	case action.LedgerFilter > 0:
		ops.ForLedger(action.LedgerFilter)
	case action.TransactionFilter != "":
		ops.ForTransaction(action.TransactionFilter)
	}

	// When querying operations for transaction return both successful
	// and failed operations. We asume that because user is querying
	// this specific transactions, she knows it's status.
	if action.TransactionFilter == "" && !action.IncludeFailed {
		ops.SuccessfulOnly()
	}

	action.Err = ops.Page(action.PagingParams).Select(&action.Records)
}

// loadLedgers populates the ledger cache for this action
func (action *PaymentsIndexAction) loadLedgers() {
	action.Ledgers = &history.LedgerCache{}

	for _, op := range action.Records {
		action.Ledgers.Queue(op.LedgerSequence())
	}

	action.Err = action.Ledgers.Load(action.HistoryQ())
}

func (action *PaymentsIndexAction) loadPage() {
	for _, record := range action.Records {
		var res hal.Pageable

		ledger, found := action.Ledgers.Records[record.LedgerSequence()]
		if !found {
			msg := fmt.Sprintf("could not find ledger data for sequence %d", record.LedgerSequence())
			action.Err = errors.New(msg)
			return
		}

		res, action.Err = resourceadapter.NewOperation(action.R.Context(), record, ledger)
		if action.Err != nil {
			return
		}
		action.Page.Add(res)
	}

	action.Page.FullURL = action.FullURL()
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}
