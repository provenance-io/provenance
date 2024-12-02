package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/nav"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService
	logger       log.Logger

	navs collections.Map[collections.Pair[string, string], nav.NetAssetValueRecord]
}

// NewKeeper returns a new nav keeper.
func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, logger log.Logger) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	return &Keeper{
		cdc:          cdc,
		storeService: storeService,
		logger:       logger.With("module", "x/nav"),
		navs: collections.NewMap(sb, []byte{nav.StorePrefixNAVs}, "navs",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			codec.CollValue[nav.NetAssetValueRecord](cdc)),
	}
}

// SetNAVs stores the provided navs in state. All will have the provided source and the height from the ctx.
// An error is returned if the source or any navs are invalid.
func (k Keeper) SetNAVs(ctx context.Context, source string, navs ...*nav.NetAssetValue) error {
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()
	return k.SetNAVsAtHeight(ctx, source, height, navs...)
}

// SetNAVsAtHeight stores the provided navs in state. All with have the provided info.
// An error is returned if the source or any navs are invalid.
func (k Keeper) SetNAVsAtHeight(ctx context.Context, source string, height int64, navs ...*nav.NetAssetValue) error {
	if len(navs) == 0 {
		return nil
	}
	if err := nav.ValidateSource(source); err != nil {
		return err
	}
	if err := nav.ValidateNAVs(navs); err != nil {
		return err
	}
	nrs := nav.NAVsAsRecords(navs, height, source)
	return k.setNAVRecordsRaw(ctx, nrs)
}

// SetNAVRecords stores the provided navs in state.
// An error is returned if any navs are invalid.
func (k Keeper) SetNAVRecords(ctx context.Context, navs nav.NAVRecords) error {
	if len(navs) == 0 {
		return nil
	}
	if err := nav.ValidateNAVRecords(navs); err != nil {
		return err
	}
	return k.setNAVRecordsRaw(ctx, navs)
}

// setNAVRecordsRaw does the actual storage of NAVs in state (without any validation).
func (k Keeper) setNAVRecordsRaw(ctx context.Context, navs nav.NAVRecords) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for i, navr := range navs {
		key := navr.Key()
		err := k.navs.Set(ctx, key, *navr)
		if err != nil {
			return fmt.Errorf("error setting nav[%d]: %w", i, err)
		}

		err = sdkCtx.EventManager().EmitTypedEvent(nav.NewEventSetNetAssetValue(navr))
		if err != nil {
			k.logger.Error(fmt.Sprintf("Error emitting NAV event: %v", err), "NAV", navr)
		}
	}
	return nil
}

// GetNAVRecord returns the navs with the given asset and price denoms.
func (k Keeper) GetNAVRecord(ctx context.Context, assetDenom, priceDenom string) *nav.NetAssetValueRecord {
	if len(assetDenom) == 0 || len(priceDenom) == 0 {
		return nil
	}
	rv, err := k.navs.Get(ctx, collections.Join(assetDenom, priceDenom))
	if err != nil {
		return nil
	}
	return &rv
}

// GetNAVRecords returns all navs with the given asset denom.
// If the assetDenom is empty, all navs are returned regardless of asset denom.
func (k Keeper) GetNAVRecords(ctx context.Context, assetDenom string) nav.NAVRecords {
	var ranger collections.Ranger[collections.Pair[string, string]]
	if len(assetDenom) > 0 {
		ranger = collections.NewPrefixedPairRange[string, string](assetDenom)
	}
	var rv nav.NAVRecords
	err := k.navs.Walk(ctx, ranger, func(_ collections.Pair[string, string], nav nav.NetAssetValueRecord) (stop bool, err error) {
		rv = append(rv, &nav)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return rv
}

// GetAllNAVRecords gets all navs. This is the same as GetNAVRecords(ctx, "").
func (k Keeper) GetAllNAVRecords(ctx context.Context) nav.NAVRecords {
	return k.GetNAVRecords(ctx, "")
}
