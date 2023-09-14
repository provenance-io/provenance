package ibchooks

import (
	"encoding/json"
	"strconv"
	"strings"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"

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

func (h MarkerHooks) ProcessMarkerMemo(ctx sdktypes.Context, packet exported.PacketI) error {
	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return err
	}
	ibcDenom := MustExtractDenomFromPacketOnRecv(packet)
	if strings.HasPrefix(ibcDenom, "ibc/") {
		markerAddress, err := markertypes.MarkerAddress(ibcDenom)
		if err != nil {
			return err
		}
		marker, err := h.MarkerKeeper.GetMarker(ctx, markerAddress)
		if err != nil {
			return err
		}
		var transferAuthAddrs []sdk.AccAddress
		transferAuthAddrs, err = ProcessMarkerMemo(data.GetDenom())
		if err != nil {
			return err
		}
		if marker == nil {
			amount, err := strconv.ParseInt(data.Amount, 10, 64)
			if err != nil {
				return err
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
			ResetMarkerAccessGrants(transferAuthAddrs, marker)

			if err = h.MarkerKeeper.AddMarkerAccount(ctx, marker); err != nil {
				return err
			}
		} else {
			ResetMarkerAccessGrants(transferAuthAddrs, marker)
			h.MarkerKeeper.SetMarker(ctx, marker)
		}
	}

	// TODO: add metadata for marker

	return nil
}

// ResetMarkerAccessGrants removes all current access grants from marker and adds new transfer grants for transferAuths
func ResetMarkerAccessGrants(transferAuths []sdk.AccAddress, marker markertypes.MarkerAccountI) {
	for _, currentAuth := range marker.GetAccessList() {
		marker.RevokeAccess(currentAuth.GetAddress())
	}
	for _, transfAuth := range transferAuths {
		marker.GrantAccess(markertypes.NewAccessGrant(transfAuth, markertypes.AccessList{markertypes.Access_Transfer}))
	}
}

// ProcessMarkerMemo extracts the list of transfer auth address from marker part of packet memo
func ProcessMarkerMemo(memo string) ([]sdk.AccAddress, error) {
	found, jsonObject := jsonStringHasKey(memo, "marker")
	if !found {
		return []sdk.AccAddress{}, nil
	}
	markerPayload, ok := jsonObject["marker"].(string)
	if !ok {
		return []sdk.AccAddress{}, nil
	}
	var markerMemo types.MarkerPayload
	err := json.Unmarshal([]byte(markerPayload), &markerMemo)
	if err != nil {
		return nil, err
	}

	transferAuths := make([]sdk.AccAddress, len(markerMemo.TransferAuth))
	for i, addr := range markerMemo.TransferAuth {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, err
		}
		transferAuths[i] = accAddr
	}
	return transferAuths, nil
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
	processData map[string]interface{},
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
