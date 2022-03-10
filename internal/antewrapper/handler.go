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

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper   cosmosante.AccountKeeper
	BankKeeper      banktypes.Keeper
	FeegrantKeeper  msgfeestypes.FeegrantKeeper
	MsgFeesKeeper   msgfeestypes.MsgFeesKeeper
	SignModeHandler authsigning.SignModeHandler
	SigGasConsumer  func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = cosmosante.DefaultSigVerificationGasConsumer
	}

	decorators := []sdk.AnteDecorator{
		cosmosante.NewSetUpContextDecorator(),
		// outermost AnteDecorator. SetUpContext must be called first
		NewFeeMeterContextDecorator(), // NOTE : fee gas meter also has the functionality of GasTracerContextDecorator in previous versions
		cosmosante.NewRejectExtensionOptionsDecorator(),
		NewTxGasLimitDecorator(),
		cosmosante.NewMempoolFeeDecorator(),
		// Fee Decorator works to augment NewMempoolFeeDecorator and also check that enough fees are paid
		NewMsgFeesDecorator(options.BankKeeper, options.AccountKeeper, options.FeegrantKeeper, options.MsgFeesKeeper),
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
