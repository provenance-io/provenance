package ibchooks

import (
	"encoding/json"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	tendermintclient "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

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

// AddMarker will add a marker and denom metadata for an ibc denom.
func (h MarkerHooks) AddMarker(ctx sdktypes.Context, packet exported.PacketI, ibcKeeper *ibckeeper.Keeper) error {
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
	// If we already have a marker, we're good here.
	if marker != nil {
		return nil
	}
	return h.createNewIbcMarker(ctx, data, ibcDenom, packet, ibcKeeper)
}

// createNewIbcMarker creates a new marker account for ibc token
func (h MarkerHooks) createNewIbcMarker(ctx sdktypes.Context, data transfertypes.FungibleTokenPacketData, ibcDenom string, packet exported.PacketI, ibcKeeper *ibckeeper.Keeper) error {
	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		return fmt.Errorf("invalid amount %q", data.Amount)
	}
	marker := markertypes.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(markertypes.MustGetMarkerAddress(ibcDenom)),
		sdktypes.NewCoin(ibcDenom, amount),
		nil,
		nil,
		markertypes.StatusActive,
		markertypes.MarkerType_Coin,
		false, // supply fixed
		false, // allow gov
		false, // force transfer not allowed.
		[]string{},
	)
	existingSupply := h.getExistingSupply(ctx, marker)
	_ = marker.SetSupply(marker.GetSupply().Add(existingSupply))
	if err := h.MarkerKeeper.AddMarkerAccount(ctx, marker); err != nil {
		return err
	}
	return h.addDenomMetaData(ctx, packet, ibcKeeper, ibcDenom, data)
}

// getExistingSupply returns current supply coin, if coin does not exist amount will be 0
func (h MarkerHooks) getExistingSupply(ctx sdktypes.Context, marker *markertypes.MarkerAccount) sdktypes.Coin {
	return sdktypes.NewCoin(marker.Denom, h.MarkerKeeper.CurrentCirculation(ctx, marker))
}

// addDenomMetaData adds denom metadata for ibc token
func (h MarkerHooks) addDenomMetaData(ctx sdktypes.Context, packet exported.PacketI, ibcKeeper *ibckeeper.Keeper, ibcDenom string, data transfertypes.FungibleTokenPacketData) error {
	chainID := h.GetChainID(ctx, packet.GetDestPort(), packet.GetDestChannel(), ibcKeeper)
	markerMetadata := banktypes.Metadata{
		Base:        ibcDenom,
		Name:        chainID + "/" + data.Denom,
		Display:     chainID + "/" + data.Denom,
		Description: data.Denom + " from " + chainID,
	}
	return h.MarkerKeeper.SetDenomMetaData(ctx, markerMetadata, authtypes.NewModuleAddress(types.ModuleName))
}

// GetChainID returns the source chain id from packet for a `07-tendermint` client connection or returns `unknown`
func (h MarkerHooks) GetChainID(ctx sdktypes.Context, ibcPort, ibcChannel string, ibcKeeper *ibckeeper.Keeper) string {
	chainID := "unknown"
	channel, found := ibcKeeper.ChannelKeeper.GetChannel(ctx, ibcPort, ibcChannel)
	if !found {
		return chainID
	}
	connectionEnd, found := ibcKeeper.ConnectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return chainID
	}
	clientState, found := ibcKeeper.ClientKeeper.GetClientState(ctx, connectionEnd.ClientId)
	if !found {
		return chainID
	}
	tmClientState, ok := clientState.(*tendermintclient.ClientState)
	if ok {
		return tmClientState.ChainId
	}
	return chainID
}
