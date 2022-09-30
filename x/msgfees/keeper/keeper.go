package keeper

import (
	"sort"

	"golang.org/x/exp/constraints"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	cosmosauthtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

const StoreKey = types.ModuleName

type baseAppSimulateFunc func(txBytes []byte) (sdk.GasInfo, *sdk.Result, sdk.Context, error)

// Keeper of the Additional fee store
type Keeper struct {
	storeKey         storetypes.StoreKey
	cdc              codec.BinaryCodec
	paramSpace       paramtypes.Subspace
	feeCollectorName string // name of the FeeCollector ModuleAccount
	defaultFeeDenom  string
	simulateFunc     baseAppSimulateFunc
	txDecoder        sdk.TxDecoder
}

// NewKeeper returns a AdditionalFeeKeeper. It handles:
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	feeCollectorName string,
	defaultFeeDenom string,
	simulateFunc baseAppSimulateFunc,
	txDecoder sdk.TxDecoder,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		paramSpace:       paramSpace,
		feeCollectorName: feeCollectorName,
		defaultFeeDenom:  defaultFeeDenom,
		simulateFunc:     simulateFunc,
		txDecoder:        txDecoder,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) GetFeeCollectorName() string {
	return k.feeCollectorName
}

// GetFloorGasPrice returns the current minimum gas price in sdk.Coin used in calculations for charging additional fees
func (k Keeper) GetFloorGasPrice(ctx sdk.Context) sdk.Coin {
	min := types.DefaultFloorGasPrice()
	if k.paramSpace.Has(ctx, types.ParamStoreKeyFloorGasPrice) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyFloorGasPrice, &min)
	}
	return min
}

// GetNhashPerUsdMil returns the current nhash amount per usd mil
func (k Keeper) GetNhashPerUsdMil(ctx sdk.Context) uint64 {
	rateInMils := types.DefaultParams().NhashPerUsdMil
	if k.paramSpace.Has(ctx, types.ParamStoreKeyNhashPerUsdMil) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyNhashPerUsdMil, &rateInMils)
	}
	return rateInMils
}

// GetConversionFeeDenom returns the conversion fee denom
func (k Keeper) GetConversionFeeDenom(ctx sdk.Context) string {
	conversionFeeDenom := types.DefaultParams().ConversionFeeDenom
	if k.paramSpace.Has(ctx, types.ParamStoreKeyConversionFeeDenom) {
		k.paramSpace.Get(ctx, types.ParamStoreKeyConversionFeeDenom, &conversionFeeDenom)
	}
	return conversionFeeDenom
}

// SetMsgFee sets the additional fee schedule for a Msg
func (k Keeper) SetMsgFee(ctx sdk.Context, msgFees types.MsgFee) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&msgFees)
	store.Set(types.GetMsgFeeKey(msgFees.MsgTypeUrl), bz)
	return nil
}

// GetMsgFee returns a MsgFee for the msg type if it exists nil if it does not
func (k Keeper) GetMsgFee(ctx sdk.Context, msgType string) (*types.MsgFee, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMsgFeeKey(msgType)
	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, nil
	}

	var msgFee types.MsgFee
	if err := k.cdc.Unmarshal(bz, &msgFee); err != nil {
		return nil, err
	}

	return &msgFee, nil
}

// RemoveMsgFee removes MsgFee or returns an error if it does not exist
func (k Keeper) RemoveMsgFee(ctx sdk.Context, msgType string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMsgFeeKey(msgType)
	bz := store.Get(key)
	if len(bz) == 0 {
		return types.ErrMsgFeeDoesNotExist
	}

	store.Delete(key)

	return nil
}

type Handler func(record types.MsgFee) (stop bool)

