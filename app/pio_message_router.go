package app

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PioMessageRouter pio wasmd MessageRouter
type PioMessageRouter struct {
	HandlerFn func(msg sdk.Msg) baseapp.MsgServiceHandler
}

// Handler is the entry point
func (m PioMessageRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	if m.HandlerFn == nil {
		panic("PioMessageRouter Handler function not expected to be called")
	}
	return m.HandlerFn(msg)
}

// MessageRouterFunc convenient type to match the keeper.MessageRouter interface
type MessageRouterFunc func(msg sdk.Msg) baseapp.MsgServiceHandler

// Handler is the entry point
func (m MessageRouterFunc) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	return m(msg)
}
