package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/provenance-io/provenance/x/oracle/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	icqtypes "github.com/strangelove-ventures/async-icq/v6/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (k Keeper) SendQuery(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	chanCap *capabilitytypes.Capability,
	reqs []abci.RequestQuery,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return 0, sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	data, err := icqtypes.SerializeCosmosQuery(reqs)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "could not serialize reqs into cosmos query")
	}
	icqPacketData := icqtypes.InterchainQueryPacketData{
		Data: data,
	}

	return k.createOutgoingPacket(ctx, sourcePort, sourceChannel, destinationPort, destinationChannel, chanCap, icqPacketData, timeoutTimestamp)
}

func (k Keeper) createOutgoingPacket(
	ctx sdk.Context,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	chanCap *capabilitytypes.Capability,
	icqPacketData icqtypes.InterchainQueryPacketData,
	timeoutTimestamp uint64,
) (uint64, error) {
	if err := icqPacketData.ValidateBasic(); err != nil {
		return 0, sdkerrors.Wrap(err, "invalid interchain query packet data")
	}

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return 0, sdkerrors.Wrapf(channeltypes.ErrSequenceSendNotFound, "failed to retrieve next sequence send for channel %s on port %s", sourceChannel, sourcePort)
	}

	packet := channeltypes.NewPacket(
		icqPacketData.GetBytes(),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		clienttypes.ZeroHeight(),
		timeoutTimestamp,
	)

	if _, err := k.ics4Wrapper.SendPacket(ctx, chanCap, packet.SourcePort, packet.SourceChannel, packet.TimeoutHeight, packet.TimeoutTimestamp, packet.Data); err != nil {
		return 0, err
	}

	return packet.Sequence, nil
}

func (k Keeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	ack channeltypes.Acknowledgement,
) error {
	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		var ackData icqtypes.InterchainQueryPacketAck
		if err := icqtypes.ModuleCdc.UnmarshalJSON(resp.Result, &ackData); err != nil {
			return sdkerrors.Wrap(err, "failed to unmarshal interchain query packet ack")
		}
		resps, err := icqtypes.DeserializeCosmosResponse(ackData.Data)
		if err != nil {
			return sdkerrors.Wrap(err, "could not deserialize data to cosmos response")
		}

		if len(resps) < 1 {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "no responses in interchain query packet ack")
		}

		var r banktypes.QueryAllBalancesResponse
		if err := k.cdc.Unmarshal(resps[0].Value, &r); err != nil {
			return sdkerrors.Wrapf(err, "failed to unmarshal interchain query response to type %T", resp)
		}

		k.SetQueryResponse(ctx, modulePacket.Sequence, r)
		k.SetLastQueryPacketSeq(ctx, modulePacket.Sequence)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeQueryResult,
				sdk.NewAttribute(types.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)

		k.Logger(ctx).Info("interchain query response", "sequence", modulePacket.Sequence, "response", r)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeQueryResult,
				sdk.NewAttribute(types.AttributeKeyAckError, resp.Error),
			),
		)

		k.Logger(ctx).Error("interchain query response", "sequence", modulePacket.Sequence, "error", resp.Error)
	}
	return nil
}
