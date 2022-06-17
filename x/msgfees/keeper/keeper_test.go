package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosauthtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	simapp "github.com/provenance-io/provenance/app"
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
	app := simapp.Setup(false)
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
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(30000000))
}

func (s *TestSuite) TestKeeper() {
	app, ctx, _ := s.app, s.ctx, s.addrs

	s.T().Log("verify that creating msg fee for type works")
	msgFee, err := app.MsgFeesKeeper.GetMsgFee(ctx, bankSendAuthMsgType)
	s.Require().Nil(msgFee)
	s.Require().Nil(err)
	newCoin := sdk.NewInt64Coin("steak", 100)
	msgFeeToCreate := types.NewMsgFee(bankSendAuthMsgType, newCoin)
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
	s.Assert().Equal(sdk.NewCoin(types.NhashDenom, sdk.NewInt(175_000_000_000)), nhash)
	usdDollar = sdk.NewCoin(types.UsdDenom, sdk.NewInt(70)) // $7 == 1hash
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin(types.NhashDenom, sdk.NewInt(1_750_000_000)), nhash)
	usdDollar = sdk.NewCoin(types.UsdDenom, sdk.NewInt(1_000)) // $1 == 14.2hash
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin(types.NhashDenom, sdk.NewInt(25_000_000_000)), nhash)

	usdDollar = sdk.NewCoin(types.UsdDenom, sdk.NewInt(10))
	nhash, err = app.MsgFeesKeeper.ConvertDenomToHash(ctx, usdDollar)
	s.Assert().NoError(err)
	s.Assert().Equal(sdk.NewCoin(types.NhashDenom, sdk.NewInt(250_000_000)), nhash)

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
	simapp.FundAccount(app, ctx, acct.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 10)))
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
	simapp.FundAccount(app, ctx, acct.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 11)))
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
	simapp.FundAccount(app, ctx, acct.GetAddress(), sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 21)))
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
	suite.Run(t, new(TestSuite))
}
