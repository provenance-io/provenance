package handlers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type PioBaseAppKeeperOptions struct {
	AccountKeeper     msgbasedfeetypes.AccountKeeper
	BankKeeper        msgbasedfeetypes.BankKeeper
	FeegrantKeeper    msgbasedfeetypes.FeegrantKeeper
	MsgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper
	Decoder           sdk.TxDecoder
}

func NewAdditionalMsgFeeHandler(options PioBaseAppKeeperOptions) (sdk.AdditionalMsgFeeHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.FeegrantKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "fee grant keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.MsgBasedFeeKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "msgbased fee keeper is required for AdditionalMsgFeeHandler builder")
	}

	if options.Decoder == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "Decoder is required for AdditionalMsgFeeHandler builder")
	}

	return NewMsgBasedFeeInvoker(options.BankKeeper, options.AccountKeeper, options.FeegrantKeeper,
		options.MsgBasedFeeKeeper, options.Decoder).Invoke, nil
}
