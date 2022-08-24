package handlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type PioBaseAppKeeperOptions struct {
	AccountKeeper  msgfeestypes.AccountKeeper
	BankKeeper     bankkeeper.Keeper
	FeegrantKeeper msgfeestypes.FeegrantKeeper
	MsgFeesKeeper  msgfeestypes.MsgFeesKeeper
	Decoder        sdk.TxDecoder
}

func NewAdditionalMsgFeeHandler(options PioBaseAppKeeperOptions) (sdk.FeeHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.FeegrantKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "fee grant keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.MsgFeesKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "msgbased fee keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.Decoder == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "Decoder is required for AdditionalMsgFeeHandler builder")
	}

	return NewMsgFeeInvoker(options.BankKeeper, options.AccountKeeper, options.FeegrantKeeper,
		options.MsgFeesKeeper, options.Decoder).Invoke, nil
}
