package keeper

import (
	"fmt"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"

	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"

	cerrs "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/provenance-io/provenance/x/expiration/types"
)

// Handler is a name record handler function for use with IterateExpirations.
type Handler func(record types.Expiration) error

// Keeper defines the name module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set account specific params
	paramSpace paramtypes.Subspace

	// Key to access the key-value store from sdk.Context.
	storeKey storetypes.StoreKey

	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// To check granter grantee authorization of messages
	authzKeeper authzkeeper.Keeper

	// To handle account interactions.
	acctKeeper banktypes.AccountKeeper

	// To handle expiration deposit processing.
	bankKeeper bankkeeper.Keeper

	// Message service router
	router baseapp.IMsgServiceRouter
}

// NewKeeper returns an expiration keeper. It handles:
// - managing a hierarchy of expiration
// - enforcing permissions for expiration creation/extension/deletion
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	authzKeeper authzkeeper.Keeper,
	acctKeeper banktypes.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	router baseapp.IMsgServiceRouter,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	// ensure expiration module account is set
	if acctKeeper.GetModuleAddress(types.ModuleName) == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		storeKey:    key,
		paramSpace:  paramSpace,
		cdc:         cdc,
		authzKeeper: authzKeeper,
		acctKeeper:  acctKeeper,
		bankKeeper:  bankKeeper,
		router:      router,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetModuleAccount returns the expiration ModuleAccount
func (k Keeper) GetModuleAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return k.acctKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// GetDeposit returns the deposit used in setting module asset expirations.
func (k Keeper) GetDeposit(ctx sdk.Context) sdk.Coin {
	deposit := &types.DefaultDeposit
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyDeposit, deposit)
	return *deposit
}

// GetDuration returns the duration used in setting module asset expirations.
func (k Keeper) GetDuration(ctx sdk.Context) string {
	duration := &types.DefaultDuration
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyDuration, duration)
	return *duration
}

// GetExpiration returns the expiration with the given module asset id.
// In case of not found, (nil, nil) is returned.
func (k Keeper) GetExpiration(ctx sdk.Context, moduleAssetID string) (*types.Expiration, error) {
	key, err := types.GetModuleAssetKey(moduleAssetID)
	if err != nil {
		return nil, types.ErrInvalidKey.Wrap(err.Error())
	}

	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return nil, nil
	}

	b := store.Get(key)
	expiration := &types.Expiration{}
	err = k.cdc.Unmarshal(b, expiration)
	if err != nil {
		return nil, types.ErrUnmarshal.Wrap(err.Error())
	}

	return expiration, nil
}

// SetExpiration creates an expiration record for a module asset.
func (k Keeper) SetExpiration(ctx sdk.Context, expiration types.Expiration) error {
	// when modules set expirations for their assets they call SetExpiration
	// therefore, we need to check message is valid by running it through ValidateBasic
	if err := expiration.ValidateBasic(); err != nil {
		return types.ErrSetExpiration.Wrapf("failed ValidateBasic: %b", err)
	}
	// unpack expiration message
	var msg sdk.Msg
	if err := k.cdc.UnpackAny(&expiration.Message, &msg); err != nil {
		return types.ErrInvalidMessage.Wrapf("failed to unpack msg: %v", err)
	}
	if err := msg.ValidateBasic(); err != nil {
		return types.ErrInvalidMessage.Wrapf("failed ValidateBasic: %v", err)
	}

	// get store key prefix
	store := ctx.KVStore(k.storeKey)
	key, err := types.GetModuleAssetKey(expiration.ModuleAssetId)
	if err != nil {
		return cerrs.Wrapf(err, "unable to retrieve module asset key for: %s", expiration.ModuleAssetId)
	}

	// move deposit from owner account into expiration module account
	ownerAddr, addrErr := sdk.AccAddressFromBech32(expiration.Owner)
	if addrErr != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid owner: %s", expiration.Owner)
	}

	// attempt to send coins from owner account to expiration module account
	depErr := k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAddr, types.ModuleName, expiration.Deposit)
	if depErr != nil {
		return types.ErrInsufficientDeposit.Wrap(depErr.Error())
	}

	// emit event indicating that a deposit was collected
	k.emitEvent(ctx, types.NewEventExpirationDeposit(expiration.ModuleAssetId, expiration.Owner, expiration.Deposit))
	if err != nil {
		return types.ErrSetExpiration.Wrapf(
			"failed to emit EventExpirationDeposit for [%s]: %v", expiration.Owner, err)
	}

	// marshal expiration record and store
	b, err := k.cdc.Marshal(&expiration)
	if err != nil {
		return types.ErrSetExpiration.Wrap(
			types.ErrUnmarshal.Wrapf("expiration: %v: %v", expiration, err).Error())
	}
	store.Set(key, b)

	// emit event indicating the expiration was added
	k.emitEvent(ctx, types.NewEventExpirationAdd(expiration.ModuleAssetId))

	return nil
}

