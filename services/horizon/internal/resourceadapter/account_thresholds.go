package resourceadapter

import (
	. "github.com/cowry-network/go/protocols/horizon"
	"github.com/cowry-network/go/services/horizon/internal/db2/core"
)

func PopulateAccountThresholds(dest *AccountThresholds, row core.Account) {
	dest.LowThreshold = row.Thresholds[1]
	dest.MedThreshold = row.Thresholds[2]
	dest.HighThreshold = row.Thresholds[3]
}
