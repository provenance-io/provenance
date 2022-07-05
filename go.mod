module github.com/provenance-io/provenance

go 1.17

require (
	github.com/CosmWasm/wasmd v0.22.0
	github.com/armon/go-metrics v0.3.10
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/cosmos/cosmos-sdk v0.45.0
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ibc-go/v2 v2.0.3
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.26.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.15
	github.com/tendermint/tm-db v0.6.7
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0

)

require (
	github.com/coinbase/rosetta-sdk-go v0.7.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
	golang.org/x/sys v0.0.0-20220408201424-a24fb2fb8a0f // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/cosmos/cosmos-sdk => github.com/provenance-io/cosmos-sdk v0.45.1-0.20220630230747-5e7656c5f8c0

replace github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76

// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0

// Introduces a way to build for aarch64 and includes a prebuilt shared library for it
// NOTE Make sure to update wasmd's wasmvm depdency if this ever changes
replace github.com/CosmWasm/wasmvm => github.com/provenance-io/wasmvm v1.0.1-betaprov

// NOTE Uses our wasmvm dependency based off of beta7 tag
replace github.com/CosmWasm/wasmd => github.com/provenance-io/wasmd v0.22.0-m1supportv3
