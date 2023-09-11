package ibchooks

import (
	"github.com/provenance-io/provenance/x/ibchooks/keeper"
)

type IbcHooks struct {
	ibcHooksKeeper *keeper.Keeper
	wasmHooks      *WasmHooks
	markerHooks    *MarkerHooks
}

func NewIbcHooks(ibcHooksKeeper *keeper.Keeper, wasmHooks *WasmHooks, markerHooks *MarkerHooks) IbcHooks {
	return IbcHooks{
		ibcHooksKeeper: ibcHooksKeeper,
		wasmHooks:      wasmHooks,
		markerHooks:    markerHooks,
	}
}

func (h IbcHooks) ProperlyConfigured() bool {
	return h.wasmHooks.ProperlyConfigured()
}
