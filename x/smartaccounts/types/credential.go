package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdktypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

type Authenticator interface {
	/// VerifySignature verifies the signature of a credential using the provided sign bytes and signature data.
	VerifySignature(ctx sdk.Context, credential Credential, signBytes []byte, codec codec.Codec, signature []byte) error
}

type CredentialI interface {
	proto.Message
	GetCredentialNumber() uint64
	GetPublicKey() *sdktypes.Any
}
