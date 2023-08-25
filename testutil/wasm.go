package testutil

import (
	"encoding/json"
	"os"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/stretchr/testify/suite"
)

func (chain *TestChain) StoreContractCodeDirect(suite *suite.Suite, path string) uint64 {
	provenanceApp := chain.GetProvenanceApp()
	govKeeper := wasmkeeper.NewGovPermissionKeeper(provenanceApp.WasmKeeper)
	creator := provenanceApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	wasmCode, err := os.ReadFile(path)
	suite.Require().NoError(err, "Read of wasm file failed", err)
	accessEveryone := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody}
	codeID, _, err := govKeeper.Create(chain.GetContext(), creator, wasmCode, &accessEveryone)
	suite.Require().NoError(err, "contract direct code load failed", err)
	println("loaded contract '", path, "' with code id: ", codeID)
	return codeID
}

func (chain *TestChain) InstantiateContract(suite *suite.Suite, msg string, codeID uint64) sdk.AccAddress {
	provenanceApp := chain.GetProvenanceApp()
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(provenanceApp.WasmKeeper)
	creator := provenanceApp.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(chain.GetContext(), codeID, creator, creator, []byte(msg), "contract", nil)
	suite.Require().NoError(err, "contract instantiation failed", err)
	println("instantiated contract '", codeID, "' with address: ", addr)
	return addr
}

func (chain *TestChain) QueryContract(suite *suite.Suite, contract sdk.AccAddress, key []byte) string {
	provenanceApp := chain.GetProvenanceApp()
	state, err := provenanceApp.WasmKeeper.QuerySmart(chain.GetContext(), contract, key)
	suite.Require().NoError(err, "contract query failed", err)
	println("got query result of ", string(state))
	return string(state)
}

func (chain *TestChain) QueryContractJson(suite *suite.Suite, contract sdk.AccAddress, key []byte) []byte {
	provenanceApp := chain.GetProvenanceApp()
	state, err := provenanceApp.WasmKeeper.QuerySmart(chain.GetContext(), contract, key)
	suite.Require().NoError(err, "contract query json failed", err)
	suite.Require().True(json.Valid(state))
	println("got query result of ", state)
	return state
}
