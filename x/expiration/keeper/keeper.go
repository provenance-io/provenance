package keeper

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/provenance-io/provenance/x/expiration/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

// Handler is a name record handler function for use with IterateExpirations.
type Handler func(record types.Expiration) error

// Keeper defines the name module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set account specific params
	paramSpace paramtypes.Subspace

	// Key to access the key-value store from sdk.Context.
	storeKey sdk.StoreKey

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
	key sdk.StoreKey,
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

	// ensure governance module account is set
	if addr := acctKeeper.GetModuleAddress(types.ModuleName); addr == nil {
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

// GetDeposit returns the default deposit used in setting module asset expirations
func (k Keeper) GetDeposit(ctx sdk.Context) sdk.Coin {
	deposit := &types.DefaultDeposit
	k.paramSpace.GetIfExists(ctx, types.ParamStoreKeyDeposit, deposit)
	return *deposit
}

// GetExpiration returns the expiration with the given module asset id.
func (k Keeper) GetExpiration(ctx sdk.Context, moduleAssetID string) (*types.Expiration, error) {
	key, err := types.GetModuleAssetKeyPrefix(moduleAssetID)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidKeyPrefix, err.Error())
	}

	store := ctx.KVStore(k.storeKey)
	if !store.Has(key) {
		return nil, sdkerrors.Wrap(types.ErrNotFound,
			fmt.Sprintf("expiration for module asset id [%s] does not exist", moduleAssetID))
	}

	b := store.Get(key)
	expiration := &types.Expiration{}
	err = k.cdc.Unmarshal(b, expiration)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrUnmarshal, err.Error())
	}

	return expiration, nil
}

// SetExpiration creates an expiration record for a module asset
func (k Keeper) SetExpiration(ctx sdk.Context, expiration types.Expiration) error {
	// get store key prefix
	store := ctx.KVStore(k.storeKey)
	key, err := types.GetModuleAssetKeyPrefix(expiration.ModuleAssetId)
	if err != nil {
		return err
	}

	// move deposit from owner account into expiration module account
	ownerAddr, addrErr := sdk.AccAddressFromBech32(expiration.Owner)
	if addrErr != nil {
		return addrErr
	}

	// attempt to send coins from owner account to expiration module account
	depErr := k.bankKeeper.SendCoinsFromAccountToModule(ctx, ownerAddr,
		types.ModuleName, sdk.NewCoins(expiration.Deposit))
	if depErr != nil {
		return sdkerrors.Wrapf(types.ErrInsufficientDeposit, fmt.Sprintf(": %v", depErr))
	}

	// marshal expiration record and store
	b, err := k.cdc.Marshal(&expiration)
	if err != nil {
		return err
	}
	store.Set(key, b)

	// emit event indicating the expiration was added
	addEvent := types.NewEventExpirationAdd(expiration.ModuleAssetId)
	if err := k.emitEvent(ctx, addEvent); err != nil {
		return err
	}

	// emit event indicating that a deposit was collected
	return k.emitEvent(ctx, types.NewEventExpirationDeposit(expiration.ModuleAssetId, expiration.Owner, expiration.Deposit))
}

func (k Keeper) ExtendExpiration(ctx sdk.Context, expiration types.Expiration) error {
	// get key prefix
	key, err := types.GetModuleAssetKeyPrefix(expiration.ModuleAssetId)
	if err != nil {
		return err
	}

	// lookup old expiration
	oldExpiration, err := k.GetExpiration(ctx, expiration.ModuleAssetId)
	if err != nil {
		return err
	}

	// Make sure that the new block height is higher than the old block height
	if oldExpiration.Time.After(expiration.Time) {
		k.Logger(ctx).Error("new expiration time must be after old expiration time", "err", err, "expiration", expiration.String(), "oldExpiration", oldExpiration.String())
		return types.ErrExtendExpiration
	}
	// Validate owners are the same
	if expiration.Owner != oldExpiration.Owner {
		k.Logger(ctx).Error("new owner and old owner do not match", "err", err, "expiration", expiration.String(), "oldExpiration", oldExpiration.String())
		return types.ErrNewOwnerNoMatch
	}

	// Marshal expiration record and store
	store := ctx.KVStore(k.storeKey)
	b, err := k.cdc.Marshal(&expiration)
	if err != nil {
		return err
	}
	store.Set(key, b)

	// emit Extend event
	extendEvent := types.NewEventExpirationExtend(expiration.ModuleAssetId)
	return k.emitEvent(ctx, extendEvent)
}

// deleteExpiration removes an expiration record from the kvstore.
func (k Keeper) deleteExpiration(ctx sdk.Context, moduleAssetID string) error {
	key, err := types.GetModuleAssetKeyPrefix(moduleAssetID)
	if err != nil {
		return err
	}

	// delete record from store
	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		store.Delete(key)
	}

	return nil
}

