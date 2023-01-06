package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosauthtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

type TestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	queryClient types.QueryClient
}

var bankSendAuthMsgType = banktypes.SendAuthorization{}.MsgTypeURL()

func (s *TestSuite) SetupTest() {
	app := simapp.Setup(s.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.MsgFeesKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient

	s.app = app
	s.ctx = ctx
	s.queryClient = queryClient
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 4, sdk.NewInt(30000000))
}

func (s *TestSuite) TestKeeper() {
	app, ctx, _ := s.app, s.ctx, s.addrs

	s.T().Log("verify that creating msg fee for type works")
	msgFee, err := app.MsgFeesKeeper.GetMsgFee(ctx, bankSendAuthMsgType)
	s.Require().Nil(msgFee)
	s.Require().Nil(err)
	newCoin := sdk.NewInt64Coin("steak", 100)
	msgFeeToCreate := types.NewMsgFee(bankSendAuthMsgType, newCoin, "", types.DefaultMsgFeeBips)
	app.MsgFeesKeeper.SetMsgFee(ctx, msgFeeToCreate)

	msgFee, err = app.MsgFeesKeeper.GetMsgFee(ctx, bankSendAuthMsgType)
	s.Require().NotNil(msgFee)
	s.Require().Nil(err)
	s.Require().Equal(bankSendAuthMsgType, msgFee.MsgTypeUrl)

	msgFee, err = app.MsgFeesKeeper.GetMsgFee(ctx, "does-not-exist")
	s.Require().Nil(err)
	s.Require().Nil(msgFee)

	err = app.MsgFeesKeeper.RemoveMsgFee(ctx, bankSendAuthMsgType)
	s.Require().Nil(err)
	msgFee, err = app.MsgFeesKeeper.GetMsgFee(ctx, bankSendAuthMsgType)
	s.Require().Nil(msgFee)
	s.Require().Nil(err)

	err = app.MsgFeesKeeper.RemoveMsgFee(ctx, "does-not-exist")
	s.Require().ErrorIs(err, types.ErrMsgFeeDoesNotExist)

}

func (s *TestSuite) TestConvertDenomToHash() {
	app, ctx, _ := s.app, s.ctx, s.addrs
	usdDollar := sdk.NewCoin(types.UsdDenom, sdk.NewInt(7_000)) // $7.00 == 100hash
	nhash, err := app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin(pioconfig.GetProvenanceConfig().FeeDenom, sdk.NewInt(175_000_000_000)), nhash)
	usdDollar = sdk.NewCoin(types.UsdDenom, sdk.NewInt(70)) // $7 == 1hash
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin(pioconfig.GetProvenanceConfig().FeeDenom, sdk.NewInt(1_750_000_000)), nhash)
	usdDollar = sdk.NewCoin(types.UsdDenom, sdk.NewInt(1_000)) // $1 == 14.2hash
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin(pioconfig.GetProvenanceConfig().FeeDenom, sdk.NewInt(25_000_000_000)), nhash)

	usdDollar = sdk.NewCoin(types.UsdDenom, sdk.NewInt(10))
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin(pioconfig.GetProvenanceConfig().FeeDenom, sdk.NewInt(250_000_000)), nhash)

	jackTheCat := sdk.NewCoin("jackThecat", sdk.NewInt(70))
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, jackTheCat)
	s.Assert().Equal("denom not supported for conversion jackThecat: invalid type", err.Error())
	s.Assert().Equal(sdk.Coin{}, nhash)
}

