package keeper

import "github.com/provenance-io/provenance/x/ledger"

type MsgServer struct {
	Keeper
}

func NewMsgServer(k Keeper) ledger.MsgServer {
	return MsgServer{
		Keeper: k,
	}
}
