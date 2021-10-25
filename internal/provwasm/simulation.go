package provwasm

import (
	"crypto/sha256"
	"fmt"
	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"io/ioutil"

	//"fmt"
	//"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	//"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
)

type ProvwasmWrapper struct {
	cdc codec.Codec
	wasm module.AppModuleSimulation
}

func NewProvwasmWrapper(cdc codec.Codec, keeper *wasm.Keeper, validatorSetSource keeper.ValidatorSetSource) *ProvwasmWrapper {

	return &ProvwasmWrapper{
		cdc: cdc,
		wasm: wasm.NewAppModule(cdc, keeper, validatorSetSource),
	}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the wasm module.
func (pw ProvwasmWrapper) GenerateGenesisState(input *module.SimulationState) {
	//pw.wasm.GenerateGenesisState(input)
	fmt.Println("asdf - GenerateGenesisState for provwasm")
	//params := wasmsim.RandomParams(input.Rand)
	contracts := make([]types.Contract, 1)
	//var contractInfo types.ContractInfo

	// get this working!
	myCodeInfo := wasmTypes.CodeInfoFixture(wasmTypes.WithSHA256CodeHash(wasmCode))


	contractInfo := types.ContractInfo {
		// CodeID is the reference to the stored Wasm code
		1,
		// Creator address who initially instantiated the contract
		//Creator string `protobuf:"bytes,2,opt,name=creator,proto3" json:"creator,omitempty"
		input.Accounts[0].Address.String(),
		// Admin is an optional address that can execute migrations
		//Admin string `protobuf:"bytes,3,opt,name=admin,proto3" json:"admin,omitempty"`
		input.Accounts[0].Address.String(),
		// Label is optional metadata to be stored with a contract instance.
		//Label string `protobuf:"bytes,4,opt,name=label,proto3" json:"label,omitempty"`
		"simple-contract",
		// Created Tx position when the contract was instantiated.
		// This data should kept internal and not be exposed via query results. Just
		// use for sorting
		//Created   *AbsoluteTxPosition `protobuf:"bytes,5,opt,name=created,proto3" json:"created,omitempty"`
		nil,
		//IBCPortID string              `protobuf:"bytes,6,opt,name=ibc_port_id,json=ibcPortId,proto3" json:"ibc_port_id,omitempty"`
		"IBCPortID",
		// Extension is an extension point to store custom metadata within the
		// persistence model.
		//Extension *types.Any `protobuf:"bytes,7,opt,name=extension,proto3" json:"extension,omitempty"`
		nil,
	}

	//var codeBytes []byte

	codeBytes, err := ioutil.ReadFile("/Users/fredkneeland/code/provenance/tutorial.wasm") // b has type []byte
	if err != nil {
		panic("failed to read file")
	}

	fmt.Println("asdf")
	fmt.Println(codeBytes)

	// TODO: how do I get the code bytes from file?

	hash := sha256.Sum256(codeBytes)
	fmt.Println("Hash: ")
	fmt.Println(hash)

	// okay... how do I get these from the tutorial.wasm file?
	codeInfo := types.CodeInfo{
		CodeHash: hash[:],
		Creator:  input.Accounts[0].Address.String(),
		InstantiateConfig: types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
	}

	codes := make([]types.Code, 1)
	codes[0] = types.Code{
		// Code struct encompasses CodeInfo and CodeBytes
		//CodeID    uint64   `protobuf:"varint,1,opt,name=code_id,json=codeId,proto3" json:"code_id,omitempty"`
		CodeID: 1,
		//CodeInfo  CodeInfo `protobuf:"bytes,2,opt,name=code_info,json=codeInfo,proto3" json:"code_info"`
		CodeInfo: codeInfo,
		//CodeBytes []byte   `protobuf:"bytes,3,opt,name=code_bytes,json=codeBytes,proto3" json:"code_bytes,omitempty"`
		CodeBytes: codeBytes,
		// Pinned to wasmvm cache
		//Pinned bool `protobuf:"varint,4,opt,name=pinned,proto3" json:"pinned,omitempty"`
		//,
	}

	//state := make([]types.Model{}, 1)

	contracts[0] = types.Contract{
		ContractAddress: input.Accounts[0].Address.String(),
		ContractInfo:    contractInfo,
	}
	wasmGenesis := types.GenesisState{
		Params:    types.DefaultParams(),
		Codes:     codes,
		Contracts: contracts, // TODO: add contract specific code here
		Sequences: []types.Sequence{
			{IDKey: types.KeyLastCodeID, Value: 10},
			{IDKey: types.KeyLastInstanceID, Value: 1},
		},
		GenMsgs:   nil,
	}

	_, err = input.Cdc.MarshalJSON(&wasmGenesis)
	if err != nil {
		panic(err)
	}

	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(&wasmGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals.
func (ProvwasmWrapper) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized bank param changes for the simulator.
func (pw ProvwasmWrapper) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {

	return pw.wasm.RandomizedParams(r)

	//params := wasmsim.RandomParams(r)
	//return []simtypes.ParamChange{
	//	simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyUploadAccess),
	//		func(r *rand.Rand) string {
	//			jsonBz, err := am.cdc.MarshalJSON(&params.CodeUploadAccess)
	//			if err != nil {
	//				panic(err)
	//			}
	//			return string(jsonBz)
	//		},
	//	),
	//	simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyInstantiateAccess),
	//		func(r *rand.Rand) string {
	//			return fmt.Sprintf("%q", params.CodeUploadAccess.Permission.String())
	//		},
	//	),
	//	simulation.NewSimParamChange(types.ModuleName, string(types.ParamStoreKeyMaxWasmCodeSize),
	//		func(r *rand.Rand) string {
	//			return fmt.Sprintf(`"%d"`, params.MaxWasmCodeSize)
	//		},
	//	),
	//}
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (pw ProvwasmWrapper) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (pw ProvwasmWrapper) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
