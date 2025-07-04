package keeper_test

import (
	"fmt"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/stretchr/testify/suite"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/hold"
	holdkeeper "github.com/provenance-io/provenance/x/hold/keeper"
)

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

type MsgServerTestSuite struct {
	suite.Suite

	app            *simapp.App
	ctx            sdk.Context
	msgServer      hold.MsgServer
	blockStartTime time.Time

	// Test accounts
	owner1Addr sdk.AccAddress
	owner2Addr sdk.AccAddress
	govAddr    string

	// Account addresses
	vestingAddrs   []sdk.AccAddress
	baseAddr       sdk.AccAddress
	permLockedAddr sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {
	s.blockStartTime = time.Now()
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: s.blockStartTime})
	s.msgServer = holdkeeper.NewMsgServerImpl(s.app.HoldKeeper)

	s.owner1Addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	s.owner2Addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	govModAddr := s.app.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
	s.govAddr = govModAddr.String()
}

func (s *MsgServerTestSuite) TestUnlockVestingAccounts() {
	ba := func(addr string, seq uint64) *authtypes.BaseAccount {
		ba := &authtypes.BaseAccount{Address: addr, Sequence: seq}
		rv := s.app.AccountKeeper.NewAccount(s.ctx, ba)
		return rv.(*authtypes.BaseAccount)
	}
	bva := func(addr string, seq uint64) *vesting.BaseVestingAccount {
		return &vesting.BaseVestingAccount{
			BaseAccount:     ba(addr, seq),
			OriginalVesting: sdk.NewCoins(sdk.NewInt64Coin("banana", 600)),
			EndTime:         s.blockStartTime.Add(time.Minute * 600).Unix(),
		}
	}

	var allAccounts []sdk.AccountI
	setupAccounts := func() {
		for _, acct := range allAccounts {
			s.app.AccountKeeper.SetAccount(s.ctx, acct)
		}
	}

	addrVestCont := sdk.AccAddress("addrVestCont________").String()
	acctVestCont := &vesting.ContinuousVestingAccount{
		BaseVestingAccount: bva(addrVestCont, 12),
		StartTime:          s.blockStartTime.Unix(),
	}
	allAccounts = append(allAccounts, acctVestCont)

	addrVestDel := sdk.AccAddress("addrVestDel_________").String()
	acctVestDel := &vesting.DelayedVestingAccount{BaseVestingAccount: bva(addrVestDel, 4)}
	allAccounts = append(allAccounts, acctVestDel)

	addrVestPer := sdk.AccAddress("addrVestPer_________").String()
	acctVestPer := &vesting.PeriodicVestingAccount{
		BaseVestingAccount: bva(addrVestPer, 44),
		StartTime:          s.blockStartTime.Unix(),
		VestingPeriods: []vesting.Period{
			{Length: 200 * 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("banana", 100))},
			{Length: 400 * 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("banana", 500))},
		},
	}
	allAccounts = append(allAccounts, acctVestPer)

	addrPermLock := sdk.AccAddress("addrPermLock________").String()
	acctPermLock := &vesting.PermanentLockedAccount{BaseVestingAccount: bva(addrPermLock, 173)}
	allAccounts = append(allAccounts, acctPermLock)

	addrBase := sdk.AccAddress("addrBase____________").String()
	acctBase := ba(addrBase, 51)
	allAccounts = append(allAccounts, acctBase)

	addrNotExists := sdk.AccAddress("addrNotExists_______").String()

	authority := s.app.HoldKeeper.GetAuthority()

	tests := []struct {
		name         string
		req          *hold.MsgUnlockVestingAccountsRequest
		expErr       string
		expConverted []string
	}{
		{
			name:   "no authority",
			req:    &hold.MsgUnlockVestingAccountsRequest{Addresses: []string{addrVestCont}},
			expErr: "expected \"" + authority + "\" got \"\": expected gov account as only signer for proposal message",
		},
		{
			name:   "wrong authority",
			req:    &hold.MsgUnlockVestingAccountsRequest{Authority: addrVestCont, Addresses: []string{addrVestCont}},
			expErr: "expected \"" + authority + "\" got \"" + addrVestCont + "\": expected gov account as only signer for proposal message",
		},
		{
			name:         "continuous vesting account",
			req:          &hold.MsgUnlockVestingAccountsRequest{Authority: authority, Addresses: []string{addrVestCont}},
			expConverted: []string{addrVestCont},
		},
		{
			name:         "delayed vesting account",
			req:          &hold.MsgUnlockVestingAccountsRequest{Authority: authority, Addresses: []string{addrVestDel}},
			expConverted: []string{addrVestDel},
		},
		{
			name:         "periodic vesting account",
			req:          &hold.MsgUnlockVestingAccountsRequest{Authority: authority, Addresses: []string{addrVestPer}},
			expConverted: []string{addrVestPer},
		},
		{
			name:         "permanent locked account",
			req:          &hold.MsgUnlockVestingAccountsRequest{Authority: authority, Addresses: []string{addrPermLock}},
			expConverted: []string{addrPermLock},
		},
		{
			name: "base account",
			req:  &hold.MsgUnlockVestingAccountsRequest{Authority: authority, Addresses: []string{addrBase}},
		},
		{
			name: "account does not exist",
			req:  &hold.MsgUnlockVestingAccountsRequest{Authority: authority, Addresses: []string{addrNotExists}},
		},
		{
			name: "multiple vesting accounts",
			req: &hold.MsgUnlockVestingAccountsRequest{
				Authority: authority,
				Addresses: []string{addrVestCont, addrVestDel, addrVestPer, addrPermLock},
			},
			expConverted: []string{addrVestCont, addrVestDel, addrVestPer, addrPermLock},
		},
		{
			name: "partial success",
			req: &hold.MsgUnlockVestingAccountsRequest{
				Authority: authority,
				Addresses: []string{addrVestCont, addrVestDel, addrNotExists, addrVestPer, addrPermLock},
			},
			expConverted: []string{addrVestCont, addrVestDel, addrVestPer, addrPermLock},
		},
	}

	hasStr := func(vals []string, toFind string) bool {
		for _, val := range vals {
			if val == toFind {
				return true
			}
		}
		return false
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			setupAccounts()

			expTypes := make([]string, len(tc.req.Addresses))
			for i, addr := range tc.req.Addresses {
				var expType string
				if hasStr(tc.expConverted, addr) {
					expType = fmt.Sprintf("%T", &authtypes.BaseAccount{})
				} else {
					a, err := sdk.AccAddressFromBech32(addr)
					s.Require().NoError(err, "sdk.AccAddressFromBech32(%q)", addr)
					acct := s.app.AccountKeeper.GetAccount(s.ctx, a)
					expType = fmt.Sprintf("%T", acct)
				}
				expTypes[i] = fmt.Sprintf("%s = %s", addr, expType)
			}

			var expResp *hold.MsgUnlockVestingAccountsResponse
			if len(tc.expErr) == 0 {
				expResp = &hold.MsgUnlockVestingAccountsResponse{}
			}

			var actResp *hold.MsgUnlockVestingAccountsResponse
			var err error
			testFunc := func() {
				actResp, err = s.msgServer.UnlockVestingAccounts(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "UnlockVestingAccounts")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "UnlockVestingAccounts error")
			s.Assert().Equal(expResp, actResp, "UnlockVestingAccounts response")

			actTypes := make([]string, len(tc.req.Addresses))
			for i, addr := range tc.req.Addresses {
				a, err := sdk.AccAddressFromBech32(addr)
				s.Require().NoError(err, "AccAddressFromBech32(%q)", addr)
				acct := s.app.AccountKeeper.GetAccount(s.ctx, a)
				actTypes[i] = fmt.Sprintf("%s = %T", addr, acct)
			}
			s.Assert().Equal(expTypes, actTypes, "account types after UnlockVestingAccounts")
		})
	}
}

