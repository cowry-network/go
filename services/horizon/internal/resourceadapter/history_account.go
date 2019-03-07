package resourceadapter

import (
	"context"

	. "github.com/cowry-network/go/protocols/horizon"
	"github.com/cowry-network/go/services/horizon/internal/db2/history"
)

func PopulateHistoryAccount(ctx context.Context, dest *HistoryAccount, row history.Account) {
	dest.ID = row.Address
	dest.AccountID = row.Address
}
