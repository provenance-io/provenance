package exchange

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// RegisterInterfaces registers implementations for the tx messages
func RegisterInterfaces(registry types.InterfaceRegistry) {
	messages := make([]proto.Message, len(allRequestMsgs))
	for i, msg := range allRequestMsgs {
		messages[i] = msg
	}
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterInterface(
		"provenance.exchange.v1.MarketAccount",
		(*authtypes.AccountI)(nil),
		&MarketAccount{},
	)

	registry.RegisterInterface(
		"provenance.exchange.v1.MarketAccount",
		(*authtypes.GenesisAccount)(nil),
		&MarketAccount{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
