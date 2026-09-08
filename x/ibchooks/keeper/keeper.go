package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cometbft/cometbft/crypto/tmhash"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type (
	Keeper struct {
		cdc            codec.BinaryCodec
		storeService   corestore.KVStoreService
		schema         collections.Schema
		channelKeeper  types.ChannelKeeper
		ContractKeeper *wasmkeeper.PermissionedKeeper
		authority      string
		// Params stores the module's persistent parameters using the legacy key (0x01).
		params collections.Item[types.Params]
		// PacketCallbacks stores temporary callback state for packets.
		// Maps (channel, sequence) to a contract bech32 address (prefix 0x02).
		packetCallbacks collections.Map[collections.Pair[string, uint64], string]
		// PacketAckActors maps (channel, sequence) to "contract::hash" bytes.
		// Used to store temporary acknowledgment state for each packet (prefix 0x03).
		packetAckActors collections.Map[collections.Pair[string, uint64], []byte]
	}
)

// NewKeeper returns a new instance of the x/ibchooks keeper
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestore.KVStoreService,
	channelKeeper types.ChannelKeeper,
	contractKeeper *wasmkeeper.PermissionedKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	keeper := Keeper{
		cdc:             cdc,
		storeService:    storeService,
		channelKeeper:   channelKeeper,
		ContractKeeper:  contractKeeper,
		authority:       authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		params:          collections.NewItem(sb, types.ParamsKeyPrefix, "params", codec.CollValue[types.Params](cdc)),
		packetCallbacks: collections.NewMap(sb, types.PacketCallbackKeyPrefix, "packet_callbacks", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key), collections.StringValue),
		packetAckActors: collections.NewMap(sb, types.PacketAckKeyPrefix, "packet_ack_actors", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key), collections.BytesValue),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	keeper.schema = schema
	return keeper
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority is signer of the proposal
func (k Keeper) GetAuthority() string {
	return k.authority
}

// IsAuthority returns true if the provided address bech32 string is the authority address.
func (k Keeper) IsAuthority(addr string) bool {
	return strings.EqualFold(k.authority, addr)
}

// ValidateAuthority returns an error if the provided address is not the authority.
func (k Keeper) ValidateAuthority(addr string) error {
	if !k.IsAuthority(addr) {
		return govtypes.ErrInvalidSigner.Wrapf("expected %q got %q", k.GetAuthority(), addr)
	}
	return nil
}

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}

func GetPacketCallbackKey(channel string, packetSequence uint64) []byte {
	return []byte(fmt.Sprintf("%s::%d", channel, packetSequence))
}

func GetPacketAckKey(channel string, packetSequence uint64) []byte {
	return []byte(fmt.Sprintf("%s::%d::ack", channel, packetSequence))
}

func GeneratePacketAckValue(packet channeltypes.Packet, contract string) ([]byte, error) {
	if _, err := sdk.AccAddressFromBech32(contract); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidContractAddr, contract)
	}

	packetHash, err := hashPacket(packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not hash packet")
	}

	return []byte(fmt.Sprintf("%s::%s", contract, packetHash)), nil
}

// StorePacketCallback stores which contract will be listening for the ack or timeout of a packet
func (k Keeper) StorePacketCallback(ctx sdk.Context, channel string, packetSequence uint64, contract string) {
	if err := k.packetCallbacks.Set(ctx, collections.Join(channel, packetSequence), contract); err != nil {
		panic(err)
	}
}

// GetPacketCallback returns the bech32 addr of the contract that is expecting a callback from a packet
func (k Keeper) GetPacketCallback(ctx sdk.Context, channel string, packetSequence uint64) string {
	v, err := k.packetCallbacks.Get(ctx, collections.Join(channel, packetSequence))
	if err != nil {
		return ""
	}
	return v
}

// DeletePacketCallback deletes the callback from storage once it has been processed
func (k Keeper) DeletePacketCallback(ctx sdk.Context, channel string, packetSequence uint64) {
	if err := k.packetCallbacks.Remove(ctx, collections.Join(channel, packetSequence)); err != nil {
		panic(err)
	}
}

// StorePacketAckActor stores which contract is allowed to send an ack for the packet
func (k Keeper) StorePacketAckActor(ctx sdk.Context, packet channeltypes.Packet, contract string) {
	val, err := GeneratePacketAckValue(packet, contract)
	if err != nil {
		panic(err)
	}
	if err := k.packetAckActors.Set(ctx, collections.Join(packet.GetSourceChannel(), packet.GetSequence()), val); err != nil {
		panic(err)
	}
}

// GetPacketAckActor returns the bech32 addr  of the contract that is allowed to send an ack for the packet and the packet hash
func (k Keeper) GetPacketAckActor(ctx sdk.Context, channel string, packetSequence uint64) (string, string) {
	rawData, err := k.packetAckActors.Get(ctx, collections.Join(channel, packetSequence))
	if err != nil {
		return "", ""
	}
	if rawData == nil {
		return "", ""
	}
	data := strings.Split(string(rawData), "::")
	if len(data) != 2 {
		return "", ""
	}
	// validate that the contract is a valid bech32 addr
	if _, err := sdk.AccAddressFromBech32(data[0]); err != nil {
		return "", ""
	}
	// validate that the hash is a valid sha256sum hash
	if _, err := hex.DecodeString(data[1]); err != nil {
		return "", ""
	}

	return data[0], data[1]
}

