package contracts

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	provenanceapp "github.com/provenance-io/provenance/app"

	_ "embed"
)

//go:embed counter/artifacts/counter.wasm
var counterWasm []byte

//go:embed echo/artifacts/echo.wasm
var echoWasm []byte

// EchoWasm returns the echo contract wasm byte data
func EchoWasm() []byte {
	return echoWasm
}

// CounterWasm returns the counter contract wasm byte data
func CounterWasm() []byte {
	return counterWasm
}

func StoreContractCode(app *provenanceapp.App, ctx sdk.Context, wasmCode []byte) (uint64, error) {
	govKeeper := wasmkeeper.NewGovPermissionKeeper(app.WasmKeeper)
	creator := app.AccountKeeper.GetModuleAddress(govtypes.ModuleName)

	accessEveryone := wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody}
	codeID, _, err := govKeeper.Create(ctx, creator, wasmCode, &accessEveryone)
	return codeID, err
}

func InstantiateContract(app *provenanceapp.App, ctx sdk.Context, msg string, codeID uint64) (sdk.AccAddress, error) {
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	creator := app.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, creator, creator, []byte(msg), "contract", nil)
	return addr, err
}

func QueryContract(app *provenanceapp.App, ctx sdk.Context, contract sdk.AccAddress, key []byte) (string, error) {
	state, err := app.WasmKeeper.QuerySmart(ctx, contract, key)
	return string(state), err
}
