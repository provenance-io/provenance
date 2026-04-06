package v2

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	"github.com/cosmos/ibc-go/v10/modules/core/api"

	"github.com/provenance-io/provenance/internal/ibc"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/provenance-io/provenance/x/ibcratelimit/keeper"
)

var _ api.IBCModule = (*IBCMiddleware)(nil)

// IBCMiddleware is the IBC v2 middleware for the ibcratelimit module.
type IBCMiddleware struct {
	app    api.IBCModule
	keeper *keeper.Keeper
}

// NewIBCMiddleware creates a new IBC v2 rate-limiting middleware.
func NewIBCMiddleware(k *keeper.Keeper, app api.IBCModule) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// OnSendPacket implements api.IBCModule.
func (im IBCMiddleware) OnSendPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	signer sdk.AccAddress,
) error {
	if !im.keeper.IsContractConfigured(ctx) {
		return im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, payload, signer)
	}

	packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("rate limit OnSendPacket: failed to convert v2 packet to v1: %s", err.Error()))
		return err
	}

	if err := im.keeper.CheckAndUpdateRateLimits(ctx, "send_packet", packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 packet send was denied: %s", err.Error()))
		return err
	}

	return im.app.OnSendPacket(ctx, sourceClient, destinationClient, sequence, payload, signer)
}

// OnRecvPacket implements api.IBCModule.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) channeltypesv2.RecvPacketResult {
	if !im.keeper.IsContractConfigured(ctx) {
		return im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
	}

	packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("rate limit OnRecvPacket: failed to convert v2 packet to v1: %s", err.Error()))
		return channeltypesv2.RecvPacketResult{
			Status:          channeltypesv2.PacketStatus_Failure,
			Acknowledgement: channeltypes.NewErrorAcknowledgement(err).Acknowledgement(),
		}
	}

	if err := im.keeper.CheckAndUpdateRateLimits(ctx, "recv_packet", packet); err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("ICS20 packet receive was denied: %s", err.Error()))
		return channeltypesv2.RecvPacketResult{
			Status:          channeltypesv2.PacketStatus_Failure,
			Acknowledgement: channeltypes.NewErrorAcknowledgement(err).Acknowledgement(),
		}
	}

	return im.app.OnRecvPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
}

// OnTimeoutPacket implements api.IBCModule.
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
	if err != nil {
		im.keeper.Logger(ctx).Error(fmt.Sprintf("rate limit OnTimeoutPacket: failed to convert v2 packet to v1: %s", err.Error()))
		return err
	}

	if err := im.keeper.RevertSentPacket(ctx, packet); err != nil {
		eventError := ctx.EventManager().EmitTypedEvent(ibcratelimit.NewEventTimeoutRevertFailure(
			ibcratelimit.ModuleName,
			string(packet.GetData()),
		))
		if eventError != nil {
			ctx.Logger().Error("unable to emit TimeoutRevertFailure event", "err", eventError)
		}
	}

	return im.app.OnTimeoutPacket(ctx, sourceClient, destinationClient, sequence, payload, relayer)
}

// OnAcknowledgementPacket implements api.IBCModule.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	sourceClient string,
	destinationClient string,
	sequence uint64,
	acknowledgement []byte,
	payload channeltypesv2.Payload,
	relayer sdk.AccAddress,
) error {
	if ibc.IsAckError(acknowledgement) {
		packet, err := v2ToV1Packet(payload, sourceClient, destinationClient, sequence)
		if err != nil {
			im.keeper.Logger(ctx).Error(fmt.Sprintf("rate limit OnAcknowledgementPacket: failed to convert v2 packet to v1: %s", err.Error()))
			return err
		}

		if err := im.keeper.RevertSentPacket(ctx, packet); err != nil {
			eventError := ctx.EventManager().EmitTypedEvent(ibcratelimit.NewEventAckRevertFailure(
				ibcratelimit.ModuleName,
				string(packet.GetData()),
				string(acknowledgement),
			))
			if eventError != nil {
				ctx.Logger().Error("unable to emit AckRevertFailure event", "err", eventError)
			}
		}
	}

	return im.app.OnAcknowledgementPacket(ctx, sourceClient, destinationClient, sequence, acknowledgement, payload, relayer)
}

// v2ToV1Packet converts a v2 Payload into a v1 channeltypes.Packet so the existing
// keeper rate-limit logic (which operates on exported.PacketI) can be reused.
func v2ToV1Packet(payload channeltypesv2.Payload, sourceClient, destinationClient string, sequence uint64) (channeltypes.Packet, error) {
	transferData, err := transfertypes.UnmarshalPacketData(payload.Value, payload.Version, payload.Encoding)
	if err != nil {
		return channeltypes.Packet{}, err
	}

	packetData := transfertypes.FungibleTokenPacketData{
		Denom:    transferData.Token.Denom.Path(),
		Amount:   transferData.Token.Amount,
		Sender:   transferData.Sender,
		Receiver: transferData.Receiver,
		Memo:     transferData.Memo,
	}

	packetDataBz, err := json.Marshal(packetData)
	if err != nil {
		return channeltypes.Packet{}, err
	}

	return channeltypes.Packet{
		Sequence:           sequence,
		SourcePort:         payload.SourcePort,
		SourceChannel:      sourceClient,
		DestinationPort:    payload.DestinationPort,
		DestinationChannel: destinationClient,
		Data:               packetDataBz,
		TimeoutHeight:      clienttypes.Height{},
		TimeoutTimestamp:   0,
	}, nil
}