// DeletePacketAckActor deletes the ack actor from storage once it has been used
func (k Keeper) DeletePacketAckActor(ctx sdk.Context, channel string, packetSequence uint64) {
	if err := k.packetAckActors.Remove(ctx, collections.Join(channel, packetSequence)); err != nil {
		panic(err)
	}
}

// DeriveIntermediateSender derives the sender address to be used when calling wasm hooks
func DeriveIntermediateSender(channel, originalSender, bech32Prefix string) (string, error) {
	senderStr := fmt.Sprintf("%s/%s", channel, originalSender)
	senderHash32 := address.Hash(types.SenderPrefix, []byte(senderStr))
	sender := sdk.AccAddress(senderHash32)
	return sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
}

// EmitIBCAck emits an event that the IBC packet has been acknowledged
func (k Keeper) EmitIBCAck(ctx sdk.Context, sender, channel string, packetSequence uint64) ([]byte, error) {
	contract, packetHash := k.GetPacketAckActor(ctx, channel, packetSequence)
	if contract == "" {
		return nil, fmt.Errorf("no ack actor set for channel %s packet %d", channel, packetSequence)
	}
	// Only the contract itself can request for the ack to be emitted. This will generally happen as a callback
	// when the result of other IBC actions has finished, but it could be exposed directly by the contract if the
	// proper checks are made
	if sender != contract {
		return nil, fmt.Errorf("sender %s is not allowed to send an ack for channel %s packet %d", sender, channel, packetSequence)
	}

	// Write the acknowledgement
	// Calling the contract. This could be made generic by using an interface if we want
	// to support other types of AckActors, but keeping it here for now for simplicity.
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not parse contract address")
	}

	msg := types.IBCAsync{
		RequestAck: types.RequestAck{RequestAckI: types.RequestAckI{
			PacketSequence: packetSequence,
			SourceChannel:  channel,
		}},
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not marshal message")
	}
	bz, err := k.ContractKeeper.Sudo(ctx, contractAddr, msgBytes)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not execute contract")
	}

	ack, err := types.UnmarshalIBCAck(bz)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not unmarshal into IBCAckResponse or IBCAckError")
	}
	var newAck channeltypes.Acknowledgement
	var packet channeltypes.Packet

	switch ack.Type {
	case "ack_response":
		jsonAck, marshallErr := json.Marshal(ack.AckResponse.ContractAck)
		if marshallErr != nil {
			return nil, sdkerrors.Wrap(marshallErr, "could not marshal acknowledgement")
		}
		packet = ack.AckResponse.Packet
		newAck = channeltypes.NewResultAcknowledgement(jsonAck)
	case "ack_error":
		packet = ack.AckError.Packet
		newAck = NewSuccessAckError(ctx, types.ErrAckFromContract, []byte(ack.AckError.ErrorResponse), ack.AckError.ErrorDescription)
	default:
		return nil, sdkerrors.Wrap(err, "could not unmarshal into IBCAckResponse or IBCAckError")
	}

	// Validate that the packet returned by the contract matches the one we stored when sending
	receivedPacketHash, err := hashPacket(packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not hash packet")
	}
	if receivedPacketHash != packetHash {
		return nil, sdkerrors.Wrap(types.ErrAckPacketMismatch, fmt.Sprintf("packet hash mismatch. Expected %s, got %s", packetHash, receivedPacketHash))
	}

	// Now we can write the acknowledgement
	err = k.channelKeeper.WriteAcknowledgement(ctx, packet, newAck)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not write acknowledgement")
	}

	response, err := json.Marshal(newAck)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not marshal acknowledgement")
	}
	return response, nil
}

func hashPacket(packet channeltypes.Packet) (string, error) {
	// ignore the data here. We only care about the channel information
	packet.Data = nil
	bz, err := json.Marshal(packet)
	if err != nil {
		return "", sdkerrors.Wrap(err, "could not marshal packet")
	}
	packetHash := tmhash.Sum(bz)
	return hex.EncodeToString(packetHash), nil
}

// NewSuccessAckError creates a new success acknowledgement that represents an error.
// This is useful for notifying the sender that an error has occurred in a way that does not allow
// the received tokens to be reverted (which means they shouldn't be released by the sender's ics20 escrow)
func NewSuccessAckError(ctx sdk.Context, err error, errorContent []byte, errorContexts ...string) channeltypes.Acknowledgement {
	logger := ctx.Logger().With("module", "ibc-acknowledgement-error")

	attributes := make([]sdk.Attribute, len(errorContexts)+1)
	attributes[0] = sdk.NewAttribute("error", err.Error())
	for i, s := range errorContexts {
		attributes[i+1] = sdk.NewAttribute("error-context", s)
		logger.Error(fmt.Sprintf("error-context: %v", s))
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"ibc-acknowledgement-error",
			attributes...,
		),
	})
	return channeltypes.NewResultAcknowledgement(errorContent)
}
