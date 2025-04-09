package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/x/flatfees/types"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// Keeper of the x/flatfees store
type Keeper struct {
	cdc              codec.Codec
	storeService     storetypes.KVStoreService
	feeCollectorName string // name of the FeeCollector ModuleAccount

	authority string

	Schema  collections.Schema
	params  collections.Item[types.Params]
	msgFees collections.Map[string, types.MsgFee]
}

func NewKeeper(
	cdc codec.Codec,
	storeService storetypes.KVStoreService,
	feeCollectorName string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	rv := Keeper{
		storeService:     storeService,
		cdc:              cdc,
		feeCollectorName: feeCollectorName,

		authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		msgFees:   collections.NewMap(sb, types.MsgFeeKeyPrefix, "msg_fees", collections.StringKey, codec.CollValue[types.MsgFee](cdc)),
		params:    collections.NewItem(sb, types.ParamsKeyPrefix, "params", codec.CollValue[types.Params](cdc)),
	}

	var err error
	rv.Schema, err = sb.Build()
	if err != nil {
		panic(err)
	}
	return rv
}

// GetAuthority is signer of the proposal
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ValidateAuthority returns an error if the provided authority does not match the keeper's authority.
func (k Keeper) ValidateAuthority(authority string) error {
	if authority == k.authority {
		return nil
	}
	return govtypes.ErrInvalidSigner.Wrapf("expected %q got %q", k.authority, authority)
}

// GetFeeCollectorName returns the name of the fee collector account.
func (k Keeper) GetFeeCollectorName() string {
	return k.feeCollectorName
}

// SetMsgFee sets the additional fee schedule for a Msg.
func (k Keeper) SetMsgFee(ctx sdk.Context, msgFee types.MsgFee) error {
	err := k.msgFees.Set(ctx, msgFee.MsgTypeUrl, msgFee)
	if err != nil {
		return fmt.Errorf("could not set msg fee for %q: %w", msgFee.MsgTypeUrl, err)
	}
	return nil
}

// GetMsgFee returns a MsgFee for the msg type if it exists nil if it does not.
func (k Keeper) GetMsgFee(ctx sdk.Context, msgType string) (*types.MsgFee, error) {
	rv, err := k.msgFees.Get(ctx, msgType)
	switch {
	case err == nil:
		return &rv, nil
	case errors.Is(err, collections.ErrNotFound):
		return nil, nil
	default:
		return nil, fmt.Errorf("could not get msg fee for %q: %w", msgType, err)
	}
}

// RemoveMsgFee removes MsgFee or returns a ErrMsgFeeDoesNotExist error if it does not exist.
func (k Keeper) RemoveMsgFee(ctx sdk.Context, msgType string) error {
	has, err := k.msgFees.Has(ctx, msgType)
	switch {
	case err != nil:
		return fmt.Errorf("cannot remove msg fee for %q: invalid key: %w", msgType, err)
	case !has:
		return fmt.Errorf("cannot remove msg fee for %q: %w", msgType, types.ErrMsgFeeDoesNotExist)
	}

	err = k.msgFees.Remove(ctx, msgType)
	if err != nil {
		return fmt.Errorf("could not remove msg fee for %q: %w", msgType, err)
	}
	return nil
}

// CalculateMsgCost calculates the up-front and on-success costs for the provided Msgs.
// The total cost of running the provided Msgs is up-front + on-success.
// The first coin returned should always be collected; the second is the amount to collect iff the tx is successful.
func (k Keeper) CalculateMsgCost(ctx sdk.Context, msgs ...sdk.Msg) (upFront sdk.Coins, onSuccess sdk.Coins, err error) {
	params := k.GetParams(ctx)

	dflt := params.ConversionFactor.ConvertCoin(params.DefaultCost)

	var msgFee *types.MsgFee
	for _, msg := range msgs {
		msgType := sdk.MsgTypeURL(msg)
		msgFee, err = k.GetMsgFee(ctx, msgType)
		switch {
		case err != nil:
			return nil, nil, fmt.Errorf("could not get msg fee for %q: %w", msgType, err)
		case msgFee == nil: // No specifically defined entry for this msg type, use the default.
			upFront = upFront.Add(dflt)
		case msgFee.Cost.IsZero(): // This message is free, move on to the next.
		default:
			msgCost := params.ConversionFactor.ConvertCoins(msgFee.Cost)
			newUpFront, newOnSuccess := splitMsgCost(msgCost, dflt)
			upFront = upFront.Add(newUpFront...)
			onSuccess = onSuccess.Add(newOnSuccess...)
		}
	}

	return upFront, onSuccess, nil
}