// ExtendExpiration extends an expiration to time in the future.
func (k Keeper) ExtendExpiration(ctx sdk.Context, expiration types.Expiration) error {
	// get key
	key, err := types.GetModuleAssetKey(expiration.ModuleAssetId)
	if err != nil {
		return types.ErrExtendExpiration.Wrapf("%s [%s]: %v",
			types.ErrInvalidKey.Error(), expiration.ModuleAssetId, err)
	}

	// lookup old expiration
	oldExpiration, err := k.GetExpiration(ctx, expiration.ModuleAssetId)
	if err != nil {
		return types.ErrExtendExpiration.Wrap(err.Error())
	}
	if oldExpiration == nil {
		return types.ErrExtendExpiration.Wrapf("%s: %s", types.ErrNotFound.Error(), expiration.ModuleAssetId)
	}

	// Validate expiration time is after old expiration time
	if oldExpiration.Time.After(expiration.Time) {
		return types.ErrExtendExpiration.Wrapf("new expiration time [%s] must be after old expiration time [%s]",
			expiration.Time.UTC(), oldExpiration.Time.UTC())
	}
	// Validate owners are the same
	if expiration.Owner != oldExpiration.Owner {
		return types.ErrNewOwnerNoMatch.Wrapf("new owner [%s] and old owner [%s] do not match",
			expiration.Owner, oldExpiration.Owner)
	}

	// Marshal expiration record and store
	store := ctx.KVStore(k.storeKey)
	b, err := k.cdc.Marshal(&expiration)
	if err != nil {
		return types.ErrExtendExpiration.Wrapf("%s: %v", types.ErrUnmarshal.Error(), err)
	}
	store.Set(key, b)

	// emit Extend event
	k.emitEvent(ctx, types.NewEventExpirationExtend(expiration.ModuleAssetId))

	return nil
}

// InvokeExpiration invokes an expiration message through the MsgServiceRouter.
// The expiration message is removed when the message invocation is successful.
//
// Caution: Do NOT call InvokeExpiration from within a module to delete a self-owned asset.
// This will lead to an infinite loop as expiration messages in this method are routed through
// the message router. Use RemoveExpiration instead.
func (k Keeper) InvokeExpiration(ctx sdk.Context, moduleAssetID string, refundTo sdk.AccAddress) error {
	// lookup expiration
	expiration, err := k.GetExpiration(ctx, moduleAssetID)
	if err != nil {
		return types.ErrInvokeExpiration.Wrap(err.Error())
	}
	if expiration == nil {
		return types.ErrInvokeExpiration.Wrapf("%s [%s]", types.ErrNotFound.Error(), expiration.ModuleAssetId)
	}

	// unpack expiration message
	var msg sdk.Msg
	if err = k.cdc.UnpackAny(&expiration.Message, &msg); err != nil {
		return types.ErrInvokeExpiration.Wrapf("failed to unpack msg: %v", err)
	}

	// route message to module
	handler := k.router.Handler(msg)
	if handler == nil {
		return types.ErrInvokeExpiration.Wrap(
			types.ErrMsgHandler.Wrapf("no message handler found for %q", sdk.MsgTypeURL(msg)).Error())
	}
	r, err := handler(ctx, msg)
	if err != nil {
		return types.ErrInvokeExpiration.Wrap(
			types.ErrMsgHandler.Wrapf("message %s: %v", sdk.MsgTypeURL(msg), err).Error())
	}
	// Handler should always return non-nil sdk.Result.
	if r == nil {
		return types.ErrInvokeExpiration.Wrap(
			types.ErrMsgHandler.Wrapf("got nil sdk.Result for message %q", msg).Error())
	}

	// refund deposit from expiration module account to depositor
	if err := k.refundDeposit(ctx, expiration.Deposit, refundTo); err != nil {
		return types.ErrInvokeExpiration.Wrapf("failed to refund deposit to [%s]: %v", refundTo, err)
	}

	// remove record from store
	if err := k.removeExpiration(ctx, moduleAssetID); err != nil {
		return types.ErrInvokeExpiration.Wrapf("failed to remove expiration: %v", err)
	}

	// emit Invoke event
	k.emitEvent(ctx, types.NewEventExpirationInvoke(moduleAssetID))

	return nil
}

