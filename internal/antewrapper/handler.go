package antewrapper

import (
	"time"

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

// HandlerOptions are the options used to construct the AnteHandler.
type HandlerOptions struct {
	// Required:

	AccountKeeper       cosmosante.AccountKeeper
	BankKeeper          banktypes.Keeper
	CircuitKeeper       circuitante.CircuitBreaker
	FeegrantKeeper      FeegrantKeeper
	FlatFeesKeeper      FlatFeesKeeper
	TxSigningHandlerMap *txsigning.HandlerMap

	// Optionals. Leave as nil to use the defaults.

	ExtensionOptionChecker cosmosante.ExtensionOptionChecker
	SigGasConsumer         func(meter storetypes.GasMeter, sig signing.SignatureV2, params types.Params) error
	SigVerifyOptions       []cosmosante.SigVerificationDecoratorOption
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for ante builder")
	}
	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for ante builder")
	}
	if options.CircuitKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("circuit keeper is required for ante builder")
	}
	if options.FeegrantKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("feegrant keeper is required for ante builder")
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

	sigVerOpts := options.SigVerifyOptions
	if sigVerOpts == nil {
		sigVerOpts = []cosmosante.SigVerificationDecoratorOption{
			cosmosante.WithUnorderedTxGasCost(cosmosante.DefaultUnorderedTxGasCost), // Consume 2240 gas when using an unordered tx.
			cosmosante.WithMaxUnorderedTxTimeoutDuration(5 * time.Minute),
		}
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
		NewDeductUpFrontCostDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		cosmosante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		cosmosante.NewValidateSigCountDecorator(options.AccountKeeper),
		cosmosante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		cosmosante.NewSigVerificationDecorator(options.AccountKeeper, options.TxSigningHandlerMap, sigVerOpts...),
		cosmosante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(decorators...), nil
}
