package keeper_test

import (
	"fmt"
	"strings"
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
	"github.com/provenance-io/provenance/x/hold"
	holdkeeper "github.com/provenance-io/provenance/x/hold/keeper"
)

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
	s.Require().NotNil(govModAddr, "governance module account must exist in genesis")
	fmt.Printf("gov : %s", govModAddr.String())
	s.govAddr = govModAddr.String()

}

// createTestAccounts generates fresh accounts for each test
func (s *MsgServerTestSuite) createTestAccounts() {
	s.vestingAddrs = []sdk.AccAddress{
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()), // Continuous
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()), // Delayed
		sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()), // Periodic
	}
	s.baseAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	s.permLockedAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	s.setupAccounts()
}

func (s *MsgServerTestSuite) setupAccounts() {

	base1 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.vestingAddrs[0])
	base1Typed := base1.(*authtypes.BaseAccount)

	vesting1, err := vesting.NewContinuousVestingAccount(
		base1Typed,
		sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		s.blockStartTime.Unix(),
		s.blockStartTime.Add(365*24*time.Hour).Unix(),
	)
	s.Require().NoError(err, "vesting1")
	s.app.AccountKeeper.SetAccount(s.ctx, vesting1)

	base2 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.vestingAddrs[1])
	base2Typed := base2.(*authtypes.BaseAccount)

	vesting2, err := vesting.NewDelayedVestingAccount(
		base2Typed,
		sdk.NewCoins(sdk.NewInt64Coin("stake", 2000)),
		s.blockStartTime.Add(30*24*time.Hour).Unix(),
	)
	s.Require().NoError(err, "vesting2")
	s.app.AccountKeeper.SetAccount(s.ctx, vesting2)

	base3 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.vestingAddrs[2])
	base3Typed := base3.(*authtypes.BaseAccount)

	vesting3, err := vesting.NewPeriodicVestingAccount(
		base3Typed,
		sdk.NewCoins(sdk.NewInt64Coin("stake", 300)),
		s.blockStartTime.Unix(),
		[]vesting.Period{
			{Length: 86400, Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 100))},
			{Length: 86400, Amount: sdk.NewCoins(sdk.NewInt64Coin("stake", 200))},
		},
	)
	s.Require().NoError(err, "vesting3")
	s.app.AccountKeeper.SetAccount(s.ctx, vesting3)

	base4 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.baseAddr)
	s.app.AccountKeeper.SetAccount(s.ctx, base4)

	base5 := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.permLockedAddr)
	base5Typed := base5.(*authtypes.BaseAccount)

	permLocked, err := vesting.NewPermanentLockedAccount(
		base5Typed,
		sdk.NewCoins(sdk.NewInt64Coin("stake", 5000)),
	)
	s.Require().NoError(err, "permLocked")
	s.app.AccountKeeper.SetAccount(s.ctx, permLocked)

	for i, addr := range s.vestingAddrs[:3] {
		acc := s.app.AccountKeeper.GetAccount(s.ctx, addr)
		s.T().Logf("Account %d: %s, type: %T", i, addr.String(), acc)
	}
}

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) TestMsgUnlockVestingAccounts_SuccessCases() {
	testCases := []string{
		"single continuous vesting account",
		"single delayed vesting account",
		"single periodic vesting account",
		"multiple vesting accounts",
	}

	for _, name := range testCases {
		s.Run(name, func() {
			// Setup fresh state
			s.createTestAccounts()

			var (
				addresses   []string
				expUnlocked int
			)

			switch name {
			case "single continuous vesting account":
				addresses = []string{s.vestingAddrs[0].String()}
				expUnlocked = 1
			case "single delayed vesting account":
				addresses = []string{s.vestingAddrs[1].String()}
				expUnlocked = 1
			case "single periodic vesting account":
				addresses = []string{s.vestingAddrs[2].String()}
				expUnlocked = 1
			case "multiple vesting accounts":
				addresses = []string{
					s.vestingAddrs[0].String(),
					s.vestingAddrs[1].String(),
					s.vestingAddrs[2].String(),
				}
				expUnlocked = 3
			default:
				s.T().Fatalf("unsupported test case: %s", name)
			}

			ctx := s.ctx
			msg := hold.MsgUnlockVestingAccountsRequest{
				Authority: s.govAddr,
				Addresses: addresses,
			}

			res, err := s.msgServer.UnlockVestingAccounts(ctx, &msg)
			s.Require().NoError(err, "unexpected error during UnlockVestingAccounts")
			s.Require().NotNil(res, "response should not be nil")

			s.Assert().Len(res.UnlockedAddresses, expUnlocked, "unexpected number of unlocked addresses")
			s.Assert().Len(res.FailedAddresses, 0, "expected no failed addresses")

			for _, addrStr := range res.UnlockedAddresses {
				accAddr, err := sdk.AccAddressFromBech32(addrStr)
				if s.Assert().NoError(err, "failed to parse address: %s", addrStr) {
					acc := s.app.AccountKeeper.GetAccount(ctx, accAddr)
					if s.Assert().NotNil(acc, "account should not be nil: %s", addrStr) {
						_, isBase := acc.(*authtypes.BaseAccount)
						s.Assert().True(isBase, "account %s should be a base account", addrStr)
					}
				}
			}
		})
	}
}

