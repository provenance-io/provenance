package antewrapper

import (
	storetypes "cosmossdk.io/store/types"
	txsigning "cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper          cosmosante.AccountKeeper
	BankKeeper             banktypes.Keeper
	ExtensionOptionChecker cosmosante.ExtensionOptionChecker
	FeegrantKeeper         msgfeestypes.FeegrantKeeper
	MsgFeesKeeper          msgfeestypes.MsgFeesKeeper
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

	if options.TxSigningHandlerMap == nil {
		return nil, sdkerrors.ErrLogic.Wrap("tx signing handler map is required for ante builder")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = cosmosante.DefaultSigVerificationGasConsumer
	}

	decorators := []sdk.AnteDecorator{
		cosmosante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewFeeMeterContextDecorator(),         // NOTE : fee gas meter also has the functionality of GasTracerContextDecorator in previous versions
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
		cosmosante.NewSigVerificationDecorator(options.AccountKeeper, options.TxSigningHandlerMap),
		cosmosante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(decorators...), nil
}
