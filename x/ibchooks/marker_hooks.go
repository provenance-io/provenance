package ibchooks

import (
	"encoding/json"
	"strconv"
	"strings"

	sdkerrors "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	tendermintclient "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	"github.com/provenance-io/provenance/x/ibchooks/types"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

type MarkerHooks struct {
	MarkerKeeper *markerkeeper.Keeper
}

func NewMarkerHooks(markerkeeper *markerkeeper.Keeper, bankKeeper *markertypes.BankKeeper) MarkerHooks {
	return MarkerHooks{
		MarkerKeeper: markerkeeper,
	}
}

func (h MarkerHooks) ProperlyConfigured() bool {
	return h.MarkerKeeper != nil
}

func (h MarkerHooks) ProcessMarkerMemo(ctx sdktypes.Context, packet exported.PacketI, cdc codec.BinaryCodec, ibcKeeper *ibckeeper.Keeper) error {
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
		transferAuthAddrs, coinType, err := ProcessMarkerMemo(data.GetMemo())
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
				coinType,
				false, // supply fixed
				false, // allow gov
				false, // allow force transfer
				[]string{},
			)
			ResetMarkerAccessGrants(transferAuthAddrs, marker)

			if err = h.MarkerKeeper.AddMarkerAccount(ctx, marker); err != nil {
				return err
			}
			chainId := h.GetChainId(ctx, packet, cdc, ibcKeeper)
			markerMetadata := banktypes.Metadata{
				Base:        ibcDenom,
				Name:        chainId + "/" + data.Denom,
				Display:     chainId + "/" + data.Denom,
				Description: data.Denom + " from chain " + chainId,
			}
			h.MarkerKeeper.SetDenomMetaData(ctx, markerMetadata, authtypes.NewModuleAddress(types.ModuleName))
		} else {
			ResetMarkerAccessGrants(transferAuthAddrs, marker)
			h.MarkerKeeper.SetMarker(ctx, marker)
		}
	}

	// TODO: add metadata for marker

	return nil
}

func (h MarkerHooks) GetChainId(ctx sdktypes.Context, packet exported.PacketI, cdc codec.BinaryCodec, ibcKeeper *ibckeeper.Keeper) string {
	chainId := "unknown"
	channel, found := ibcKeeper.ChannelKeeper.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return chainId
	}
	connectionEnd, found := ibcKeeper.ConnectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return chainId
	}
	clientState, found := ibcKeeper.ClientKeeper.GetClientState(ctx, connectionEnd.GetClientID())
	if !found {
		return chainId
	}
	if clientState.ClientType() == "07-tendermint" {
		tmClientState, ok := clientState.(*tendermintclient.ClientState)
		if ok {
			chainId = tmClientState.ChainId
		}
	}
	return chainId
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
func ProcessMarkerMemo(memo string) ([]sdk.AccAddress, markertypes.MarkerType, error) {
	found, jsonObject := jsonStringHasKey(memo, "marker")
	if !found {
		return []sdk.AccAddress{}, markertypes.MarkerType_Coin, nil
	}
	markerPayload, ok := jsonObject["marker"].(string)
	if !ok {
		return []sdk.AccAddress{}, markertypes.MarkerType_Coin, nil
	}
	var markerMemo types.MarkerPayload
	err := json.Unmarshal([]byte(markerPayload), &markerMemo)
	if err != nil {
		return nil, markertypes.MarkerType_Unknown, err
	}
	if markerMemo.TransferAuth == nil {
		return []sdk.AccAddress{}, markertypes.MarkerType_Coin, nil
	}

	transferAuths := make([]sdk.AccAddress, len(markerMemo.TransferAuth))
	for i, addr := range markerMemo.TransferAuth {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, markertypes.MarkerType_Unknown, err
		}
		transferAuths[i] = accAddr
	}
	return transferAuths, markertypes.MarkerType_RestrictedCoin, nil
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
	if marker == nil || marker.GetMarkerType() != markertypes.MarkerType_RestrictedCoin {
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
