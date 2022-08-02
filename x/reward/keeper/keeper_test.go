package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/reward"
	"github.com/provenance-io/provenance/x/reward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	PKs = simapp.CreateTestPubKeys(500)
)

type KeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	queryClient types.QueryClient
	handler     sdk.Handler

	accountAddr      sdk.AccAddress
	accountKey       *secp256k1.PrivKey
	keyring          keyring.Keyring
	keyringDir       string
	accountAddresses []sdk.AccAddress
}

func (suite *KeeperTestSuite) CreateAccounts(number int) {
	for i := 0; i < number; i++ {
		accountKey := secp256k1.GenPrivKeyFromSecret([]byte(fmt.Sprintf("acc%d", i+2)))
		addr, err := sdk.AccAddressFromHex(accountKey.PubKey().Address().String())
		suite.Require().NoError(err)
		suite.accountAddr = addr
		suite.accountAddresses = append(suite.accountAddresses, addr)
	}
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = app.Setup(false)
	suite.CreateAccounts(4)
	suite.handler = reward.NewHandler(suite.app.RewardKeeper)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now().UTC()})
	suite.createTestValidators(10)
	simapp.FundModuleAccount(suite.app, suite.ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("nhash", 10000000000000)))

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, suite.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, suite.app.RewardKeeper)
	suite.queryClient = types.NewQueryClient(queryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// Test no reward programs. Nothing should happen
func (suite *KeeperTestSuite) TestDelegateAgainstNoRewardPrograms() {
	suite.SetupTest()
	SetupEventHistoryWithDelegates(suite)

	// Advance one day
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + (24 * 60 * 60))
	suite.Assert().NotPanics(func() {
		reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)
	})

	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += 1
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no shares should be created")
}

// Test against inactive reward programs. They should not be updated
func (suite *KeeperTestSuite) TestDelegateAgainstInactiveRewardPrograms() {
	suite.SetupTest()
	SetupEventHistoryWithDelegates(suite)

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)
	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no shares are granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += 1
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no shares should be created")

	programs, err := suite.app.RewardKeeper.GetAllRewardPrograms(suite.ctx)
	suite.Assert().Equal(1, len(programs))
	suite.Assert().NoError(err, "get all reward programs should not throw error")

	activePrograms, err := suite.app.RewardKeeper.GetAllActiveRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "get all active reward programs should not throw error")
	suite.Assert().Equal(0, len(activePrograms))
}

// Test against delegate reward program. No delegate happens.
func (suite *KeeperTestSuite) TestNonDelegateAgainstRewardProgram() {
	suite.SetupTest()
	setupEventHistory(suite)

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)
	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.State = types.RewardProgram_STATE_STARTED
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no shares are granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += 1
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no shares should be created")

	programs, err := suite.app.RewardKeeper.GetAllRewardPrograms(suite.ctx)
	suite.Assert().Equal(1, len(programs))
	suite.Assert().NoError(err, "get all reward programs should not throw error")

	activePrograms, err := suite.app.RewardKeeper.GetAllActiveRewardPrograms(suite.ctx)
	suite.Assert().NoError(err, "get all active reward programs should not throw error")
	suite.Assert().Equal(1, len(activePrograms))
}

func (suite *KeeperTestSuite) createDelegateEvents(validator, amount, sender, shares string) sdk.Events {
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("module", "staking"),
		sdk.NewAttribute("sender", sender),
	}
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("validator", validator),
		sdk.NewAttribute("amount", amount),
		sdk.NewAttribute("new_shares", shares),
	}
	event1 := sdk.NewEvent("delegate", attributes1...)
	event2 := sdk.NewEvent("message", attributes2...)
	events := sdk.Events{
		event1,
		event2,
	}
	return events
}

func (suite *KeeperTestSuite) createValidatorEvent(validator, amount, sender string) sdk.Events {
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("validator", validator),
		sdk.NewAttribute("amount", amount),
	}
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("module", "staking"),
		sdk.NewAttribute("sender", sender),
	}
	event1 := sdk.NewEvent("create_validator", attributes1...)
	event2 := sdk.NewEvent("message", attributes2...)
	events := sdk.Events{
		event1,
		event2,
	}
	return events
}