// RemoveExpiration removes an expiration message through calls from other modules.
// Refunds deposit to original owner. In case of not found, nil is returned.
func (k Keeper) RemoveExpiration(ctx sdk.Context, moduleAssetID string, refundTo sdk.AccAddress) error {
	// lookup expiration
	expiration, err := k.GetExpiration(ctx, moduleAssetID)
	if err != nil {
		return types.ErrInvokeExpiration.Wrap(err.Error())
	}
	if expiration == nil {
		return nil
	}

	// refund deposit from expiration module account to depositor
	if err := k.refundDeposit(ctx, expiration.Deposit, refundTo); err != nil {
		return types.ErrRemoveExpiration.Wrapf("failed to refund deposit to [%s]: %v", refundTo, err)
	}

	// remove record from store
	if err := k.removeExpiration(ctx, moduleAssetID); err != nil {
		return types.ErrRemoveExpiration.Wrapf("failed to remove expiration: %v", err)
	}

	// emit Remove event
	k.emitEvent(ctx, types.NewEventExpirationRemove(moduleAssetID))

	return nil
}

// ValidateSetExpiration validates an expiration message.
func (k Keeper) ValidateSetExpiration(
	ctx sdk.Context,
	expiration types.Expiration,
	signers []string,
	msgTypeURL string,
) error {
	// validate basic
	if err := expiration.ValidateBasic(); err != nil {
		return err
	}

	// validate module asset id
	if _, err := sdk.AccAddressFromBech32(expiration.ModuleAssetId); err != nil {
		// check if we're dealing with a MetadataAddress
		if _, err2 := metadatatypes.MetadataAddressFromBech32(expiration.ModuleAssetId); err2 != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid module asset id: %v", err)
		}
	}

	// validate block time is in the future
	if expiration.Time.Before(ctx.BlockTime()) {
		return types.ErrExpirationTime.Wrapf("expiration time %s must be in the future", expiration.Time)
	}

	// validate deposit
	if err := expiration.Deposit.Validate(); err != nil {
		return sdkerrors.ErrInvalidCoins.Wrap(err.Error())
	}
	deposit := expiration.Deposit
	defaultDeposit := types.DefaultDeposit
	if deposit.IsZero() {
		return types.ErrInvalidDeposit.Wrap("deposit amount cannot be zero")
	}
	if deposit.IsAllLT(sdk.NewCoins(defaultDeposit)) {
		return types.ErrInvalidDeposit.Wrapf("deposit amount %s is less than minimum deposit amount %s",
			deposit.String(), defaultDeposit.String())
	}

	// validate msg
	var msg sdk.Msg
	if err := k.cdc.UnpackAny(&expiration.Message, &msg); err != nil {
		return types.ErrInvalidMessage.Wrapf("failed to unpack msg: %v", err)
	}
	if err := msg.ValidateBasic(); err != nil {
		return types.ErrInvalidMessage.Wrapf("failed ValidateBasic: %v", err)
	}

	// validate signers
	if err := k.validateSigners(ctx, expiration.Owner, signers, msgTypeURL); err != nil {
		return types.ErrInvalidSigners.Wrap(err.Error())
	}

	return nil
}

// ValidateInvokeExpiration validates an expiration record.
func (k Keeper) ValidateInvokeExpiration(
	ctx sdk.Context,
	moduleAssetID string,
	signers []string,
	msgTypeURL string,
) (*types.Expiration, error) {
	expiration, err := k.GetExpiration(ctx, moduleAssetID)
	if err != nil {
		return expiration, err
	}
	if expiration == nil {
		return expiration, types.ErrNotFound.Wrapf("module asset id [%s]", moduleAssetID)
	}

	// validate before routing as things may have
	// changed from the time this was set to now
	var msg sdk.Msg
	if err := k.cdc.UnpackAny(&expiration.Message, &msg); err != nil {
		return nil, types.ErrInvalidMessage.Wrapf("failed to unpack msg: %v", err)
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, types.ErrInvalidMessage.Wrapf("failed ValidateBasic: %v", err)
	}

	// anyone can delete an expired expiration
	if ctx.BlockTime().After(expiration.Time) {
		return expiration, nil
	}

	// validate signers and owner
	if err := k.validateSigners(ctx, expiration.Owner, signers, msgTypeURL); err != nil {
		return expiration, types.ErrInvalidSigners.Wrap(err.Error())
	}

	return expiration, nil
}

