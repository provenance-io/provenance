package ibchooks

import (
	"encoding/json"
	"strconv"
	"strings"

	sdkerrors "cosmossdk.io/errors"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"

	"github.com/provenance-io/provenance/x/ibchooks/types"
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
	isIcs20, data := isIcs20Packet(packet.GetData())
	if !isIcs20 {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	ibcDenom := MustExtractDenomFromPacketOnRecv(packet)
	if strings.HasPrefix(ibcDenom, "ibc/") {
		markerAddress, err := markertypes.MarkerAddress(ibcDenom)
		if err != nil {
			//TODO: emit some kind of event, proceed as normal
			return im.App.OnRecvPacket(ctx, packet, relayer)
		}
		marker, err := h.MarkerKeeper.GetMarker(ctx, markerAddress)
		if err != nil {
			// TODO: emit some kind of event, proceed as normal
			return im.App.OnRecvPacket(ctx, packet, relayer)
		}
		if marker == nil {
			amount, err := strconv.ParseInt(data.Amount, 10, 64)
			if err != nil {
				//TODO: emit some kind of event, proceed as normal
				return im.App.OnRecvPacket(ctx, packet, relayer)
			}
			marker = markertypes.NewMarkerAccount(
				authtypes.NewBaseAccountWithAddress(markertypes.MustGetMarkerAddress(ibcDenom)),
				sdktypes.NewInt64Coin(ibcDenom, amount),
				nil,
				nil,
				markertypes.StatusActive,
				markertypes.MarkerType_Coin,
				false, // supply fixed
				false, // allow gov
				false, // allow force transfer
				[]string{},
			)
			if err = h.MarkerKeeper.AddMarkerAccount(ctx, marker); err != nil {
				//TODO: emit some kind of event, proceed as normal
				return im.App.OnRecvPacket(ctx, packet, relayer)
			}
			// metadata := banktypes.Metadata{Base: ibcDenom, Name: "chain-id/" + data.Denom, Display: "chain-id/" + data.Denom}
			// im.bankKeeper.SetDenomMetaData(ctx, metadata)
		}
	}

	// TODO: check if there is a memo with marker key and transfer auths to update.

	return im.App.OnRecvPacket(ctx, packet, relayer)
}

// func ProcessMemo(memo string) ([]markertypes.AccessGrant, error) {
// 	var markerMemo types.MarkerMemo
// 	err := json.Unmarshal([]byte(memo), &markerMemo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	len := len(markerMemo.Marker.TransferAuth)
// 	accessGrants := make(markertypes.AccessGrant{}, len)
// }

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
	if !isIcs20 {
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	markerAddress, err := markertypes.MarkerAddress(ics20Packet.Denom)
	if err != nil {
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}
	marker, err := h.MarkerKeeper.GetMarker(ctx, markerAddress)
	if err != nil {
		return i.channel.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}
	ics20Packet.Memo = ""
	if marker != nil {
		ics20Packet.Memo, err = CreateMarkerMemo(marker)
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

func (h MarkerHooks) SendPacketFn(
	ctx sdktypes.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) ([]byte, error) {
	isIcs20, ics20Packet := isIcs20Packet(data)
	if !isIcs20 {
		return data, nil
	}

	memoAsJson := SanitizeMemo(ics20Packet.GetMemo())

	markerAddress, err := markertypes.MarkerAddress(ics20Packet.Denom)
	if err != nil {
		return nil, err
	}
	marker, err := h.MarkerKeeper.GetMarker(ctx, markerAddress)
	if err != nil {
		return nil, err
	}
	memoAsJson["marker"], err = CreateMarkerMemo(marker)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "ics20data marshall error")
	}
	memo, err := json.Marshal(memoAsJson)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "ics20data marshall error")
	}
	ics20Packet.Memo = string(memo)
	dataBytes, err := json.Marshal(ics20Packet)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "ics20data marshall error")
	}

	return dataBytes, nil
}

// SanitizeMemo returns a keyed json object for memo
func SanitizeMemo(memo string) map[string]interface{} {
	jsonObject := make(map[string]interface{})
	if len(memo) != 0 {
		err := json.Unmarshal([]byte(memo), &jsonObject)
		if err != nil {
			jsonObject["memo"] = memo
		}
	}
	return jsonObject
}

// CreateMarkerMemo returns a json memo for marker
func CreateMarkerMemo(marker markertypes.MarkerAccountI) (string, error) {
	if marker == nil {
		return "{}", nil
	}
	transferAuthAddrs := marker.AddressListForPermission(markertypes.Access_Transfer)
	markerPayload := types.NewMarkerPayload(transferAuthAddrs)
	jsonData, err := json.Marshal(markerPayload)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
