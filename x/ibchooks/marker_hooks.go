package ibchooks

import (
	"encoding/json"
	"strconv"
	"strings"

	sdkerrors "cosmossdk.io/errors"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
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

func NewMarkerHooks(markerkeeper *markerkeeper.Keeper) MarkerHooks {
	return MarkerHooks{
		MarkerKeeper: markerkeeper,
	}
}

// ProperlyConfigured returns false when marker hooks are configured incorrectly
func (h MarkerHooks) ProperlyConfigured() bool {
	return h.MarkerKeeper != nil
}

// AddUpdateMarker will add or update ibc Marker with transfer authorities
func (h MarkerHooks) AddUpdateMarker(ctx sdktypes.Context, packet exported.PacketI, ibcKeeper *ibckeeper.Keeper) error {
	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return err
	}
	ibcDenom := MustExtractDenomFromPacketOnRecv(packet)
	if !strings.HasPrefix(ibcDenom, "ibc/") {
		return nil
	}

	markerAddress, err := markertypes.MarkerAddress(ibcDenom)
	if err != nil {
		return err
	}
	marker, err := h.MarkerKeeper.GetMarker(ctx, markerAddress)
	if err != nil {
		return err
	}
	var transferAuthAddrs []sdktypes.AccAddress
	transferAuthAddrs, coinType, err := ProcessMarkerMemo(data.GetMemo())
	if err != nil {
		return err
	}

	if marker != nil {
		if err = ResetMarkerAccessGrants(transferAuthAddrs, marker); err != nil {
			return err
		}
		h.MarkerKeeper.SetMarker(ctx, marker)
		return nil
	}
	return h.createNewIbcMarker(data, marker, ibcDenom, coinType, err, transferAuthAddrs, ctx, packet, ibcKeeper)
}

// createNewIbcMarker creates a new marker account for ibc token
func (h MarkerHooks) createNewIbcMarker(data transfertypes.FungibleTokenPacketData, marker markertypes.MarkerAccountI, ibcDenom string, coinType markertypes.MarkerType, err error, transferAuthAddrs []sdktypes.AccAddress, ctx sdktypes.Context, packet exported.PacketI, ibcKeeper *ibckeeper.Keeper) error {
	amount, errParse := strconv.ParseInt(data.Amount, 10, 64)
	if errParse != nil {
		return errParse
	}
	marker = markertypes.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(markertypes.MustGetMarkerAddress(ibcDenom)),
		sdktypes.NewInt64Coin(ibcDenom, amount),
		nil,
		nil,
		markertypes.StatusActive,
		coinType,
		false, //supply fixed
		false, // allow gov
		false, // allow force transfer
		[]string{},
	)
	if err = ResetMarkerAccessGrants(transferAuthAddrs, marker); err != nil {
		return err
	}
	if err = h.MarkerKeeper.AddMarkerAccount(ctx, marker); err != nil {
		return err
	}
	return h.addDenomMetaData(ctx, packet, ibcKeeper, ibcDenom, data, err)
}

// addDenomMetaData adds denom metadata for ibc token
func (h MarkerHooks) addDenomMetaData(ctx sdktypes.Context, packet exported.PacketI, ibcKeeper *ibckeeper.Keeper, ibcDenom string, data transfertypes.FungibleTokenPacketData, err error) error {
	chainID := h.GetChainID(ctx, packet, ibcKeeper)
	markerMetadata := banktypes.Metadata{
		Base:        ibcDenom,
		Name:        chainID + "/" + data.Denom,
		Display:     chainID + "/" + data.Denom,
		Description: data.Denom + " from chain " + chainID,
	}
	if err = h.MarkerKeeper.SetDenomMetaData(ctx, markerMetadata, authtypes.NewModuleAddress(types.ModuleName)); err != nil {
		return err
	}
	return nil
}

