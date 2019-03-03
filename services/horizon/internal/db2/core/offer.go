package core

import (
	"fmt"
	"math/big"

	sq "github.com/Masterminds/squirrel"
	"github.com/guregu/null"
	"github.com/stellar/go/services/horizon/internal/db2"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/xdr"
)

// PagingToken returns a suitable paging token for the Offer
func (r Offer) PagingToken() string {
	return fmt.Sprintf("%d", r.OfferID)
}

// PriceAsString return the price fraction as a floating point approximate.
func (r Offer) PriceAsString() string {
	return big.NewRat(int64(r.Pricen), int64(r.Priced)).FloatString(7)
}

// ConnectedAssets loads xdr.Asset records for the purposes of path
// finding.  Given the input asset type, a list of xdr.Assets is returned that
// each have some available trades for the input asset.
func (q *Q) ConnectedAssets(dest interface{}, selling xdr.Asset) error {
	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	if schemaVersion < 9 {
		return q.connectedAssetsSchema8(dest, selling)
	} else {
		return q.connectedAssetsSchema9(dest, selling)
	}
}

func (q *Q) connectedAssetsSchema9(dest interface{}, selling xdr.Asset) error {
	assets, ok := dest.(*[]xdr.Asset)
	if !ok {
		return errors.New("dest is not *[]xdr.Asset")
	}

	sellingAssetXDRString, err := xdr.MarshalBase64(selling)
	if err != nil {
		return errors.Wrap(err, "Error marshaling selling")
	}

	sql := sq.Select("buyingasset").
		From("offers").
		Where(sq.Eq{"sellingasset": sellingAssetXDRString}).
		GroupBy("buyingasset")

	var rows []struct {
		Asset string `db:"buyingasset"`
	}

	err = q.Select(&rows, sql)

	if err != nil {
		return err
	}

	results := make([]xdr.Asset, len(rows))
	*assets = results

	for i, r := range rows {
		var asset xdr.Asset
		err = xdr.SafeUnmarshalBase64(r.Asset, &asset)
		if err != nil {
			return errors.Wrap(err, "Error decoding asset")
		}

		results[i] = asset
	}

	return nil
}

// ConnectedAssets loads xdr.Asset records for the purposes of path
// finding.  Given the input asset type, a list of xdr.Assets is returned that
// each have some available trades for the input asset.
func (q *Q) connectedAssetsSchema8(dest interface{}, selling xdr.Asset) error {
	assets, ok := dest.(*[]xdr.Asset)
	if !ok {
		return errors.New("dest is not *[]xdr.Asset")
	}

	var (
		t xdr.AssetType
		c string
		i string
	)

	err := selling.Extract(&t, &c, &i)
	if err != nil {
		return err
	}

	sql := sq.Select(
		"buyingassettype AS type",
		"coalesce(buyingassetcode, '') AS code",
		"coalesce(buyingissuer, '') AS issuer").
		From("offers").
		Where(sq.Eq{"sellingassettype": t}).
		GroupBy("buyingassettype", "buyingassetcode", "buyingissuer")

	if t != xdr.AssetTypeAssetTypeNative {
		sql = sql.Where(sq.Eq{"sellingassetcode": c, "sellingissuer": i})
	}

	var rows []struct {
		Type   xdr.AssetType
		Code   string
		Issuer string
	}

	err = q.Select(&rows, sql)

	if err != nil {
		return err
	}

	results := make([]xdr.Asset, len(rows))
	*assets = results

	for i, r := range rows {
		results[i], err = AssetFromDB(r.Type, r.Code, r.Issuer)
		if err != nil {
			return err
		}
	}

	return nil
}

// OffersByAddress loads a page of active offers for the given
// address.
func (q *Q) OffersByAddress(dest interface{}, addy string, pq db2.PageQuery) error {
	schemaVersion, err := q.SchemaVersion()
	if err != nil {
		return err
	}

	sql := sq.Select("co.*").
		From("offers co").
		Where("co.sellerid = ?", addy).
		Limit(uint64(pq.Limit))

	cursor, err := pq.CursorInt64()
	if err != nil {
		return err
	}

	switch pq.Order {
	case "asc":
		sql = sql.Where("co.offerid > ?", cursor).OrderBy("co.offerid asc")
	case "desc":
		sql = sql.Where("co.offerid < ?", cursor).OrderBy("co.offerid desc")
	}

	err = q.Select(dest, sql)
	if err != nil {
		return err
	}

	if schemaVersion < 9 {
		return nil
	}

	// In schema 9 we need to decode XDR-encoded assets
	offers, ok := dest.(*[]Offer)
	if !ok {
		return errors.New("dest is not []Offer")
	}

	newOffers := make([]Offer, len(*offers))
	for i, offer := range *offers {
		var sellingAsset, buyingAsset xdr.Asset

		err = xdr.SafeUnmarshalBase64(offer.SellingAsset, &sellingAsset)
		if err != nil {
			return errors.Wrap(err, "Error decoding sellingasset")
		}

		err = xdr.SafeUnmarshalBase64(offer.BuyingAsset, &buyingAsset)
		if err != nil {
			return errors.Wrap(err, "Error decoding buyingasset")
		}

		newOffers[i] = offer

		var sellingAssetCode, sellingIssuer string
		err = sellingAsset.Extract(
			&newOffers[i].SellingAssetType,
			&sellingAssetCode,
			&sellingIssuer,
		)
		if err != nil {
			return errors.Wrap(err, "Error extracting sellingasset")
		}

		if newOffers[i].SellingAssetType != xdr.AssetTypeAssetTypeNative {
			newOffers[i].SellingAssetCode = null.StringFrom(sellingAssetCode)
			newOffers[i].SellingIssuer = null.StringFrom(sellingIssuer)
		}

		var buyingAssetCode, buyingIssuer string
		err = buyingAsset.Extract(
			&newOffers[i].BuyingAssetType,
			&buyingAssetCode,
			&buyingIssuer,
		)
		if err != nil {
			return errors.Wrap(err, "Error extracting buyingasset")
		}

		if newOffers[i].BuyingAssetType != xdr.AssetTypeAssetTypeNative {
			newOffers[i].BuyingAssetCode = null.StringFrom(buyingAssetCode)
			newOffers[i].BuyingIssuer = null.StringFrom(buyingIssuer)
		}
	}

	*dest.(*[]Offer) = newOffers
	return nil
}
