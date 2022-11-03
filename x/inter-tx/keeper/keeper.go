package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/tendermint/tendermint/libs/log"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/keeper"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	channelkeeper "github.com/cosmos/ibc-go/v5/modules/core/04-channel/keeper"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	"github.com/provenance-io/provenance/x/inter-tx/types"
)

type InterTxKeeperI interface {
	ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error
	SubmitTx(ctx sdk.Context, msg *types.MsgSubmitTx, timeout time.Duration, success SuccessCallbackFunc, failure FailureCallbackFunc) error
	RegisterAccount(ctx sdk.Context, msg *types.MsgRegisterAccount) error
}

type InterTxCallbackI interface {
	SuccessCallback(ctx sdk.Context, packet channeltypes.Packet, message *codectypes.Any) error
	FailureCallback(ctx sdk.Context, packet channeltypes.Packet, err string) error
}

type SuccessCallbackFunc = (func(ctx sdk.Context, message *codectypes.Any) error)
type FailureCallbackFunc = (func(ctx sdk.Context, err string) error)

type Keeper struct {
	cdc codec.Codec

	storeKey storetypes.StoreKey

	scopedKeeper        capabilitykeeper.ScopedKeeper
	icaControllerKeeper icacontrollerkeeper.Keeper
	successCallbacks    map[uint64]SuccessCallbackFunc
	failureCallbacks    map[uint64]FailureCallbackFunc
	channelKeeper       channelkeeper.Keeper
}

func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, iaKeeper icacontrollerkeeper.Keeper, scopedKeeper capabilitykeeper.ScopedKeeper, channelKeeper channelkeeper.Keeper) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		scopedKeeper:        scopedKeeper,
		icaControllerKeeper: iaKeeper,
		successCallbacks:    make(map[uint64]SuccessCallbackFunc),
		failureCallbacks:    make(map[uint64]FailureCallbackFunc),
		channelKeeper:       channelKeeper,
	}
}

// Logger returns the application logger, scoped to the associated module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

var _ InterTxKeeperI = &Keeper{}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k *Keeper) GetInterChainAccountAddress(ctx sdk.Context, connectionID, ownerId string) (string, bool) {
	portID, err := icatypes.NewControllerPortID(ownerId)
	if err != nil {
		return "", false
	}
	return k.icaControllerKeeper.GetInterchainAccountAddress(ctx, connectionID, portID)

}

func (k *Keeper) SubmitTx(ctx sdk.Context, msg *types.MsgSubmitTx, timeout time.Duration, success SuccessCallbackFunc, failure FailureCallbackFunc) error {
	portID, err := icatypes.NewControllerPortID(msg.Owner)
	if err != nil {
		return err
	}

	channelID, found := k.icaControllerKeeper.GetActiveChannelID(ctx, msg.ConnectionId, portID)
	if !found {
		return sdkerrors.Wrapf(icatypes.ErrActiveChannelNotFound, "failed to retrieve active channel for port %s", portID)
	}

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(portID, channelID))
	if !found {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	data, err := icatypes.SerializeCosmosTx(k.cdc, []sdk.Msg{msg.GetTxMsg()})
	if err != nil {
		return err
	}

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to satisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	timeoutTimestamp := ctx.BlockTime().Add(timeout).UnixNano()

	nextSequence, _ := k.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)
	k.successCallbacks[nextSequence] = success
	k.failureCallbacks[nextSequence] = failure

	_, err = k.icaControllerKeeper.SendTx(ctx, chanCap, msg.ConnectionId, portID, packetData, uint64(timeoutTimestamp))
	if err != nil {
		return err
	}

	return nil
}

// RegisterAccount implements the Msg/RegisterAccount interface
func (k *Keeper) RegisterAccount(ctx sdk.Context, msg *types.MsgRegisterAccount) error {
	if err := k.icaControllerKeeper.RegisterInterchainAccount(ctx, msg.ConnectionId, msg.Owner, msg.Version); err != nil {
		return err
	}
	return nil
}

func (k *Keeper) SuccessCallback(ctx sdk.Context, packet channeltypes.Packet, message *codectypes.Any) error {
	if callback, ok := k.successCallbacks[packet.Sequence]; ok && callback != nil {
		err := callback(ctx, message)
		delete(k.successCallbacks, packet.Sequence)
		delete(k.failureCallbacks, packet.Sequence)
		return err
	}

	// We received a packet that does not have a callback. This is an error
	return nil
}

func (k *Keeper) FailureCallback(ctx sdk.Context, packet channeltypes.Packet, err string) error {
	if callback, ok := k.failureCallbacks[packet.Sequence]; ok && callback != nil {
		err := callback(ctx, err)
		delete(k.successCallbacks, packet.Sequence)
		delete(k.failureCallbacks, packet.Sequence)
		return err
	}

	// We received a packet that does not have a callback. This is an error
	return nil
}