// GetChainID returns the source chain id from packet for `07-tendermint` client connection or returns `unknown`
func (h MarkerHooks) GetChainID(ctx sdktypes.Context, packet exported.PacketI, ibcKeeper *ibckeeper.Keeper) string {
	chainID := "unknown"
	channel, found := ibcKeeper.ChannelKeeper.GetChannel(ctx, packet.GetSourcePort(), packet.GetSourceChannel())
	if !found {
		return chainID
	}
	connectionEnd, found := ibcKeeper.ConnectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return chainID
	}
	clientState, found := ibcKeeper.ClientKeeper.GetClientState(ctx, connectionEnd.GetClientID())
	if !found {
		return chainID
	}
	if clientState.ClientType() == "07-tendermint" {
		tmClientState, ok := clientState.(*tendermintclient.ClientState)
		if ok {
			chainID = tmClientState.ChainId
		}
	}
	return chainID
}

// ResetMarkerAccessGrants removes all current access grants from marker and adds new transfer grants for transferAuths
func ResetMarkerAccessGrants(transferAuths []sdktypes.AccAddress, marker markertypes.MarkerAccountI) error {
	for _, currentAuth := range marker.GetAccessList() {
		if err := marker.RevokeAccess(currentAuth.GetAddress()); err != nil {
			return err
		}
	}
	for _, transfAuth := range transferAuths {
		if err := marker.GrantAccess(markertypes.NewAccessGrant(transfAuth, markertypes.AccessList{markertypes.Access_Transfer})); err != nil {
			return err
		}
	}
	return nil
}

// ProcessMarkerMemo extracts the list of transfer auth address from marker part of packet memo
func ProcessMarkerMemo(memo string) ([]sdktypes.AccAddress, markertypes.MarkerType, error) {
	found, jsonObject := jsonStringHasKey(memo, "marker")
	if !found {
		return []sdktypes.AccAddress{}, markertypes.MarkerType_Coin, nil
	}
	jsonBytes, err := json.Marshal(jsonObject["marker"])
	if err != nil {
		return nil, markertypes.MarkerType_Unknown, err
	}

	var markerMemo types.MarkerPayload
	err = json.Unmarshal(jsonBytes, &markerMemo)
	if err != nil {
		return nil, markertypes.MarkerType_Unknown, err
	}
	if markerMemo.TransferAuths == nil {
		return []sdktypes.AccAddress{}, markertypes.MarkerType_Coin, nil
	}

	transferAuths := make([]sdktypes.AccAddress, len(markerMemo.TransferAuths))
	for i, addr := range markerMemo.TransferAuths {
		accAddr, err := sdktypes.AccAddressFromBech32(addr)
		if err != nil {
			return nil, markertypes.MarkerType_Unknown, err
		}
		transferAuths[i] = accAddr
	}
	return transferAuths, markertypes.MarkerType_RestrictedCoin, nil
}

// ProcessMarkerMemoFn processes a ics20 packets memo part to have `marker` setup information for receiving chain
func (h MarkerHooks) ProcessMarkerMemoFn(
	ctx sdktypes.Context,
	data []byte,
	_ map[string]interface{},
) ([]byte, error) {
	isIcs20, ics20Packet := isIcs20Packet(data)
	if !isIcs20 {
		return data, nil
	}

	memoAsJSON := SanitizeMemo(ics20Packet.GetMemo())

	markerAddress, err := markertypes.MarkerAddress(ics20Packet.Denom)
	if err != nil {
		return nil, err
	}
	marker, err := h.MarkerKeeper.GetMarker(ctx, markerAddress)
	if err != nil {
		return nil, err
	}
	memoAsJSON["marker"], err = CreateMarkerMemo(marker)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "ics20data marshall error")
	}
	memo, err := json.Marshal(memoAsJSON)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "ics20data marshall error")
	}
	ics20Packet.Memo = string(memo)

	return ics20Packet.GetBytes(), nil
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
func CreateMarkerMemo(marker markertypes.MarkerAccountI) (interface{}, error) {
	if marker == nil || marker.GetMarkerType() != markertypes.MarkerType_RestrictedCoin {
		return make(map[string]interface{}), nil
	}
	transferAuthAddrs := marker.AddressListForPermission(markertypes.Access_Transfer)
	return types.NewMarkerPayload(transferAuthAddrs), nil
}
