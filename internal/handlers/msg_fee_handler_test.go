package handlers_test

import (
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
)

func (suite *HandlerTestSuite) TestMsgFeeHandlerSetUp() {
	encodingConfig, err := setUpApp(suite, false, "atom", 100)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:     suite.app.AccountKeeper,
		BankKeeper:        suite.app.BankKeeper,
		FeegrantKeeper:    suite.app.FeeGrantKeeper,
		MsgBasedFeeKeeper: suite.app.MsgBasedFeeKeeper,
		Decoder:           encodingConfig.TxConfig.TxDecoder(),
	})
	suite.Require().NoError(err)

}

func (suite *HandlerTestSuite) TestMsgFeeHandlerSetUpIncorrect() {
	encodingConfig, err := setUpApp(suite, false, "atom", 100)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:     suite.app.AccountKeeper,
		BankKeeper:        suite.app.BankKeeper,
		FeegrantKeeper:    suite.app.FeeGrantKeeper,
		MsgBasedFeeKeeper: nil,
		Decoder:           encodingConfig.TxConfig.TxDecoder(),
	})
	suite.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:     suite.app.AccountKeeper,
		BankKeeper:        suite.app.BankKeeper,
		FeegrantKeeper:    nil,
		MsgBasedFeeKeeper: suite.app.MsgBasedFeeKeeper,
		Decoder:           encodingConfig.TxConfig.TxDecoder(),
	})
	suite.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:     suite.app.AccountKeeper,
		BankKeeper:        nil,
		FeegrantKeeper:    suite.app.FeeGrantKeeper,
		MsgBasedFeeKeeper: suite.app.MsgBasedFeeKeeper,
		Decoder:           encodingConfig.TxConfig.TxDecoder(),
	})
	suite.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:     nil,
		BankKeeper:        suite.app.BankKeeper,
		FeegrantKeeper:    suite.app.FeeGrantKeeper,
		MsgBasedFeeKeeper: suite.app.MsgBasedFeeKeeper,
		Decoder:           encodingConfig.TxConfig.TxDecoder(),
	})
	suite.Require().Error(err)

	_, err = piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:     suite.app.AccountKeeper,
		BankKeeper:        suite.app.BankKeeper,
		FeegrantKeeper:    suite.app.FeeGrantKeeper,
		MsgBasedFeeKeeper: suite.app.MsgBasedFeeKeeper,
		Decoder:           nil,
	})
	suite.Require().Error(err)

}
