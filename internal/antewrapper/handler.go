package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type BaseAppSimulateFunc func(txBytes []byte) (sdk.GasInfo, *sdk.Result, sdk.Context, error)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper          cosmosante.AccountKeeper
	BankKeeper             banktypes.Keeper
	ExtensionOptionChecker cosmosante.ExtensionOptionChecker
	FeegrantKeeper         msgfeestypes.FeegrantKeeper
	MsgFeesKeeper          msgfeestypes.MsgFeesKeeper
	SignModeHandler        authsigning.SignModeHandler
	SigGasConsumer         func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error
	SimulateFunc           BaseAppSimulateFunc
	TxEncoder              sdk.TxEncoder
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.ErrLogic.Wrap("bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.ErrLogic.Wrap("sign mode handler is required for ante builder")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = cosmosante.DefaultSigVerificationGasConsumer
	}

	decorators := []sdk.AnteDecorator{
		cosmosante.NewSetUpContextDecorator(),                                // outermost AnteDecorator. SetUpContext must be called first
		NewFeeMeterContextDecorator(options.SimulateFunc, options.TxEncoder), // NOTE : fee gas meter also has the functionality of GasTracerContextDecorator in previous versions
		NewTxGasLimitDecorator(),
		NewMinGasPricesDecorator(),
		NewMsgFeesDecorator(options.MsgFeesKeeper),
		cosmosante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		cosmosante.NewValidateBasicDecorator(),
		cosmosante.NewTxTimeoutHeightDecorator(),
		cosmosante.NewValidateMemoDecorator(options.AccountKeeper),
		cosmosante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		NewProvenanceDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.MsgFeesKeeper),
		cosmosante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		cosmosante.NewValidateSigCountDecorator(options.AccountKeeper),
		cosmosante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		cosmosante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		cosmosante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(decorators...), nil
}
