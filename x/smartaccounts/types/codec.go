package types

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

func RegisterInterfaces(registry types.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations((*proto.Message)(nil),
		&SmartAccountQueryRequest{},
		&SmartAccountResponse{},
	)
	registry.RegisterImplementations((*proto.Message)(nil),
		&EC2PublicKeyData{},
		&EdDSAPublicKeyData{},
		&PublicKeyData{},
	)
}
