package ibchooks

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"

	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

type MarkerHooks struct {
	MarkerKeeper *markerkeeper.Keeper
}

func NewMarkerHooks(markerkeeper *markerkeeper.Keeper) MarkerHooks {
	return MarkerHooks{
		MarkerKeeper: markerkeeper,
	}
}

func (h MarkerHooks) ProperlyConfigured() bool {
	return h.MarkerKeeper != nil
}

func (h MarkerHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, relayer sdktypes.AccAddress) ibcexported.Acknowledgement {
	// isIcs20, data := isIcs20Packet(packet.GetData())
	// if !isIcs20 {
	// 	return im.App.OnRecvPacket(ctx, packet, relayer)
	// }
	// packet.
	return im.App.OnRecvPacket(ctx, packet, relayer)
}

func (h MarkerHooks) SendPacketOverride(
	i ICS4Middleware,
	ctx sdktypes.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	isIcs20, ics20Packet := isIcs20Packet(data)
	if !isIcs20 || ics20Packet.Memo != "" {
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}
	markerAddr, err := markertypes.MarkerAddress(ics20Packet.Denom)
	if err != nil {
		//TODO: emit some kind of event, proceed as normal
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)

	}
	marker, err := h.MarkerKeeper.GetMarker(ctx, markerAddr)
	if err != nil {
		//TODO: emit some kind of event, proceed as normal
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}
	if marker != nil {
		ics20Packet.Memo, err = CreateMarkerMemo(ctx, marker)
		if err != nil {
			return 0, sdkerrors.Wrap(err, "ics20data marshall error")
		}
	}
	dataBytes, err := json.Marshal(ics20Packet)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "ics20data marshall error")
	}

	return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, dataBytes)
}

type MarkerMemo struct {
	Marker MarkerPayload `json:"marker"`
}

type MarkerPayload struct {
	ChainId      string   `json:"chain-id"`
	TransferAuth []string `json:"transfer-auth"`
}

func CreateMarkerMemo(ctx sdktypes.Context, marker markertypes.MarkerAccountI) (string, error) {
	transferAuthAddr := marker.AddressListForPermission(markertypes.Access_Transfer)
	addresses := make([]string, len(transferAuthAddr))
	for i := 0; i < len(transferAuthAddr); i++ {
		addresses[i] = transferAuthAddr[i].String()
	}

	memo := MarkerMemo{Marker: MarkerPayload{
		ChainId:      ctx.ChainID(),
		TransferAuth: addresses,
	}}

	jsonData, err := json.Marshal(memo)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