// splitMsgCost will split the provided msgCost into the amounts to charge up-front and upon success.
func splitMsgCost(msgCost sdk.Coins, defaultCost sdk.Coin) (upFront sdk.Coins, onSuccess sdk.Coins) {
	for _, coin := range msgCost {
		switch {
		case coin.Denom != defaultCost.Denom:
			// A coin in a denom other than the default is all collected upon success.
			onSuccess = append(onSuccess, coin)
		case coin.Amount.LTE(defaultCost.Amount):
			// The coin is at most the default, so charge all of it up-front.
			upFront = append(upFront, coin)
		default:
			// The coin is more than the default. Collect the default up-front and the rest upon success.
			upFront = append(upFront, defaultCost)
			onSuccess = append(onSuccess, coin.Sub(defaultCost))
		}
	}
	return upFront, onSuccess
}

// ExpandMsgs returns all the provided msgs as well as all Msgs (of interest) that they contain.
func (k Keeper) ExpandMsgs(msgs []sdk.Msg) ([]sdk.Msg, error) {
	return k.expandMsgs(msgs, codectypes.MaxUnpackAnyRecursionDepth)
}

// expandMsgs recursively (up to the provided depthLeft) extracts sub-Msgs from some specific Msg types.
// All provided msgs are returned as well as any sub-msgs.
func (k Keeper) expandMsgs(msgs []sdk.Msg, depthLeft int) ([]sdk.Msg, error) {
	if depthLeft < 0 {
		return nil, fmt.Errorf("could not expand sub-messages: max depth exceeded")
	}

	rv := make([]sdk.Msg, 0, len(msgs))
	var subMsgs []sdk.Msg
	var err error
	for _, msg := range msgs {
		rv = append(rv, msg)

		switch m := msg.(type) {
		case *authztypes.MsgExec:
			// An authz MsgExec executes the messages right now, so we charge for them now.
			subMsgs, err = k.anysToMsgs(m.Msgs)
		case *govv1.MsgSubmitProposal:
			// Pay for the gov prop msgs at the time it's submitted because there's no way to collect it later.
			subMsgs, err = k.anysToMsgs(m.Messages)
		case *triggertypes.MsgCreateTriggerRequest:
			// Gotta pay the trigger costs upon creation too.
			subMsgs, err = k.anysToMsgs(m.Actions)
		default:
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("could not extract sub-messages from %T: %w", msg, err)
		}
		// Other msg types considered, but not treated specially:
		//  - /cosmos.group.v1.MsgSubmitProposal - Msgs in here don't get executed until MsgExec.
		//       So we charge for them during the MsgExec (since that's what'd happen with fees for gas).
		//  - /cosmos.gov.v1.MsgExecLegacyContent - We've disabled all the old proposals. Not really a Msg anyway.
		//  - /cosmos.gov.v1beta1.MsgSubmitProposal - No longer usable.

		subMsgs, err = k.expandMsgs(subMsgs, depthLeft-1)
		if err != nil {
			return nil, err
		}
		rv = append(rv, subMsgs...)
	}

	return rv, nil
}

// anysToMsgs is similar to sdktx.GetMsgs, but will use this keeper's codec to unpack entries without a cached value.
func (k Keeper) anysToMsgs(anys []*codectypes.Any) ([]sdk.Msg, error) {
	rv := make([]sdk.Msg, len(anys))
	var ok bool
	var err error
	for i, msgAny := range anys {
		cached := msgAny.GetCachedValue()
		if cached != nil {
			rv[i], ok = cached.(sdk.Msg)
			if !ok {
				return nil, fmt.Errorf("could not cast %T %q as %T", cached, msgAny.TypeUrl, rv[i])
			}
			continue
		}

		err = k.cdc.UnpackAny(msgAny, &rv[i])
		if err != nil {
			return nil, fmt.Errorf("could not unpack %T with a %q: %w", msgAny, msgAny.TypeUrl, err)
		}
	}

	return rv, nil
}

// GetParams returns the x/flatfees parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	rv, err := k.params.Get(ctx)
	if err != nil {
		panic(fmt.Errorf("error getting params: %w", err))
	}
	return rv
}

// SetParams stores/updates the x/flatfees parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	err := k.params.Set(ctx, params)
	if err != nil {
		return fmt.Errorf("error setting params: %w", err)
	}
	return nil
}
