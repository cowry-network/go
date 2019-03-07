package resourceadapter

import (
	"context"

	. "github.com/cowry-network/go/protocols/horizon"
	"github.com/cowry-network/go/xdr"
)

func PopulateAsset(ctx context.Context, dest *Asset, asset xdr.Asset) error {
	return asset.Extract(&dest.Type, &dest.Code, &dest.Issuer)
}
