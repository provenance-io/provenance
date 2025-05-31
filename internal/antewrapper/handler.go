package antewrapper

import (
	storetypes "cosmossdk.io/store/types"
	circuitante "cosmossdk.io/x/circuit/ante"
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper          cosmosante.AccountKeeper
	BankKeeper             banktypes.Keeper
	ExtensionOptionChecker cosmosante.ExtensionOptionChecker
	FeegrantKeeper         FeegrantKeeper
	FlatFeesKeeper         FlatFeesKeeper
	CircuitKeeper          circuitante.CircuitBreaker
	TxSigningHandlerMap    *txsigning.HandlerMap
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for ante builder")
	}

	if options.FlatFeesKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("flatfees keeper is required for ante builder")
	}
	if options.TxSigningHandlerMap == nil {
		return nil, sdkerrors.ErrLogic.Wrap("tx signing handler map is required for ante builder")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = cosmosante.DefaultSigVerificationGasConsumer
	}

	decorators := []sdk.AnteDecorator{
		NewProvSetUpContextDecorator(options.FlatFeesKeeper), // outermost AnteDecorator. SetUpContext must be called first
		circuitante.NewCircuitBreakerDecorator(options.CircuitKeeper),
		NewFlatFeeSetupDecorator(),
		cosmosante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		cosmosante.NewValidateBasicDecorator(),
		cosmosante.NewTxTimeoutHeightDecorator(),
		cosmosante.NewValidateMemoDecorator(options.AccountKeeper),
		cosmosante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		cosmosante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		cosmosante.NewValidateSigCountDecorator(options.AccountKeeper),
		cosmosante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		cosmosante.NewSigVerificationDecorator(options.AccountKeeper, options.TxSigningHandlerMap),
		cosmosante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(decorators...), nil
}

// GetFeeTx coverts the provided Tx to a FeeTx if possible.
func GetFeeTx(tx sdk.Tx) (sdk.FeeTx, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, sdkerrors.ErrTxDecode.Wrapf("Tx must be a FeeTx: %T", tx)
	}
	return feeTx, nil
}

// IsInitGenesis returns true if the context indicates we're in InitGenesis.
func IsInitGenesis(ctx sdk.Context) bool {
	// Note: This isn't fully accurate since you can initialize a chain at a height other than zero.
	// But it should be good enough for our stuff. Ideally we'd want something specifically set in
	// the context during InitGenesis to check, but that'd probably involve some SDK work.
	return ctx.BlockHeight() <= 0
}

const (
	SimAppChainID = "simapp-unit-testing"
)
