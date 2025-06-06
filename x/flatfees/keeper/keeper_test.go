package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/gogoproto/proto"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/mocks"
	"github.com/provenance-io/provenance/x/flatfees/keeper"
	"github.com/provenance-io/provenance/x/flatfees/types"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

type KeeperTestSuite struct {
	suite.Suite

	app *simapp.App
	ctx sdk.Context
	kpr keeper.Keeper

	bankSendAuthMsgType string
}

func (s *KeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.kpr = s.app.FlatFeesKeeper
	s.ctx = s.app.BaseApp.NewContext(false)
	s.bankSendAuthMsgType = banktypes.SendAuthorization{}.MsgTypeURL()
}

// assertEqualMsgFee asserts that a MsgFee is as expected, returning true iff so.
func assertEqualMsgFee(t testing.TB, expected, actual *types.MsgFee) bool {
	t.Helper()
	expStr := expected.String()
	actStr := actual.String()
	if !assert.Equal(t, expStr, actStr, "MsgFee as strings") {
		return false
	}
	return assert.Equal(t, expected, actual, "MsgFee")
}

// assertEqualMsgFees asserts that two slices of MsgFees are equal, returning true iff equal.
func assertEqualMsgFees(t testing.TB, expected, actual []*types.MsgFee, msgAndArgs ...interface{}) bool {
	t.Helper()
	msg, args := splitMsgAndArgs(t, "MsgFees", msgAndArgs)
	expStrs := sliceToString(expected)
	actStrs := sliceToString(actual)
	if !assert.Equalf(t, expStrs, actStrs, msg+" (as strings)", args...) {
		return false
	}
	return assert.Equalf(t, expected, actual, msg, args...)
}

// assertEqualParams asserts that two params structs are equal, returning true iff equal.
func assertEqualParams(t testing.TB, expected, actual types.Params, msgAndArgs ...interface{}) bool {
	t.Helper()
	msg, args := splitMsgAndArgs(t, "Params", msgAndArgs)
	ok := assert.Equalf(t, expected.DefaultCost.String(), actual.DefaultCost.String(), msg+" DefaultCost", args...)
	ok = assert.Equalf(t, expected.ConversionFactor.String(), actual.ConversionFactor.String(), msg+" ConversionFactor", args...) && ok
	if !ok {
		return false
	}
	return assert.Equalf(t, expected, actual, msg, args...)
}

// sliceToString returns a slice with each of the provided vals converted to a string.
func sliceToString[S ~[]E, E fmt.Stringer](vals S) []string {
	if vals == nil {
		return nil
	}
	rv := make([]string, len(vals))
	for i, val := range vals {
		rv[i] = val.String()
	}
	return rv
}

// reversed returns a copy of the provided slice with the entries reversed.
func reversed[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}
	rv := make(S, len(s))
	for i, val := range s {
		rv[len(s)-i-1] = val
	}
	return rv
}

// splitMsgAndArgs will split the provided msgAndArgs into a msg string and args slice.
// If msgAndArgs is empty, the defaultMsg is used. Fails the test if the first arg isn't a string.
func splitMsgAndArgs(t testing.TB, defaultMsg string, msgAndArgs []interface{}) (string, []interface{}) {
	t.Helper()
	if len(msgAndArgs) == 0 {
		return defaultMsg, nil
	}
	msg, ok := msgAndArgs[0].(string)
	require.True(t, ok, "The first entry in msgAndArgs must be a string, found %T", msgAndArgs[0])
	return msg, msgAndArgs[1:]
}

// msgTypeURLs returns a slice of MsgTypeURL that correspond to the provided Msgs.
func msgTypeURLs(msgs []sdk.Msg) []string {
	if msgs == nil {
		return nil
	}
	rv := make([]string, len(msgs))
	for i, msg := range msgs {
		rv[i] = sdk.MsgTypeURL(msg)
	}
	return rv
}

func (s *KeeperTestSuite) TestKeeper_GetAuthority() {
	exp := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	var act string
	testFunc := func() {
		act = s.kpr.GetAuthority()
	}
	s.Require().NotPanics(testFunc, "GetAuthority()")
	s.Assert().Equal(exp, act, "GetAuthority() result")
}

