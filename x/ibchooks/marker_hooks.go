package ibchooks

import (
	"fmt"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"

	"github.com/provenance-io/provenance/x/ibchooks/keeper"
	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type MarkerHooks struct {
	MarkerKeeper        *markerkeeper.Keeper
	ibcHooksKeeper      *keeper.Keeper
	bech32PrefixAccAddr string
}

func NewMarkerHooks(ibcHooksKeeper *keeper.Keeper, markerKeeper *markerkeeper.Keeper, bech32PrefixAccAddr string) MarkerHooks {
	return MarkerHooks{
		MarkerKeeper:        markerKeeper,
		ibcHooksKeeper:      ibcHooksKeeper,
		bech32PrefixAccAddr: bech32PrefixAccAddr,
	}
}

func (h MarkerHooks) ProperlyConfigured() bool {
	return h.MarkerKeeper != nil && h.ibcHooksKeeper != nil
}

func (h MarkerHooks) OnRecvPacketOverride(im IBCMiddleware, ctx sdktypes.Context, packet channeltypes.Packet, relayer sdktypes.AccAddress) ibcexported.Acknowledgement {
	if !h.ProperlyConfigured() {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	// TODO: this does nothing yet...
	isIcs20, _ := isIcs20Packet(packet)
	if !isIcs20 {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	return im.App.OnRecvPacket(ctx, packet, relayer)

}

func ValidateAndParseMarkerMemo(memo string, receiver string) (isMarkerRouted bool, err error) {
	isMarkerRouted, metadata := jsonStringHasKey(memo, types.MarkerHookKey)
	if !isMarkerRouted {
		return isMarkerRouted, nil
	}

	markerRaw := metadata[types.MarkerHookKey]

	// Make sure the marker key is a map. If it isn't, ignore this packet
	_, ok := markerRaw.(map[string]interface{})
	if !ok {
		return isMarkerRouted,
			fmt.Errorf(types.ErrBadMetadataFormatMsg, memo, "wasm metadata is not a valid JSON map object")
	}

	// TODO: do stuff with marker data

	return isMarkerRouted, nil
}
