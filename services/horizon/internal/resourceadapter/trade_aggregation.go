package resourceadapter

import (
	"context"

	"github.com/cowry-network/go/amount"
	"github.com/cowry-network/go/price"
	. "github.com/cowry-network/go/protocols/horizon"
	"github.com/cowry-network/go/services/horizon/internal/db2/history"
)

// Populate fills out the details of a trade using a row from the history_trades
// table.
func PopulateTradeAggregation(
	ctx context.Context,
	dest *TradeAggregation,
	row history.TradeAggregation,
) (err error) {
	dest.Timestamp = row.Timestamp
	dest.TradeCount = row.TradeCount
	dest.BaseVolume = amount.StringFromInt64(row.BaseVolume)
	dest.CounterVolume = amount.StringFromInt64(row.CounterVolume)
	dest.Average = price.StringFromFloat64(row.Average)
	dest.High = row.High.String()
	dest.HighR = row.High
	dest.Low = row.Low.String()
	dest.LowR = row.Low
	dest.Open = row.Open.String()
	dest.OpenR = row.Open
	dest.Close = row.Close.String()
	dest.CloseR = row.Close
	return
}