func (s *MsgServerTestSuite) TestMsgUnlockVestingAccounts_PartialFailures() {
	s.Run("partial success", func() {
		s.createTestAccounts()
		ctx := s.ctx
		nonExistentAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		msg := hold.MsgUnlockVestingAccountsRequest{
			Authority: s.govAddr,
			Addresses: []string{
				s.vestingAddrs[0].String(), // Valid
				s.baseAddr.String(),        // Base account (invalid)
				nonExistentAddr.String(),   // Non-existent (valid format, no state)
				s.permLockedAddr.String(),  // Permanent locked (invalid)
			},
		}

		res, err := s.msgServer.UnlockVestingAccounts(ctx, &msg)
		s.Require().NoError(err, "UnlockVestingAccounts")
		s.Require().NotNil(res, "UnlockVestingAccounts")
		s.Assert().Len(res.UnlockedAddresses, 1, "only one account should be successfully unlocked")
		s.Assert().Len(res.FailedAddresses, 3, "three accounts should have failed to unlock")
		acc := s.app.AccountKeeper.GetAccount(ctx, s.vestingAddrs[0])
		_, isBase := acc.(*authtypes.BaseAccount)
		s.Assert().True(isBase, "account %s should be unlocked as base account", s.vestingAddrs[0].String())

		for _, failure := range res.FailedAddresses {
			s.Assert().NotEmpty(failure.Reason, "failure reason must not be empty")

			switch failure.Address {
			case s.baseAddr.String():
				s.Assert().Contains(failure.Reason, "not a supported vesting account type", "base account failure reason mismatch")

			case s.permLockedAddr.String():
				s.Assert().Contains(failure.Reason, "not a supported vesting account type", "permanent locked account reason mismatch")

			case nonExistentAddr.String():
				s.T().Logf("Non-existent account failure: %s", failure.Reason)
				s.Assert().True(
					strings.Contains(failure.Reason, "account not found") ||
						strings.Contains(failure.Reason, "not found") ||
						strings.Contains(failure.Reason, "invalid address"),
					"unexpected failure reason for non-existent account: %s", failure.Reason,
				)

			default:
				s.T().Errorf("unexpected address in failure list: %s", failure.Address)
			}
		}
	})
}

func (s *MsgServerTestSuite) TestMsgUnlockVestingAccounts_EdgeCases() {
	s.Run("already unlocked account", func() {
		s.createTestAccounts()
		ctx := s.ctx

		msg1 := hold.MsgUnlockVestingAccountsRequest{
			Authority: s.govAddr,
			Addresses: []string{s.vestingAddrs[0].String()},
		}
		_, err := s.msgServer.UnlockVestingAccounts(ctx, &msg1)
		s.Require().NoError(err, "s.msgServer.UnlockVestingAccounts")

		msg2 := hold.MsgUnlockVestingAccountsRequest{
			Authority: s.govAddr,
			Addresses: []string{s.vestingAddrs[0].String()},
		}
		res, err := s.msgServer.UnlockVestingAccounts(ctx, &msg2)
		s.Require().NoError(err, "s.msgServer.UnlockVestingAccounts")
		s.Require().NotNil(res)
		s.Assert().Len(res.UnlockedAddresses, 0)
		s.Assert().Len(res.FailedAddresses, 1)
		s.Assert().Contains(res.FailedAddresses[0].Reason, "not a supported vesting account type")
	})

	s.Run("preserve account sequence numbers", func() {
		s.createTestAccounts()
		ctx := s.ctx

		addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		baseAccount := authtypes.NewBaseAccount(addr, nil, 100, 42)
		vestingAccount, err := vesting.NewContinuousVestingAccount(
			baseAccount,
			sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
			s.blockStartTime.Unix(),
			s.blockStartTime.Add(time.Hour*24*365).Unix(),
		)
		s.Require().NoError(err)
		s.app.AccountKeeper.SetAccount(ctx, vestingAccount)

		msg := hold.MsgUnlockVestingAccountsRequest{
			Authority: s.govAddr,
			Addresses: []string{addr.String()},
		}

		res, err := s.msgServer.UnlockVestingAccounts(ctx, &msg)
		s.Require().NoError(err, "s.msgServer.UnlockVestingAccounts")
		s.Require().NotNil(res)
		s.Require().Len(res.UnlockedAddresses, 1)
		s.Require().Len(res.FailedAddresses, 0)

		acc := s.app.AccountKeeper.GetAccount(ctx, addr)
		baseAcc, ok := acc.(*authtypes.BaseAccount)
		s.Require().True(ok, "account should be base account")
		s.Assert().Equal(uint64(100), baseAcc.GetAccountNumber())
		s.Assert().Equal(uint64(42), baseAcc.GetSequence())
	})
}

func (s *MsgServerTestSuite) TestMsgUnlockVestingAccounts_Performance() {
	s.Run("batch processing", func() {
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
			s.Require().NoError(err, "vestingAccount")
			s.app.AccountKeeper.SetAccount(ctx, vestingAccount)
			addresses = append(addresses, addr.String())
			accAddrs = append(accAddrs, addr)
		}

		msg := hold.MsgUnlockVestingAccountsRequest{
			Authority: s.govAddr,
			Addresses: addresses,
		}

		res, err := s.msgServer.UnlockVestingAccounts(ctx, &msg)
		s.Require().NoError(err, "s.msgServer.UnlockVestingAccounts")
		s.Require().NotNil(res)
		s.Assert().Len(res.UnlockedAddresses, batchSize)
		s.Assert().Len(res.FailedAddresses, 0)

		for _, addr := range accAddrs {
			acc := s.app.AccountKeeper.GetAccount(ctx, addr)
			_, isBase := acc.(*authtypes.BaseAccount)
			s.Assert().True(isBase, "account should be unlocked")
		}
	})
}