// Test against delegate reward program. Grant 1 share
func (suite *KeeperTestSuite) TestSingleDelegate() {
	suite.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure one share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(1, count, "1 share should be created")
}

// Test against delegate reward program. Grant 2 share
func (suite *KeeperTestSuite) TestMultipleDelegate() {
	suite.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure one share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(2, count, "2 shares should be created")
}

// Test against delegate reward program. Not enough actions
func (suite *KeeperTestSuite) TestDelegateBelowMinimumActions() {
	suite.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               10,
						MaximumActions:               20,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(7, 1),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no share should be created when below minimum actions")
}

// Test against delegate reward program. Too many actions
func (suite *KeeperTestSuite) TestDelegateAboveMaximumActions() {
	suite.SetupTest()

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               0,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no share should be created when above maximum actions")
}

// Test against delegate reward program. Below delegation amount
func (suite *KeeperTestSuite) TestDelegateBelowMinimumDelegation() {
	suite.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)

	minimumDelegation := sdk.NewInt64Coin("nhash", 100)
	maximumDelegation := sdk.NewInt64Coin("nhash", 200)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no share should be created when below minimum delegation amount")
}

// Test against delegate reward program. Above delegation amount
func (suite *KeeperTestSuite) TestDelegateAboveMaximumDelegation() {
	suite.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 50)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no share should be created when above maximum delegation amount")
}

// Test against delegate reward program. Below percentile
func (suite *KeeperTestSuite) TestDelegateBelowMinimumPercentile() {
	suite.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(7, 1),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no share should be created when below minimum delegation percentage")
}

// Test against delegate reward program. Above percentile
func (suite *KeeperTestSuite) TestDelegateAboveMaximumPercentile() {
	suite.SetupTest()

	// Create inactive reward program
	action := types.NewActionDelegate()
	action.MaximumActions = 10
	action.MinimumActions = 1
	action.MinimumActiveStakePercentile = sdk.NewDecWithPrec(0, 0)
	action.MaximumActiveStakePercentile = sdk.NewDecWithPrec(1, 0)

	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	action.MinimumDelegationAmount = &minimumDelegation
	action.MaximumDelegationAmount = &maximumDelegation

	coin := sdk.NewInt64Coin("hotdog", 10000)
	maxCoin := sdk.NewInt64Coin("hotdog", 100)

	now := time.Now().UTC()
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		coin,
		maxCoin,
		now,
		60*60,
		3,
		0,
		2,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               10,
						MinimumDelegationAmount:      &minimumDelegation,
						MaximumDelegationAmount:      &maximumDelegation,
						MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdk.NewDecWithPrec(2, 1),
					},
				},
			},
		},
	)
	suite.app.RewardKeeper.StartRewardProgram(suite.ctx, &rewardProgram)
	suite.app.RewardKeeper.SetRewardProgram(suite.ctx, rewardProgram)

	// We want to set the events here
	validators := getTestValidators(6, 6)
	delegates := suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000")
	delegates = delegates.AppendEvents(suite.createDelegateEvents(validators[0].OperatorAddress, "1000000000nhash", "cosmos15ky9du8a2wlstz6fpx3p4mqpjyrm5cgxwpuzvh", "50000000000000.000000000000000000"))
	SetupEventHistory(suite, delegates)
	reward.EndBlocker(suite.ctx, suite.app.RewardKeeper)

	// Ensure no share is granted
	count := 0
	err := suite.app.RewardKeeper.IterateAllRewardAccountStates(suite.ctx, func(state types.RewardAccountState) bool {
		if state.GetSharesEarned() > 0 {
			count += int(state.GetSharesEarned())
			return true
		}
		return false
	})
	suite.Assert().NoError(err, "iterate should not throw error")
	suite.Assert().Equal(0, count, "no share should be created when above maximum delegation percentage")
}
