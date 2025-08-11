package provenanceaccount

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"

	smartaccounttypes "github.com/provenance-io/provenance/x/smartaccounts/types"
)

var (
	PubKeyTypePrefix = collections.NewPrefix(1)
)

type Option func(a *ProvenanceSmartAccountHandler)

type ProvenanceSmartAccountHandler struct {
	PubKeyType collections.Item[string]

	AddrCodec address.Codec

	SupportedPubKeys map[string]pubKeyImpl

	signingHandlers *signing.HandlerMap
}

type AccountCreatorDependencies struct {
	AddressCodec address.Codec
	StoreService store.KVStoreService
	Cdc          codec.Codec
}

func NewProvenanceAccountHandler(accountCreationDependencies AccountCreatorDependencies, handlerMap *signing.HandlerMap) (*ProvenanceSmartAccountHandler, error) {
	// Initialize the SchemaBuilder
	schemaBuilder := collections.NewSchemaBuilder(accountCreationDependencies.StoreService)
	acc := &ProvenanceSmartAccountHandler{
		PubKeyType:       collections.NewItem(schemaBuilder, PubKeyTypePrefix, "pub_key_type", collections.StringValue),
		AddrCodec:        accountCreationDependencies.AddressCodec,
		SupportedPubKeys: map[string]pubKeyImpl{},
		signingHandlers:  handlerMap,
	}

	// Apply the WithSecp256K1PubKey option directly
	WithSecp256K1PubKey()(acc)
	// Apply the WithWebAuthnPubKey option for other key types(for now they are EC2 and EdDSA keys)
	WithWebAuthnPubKey[smartaccounttypes.EC2PublicKeyData, *smartaccounttypes.EC2PublicKeyData]()(acc)
	WithWebAuthnPubKey[smartaccounttypes.EdDSAPublicKeyData, *smartaccounttypes.EdDSAPublicKeyData]()(acc)

	return acc, nil
}
