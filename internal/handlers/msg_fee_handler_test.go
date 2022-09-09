package handlers_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	piohandlers "github.com/provenance-io/provenance/internal/handlers"
)

func (s *HandlerTestSuite) TestMsgFeeHandlerSetUp() {
	encodingConfig, err := setUpApp(s, false, sdk.DefaultBondDenom, 100)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().NoError(err)

}

func (s *HandlerTestSuite) TestMsgFeeHandlerSetUpIncorrect() {
	encodingConfig, err := setUpApp(s, false, sdk.DefaultBondDenom, 100)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  nil,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: nil,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     nil,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  nil,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})
	s.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  s.app.AccountKeeper,
		BankKeeper:     s.app.BankKeeper,
		FeegrantKeeper: s.app.FeeGrantKeeper,
		MsgFeesKeeper:  s.app.MsgFeesKeeper,
		Decoder:        nil,
	})
	s.Require().Error(err)

}