func (s *MsgServerTestSuite) TestUnlockVestingAccounts_Performance() {
	ctx := s.ctx
	batchSize := 100
	addresses := make([]string, 0, batchSize)
	accAddrs := make([]sdk.AccAddress, 0, batchSize)

	for i := 0; i < batchSize; i++ {
		addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		baseAccount := authtypes.NewBaseAccount(addr, nil, uint64(1000+i), 0)
		vestingAccount, err := vesting.NewContinuousVestingAccount(
			baseAccount,
			sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
			s.blockStartTime.Unix(),
			s.blockStartTime.Add(time.Hour*24*365).Unix(),
		)
		s.Require().NoError(err, "NewContinuousVestingAccount")
		s.app.AccountKeeper.SetAccount(ctx, vestingAccount)
		addresses = append(addresses, addr.String())
		accAddrs = append(accAddrs, addr)
	}

	msg := hold.MsgUnlockVestingAccountsRequest{
		Authority: s.govAddr,
		Addresses: addresses,
	}

	res, err := s.msgServer.UnlockVestingAccounts(ctx, &msg)
	s.Require().NoError(err, "s.msgServer.UnlockVestingAccounts error")
	s.Require().NotNil(res, "s.msgServer.UnlockVestingAccounts response")

	for _, addr := range accAddrs {
		acc := s.app.AccountKeeper.GetAccount(ctx, addr)
		s.Assert().IsType(acc, &authtypes.BaseAccount{}, "unlocked account")
		//_, isBase := acc.(*authtypes.BaseAccount)
		// s.Assert().True(isBase, "account should be unlocked")
	}
}
