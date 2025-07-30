package simulation

import (
	"encoding/base64"
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	simappparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/x/smartaccounts/keeper"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
	"github.com/provenance-io/provenance/x/smartaccounts/utils"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	simState module.SimulationState,
	smartaccountkeeper keeper.Keeper,
	bankkeeper bankkeeper.Keeper,
	protoCodec *codec.ProtoCodec,
) simulation.WeightedOperations {
	var operations simulation.WeightedOperations

	// Set up args
	args := &WeightedOpsArgs{
		SimState:           simState,
		ProtoCodec:         protoCodec,
		Smartaccountkeeper: smartaccountkeeper,
		Bankkeeper:         bankkeeper,
	}

	// Add operation to create a smart account first
	operations = append(operations, simulation.NewWeightedOperation(
		simappparams.DefaultWeightMsgRegisterWebAuthnCredential,
		SimulateMsgRegisterWebAuthnAccount(smartaccountkeeper, args),
	))
	// Add operation to register a cosmos credential
	operations = append(operations, simulation.NewWeightedOperation(
		simappparams.DefaultWeightMsgRegisterCosmosCredential,
		SimulateMsgRegisterCosmosCredential(smartaccountkeeper, args),
	))

	return operations
}

// WeightedOpsArgs holds all the args provided to WeightedOperations so that they can be passed on later more easily.
type WeightedOpsArgs struct {
	SimState   module.SimulationState
	ProtoCodec *codec.ProtoCodec
	// The keeper for the module
	Smartaccountkeeper keeper.Keeper
	Bankkeeper         bankkeeper.Keeper
}

func SimulateMsgRegisterCosmosCredential(k keeper.Keeper, args *WeightedOpsArgs) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get a random account to be the sender
		sender, _ := simtypes.RandomAcc(r, accs)

		// Generate random public key for credential
		credPubKey, _ := codectypes.NewAnyWithValue(GenRandomSecp256k1PubKey())
		// Create the message
		msg := &types.MsgRegisterCosmosCredential{
			Sender: sender.Address.String(),
			Pubkey: credPubKey,
		}

		return Dispatch(r, app, ctx, args.SimState, sender, chainID, msg, args.Smartaccountkeeper.AccountKeeper, args.Bankkeeper, "register cosmos credential")
	}
}

func SimulateMsgRegisterWebAuthnAccount(k keeper.Keeper, args *WeightedOpsArgs) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get a random account to be the sender (base account)
		sender, _ := simtypes.RandomAcc(r, accs)

		// Use predefined test credential that's known to be valid
		// Import the utils package that contains TestCredentialRequestResponses
		msg := &types.MsgRegisterFido2Credential{
			Sender:             sender.Address.String(),
			EncodedAttestation: base64.RawURLEncoding.EncodeToString([]byte(utils.TestCredentialRequestResponses["success"])),
			UserIdentifier:     fmt.Sprintf("user_%s", simtypes.RandStringOfLength(r, 8)),
		}

		return Dispatch(r, app, ctx, args.SimState, sender, chainID, msg, args.Smartaccountkeeper.AccountKeeper, args.Bankkeeper, "create smart account")
	}
}

// Dispatch handles the common logic for simulation operation execution
func Dispatch(
	r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	simState module.SimulationState, sender simtypes.Account, chainID string,
	msg sdk.Msg, ak authkeeper.AccountKeeperI, bk bankkeeper.Keeper, comment string,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	msgType := sdk.MsgTypeURL(msg)

	// Check if the sender has any spendable coins
	spendableCoins := bk.SpendableCoins(ctx, sender.Address)
	if spendableCoins.Empty() {
		return simtypes.NoOpMsg(types.ModuleName, msgType, "sender has no spendable coins"), nil, nil
	}

	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           simState.TxConfig,
		Cdc:             nil,
		Msg:             msg,
		CoinsSpentInMsg: sdk.Coins{}, // No coins spent in this message
		Context:         ctx,
		SimAccount:      sender,
		AccountKeeper:   ak,
		Bankkeeper:      bk,
		ModuleName:      types.ModuleName,
	}

	opMsg, fops, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
	if opMsg.Comment == "" && comment != "" {
		opMsg.Comment = comment
	}

	return opMsg, fops, err
}