func (s *TestSuite) TestDeductFeesDistributions() {
	app, ctx, addrs := s.app, s.ctx, s.addrs
	var err error
	var remainingCoins, balances sdk.Coins
	stakeCoin := sdk.NewInt64Coin("stake", 30000000)
	feeDist := make(map[string]sdk.Coins)
	feeDist[""] = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	remainingCoins = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	priv, _, _ := testdata.KeyTestPubAddr()
	acct := cosmosauthtypes.NewBaseAccount(addrs[0], priv.PubKey(), 0, 0)

	err = app.MsgFeesKeeper.DeductFeesDistributions(app.BankKeeper, ctx, acct, remainingCoins, feeDist)
	s.Assert().Error(err)
	s.Assert().Equal("0jackthecat is smaller than 10jackthecat: insufficient funds: insufficient funds", err.Error())

	feeDist = make(map[string]sdk.Coins)
	feeDist["not-an-address"] = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	err = app.MsgFeesKeeper.DeductFeesDistributions(app.BankKeeper, ctx, acct, remainingCoins, feeDist)
	s.Assert().Error(err)
	s.Assert().Equal("decoding bech32 failed: invalid separator index -1: invalid address", err.Error())

	feeDist = make(map[string]sdk.Coins)
	feeDist[addrs[0].String()] = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	err = app.MsgFeesKeeper.DeductFeesDistributions(app.BankKeeper, ctx, acct, remainingCoins, feeDist)
	s.Assert().Error(err)
	s.Assert().Equal("0jackthecat is smaller than 10jackthecat: insufficient funds: insufficient funds", err.Error())

	// Account has enough funds to pay account, but not enough to sweep remaining coins
	s.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, acct.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))), "initial fund")
	feeDist = make(map[string]sdk.Coins)
	feeDist[addrs[1].String()] = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	remainingCoins = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 11))
	err = app.MsgFeesKeeper.DeductFeesDistributions(app.BankKeeper, ctx, acct, remainingCoins, feeDist)
	s.Assert().Error(err)
	s.Assert().Equal("0jackthecat is smaller than 1jackthecat: insufficient funds: insufficient funds", err.Error())
	balances = app.BankKeeper.GetAllBalances(ctx, acct.GetAddress())
	s.Assert().Equal(balances.String(), stakeCoin.String())
	balances = app.BankKeeper.GetAllBalances(ctx, addrs[1])
	s.Assert().Equal(balances.String(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10), stakeCoin).String())

	// Account has enough to pay funds to account and to sweep the remaining coins
	s.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, acct.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 11))), "followup fund")
	feeDist = make(map[string]sdk.Coins)
	feeDist[addrs[1].String()] = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	remainingCoins = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 11))
	err = app.MsgFeesKeeper.DeductFeesDistributions(app.BankKeeper, ctx, acct, remainingCoins, feeDist)
	s.Assert().NoError(err)
	balances = app.BankKeeper.GetAllBalances(ctx, acct.GetAddress())
	s.Assert().Equal(balances.String(), stakeCoin.String())
	balances = app.BankKeeper.GetAllBalances(ctx, addrs[1])
	s.Assert().Equal(balances.String(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 20), stakeCoin).String())

	// Account has enough to pay funds to account, module, and to sweep the remaining coins
	s.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, acct.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 21))), "final fund")
	feeDist = make(map[string]sdk.Coins)
	feeDist[""] = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	feeDist[addrs[1].String()] = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10))
	remainingCoins = sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 21))
	err = app.MsgFeesKeeper.DeductFeesDistributions(app.BankKeeper, ctx, acct, remainingCoins, feeDist)
	s.Assert().NoError(err)
	balances = app.BankKeeper.GetAllBalances(ctx, acct.GetAddress())
	s.Assert().Equal(balances.String(), stakeCoin.String())
	balances = app.BankKeeper.GetAllBalances(ctx, addrs[1])
	s.Assert().Equal(balances.String(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 30), stakeCoin).String())

}