// ResolveDepositor resolves the account address where the expiration message deposit will be refunded.
func (k Keeper) ResolveDepositor(
	ctx sdk.Context,
	expiration types.Expiration,
	msg *types.MsgInvokeExpirationRequest,
) (sdk.AccAddress, error) {
	// check for owner in list first
	for _, signer := range msg.Signers {
		if signer == expiration.Owner {
			addr, err := sdk.AccAddressFromBech32(expiration.Owner)
			if err != nil {
				return nil, types.ErrResolveDepositor.Wrap(
					types.ErrInvalidSigners.Wrapf("[%s]: %v", signer, err).Error())
			}
			return addr, nil
		}
	}
	// fall back to first signer if after expiration
	if ctx.BlockTime().After(expiration.Time) && len(msg.Signers) >= 1 {
		signer := msg.Signers[0]
		addr, err := sdk.AccAddressFromBech32(signer)
		if err != nil {
			return nil, types.ErrResolveDepositor.Wrap(
				types.ErrInvalidSigners.Wrapf("[%s]: %v", signer, err).Error())
		}
		return addr, nil
	}

	// error if no qualifying depositors are found
	return nil, types.ErrResolveDepositor.Wrapf("asset [%s]", msg.ModuleAssetId)
}

// Private method that validates expiration message signatures
func (k Keeper) validateSigners(
	ctx sdk.Context,
	owner string,
	signers []string,
	msgTypeURL string,
) error {
	found := false
	for _, signer := range signers {
		if signer == owner {
			found = true
			break
		}

		// validate signer with authz
		var err error
		found, err = k.hasSignerWithAuthz(ctx, owner, signer, msgTypeURL)
		if err != nil {
			return err
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("intended signers %s do not match given signer [%s]", signers, owner)
	}

	return nil
}

// Private method that checks if any of the signers have been
// granted authorizations to perform actions on an expiration record
func (k Keeper) hasSignerWithAuthz(
	ctx sdk.Context,
	owner string,
	signer string,
	msgTypeURL string,
) (bool, error) {
	granter, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return false, fmt.Errorf("invalid owner: %w", err)
	}
	grantee, err := sdk.AccAddressFromBech32(signer)
	if err != nil {
		return false, fmt.Errorf("invalid signers: %w", err)
	}

	authorization, exp := k.authzKeeper.GetAuthorization(ctx, grantee, granter, msgTypeURL)
	if authorization != nil {
		resp, err := authorization.Accept(ctx, nil)
		if err != nil {
			return false, err
		}
		if resp.Accept {
			switch {
			case resp.Delete:
				err = k.authzKeeper.DeleteGrant(ctx, grantee, granter, msgTypeURL)
				if err != nil {
					return false, err
				}
			case resp.Updated != nil:
				if err = k.authzKeeper.SaveGrant(ctx, grantee, granter, resp.Updated, exp); err != nil {
					return false, err
				}
			}
			return true, nil
		}
	}

	return false, nil
}

// Private helper method that emits events
func (k Keeper) emitEvent(ctx sdk.Context, message proto.Message) {
	if err := ctx.EventManager().EmitTypedEvent(message); err != nil {
		k.Logger(ctx).Error("unable to emit event", "error", err, "event", message)
	}
}

// IterateExpirations iterates over all the expiration records and passes them to a callback function.
func (k Keeper) IterateExpirations(ctx sdk.Context, prefix []byte, handle Handler) error {
	// Init a name record iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer func() {
		err := iterator.Close()
		if err != nil {
			k.Logger(ctx).Error("failed to close kvStore iterator")
		}
	}()
	// Iterate over records, processing callbacks.
	for ; iterator.Valid(); iterator.Next() {
		record := types.Expiration{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if err := handle(record); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) refundDeposit(ctx sdk.Context, deposit sdk.Coins, refundTo sdk.AccAddress) error {
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, refundTo, deposit); err != nil {
		return err
	}
	return nil
}

func (k Keeper) removeExpiration(ctx sdk.Context, moduleAssetID string) error {
	key, err := types.GetModuleAssetKey(moduleAssetID)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		store.Delete(key)
	}
	return nil
}