// IterateMsgFees  iterates all msg fees with the given handler function.
func (k Keeper) IterateMsgFees(ctx sdk.Context, handle func(msgFees types.MsgFee) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.MsgFeeKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.MsgFee{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}

// DeductFeesDistributions deducts fees from the given account.  The fees map contains a key of bech32 addresses to distribute funds to.
// If the key in the map is an empty string, those will go to the fee collector.  After all the accounts in fees map are paid out,
// the remainder of remainingFees will be swept to the fee collector account.
func (k Keeper) DeductFeesDistributions(bankKeeper bankkeeper.Keeper, ctx sdk.Context, acc cosmosauthtypes.AccountI, remainingFees sdk.Coins, fees map[string]sdk.Coins) error {
	sentCoins := sdk.NewCoins()
	for _, key := range sortedKeys(fees) {
		coins := fees[key]
		if !coins.IsValid() {
			return sdkerrors.ErrInsufficientFee.Wrapf("invalid fee amount: %q", fees)
		}
		if len(key) == 0 {
			err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), k.feeCollectorName, coins)
			if err != nil {
				return sdkerrors.ErrInsufficientFunds.Wrap(err.Error())
			}
		} else {
			recipient, err := sdk.AccAddressFromBech32(key)
			if err != nil {
				return sdkerrors.ErrInvalidAddress.Wrap(err.Error())
			}
			err = bankKeeper.SendCoins(ctx, acc.GetAddress(), recipient, coins)
			if err != nil {
				return sdkerrors.ErrInsufficientFunds.Wrap(err.Error())
			}
		}
		sentCoins = sentCoins.Add(coins...)
	}
	unsentFee, neg := remainingFees.SafeSub(sentCoins...)
	if neg {
		return sdkerrors.ErrInsufficientFunds.Wrapf("negative balance after sending coins to accounts and fee collector: remainingFees: %q, sentCoins: %q, distribution: %v", remainingFees, sentCoins, fees)
	}
	if !unsentFee.IsZero() {
		// sweep the rest of the fees to module
		err := bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), k.feeCollectorName, unsentFee)
		if err != nil {
			return sdkerrors.ErrInsufficientFunds.Wrap(err.Error())
		}
	}

	return nil
}

// ConvertDenomToHash converts usd coin to nhash coin using nhash per usd mil.
// Currently, usd is only supported with nhash to usd mil coming from params
func (k Keeper) ConvertDenomToHash(ctx sdk.Context, coin sdk.Coin) (sdk.Coin, error) {
	conversionDenom := k.GetConversionFeeDenom(ctx)
	switch coin.Denom {
	case types.UsdDenom:
		nhashPerMil := int64(k.GetNhashPerUsdMil(ctx))
		amount := coin.Amount.Mul(sdk.NewInt(nhashPerMil))
		msgFeeCoin := sdk.NewInt64Coin(conversionDenom, amount.Int64())
		return msgFeeCoin, nil
	case conversionDenom:
		return coin, nil
	default:
		return sdk.Coin{}, sdkerrors.ErrInvalidType.Wrapf("denom not supported for conversion %s", coin.Denom)
	}
}

// CalculateAdditionalFeesToBePaid computes the additional fees to be paid for the provided messages.
func (k Keeper) CalculateAdditionalFeesToBePaid(ctx sdk.Context, msgs ...sdk.Msg) (types.MsgFeesDistribution, error) {
	msgFeesDistribution := types.MsgFeesDistribution{
		RecipientDistributions: make(map[string]sdk.Coins),
	}
	assessCustomMsgTypeURL := sdk.MsgTypeURL(&types.MsgAssessCustomMsgFeeRequest{})
	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		msgFees, err := k.GetMsgFee(ctx, typeURL)
		if err != nil {
			return msgFeesDistribution, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}

		if msgFees != nil {
			if err := msgFeesDistribution.Increase(msgFees.AdditionalFee, msgFees.RecipientBasisPoints, msgFees.Recipient); err != nil {
				return msgFeesDistribution, err
			}
		}

		if typeURL == assessCustomMsgTypeURL {
			assessFee, ok := msg.(*types.MsgAssessCustomMsgFeeRequest)
			if !ok {
				return msgFeesDistribution, sdkerrors.ErrInvalidType.Wrap("unable to convert msg to MsgAssessCustomMsgFeeRequest")
			}
			msgFeeCoin, err := k.ConvertDenomToHash(ctx, assessFee.Amount)
			if err != nil {
				return msgFeesDistribution, err
			}
			if err := msgFeesDistribution.Increase(msgFeeCoin, types.AssessCustomMsgFeeBips, assessFee.Recipient); err != nil {
				return msgFeesDistribution, err
			}
		}
	}

	return msgFeesDistribution, nil
}

// sortedKeys gets the keys of a map, sorts them and returns them as a slice.
func sortedKeys[K constraints.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}
