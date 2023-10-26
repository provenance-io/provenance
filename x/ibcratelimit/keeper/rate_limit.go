package keeper

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// CheckAndUpdateRateLimits Updates the rate limiter and checks if rate limit has been exceeded.
func (k Keeper) CheckAndUpdateRateLimits(ctx sdk.Context, msgType string, packet exported.PacketI) error {
	contract := k.GetContractAddress(ctx)

	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return errorsmod.Wrap(ibcratelimit.ErrContractError, err.Error())
	}

	sendPacketMsg, err := k.buildWasmExecMsg(msgType, packet)
	if err != nil {
		return errorsmod.Wrap(ibcratelimit.ErrContractError, err.Error())
	}

	_, err = k.PermissionedKeeper.Sudo(ctx, contractAddr, sendPacketMsg)
	if err != nil {
		return errorsmod.Wrap(ibcratelimit.ErrRateLimitExceeded, err.Error())
	}

	return nil
}

// UndoSendRateLimit Undos the changes made to the rate limiter.
func (k Keeper) UndoSendRateLimit(ctx sdk.Context, contract string, packet exported.PacketI) error {
	contractAddr, err := sdk.AccAddressFromBech32(contract)
	if err != nil {
		return err
	}

	unwrapped, err := ibcratelimit.UnwrapPacket(packet)
	if err != nil {
		return err
	}

	msg := ibcratelimit.UndoSendMsg{UndoSend: ibcratelimit.UndoPacketMsg{Packet: unwrapped}}
	asJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = k.PermissionedKeeper.Sudo(ctx, contractAddr, asJSON)
	if err != nil {
		return errorsmod.Wrap(ibcratelimit.ErrContractError, err.Error())
	}

	return nil
}

// buildWasmExecMsg Constructs a Wasm Execute Message from a packet and type.
func (k Keeper) buildWasmExecMsg(msgType string, packet exported.PacketI) ([]byte, error) {
	unwrapped, err := ibcratelimit.UnwrapPacket(packet)
	if err != nil {
		return []byte{}, err
	}

	var asJSON []byte
	switch {
	case msgType == ibcratelimit.MsgSendPacket:
		msg := ibcratelimit.SendPacketMsg{SendPacket: ibcratelimit.PacketMsg{
			Packet: unwrapped,
		}}
		asJSON, err = json.Marshal(msg)
	case msgType == ibcratelimit.MsgRecvPacket:
		msg := ibcratelimit.RecvPacketMsg{RecvPacket: ibcratelimit.PacketMsg{
			Packet: unwrapped,
		}}
		asJSON, err = json.Marshal(msg)
	default:
		return []byte{}, ibcratelimit.ErrBadMessage
	}

	if err != nil {
		return []byte{}, err
	}

	return asJSON, nil
}

// RevertSentPacket Notifies the contract that a sent packet wasn't properly received.
func (k Keeper) RevertSentPacket(
	ctx sdk.Context,
	packet exported.PacketI,
) error {
	if !k.ContractConfigured(ctx) {
		return nil
	}

	contract := k.GetContractAddress(ctx)
	return k.UndoSendRateLimit(ctx, contract, packet)
}