func (k Keeper) InvokeExpiration(ctx sdk.Context, moduleAssetID string, refundTo sdk.AccAddress) error {
	// lookup expiration
	expiration, err := k.GetExpiration(ctx, moduleAssetID)
	if err != nil {
		return err
	}

	// unpack expiration message
	var msg sdk.Msg
	if err = k.cdc.UnpackAny(&expiration.Message, &msg); err != nil {
		return err
	}

	// route message to module
	handler := k.router.Handler(msg)
	if handler == nil {
		return errors.Errorf("no message handler found for %q", sdk.MsgTypeURL(msg))
	}
	r, err := handler(ctx, msg)
	if err != nil {
		return errors.Wrapf(err, "message %s", sdk.MsgTypeURL(msg))
	}
	// Handler should always return non-nil sdk.Result.
	if r == nil {
		return fmt.Errorf("got nil sdk.Result for message %q", msg)
	}

	// clean up
	if err := k.deleteExpiration(ctx, moduleAssetID); err != nil {
		return err
	}

	// refund deposit from expiration module account to depositor
	refundErr := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, refundTo, sdk.NewCoins(expiration.Deposit))
	if refundErr != nil {
		return refundErr
	}

	// emit Invoke event
	invokeEvent := types.NewEventExpirationInvoke(moduleAssetID)
	return k.emitEvent(ctx, invokeEvent)
}

func (k Keeper) ValidateSetExpiration(
	ctx sdk.Context,
	expiration types.Expiration,
	signers []string,
	msgTypeURL string,
) error {
	// validate block height is in the future
	if expiration.Time.Before(ctx.BlockTime()) {
		return sdkerrors.Wrap(types.ErrTimeInPast,
			fmt.Sprintf("expiration time %s must be in the future",
				expiration.Time))
	}

	// validate deposit
	if err := expiration.Deposit.Validate(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, err.Error())
	}
	deposit := expiration.Deposit
	defaultDeposit := types.DefaultDeposit
	if deposit.IsLT(defaultDeposit) {
		return sdkerrors.Wrap(types.ErrInvalidDeposit,
			fmt.Sprintf("deposit amount %s is less than minimum deposit amount %s",
				deposit.Amount.String(), defaultDeposit.Amount.String()))
	}

	// validate module asset id
	if _, err := sdk.AccAddressFromBech32(expiration.ModuleAssetId); err != nil {
		// check if we're dealing with a MetadataAddress
		if _, err2 := metadatatypes.MetadataAddressFromBech32(expiration.ModuleAssetId); err2 != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress,
				fmt.Sprintf("invalid module asset id: %s", err.Error()))
		}
	}

	// validate signers
	if err := k.validateSigners(ctx, expiration.Owner, signers, msgTypeURL); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidSigners, err.Error())
	}

	return nil
}

func (k Keeper) ValidateInvokeExpiration(
	ctx sdk.Context,
	moduleAssetID string,
	signers []string,
	msgTypeURL string,
) (*types.Expiration, error) {
	expiration, err := k.GetExpiration(ctx, moduleAssetID)
	if err != nil {
		return nil, err
	}

	// anyone can delete an expired expiration
	if ctx.BlockTime().After(expiration.Time) {
		return expiration, nil
	}

	// validate signers
	if err := k.validateSigners(ctx, expiration.Owner, signers, msgTypeURL); err != nil {
		return expiration, sdkerrors.Wrap(types.ErrInvalidSigners, err.Error())
	}

	return expiration, nil
}

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
				return nil, err
			}
			return addr, nil
		}
	}
	// fall back to first signer if after expiration
	if ctx.BlockTime().After(expiration.Time) && len(msg.Signers) >= 1 {
		addr, err := sdk.AccAddressFromBech32(msg.Signers[0])
		if err != nil {
			return nil, err
		}
		return addr, nil
	}

	// error if no qualifying depositors are found
	return nil, fmt.Errorf("unable to resolve depositor for asset %s", msg.ModuleAssetId)
}

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

	authorization, exp := k.authzKeeper.GetCleanAuthorization(ctx, grantee, granter, msgTypeURL)
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

func (k Keeper) emitEvent(ctx sdk.Context, message proto.Message) error {
	if err := ctx.EventManager().EmitTypedEvent(message); err != nil {
		k.Logger(ctx).Error("unable to emit event", "error", err, "event", message)
		return err
	}
	return nil
}

// IterateExpirations iterates over all the stored name records and passes them to a callback function.
func (k Keeper) IterateExpirations(ctx sdk.Context, prefix []byte, handle Handler) error {
	// Init a name record iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer func(iterator sdk.Iterator) {
		err := iterator.Close()
		if err != nil {
			k.Logger(ctx).Error("failed to close kvStore iterator")
		}
	}(iterator)
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