func (s *KeeperTestSuite) TestKeeper_ValidateAuthority() {
	tests := []struct {
		name      string
		authority string
		expErr    bool
	}{
		{
			name:      "correct authority",
			authority: s.kpr.GetAuthority(),
			expErr:    false,
		},
		{
			name:      "empty string",
			authority: "",
			expErr:    true,
		},
		{
			name:      "extra leading space",
			authority: " cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			expErr:    true,
		},
		{
			name:      "extra trailing space",
			authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn ",
			expErr:    true,
		},
		{
			name:      "other authority",
			authority: sdk.AccAddress("other_authority_____").String(),
			expErr:    true,
		},
		{
			name:      "uppercase",
			authority: "COSMOS10D07Y265GMMUVT4Z0W9AW880JNSR700J6ZN9KN",
			expErr:    true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var expErr string
			if tc.expErr {
				expErr = fmt.Sprintf("expected \"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn\" got %q: %s",
					tc.authority, govtypes.ErrInvalidSigner.Error())
			}
			var err error
			testFunc := func() {
				err = s.kpr.ValidateAuthority(tc.authority)
			}
			s.Require().NotPanics(testFunc, "ValidateAuthority(%q)", tc.authority)
			assertions.AssertErrorValue(s.T(), err, expErr, "ValidateAuthority(%q) error", tc.authority)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_GetFeeCollectorName() {
	exp := authtypes.FeeCollectorName
	var act string
	testFunc := func() {
		act = s.kpr.GetFeeCollectorName()
	}
	s.Require().NotPanics(testFunc, "GetFeeCollectorName()")
	s.Assert().Equal(exp, act, "GetFeeCollectorName() result")
}

func (s *KeeperTestSuite) TestKeeper_SetMsgFee() {
	tests := []struct {
		name   string
		msgFee types.MsgFee
		expErr string
	}{
		{
			name:   "empty url",
			msgFee: *types.NewMsgFee("", sdk.NewInt64Coin("banana", 3)),
		},
		{
			name:   "long url",
			msgFee: *types.NewMsgFee(strings.Repeat("x", types.MaxMsgTypeURLLen+1), sdk.NewInt64Coin("banana", 81)),
		},
		{
			name:   "free",
			msgFee: *types.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal"),
		},
		{
			name:   "one coin",
			msgFee: *types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", sdk.NewInt64Coin("banana", 3)),
		},
		{
			name:   "two coins",
			msgFee: *types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", sdk.NewInt64Coin("apple", 14), sdk.NewInt64Coin("banana", 3)),
		},
		// Not really sure how to make the collection .Set method return an error here.
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var err error
			testSet := func() {
				err = s.kpr.SetMsgFee(s.ctx, tc.msgFee)
			}
			s.Require().NotPanics(testSet, "SetMsgFee(%s)", tc.msgFee)
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "SetMsgFee(%s) error", tc.msgFee)

			if len(tc.expErr) > 0 || err != nil {
				return
			}

			var actMsgFee *types.MsgFee
			testGet := func() {
				actMsgFee, err = s.kpr.GetMsgFee(s.ctx, tc.msgFee.MsgTypeUrl)
			}
			s.Require().NotPanics(testGet, "GetMsgFee(%q) (after setting it)", tc.msgFee.MsgTypeUrl)
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "GetMsgFee(%s) error", tc.msgFee.MsgTypeUrl)
			s.Assert().Equal(&tc.msgFee, actMsgFee, "GetMsgFee(%q) result (after setting it)", tc.msgFee.MsgTypeUrl)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_GetMsgFee() {
	notSetMsgFeeURL := "/cosmos.group.v1.MsgExec"
	freeMsgFee := types.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal")
	oneCoinMsgFee := types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", sdk.NewInt64Coin("banana", 7))
	twoCoinsMsgFee := types.NewMsgFee("/cosmos.group.v1.MsgLeaveGroup", sdk.NewInt64Coin("banana", 7), sdk.NewInt64Coin("cherry", 4))
	msgFees := []*types.MsgFee{freeMsgFee, oneCoinMsgFee, twoCoinsMsgFee}
	for i, msgFee := range msgFees {
		err := s.app.FlatFeesKeeper.SetMsgFee(s.ctx, *msgFee)
		s.Require().NoError(err, "[%d]: SetMsgFee(%s)", i, msgFee)
	}

	tests := []struct {
		name      string
		msgType   string
		expMsgFee *types.MsgFee
		expErr    string
	}{
		{
			name:      "empty url",
			msgType:   "",
			expMsgFee: nil,
			expErr:    "",
		},
		{
			name:      "long url",
			msgType:   strings.Repeat("x", types.MaxMsgTypeURLLen+1),
			expMsgFee: nil,
			expErr:    "",
		},
		{
			name:      "not found",
			msgType:   notSetMsgFeeURL,
			expMsgFee: nil,
			expErr:    "",
		},
		{
			name:      "free",
			msgType:   freeMsgFee.MsgTypeUrl,
			expMsgFee: freeMsgFee,
		},
		{
			name:      "one coin",
			msgType:   oneCoinMsgFee.MsgTypeUrl,
			expMsgFee: oneCoinMsgFee,
		},
		{
			name:      "two coins",
			msgType:   twoCoinsMsgFee.MsgTypeUrl,
			expMsgFee: twoCoinsMsgFee,
		},
		// Not really sure how to make the collection .Get method return an error here.
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actMsgFee *types.MsgFee
			var err error
			testFunc := func() {
				actMsgFee, err = s.kpr.GetMsgFee(s.ctx, tc.msgType)
			}
			s.Require().NotPanics(testFunc, "GetMsgFee(%q)", tc.msgType)
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "GetMsgFee(%q) error", tc.msgType)
			assertEqualMsgFee(s.T(), tc.expMsgFee, actMsgFee)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_RemoveMsgFee() {
	notSetMsgFeeURL := "/cosmos.group.v1.MsgExec"
	freeMsgFee := types.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal")
	oneCoinMsgFee := types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", sdk.NewInt64Coin("banana", 7))
	twoCoinsMsgFee := types.NewMsgFee("/cosmos.group.v1.MsgLeaveGroup", sdk.NewInt64Coin("banana", 7), sdk.NewInt64Coin("cherry", 4))
	msgFees := []*types.MsgFee{freeMsgFee, oneCoinMsgFee, twoCoinsMsgFee}
	for i, msgFee := range msgFees {
		err := s.app.FlatFeesKeeper.SetMsgFee(s.ctx, *msgFee)
		s.Require().NoError(err, "[%d]: SetMsgFee(%s)", i, msgFee)
	}

	tests := []struct {
		name    string
		msgType string
		expErr  string
	}{
		{
			name:    "empty string",
			msgType: "",
			expErr:  "cannot remove msg fee for \"\": fee for type does not exist",
		},
		{
			name:    "long url",
			msgType: strings.Repeat("v", types.MaxMsgTypeURLLen+1),
			expErr:  "cannot remove msg fee for \"" + strings.Repeat("v", types.MaxMsgTypeURLLen+1) + "\": fee for type does not exist",
		},
		{
			name:    "no entry",
			msgType: notSetMsgFeeURL,
			expErr:  "cannot remove msg fee for \"" + notSetMsgFeeURL + "\": fee for type does not exist",
		},
		{
			name:    "free",
			msgType: freeMsgFee.MsgTypeUrl,
		},
		{
			name:    "one coin",
			msgType: oneCoinMsgFee.MsgTypeUrl,
		},
		{
			name:    "two coins",
			msgType: twoCoinsMsgFee.MsgTypeUrl,
		},
		// Not sure how to make either Has or Remove collections methods return an error.
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var err error
			testRemove := func() {
				err = s.kpr.RemoveMsgFee(s.ctx, tc.msgType)
			}
			s.Require().NotPanics(testRemove, "RemoveMsgFee(%q)", tc.msgType)
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "RemoveMsgFee(%q) error", tc.msgType)

			if len(tc.expErr) > 0 || err != nil {
				return
			}

			var actMsgFee *types.MsgFee
			testGet := func() {
				actMsgFee, err = s.kpr.GetMsgFee(s.ctx, tc.msgType)
			}
			s.Require().NotPanics(testGet, "GetMsgFee(%q) (after remove)", tc.msgType)
			s.Assert().NoError(err, "GetMsgFee(%q) error (after remove)", tc.msgType)
			s.Assert().Nil(actMsgFee, "GetMsgFee(%q) result (after remove)", tc.msgType)
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_CalculateMsgCost() {
	defCoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin("banana", amount)
	}
	feeCoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin("cherry", amount)
	}
	otherCoin := func(amount int64) sdk.Coin {
		return sdk.NewInt64Coin("plum", amount)
	}
	cz := func(coins ...sdk.Coin) sdk.Coins {
		return coins
	}

	params := types.Params{
		DefaultCost: defCoin(10),
		ConversionFactor: types.ConversionFactor{
			BaseAmount:      defCoin(2),
			ConvertedAmount: feeCoin(5),
		},
		// converted default cost = 10banana * 5cherry / 2banana = 25cherry.
	}
	s.Require().NoError(s.kpr.SetParams(s.ctx, params), "SetParams(%s)", params)

	msgFees := []*types.MsgFee{
		types.NewMsgFee("/cosmos.gov.v1.MsgSubmitProposal"),
		types.NewMsgFee("/cosmos.gov.v1.MsgVote"),
		types.NewMsgFee("/cosmos.group.v1.MsgCreateGroup", defCoin(3)),
		types.NewMsgFee("/cosmos.group.v1.MsgCreateGroupPolicy", defCoin(4), otherCoin(104)),
		types.NewMsgFee("/cosmos.group.v1.MsgCreateGroupWithPolicy", defCoin(5)),
		types.NewMsgFee("/cosmos.group.v1.MsgExec", defCoin(6)),
		types.NewMsgFee("/cosmos.group.v1.MsgLeaveGroup", defCoin(7)),
		types.NewMsgFee("/cosmos.group.v1.MsgSubmitProposal", defCoin(8)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupAdmin", defCoin(9)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupMembers", defCoin(10)), // = default, up = less, down = more.
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupMetadata", defCoin(11)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyAdmin", defCoin(12)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyDecisionPolicy", defCoin(13)),
		types.NewMsgFee("/cosmos.group.v1.MsgUpdateGroupPolicyMetadata", defCoin(14)),
		types.NewMsgFee("/cosmos.group.v1.MsgVote", defCoin(15)),
		types.NewMsgFee("/cosmos.group.v1.MsgWithdrawProposal", defCoin(16), otherCoin(116)),
	}
	for i, msgFee := range msgFees {
		err := s.app.FlatFeesKeeper.SetMsgFee(s.ctx, *msgFee)
		s.Require().NoError(err, "[%d]: SetMsgFee(%s)", i, msgFee)
	}

	msgDefault1 := &authztypes.MsgGrant{}                        // 25 (default)
	msgDefault2 := &authztypes.MsgExec{}                         // 25 (default)
	msgFree1 := &govv1types.MsgSubmitProposal{}                  // 0 (free)
	msgFree2 := &govv1types.MsgVote{}                            // 0 (free)
	msgLessThanDefaultEven := &group.MsgExec{}                   // 15 (6*5/2)
	msgLessThanDefaultOdd := &group.MsgUpdateGroupAdmin{}        // 23 (9*5/2 = 22r1 => 23)
	msgMoreThanDefaultEven := &group.MsgUpdateGroupPolicyAdmin{} // 30 (12*5/2: 25 now, 5 later)
	msgMoreThanDefaultOdd := &group.MsgVote{}                    // 38 (15*5/2 = 37r1 => 38: 25 now, 13 later)
	msgTwoCoinsLess := &group.MsgCreateGroupPolicy{}             // 10 (4*5/2=10: 10 now, other(104) later)
	msgTwoCoinsMore := &group.MsgWithdrawProposal{}              // 40 (18*5/2=40: 25 now, 15+other(116) later)

	tests := []struct {
		name         string
		msgs         []sdk.Msg
		expUpFront   sdk.Coins
		expOnSuccess sdk.Coins
		expErr       string
	}{
		{
			name:         "no msgs",
			msgs:         nil,
			expUpFront:   nil,
			expOnSuccess: nil,
			expErr:       "",
		},
		{
			name:       "one msg: default",
			msgs:       []sdk.Msg{msgDefault1},
			expUpFront: cz(feeCoin(25)), // = 25 (default)
		},
		{
			name:       "one msg: free",
			msgs:       []sdk.Msg{msgFree1},
			expUpFront: nil, // = 0 (free)
		},
		{
			name:       "one msg: less than default, evenly divisible",
			msgs:       []sdk.Msg{msgLessThanDefaultEven},
			expUpFront: cz(feeCoin(15)), // = 15 (6*5/2)
		},
		{
			name:       "one msg: less than default, not evenly divisible",
			msgs:       []sdk.Msg{msgLessThanDefaultOdd},
			expUpFront: cz(feeCoin(23)), // = 23 (9*5/2 = 22r1 => 23)
		},
		{
			name:         "one msg: more than default, evenly divisible",
			msgs:         []sdk.Msg{msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(25)), // = 30 (12*5/2: 25 now, 5 later)
			expOnSuccess: cz(feeCoin(5)),
		},
		{
			name:         "one msg: more than default, not evenly divisible",
			msgs:         []sdk.Msg{msgMoreThanDefaultOdd},
			expUpFront:   cz(feeCoin(25)), // 38 (15*5/2 = 37r1 => 38: 25 now, 13 later)
			expOnSuccess: cz(feeCoin(13)),
		},
		{
			name:         "one msg: two coins, less than default",
			msgs:         []sdk.Msg{msgTwoCoinsLess},
			expUpFront:   cz(feeCoin(10)), // 10 (4*5/2=10: 10 now, other(104) later)
			expOnSuccess: cz(otherCoin(104)),
		},
		{
			name:         "one msg: two coins, more than default",
			msgs:         []sdk.Msg{msgTwoCoinsMore},
			expUpFront:   cz(feeCoin(25)), // = 16 * 5 / 2 = 40, 25 now, 15+other(116) later
			expOnSuccess: cz(feeCoin(15), otherCoin(116)),
		},
		{
			name:       "two msgs: default, default",
			msgs:       []sdk.Msg{msgDefault1, msgDefault2},
			expUpFront: cz(feeCoin(50)), // = 25 (default) + 25 (default)
		},
		{
			name:       "two msgs: default, free",
			msgs:       []sdk.Msg{msgDefault1, msgFree2},
			expUpFront: cz(feeCoin(25)), // = 25 (default) + 0 (free)
		},
		{
			name:       "two msgs: default, less than default",
			msgs:       []sdk.Msg{msgDefault1, msgLessThanDefaultEven},
			expUpFront: cz(feeCoin(40)), // = 25 (default) + 15 (6*5/2)
		},
		{
			name:         "two msgs: default, more than default",
			msgs:         []sdk.Msg{msgDefault1, msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(50)), // = 25 (default) + 30 (12*5/2: 25 now, 5 later)
			expOnSuccess: cz(feeCoin(5)),
		},
		{
			name:         "two msgs: default, two coins",
			msgs:         []sdk.Msg{msgDefault1, msgTwoCoinsMore},
			expUpFront:   cz(feeCoin(50)), // = 25 (default) + 40 (18*5/2=40: 25 now, 15+other(116) later)
			expOnSuccess: cz(feeCoin(15), otherCoin(116)),
		},
		{
			name:       "two msgs: free, default",
			msgs:       []sdk.Msg{msgFree1, msgDefault2},
			expUpFront: cz(feeCoin(25)), // = 0 (free) + 25 (default)
		},
		{
			name:       "two msgs: free, free",
			msgs:       []sdk.Msg{msgFree1, msgFree2},
			expUpFront: nil, // = 0 (free) + 0 (free)
		},
		{
			name:       "two msgs: free, less than default",
			msgs:       []sdk.Msg{msgFree1, msgLessThanDefaultEven},
			expUpFront: cz(feeCoin(15)), // = 0 (free) + 15 (6*5/2)
		},
		{
			name:         "two msgs: free, more than default",
			msgs:         []sdk.Msg{msgFree1, msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(25)), // = 0 (free) + 30 (12*5/2: 25 now, 5 later)
			expOnSuccess: cz(feeCoin(5)),
		},
		{
			name:         "two msgs: free, two coins",
			msgs:         []sdk.Msg{msgFree1, msgTwoCoinsLess},
			expUpFront:   cz(feeCoin(10)), // = 0 (free) + 10 (4*5/2=10: 10 now, other(104) later)
			expOnSuccess: cz(otherCoin(104)),
		},
		{
			name:       "two msgs: less than default, default",
			msgs:       []sdk.Msg{msgLessThanDefaultEven, msgDefault2},
			expUpFront: cz(feeCoin(40)), // = 15 (6*5/2) + 25 (default)
		},
		{
			name:       "two msgs: less than default, free",
			msgs:       []sdk.Msg{msgLessThanDefaultEven, msgFree1},
			expUpFront: cz(feeCoin(15)), // = 15 (6*5/2) + 0 (free)
		},
		{
			name:       "two msgs: less than default, less than default: even, even",
			msgs:       []sdk.Msg{msgLessThanDefaultEven, msgLessThanDefaultEven},
			expUpFront: cz(feeCoin(30)), // = 15 (6*5/2) + 15 (6*5/2)
		},
		{
			name:       "two msgs: less than default, less than default: even, odd",
			msgs:       []sdk.Msg{msgLessThanDefaultEven, msgLessThanDefaultOdd},
			expUpFront: cz(feeCoin(38)), // = 15 (6*5/2) + 23 (9*5/2 = 22r1 => 23)
		},
		{
			name:       "two msgs: less than default, less than default: odd, even",
			msgs:       []sdk.Msg{msgLessThanDefaultOdd, msgLessThanDefaultEven},
			expUpFront: cz(feeCoin(38)), // = 23 (9*5/2 = 22r1 => 23) + 15 (6*5/2)
		},
		{
			name:       "two msgs: less than default, less than default: odd, odd",
			msgs:       []sdk.Msg{msgLessThanDefaultOdd, msgLessThanDefaultOdd},
			expUpFront: cz(feeCoin(46)), // = 23 (9*5/2 = 22r1 => 23) + 23 (9*5/2 = 22r1 => 23)
		},
		{
			name:         "two msgs: less than default, more than default",
			msgs:         []sdk.Msg{msgLessThanDefaultEven, msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(40)), // = 15 (6*5/2) + 30 (12*5/2: 25 now, 5 later)
			expOnSuccess: cz(feeCoin(5)),
		},
		{
			name:         "two msgs: less than default, two coins",
			msgs:         []sdk.Msg{msgLessThanDefaultEven, msgTwoCoinsMore},
			expUpFront:   cz(feeCoin(40)), // = 15 (6*5/2) + 40 (18*5/2=40: 25 now, 15+other(116) later)
			expOnSuccess: cz(feeCoin(15), otherCoin(116)),
		},
		{
			name:         "two msgs: more than default, default",
			msgs:         []sdk.Msg{msgMoreThanDefaultEven, msgDefault2},
			expUpFront:   cz(feeCoin(50)), // = 30 (12*5/2: 25 now, 5 later) + 25 (default)
			expOnSuccess: cz(feeCoin(5)),
		},
		{
			name:         "two msgs: more than default, free",
			msgs:         []sdk.Msg{msgMoreThanDefaultEven, msgFree2},
			expUpFront:   cz(feeCoin(25)), // = 30 (12*5/2: 25 now, 5 later) + 0 (free)
			expOnSuccess: cz(feeCoin(5)),
		},
		{
			name:         "two msgs: more than default, less than default",
			msgs:         []sdk.Msg{msgMoreThanDefaultEven, msgLessThanDefaultEven},
			expUpFront:   cz(feeCoin(40)), // = 30 (12*5/2: 25 now, 5 later) + 15 (6*5/2)
			expOnSuccess: cz(feeCoin(5)),
		},
		{
			name:         "two msgs: more than default, more than default: even, even",
			msgs:         []sdk.Msg{msgMoreThanDefaultEven, msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(50)), // = 30 (12*5/2: 25 now, 5 later) + 30 (12*5/2: 25 now, 5 later)
			expOnSuccess: cz(feeCoin(10)),
		},
		{
			name:         "two msgs: more than default, more than default: even, odd",
			msgs:         []sdk.Msg{msgMoreThanDefaultEven, msgMoreThanDefaultOdd},
			expUpFront:   cz(feeCoin(50)), // = 30 (12*5/2: 25 now, 5 later) + 38 (15*5/2 = 37r1 => 38: 25 now, 13 later)
			expOnSuccess: cz(feeCoin(18)),
		},
		{
			name:         "two msgs: more than default, more than default: odd, even",
			msgs:         []sdk.Msg{msgMoreThanDefaultOdd, msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(50)), // = 38 (15*5/2 = 37r1 => 38: 25 now, 13 later) + 30 (12*5/2: 25 now, 5 later)
			expOnSuccess: cz(feeCoin(18)),
		},
		{
			name:         "two msgs: more than default, more than default: odd, even",
			msgs:         []sdk.Msg{msgMoreThanDefaultOdd, msgMoreThanDefaultOdd},
			expUpFront:   cz(feeCoin(50)), // = 38 (15*5/2 = 37r1 => 38: 25 now, 13 later) + 38 (15*5/2 = 37r1 => 38: 25 now, 13 later)
			expOnSuccess: cz(feeCoin(26)),
		},
		{
			name:         "two msgs: more than default, two coins",
			msgs:         []sdk.Msg{msgMoreThanDefaultEven, msgTwoCoinsLess},
			expUpFront:   cz(feeCoin(35)), // = 30 (12*5/2: 25 now, 5 later) + 10 (4*5/2=10: 10 now, other(104) later)
			expOnSuccess: cz(feeCoin(5), otherCoin(104)),
		},
		{
			name:         "two msgs: two coins, default",
			msgs:         []sdk.Msg{msgTwoCoinsLess, msgDefault2},
			expUpFront:   cz(feeCoin(35)), // = 10 (4*5/2=10: 10 now, other(104) later) + 25 (default)
			expOnSuccess: cz(otherCoin(104)),
		},
		{
			name:         "two msgs: two coins, free",
			msgs:         []sdk.Msg{msgTwoCoinsMore, msgFree2},
			expUpFront:   cz(feeCoin(25)), // = 40 (18*5/2=40: 25 now, 15+other(116) later) + 0 (free)
			expOnSuccess: cz(feeCoin(15), otherCoin(116)),
		},
		{
			name:         "two msgs: two coins, less than default",
			msgs:         []sdk.Msg{msgTwoCoinsMore, msgLessThanDefaultEven},
			expUpFront:   cz(feeCoin(40)), // = 40 (18*5/2=40: 25 now, 15+other(116) later) + 15 (6*5/2)
			expOnSuccess: cz(feeCoin(15), otherCoin(116)),
		},
		{
			name:         "two msgs: two coins, more than default",
			msgs:         []sdk.Msg{msgTwoCoinsMore, msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(50)), // = 40 (18*5/2=40: 25 now, 15+other(116) later) + 30 (12*5/2: 25 now, 5 later)
			expOnSuccess: cz(feeCoin(20), otherCoin(116)),
		},
		{
			name:         "two msgs: two coins, two coins",
			msgs:         []sdk.Msg{msgTwoCoinsMore, msgTwoCoinsLess},
			expUpFront:   cz(feeCoin(35)), // = 40 (18*5/2=40: 25 now, 15+other(116) later) + 10 (4*5/2=10: 10 now, other(104) later)
			expOnSuccess: cz(feeCoin(15), otherCoin(220)),
		},
		{
			name: "six msgs: default, free, less, more, two coins less, two coins more",
			msgs: []sdk.Msg{
				msgDefault1,            // 25 (default)
				msgFree1,               // 0 (free)
				msgLessThanDefaultEven, // 15 (6*5/2)
				msgMoreThanDefaultEven, // 30 (12*5/2=30: 25 now, 5 later)
				msgTwoCoinsLess,        // 10 (4*5/2=10: 10 now, other(104) later)
				msgTwoCoinsMore,        // 40 (18*5/2=40: 25 now, 15+other(116) later)
			},
			expUpFront:   cz(feeCoin(100)),                // = 25 + 0 + 15 + 25 + 10 + 25
			expOnSuccess: cz(feeCoin(20), otherCoin(220)), // 0 + 0 + 0 + 5 + other(104) + 15 + other(116)
		},
		{
			name:       "five msgs: all default",
			msgs:       []sdk.Msg{msgDefault1, msgDefault1, msgDefault1, msgDefault1, msgDefault1},
			expUpFront: cz(feeCoin(25 * 5)), // = 25 (default) (times five)
		},
		{
			name:       "five msgs: all free",
			msgs:       []sdk.Msg{msgFree1, msgFree1, msgFree1, msgFree1, msgFree1},
			expUpFront: nil, // = 0 (free) (times five)
		},
		{
			name: "five msgs: all less than default and even",
			msgs: []sdk.Msg{msgLessThanDefaultEven, msgLessThanDefaultEven,
				msgLessThanDefaultEven, msgLessThanDefaultEven, msgLessThanDefaultEven},
			expUpFront: cz(feeCoin(15 * 5)), // = 15 (6*5/2) (times five)
		},
		{
			name: "five msgs: all less than default and odd",
			msgs: []sdk.Msg{msgLessThanDefaultOdd, msgLessThanDefaultOdd,
				msgLessThanDefaultOdd, msgLessThanDefaultOdd, msgLessThanDefaultOdd},
			expUpFront: cz(feeCoin(23 * 5)), // = 23 (9*5/2 = 22r1 => 23) (times five)
		},
		{
			name: "five msgs: all more than default and even",
			msgs: []sdk.Msg{msgMoreThanDefaultEven, msgMoreThanDefaultEven,
				msgMoreThanDefaultEven, msgMoreThanDefaultEven, msgMoreThanDefaultEven},
			expUpFront:   cz(feeCoin(25 * 5)), // = 30 (12*5/2: 25 now, 5 later) (times five)
			expOnSuccess: cz(feeCoin(5 * 5)),
		},
		{
			name: "five msgs: all more than default and odd",
			msgs: []sdk.Msg{msgMoreThanDefaultOdd, msgMoreThanDefaultOdd,
				msgMoreThanDefaultOdd, msgMoreThanDefaultOdd, msgMoreThanDefaultOdd},
			expUpFront:   cz(feeCoin(25 * 5)), // = 38 (15*5/2 = 37r1 => 38: 25 now, 13 later) (times five)
			expOnSuccess: cz(feeCoin(13 * 5)),
		},
		{
			name:         "five msgs: all two coins and less than default",
			msgs:         []sdk.Msg{msgTwoCoinsLess, msgTwoCoinsLess, msgTwoCoinsLess, msgTwoCoinsLess, msgTwoCoinsLess},
			expUpFront:   cz(feeCoin(10 * 5)), // 10 (4*5/2=10: 10 now, other(104) later) (times five)
			expOnSuccess: cz(otherCoin(104 * 5)),
		},
		{
			name:         "five msgs: all two coins and more than default",
			msgs:         []sdk.Msg{msgTwoCoinsMore, msgTwoCoinsMore, msgTwoCoinsMore, msgTwoCoinsMore, msgTwoCoinsMore},
			expUpFront:   cz(feeCoin(25 * 5)), // 40 (18*5/2=40: 25 now, 15+other(116) later) (times five)
			expOnSuccess: cz(feeCoin(15*5), otherCoin(116*5)),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actUpFront, actOnSuccess sdk.Coins
			var err error
			testFunc := func() {
				actUpFront, actOnSuccess, err = s.kpr.CalculateMsgCost(s.ctx, tc.msgs...)
			}
			s.Require().NotPanics(testFunc, "CalculateMsgCost(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "CalculateMsgCost(...) error")
			s.Assert().Equal(tc.expUpFront.String(), actUpFront.String(), "up-front cost")
			s.Assert().Equal(tc.expOnSuccess.String(), actOnSuccess.String(), "on success cost")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_ExpandMsgs() {
	// asAnyWithCache wraps the provided msg in an Any such that it's properly cached for later retrieval.
	asAnyWithCache := func(msg proto.Message) *codectypes.Any {
		rv, err := codectypes.NewAnyWithValue(msg)
		s.Require().NoError(err, "NewAnyWithValue(%T)", msg)
		return rv
	}
	// asAnyNoCache wraps the provided msg in an Any such that it doesn't have a cached value.
	asAnyNoCache := func(msg proto.Message) *codectypes.Any {
		rv := &codectypes.Any{TypeUrl: sdk.MsgTypeURL(msg)}
		if msg2, ok := msg.(protov2.Message); ok {
			opts := protov2.MarshalOptions{Deterministic: true}
			var err error
			rv.Value, err = opts.Marshal(msg2)
			s.Require().NoError(err, "protov2 Marshal(%T)", msg)
		} else {
			var err error
			rv.Value, err = proto.Marshal(msg)
			s.Require().NoError(err, "proto.Marshal(%T)", msg)
		}
		return rv
	}
	// anys is a shorthand way to create a slice of Any entries.
	anys := func(vals ...*codectypes.Any) []*codectypes.Any {
		return vals
	}
	// nestedProp creates a MsgSubmitProposal with nested MsgSubmitProposal to the provided depth.
	// The bottom-most one will have the provided endMsgs.
	// A depth of 1 means a MsgSubmitProposal with the provided endMsgs.
	// A depth of 2 means a MsgSubmitProposal with a MsgSubmitProposal with the provided endMsgs.
	// The Anys are packed properly (with the cached value).
	var nestedProp func(depth int, endMsgs ...*codectypes.Any) *govv1.MsgSubmitProposal
	nestedProp = func(depth int, endMsgs ...*codectypes.Any) *govv1.MsgSubmitProposal {
		msg := &govv1.MsgSubmitProposal{}
		if depth > 1 {
			msg.Messages = []*codectypes.Any{asAnyWithCache(nestedProp(depth-1, endMsgs...))}
		} else {
			msg.Messages = endMsgs
		}
		return msg
	}

	tests := []struct {
		name       string
		unpackErrs []string
		msgs       []sdk.Msg
		expMsgs    []sdk.Msg
		expErr     string
	}{
		{
			name:    "nil",
			msgs:    nil,
			expMsgs: []sdk.Msg{},
		},
		{
			name:    "empty",
			msgs:    []sdk.Msg{},
			expMsgs: []sdk.Msg{},
		},
		{
			name:    "one msg: nothing to expand",
			msgs:    []sdk.Msg{&govv1.MsgVote{}},
			expMsgs: []sdk.Msg{&govv1.MsgVote{}},
		},
		{
			name: "one msg: authz.MsgExec",
			msgs: []sdk.Msg{&authztypes.MsgExec{
				Msgs: anys(asAnyWithCache(&govv1.MsgVote{}), asAnyWithCache(&govv1.MsgVote{})),
			}},
			expMsgs: []sdk.Msg{&authztypes.MsgExec{}, &govv1.MsgVote{}, &govv1.MsgVote{}},
		},
		{
			name: "one msg: govv1.MsgSubmitProposal",
			msgs: []sdk.Msg{&govv1.MsgSubmitProposal{
				Messages: anys(asAnyWithCache(&govv1.MsgVote{}), asAnyWithCache(&govv1.MsgVote{})),
			}},
			expMsgs: []sdk.Msg{&govv1.MsgSubmitProposal{}, &govv1.MsgVote{}, &govv1.MsgVote{}},
		},
		{
			name: "one msg: triggertypes.MsgCreateTriggerRequest",
			msgs: []sdk.Msg{&triggertypes.MsgCreateTriggerRequest{
				Actions: anys(asAnyWithCache(&govv1.MsgVote{}), asAnyWithCache(&govv1.MsgDeposit{})),
			}},
			expMsgs: []sdk.Msg{&triggertypes.MsgCreateTriggerRequest{}, &govv1.MsgVote{}, &govv1.MsgDeposit{}},
		},
		{
			name: "one msg: just under max depth",
			msgs: []sdk.Msg{nestedProp(codectypes.MaxUnpackAnyRecursionDepth, asAnyWithCache(&govv1.MsgVote{}))},
			expMsgs: []sdk.Msg{
				&govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{}, &govv1.MsgSubmitProposal{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgVote{},
			},
		},
		{
			name:   "one msg: just over max depth",
			msgs:   []sdk.Msg{nestedProp(codectypes.MaxUnpackAnyRecursionDepth+1, asAnyWithCache(&govv1.MsgVote{}))},
			expErr: "could not expand sub-messages: max depth exceeded",
		},
		{
			name: "one msg: expands: no cached value",
			msgs: []sdk.Msg{&govv1.MsgSubmitProposal{
				Messages: anys(asAnyNoCache(&govv1.MsgCancelProposal{}), asAnyWithCache(&govv1.MsgVote{})),
			}},
			expMsgs: []sdk.Msg{&govv1.MsgSubmitProposal{}, &govv1.MsgCancelProposal{}, &govv1.MsgVote{}},
		},
		// Missing test case: "one msg: expands: cached value not Msg".
		// As of writing this, the sdk.Msg type is an alise for proto.Message.
		// Since an any can only wrap a proto.Message, there's no way to put a non Msg in it.
		{
			name:       "one msg: expands: error unpacking",
			unpackErrs: []string{"not a real error"},
			msgs:       []sdk.Msg{&govv1.MsgSubmitProposal{Messages: anys(asAnyNoCache(&govv1.MsgVote{}))}},
			expErr: "could not extract sub-messages from *v1.MsgSubmitProposal: " +
				"could not unpack *types.Any with a \"/cosmos.gov.v1.MsgVote\": not a real error",
		},
		{
			name:       "error from two deep",
			unpackErrs: []string{"not a real error"},
			msgs: []sdk.Msg{&govv1.MsgSubmitProposal{
				Messages: anys(asAnyWithCache(&govv1.MsgSubmitProposal{Messages: anys(asAnyNoCache(&govv1.MsgVote{}))})),
			}},
			expErr: "could not extract sub-messages from *v1.MsgSubmitProposal: " +
				"could not unpack *types.Any with a \"/cosmos.gov.v1.MsgVote\": not a real error",
		},
		{
			name:    "two msgs: neither expand",
			msgs:    []sdk.Msg{&govv1.MsgVote{}, &govv1.MsgDeposit{}},
			expMsgs: []sdk.Msg{&govv1.MsgVote{}, &govv1.MsgDeposit{}},
		},
		{
			name: "two msgs: normal, expand",
			msgs: []sdk.Msg{
				&govv1.MsgVote{},
				&triggertypes.MsgCreateTriggerRequest{
					Actions: anys(asAnyWithCache(&govv1.MsgUpdateParams{}), asAnyWithCache(&govv1.MsgDeposit{})),
				},
			},
			expMsgs: []sdk.Msg{
				&govv1.MsgVote{},
				&triggertypes.MsgCreateTriggerRequest{}, &govv1.MsgUpdateParams{}, &govv1.MsgDeposit{},
			},
		},
		{
			name: "two msgs: expand, normal",
			msgs: []sdk.Msg{
				&authztypes.MsgExec{Msgs: anys(asAnyWithCache(&govv1.MsgVote{}), asAnyWithCache(&govv1.MsgVote{}))},
				&govv1.MsgDeposit{},
			},
			expMsgs: []sdk.Msg{
				&authztypes.MsgExec{},
				&govv1.MsgVote{},
				&govv1.MsgVote{},
				&govv1.MsgDeposit{},
			},
		},
		{
			name: "three msgs: all expand",
			msgs: []sdk.Msg{
				&authztypes.MsgExec{Msgs: anys(asAnyWithCache(&govv1.MsgVote{}), asAnyWithCache(&govv1.MsgCancelProposal{}))},
				&govv1.MsgSubmitProposal{Messages: anys(asAnyWithCache(&govv1.MsgVote{}))},
				&triggertypes.MsgCreateTriggerRequest{
					Actions: anys(asAnyWithCache(&govv1.MsgUpdateParams{}), asAnyWithCache(&govv1.MsgDeposit{})),
				},
			},
			expMsgs: []sdk.Msg{
				&authztypes.MsgExec{}, &govv1.MsgVote{}, &govv1.MsgCancelProposal{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgVote{},
				&triggertypes.MsgCreateTriggerRequest{}, &govv1.MsgUpdateParams{}, &govv1.MsgDeposit{},
			},
		},
		{
			name: "five msgs: one expands",
			msgs: []sdk.Msg{
				&govv1.MsgVote{}, &govv1.MsgDeposit{}, &govv1.MsgUpdateParams{},
				&govv1.MsgSubmitProposal{Messages: anys(asAnyWithCache(&govv1.MsgCancelProposal{}))},
				&govv1.MsgExecLegacyContent{},
			},
			expMsgs: []sdk.Msg{
				&govv1.MsgVote{}, &govv1.MsgDeposit{}, &govv1.MsgUpdateParams{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgCancelProposal{},
				&govv1.MsgExecLegacyContent{},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			kpr := s.kpr
			if len(tc.unpackErrs) > 0 {
				cdc := mocks.NewMockCodec(kpr.GetCodec()).WithUnpackAnyErrs(tc.unpackErrs...)
				kpr = kpr.WithCodec(cdc)
			}

			var actMsgs []sdk.Msg
			var err error
			testFunc := func() {
				actMsgs, err = kpr.ExpandMsgs(tc.msgs)
			}
			s.Require().NotPanics(testFunc, "ExpandMsgs(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "ExpandMsgs(...) error")
			// All we really care about are the msg type urls, so let's just compare those.
			expURLs := msgTypeURLs(tc.expMsgs)
			actURLs := msgTypeURLs(actMsgs)
			s.Assert().Equal(expURLs, actURLs, "ExpandMsgs(...) result")
		})
	}
}

func (s *KeeperTestSuite) TestKeeper_GetParams() {
	exp := types.DefaultParams()
	var act types.Params
	testFunc := func() {
		act = s.kpr.GetParams(s.ctx)
	}
	s.Require().NotPanics(testFunc, "GetParams()")
	assertEqualParams(s.T(), exp, act)
	// Not really sure how to make the collections Get method return an error.
}

func (s *KeeperTestSuite) TestKeeper_SetParams() {
	defaultParams := types.DefaultParams()
	tests := []struct {
		name   string
		params types.Params
		expErr string
	}{
		{
			name: "default free",
			params: types.Params{
				DefaultCost:      sdk.NewInt64Coin(defaultParams.DefaultCost.Denom, 0),
				ConversionFactor: defaultParams.ConversionFactor,
			},
		},
		{
			name: "no conversion",
			params: types.Params{
				DefaultCost: defaultParams.DefaultCost,
				ConversionFactor: types.ConversionFactor{
					BaseAmount:      sdk.NewInt64Coin(defaultParams.DefaultCost.Denom, 1),
					ConvertedAmount: sdk.NewInt64Coin(defaultParams.DefaultCost.Denom, 1),
				},
			},
		},
		{
			name:   "defaults",
			params: defaultParams,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var err error
			testSet := func() {
				err = s.kpr.SetParams(s.ctx, tc.params)
			}
			s.Require().NotPanics(testSet, "SetParams(%s)", tc.params)
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "SetParams(%s) error", tc.params)

			if len(tc.expErr) > 0 || err != nil {
				return
			}

			var act types.Params
			testGet := func() {
				act = s.kpr.GetParams(s.ctx)
			}
			s.Require().NotPanics(testGet, "GetParams()")
			assertEqualParams(s.T(), tc.params, act)
		})
	}
}
