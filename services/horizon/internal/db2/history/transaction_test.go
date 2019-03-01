package history

import (
	"database/sql"
	"testing"

	"github.com/stellar/go/services/horizon/internal/test"
)

func TestTransactionQueries(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.HorizonSession()}

	// Test TransactionByHash
	var tx Transaction
	real := "2374e99349b9ef7dba9a5db3339b78fda8f34777b1af33ba468ad5c0df946d4d"
	err := q.TransactionByHash(&tx, real)
	tt.Assert.NoError(err)

	fake := "not_real"
	err = q.TransactionByHash(&tx, fake)
	tt.Assert.Equal(err, sql.ErrNoRows)
}

// TestTransactionSuccessfulOnly tests if default query returns successful
// transactions only.
// If it's not enclosed in brackets, it may return incorrect result when mixed
// with `ForAccount` or `ForLedger` filters.
func TestTransactionSuccessfulOnly(t *testing.T) {
	tt := test.Start(t).Scenario("failed_transactions")
	defer tt.Finish()

	var transactions []Transaction

	q := &Q{tt.HorizonSession()}
	query := q.Transactions().
		ForAccount("GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2")

	err := query.Select(&transactions)
	tt.Assert.NoError(err)

	tt.Assert.Equal(3, len(transactions))

	for _, transaction := range transactions {
		tt.Assert.True(*transaction.Successful)
	}

	sql, _, err := query.sql.ToSql()
	tt.Assert.NoError(err)
	// Note: brackets around `(ht.successful = true OR ht.successful IS NULL)` are critical!
	tt.Assert.Contains(sql, "WHERE htp.history_account_id = ? AND (ht.successful = true OR ht.successful IS NULL)")
}

// TestTransactionIncludeFailed tests `IncludeFailed` method.
func TestTransactionIncludeFailed(t *testing.T) {
	tt := test.Start(t).Scenario("failed_transactions")
	defer tt.Finish()

	var transactions []Transaction

	q := &Q{tt.HorizonSession()}
	query := q.Transactions().
		ForAccount("GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2").
		IncludeFailed()

	err := query.Select(&transactions)
	tt.Assert.NoError(err)

	var failed, successful int
	for _, transaction := range transactions {
		if *transaction.Successful {
			successful++
		} else {
			failed++
		}
	}

	tt.Assert.Equal(3, successful)
	tt.Assert.Equal(1, failed)

	sql, _, err := query.sql.ToSql()
	tt.Assert.NoError(err)
	tt.Assert.Equal("SELECT ht.id, ht.transaction_hash, ht.ledger_sequence, ht.application_order, ht.account, ht.account_sequence, ht.fee_paid, ht.operation_count, ht.tx_envelope, ht.tx_result, ht.tx_meta, ht.tx_fee_meta, ht.created_at, ht.updated_at, ht.successful, array_to_string(ht.signatures, ',') AS signatures, ht.memo_type, ht.memo, lower(ht.time_bounds) AS valid_after, upper(ht.time_bounds) AS valid_before, hl.closed_at AS ledger_close_time FROM history_transactions ht LEFT JOIN history_ledgers hl ON ht.ledger_sequence = hl.sequence JOIN history_transaction_participants htp ON htp.history_transaction_id = ht.id WHERE htp.history_account_id = ?", sql)
}