func TestTestSuite(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestCalculateAdditionalFeesToBePaid() {
	// nhashCoins = a shorter way to create sdk.Coins with a single entry for nhash in the given amount.
	nhashCoins := func(amount int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, amount))
	}
	// nhashCoin = a shorter way to create sdk.Coin for nhash in the given amount.
	nhashCoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, amount)
	}
	someAddress := s.addrs[3]
	sendTypeURL := sdk.MsgTypeURL(&banktypes.MsgSend{})
	assessFeeTypeURL := sdk.MsgTypeURL(&types.MsgAssessCustomMsgFeeRequest{})
	oneHash := nhashCoin(1_000_000_000)

	msgSend := banktypes.NewMsgSend(someAddress, someAddress, nhashCoins(1_234_567_890))
	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee(sendTypeURL, oneHash, "", 0)), "setting MsgSend fee")

	assertEqualDist := func(t *testing.T, expected, actual types.MsgFeesDistribution) bool {
		t.Helper()
		failed := false
		failed = assert.Equal(t, expected.TotalAdditionalFees, actual.TotalAdditionalFees, "TotalAdditionalFees") || failed
		failed = assert.Equal(t, expected.AdditionalModuleFees, actual.AdditionalModuleFees, "AdditionalModuleFees") || failed
		failed = assert.Equal(t, expected.RecipientDistributions, actual.RecipientDistributions, "RecipientDistributions") || failed
		return failed
	}

	s.Run("single send", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:    nhashCoins(1_000_000_000),
			AdditionalModuleFees:   nhashCoins(1_000_000_000),
			RecipientDistributions: map[string]sdk.Coins{},
		}
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Run("double send", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:    nhashCoins(2_000_000_000),
			AdditionalModuleFees:   nhashCoins(2_000_000_000),
			RecipientDistributions: map[string]sdk.Coins{},
		}
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, msgSend)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Run("send and custom with recipient", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:  nhashCoins(2_000_000_000),
			AdditionalModuleFees: nhashCoins(1_000_000_000),
			RecipientDistributions: map[string]sdk.Coins{
				"recipient1": nhashCoins(1_000_000_000),
			},
		}
		assessFee := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "recipient1", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Run("send and custom without recipient", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:    nhashCoins(2_000_000_000),
			AdditionalModuleFees:   nhashCoins(2_000_000_000),
			RecipientDistributions: map[string]sdk.Coins{},
		}
		assessFee := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Run("send and two customs with different recipients", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:  nhashCoins(2_500_000_000),
			AdditionalModuleFees: nhashCoins(1_000_000_000), // only gets 1 hash from msgfees.
			RecipientDistributions: map[string]sdk.Coins{
				"recipient1": nhashCoins(1_000_000_000), // gets 1 hash
				"recipient2": nhashCoins(500_000_000),   // gets 0.5 hash
			},
		}
		assessFee1 := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "recipient1", someAddress.String(), "")
		assessFee2 := types.NewMsgAssessCustomMsgFeeRequest("", nhashCoin(500_000_000), "recipient2", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee1, &assessFee2)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Run("send and two customs with same recipient", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:  nhashCoins(2_500_000_000),
			AdditionalModuleFees: nhashCoins(1_000_000_000), // 1 hash from msg fees.
			RecipientDistributions: map[string]sdk.Coins{
				"recipient1": nhashCoins(1_500_000_000), // 1.5 hash from MsgAssessCustomMsgFee
			},
		}
		assessFee1 := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "recipient1", someAddress.String(), "")
		assessFee2 := types.NewMsgAssessCustomMsgFeeRequest("", nhashCoin(500_000_000), "recipient1", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee1, &assessFee2)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee(sendTypeURL, oneHash, "sendrecipient", 2_500)), "setting MsgSend fee with recipient")

	s.Run("send with recipient at 2500", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:  nhashCoins(1_000_000_000),
			AdditionalModuleFees: nhashCoins(750_000_000),
			RecipientDistributions: map[string]sdk.Coins{
				"sendrecipient": nhashCoins(250_000_000),
			},
		}
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee(assessFeeTypeURL, oneHash, "sendrecipient", 1_000)), "setting MsgAssessCustomMsgFeeRequest fee")

	s.Run("send and two customs all with fees and same recipient", func() {
		// The Send will have a fee of 750_000_000 to the module and 250_000_000 to sendrecipient.
		// The 1st assess will have a fee of  900_000_000 to the module, and 100_000_000 to sendrecipient.(this still holds)
		// then it will send 0 to the module and 1_000_000_000 to sendrecipient
		// The 2nd assess will have a fee of  900_000_000 to the module, and 100_000_000 to sendrecipient.
		// then it will send 0 to the module and 500_000_000 to sendrecipient
		// module = (900_000_000 +750_000_000+ 900_000_000=2_550_000_000)
		// recipient = (250_000_000 + 100_000_000 + 1_000_000_000+ 100_000_000 + 500_000_000 = 1_950_000_000)
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:  nhashCoins(4_500_000_000),
			AdditionalModuleFees: nhashCoins(2_550_000_000),
			RecipientDistributions: map[string]sdk.Coins{
				"sendrecipient": nhashCoins(1_950_000_000),
			},
		}
		assessFee1 := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "sendrecipient", someAddress.String(), "")
		assessFee2 := types.NewMsgAssessCustomMsgFeeRequest("", nhashCoin(500_000_000), "sendrecipient", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee1, &assessFee2)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee(assessFeeTypeURL, oneHash, "customrecipient", 1_000)), "setting MsgAssessCustomMsgFeeRequest fee")

	s.Run("send and custom all different recipients", func() {
		// The Send will have a fee of 750_000_000 to the module and 250_000_000 to sendrecipient.
		// The 1st assess will have a fee of 900_000_000 to the module, and 100_000_000 to customrecipient.
		// then it will send 0 to the module and 1_000_000_000 to anotherrecipient
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:  nhashCoins(3_000_000_000),
			AdditionalModuleFees: nhashCoins(1_650_000_000),
			RecipientDistributions: map[string]sdk.Coins{
				"sendrecipient":    nhashCoins(250_000_000),
				"customrecipient":  nhashCoins(100_000_000),
				"anotherrecipient": nhashCoins(1_000_000_000),
			},
		}
		assessFee1 := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "anotherrecipient", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee1)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)

	})

	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee(assessFeeTypeURL, oneHash, "", 0)), "setting MsgAssessCustomMsgFeeRequest fee without split")
	s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, types.NewMsgFee(sendTypeURL, oneHash, "", 0)), "setting MsgSend fee back to no recipient")

	s.Run("with fee on custom assess too do send and custom with no recipient", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:    nhashCoins(3_000_000_000),
			AdditionalModuleFees:   nhashCoins(3_000_000_000),
			RecipientDistributions: map[string]sdk.Coins{},
		}
		assessFee := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})

	s.Run("with fee on custom assess too do send and custom with recipient", func() {
		expected := types.MsgFeesDistribution{
			TotalAdditionalFees:  nhashCoins(3_000_000_000),
			AdditionalModuleFees: nhashCoins(2_000_000_000),
			RecipientDistributions: map[string]sdk.Coins{
				"recipient1": nhashCoins(1_000_000_000), // 1 hash goes to recipient1
			},
		}
		assessFee := types.NewMsgAssessCustomMsgFeeRequest("", oneHash, "recipient1", someAddress.String(), "")
		actual, err := s.app.MsgFeesKeeper.CalculateAdditionalFeesToBePaid(s.ctx, msgSend, &assessFee)
		s.Require().NoError(err)
		assertEqualDist(s.T(), expected, actual)
	})
}
