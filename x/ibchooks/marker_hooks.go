package ibchooks

import (
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
)

type MarkerHooks struct {
	markerKeeper *markerkeeper.Keeper
}

func NewMarkerHooks(markerkeeper *markerkeeper.Keeper) MarkerHooks {
	return MarkerHooks{
		markerKeeper: markerkeeper,
	}
}

func (h MarkerHooks) ProperlyConfigured() bool {
	return h.markerKeeper != nil
}
