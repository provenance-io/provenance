package antewrapper

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal"
	"github.com/provenance-io/provenance/testutil/assertions"
)

// AssertEqualFlatFeeGasMeters will assert that the provided FlatFeeGasMeters are equal, returning false if not equal.
func AssertEqualFlatFeeGasMeters(t *testing.T, expected, actual *FlatFeeGasMeter) bool {
	t.Helper()
	if !assert.Equal(t, expected == nil, actual == nil, "FlatFeeGasMeter: is nil?") {
		return false
	}
	if expected == nil {
		return true
	}

	rv := true

	// We don't care about checking anything in the underlying gas meter.
	// And since it's an interface type, it's not even worth checking if they're both not nil.

	rv = assert.Equal(t, expected.upFrontCost.String(), actual.upFrontCost.String(), "FlatFeeGasMeter.upFrontCost") && rv
	rv = assert.Equal(t, expected.onSuccessCost.String(), actual.onSuccessCost.String(), "FlatFeeGasMeter.onSuccessCost") && rv
	rv = assert.Equal(t, expected.extraMsgsCost.String(), actual.extraMsgsCost.String(), "FlatFeeGasMeter.extraMsgsCost") && rv
	rv = assert.Equal(t, expected.addedFees.String(), actual.addedFees.String(), "FlatFeeGasMeter.addedFees") && rv

	rv = assert.Equal(t, expected.knownMsgs, actual.knownMsgs, "FlatFeeGasMeter.knownMsgs") && rv
	rv = assertEqualMsgTypes(t, expected.extraMsgs, actual.extraMsgs, "FlatFeeGasMeter.extraMsgs") && rv

	// We don't care about checking the FlatFeesKeeper or logger since they don't change over the life of the gas meter.

	rv = assert.Equal(t, expected.msgTypeURLs, actual.msgTypeURLs, "FlatFeeGasMeter.msgTypeURLs") && rv
	rv = assert.Equal(t, expected.used, actual.used, "FlatFeeGasMeter.used") && rv
	rv = assert.Equal(t, expected.counts, actual.counts, "FlatFeeGasMeter.counts") && rv

	return rv
}

// AssertEqualGas will assert that the provided gases are equal, returning false if not equal.
func AssertEqualGas(t *testing.T, expected, actual storetypes.Gas, msgAndArgs ...interface{}) bool {
	t.Helper()
	// If assert.Equal is given unsigned ints and fails, the message shows the values in hex.
	// That's not how we refer to the values, though, so we do the comparison as decimal strings.
	// A lot of times we just cast them to int in the assertion, but there's a chance that'd overflow here.
	return assert.Equal(t, strconv.FormatUint(expected, 10), strconv.FormatUint(actual, 10), msgAndArgs...)
}

func TestNewFlatFeeGasMeter(t *testing.T) {
	baseGM := storetypes.NewGasMeter(100)
	logger := log.NewNopLogger()
	ffk := NewMockFlatFeesKeeper()

	var gasMeter *FlatFeeGasMeter
	testFunc := func() {
		gasMeter = NewFlatFeeGasMeter(baseGM, logger, ffk)
	}
	require.NotPanics(t, testFunc, "NewFlatFeeGasMeter")
	require.NotNil(t, gasMeter, "NewFlatFeeGasMeter return value")

	assertNotNilInterface(t, gasMeter.GasMeter, "NewFlatFeeGasMeter.GasMeter")
	assertNotNilInterface(t, gasMeter.logger, "NewFlatFeeGasMeter.logger")
	assertNotNilInterface(t, gasMeter.ffk, "NewFlatFeeGasMeter.ffk")

	if assert.NotNil(t, gasMeter.knownMsgs, "NewFlatFeeGasMeter.knownMsgs") {
		assert.Empty(t, gasMeter.knownMsgs, "NewFlatFeeGasMeter.knownMsgs")
	}
	if assert.NotNil(t, gasMeter.used, "NewFlatFeeGasMeter.used") {
		assert.Empty(t, gasMeter.used, "NewFlatFeeGasMeter.used")
	}
	if assert.NotNil(t, gasMeter.counts, "NewFlatFeeGasMeter.counts") {
		assert.Empty(t, gasMeter.counts, "NewFlatFeeGasMeter.counts")
	}
}

func TestFlatFeeGasMeter_SetCosts(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name        string
		gm          *FlatFeeGasMeter
		ffk         *MockFlatFeesKeeper // Will be set in gm automatically.
		msgs        []sdk.Msg
		expErr      string
		expGM       *FlatFeeGasMeter
		expCalcArgs []sdk.Msg
	}{
		{
			name: "error expanding messages",
			gm: &FlatFeeGasMeter{
				// All of these should be reset
				upFrontCost:   cz("1upfrontcost"),
				onSuccessCost: cz("1onsuccesscost"),
				extraMsgsCost: cz("1extramsgscost"),
				addedFees:     cz("1addedfees"),
				knownMsgs:     map[string]int{"dummymsg": 3},
				extraMsgs:     []sdk.Msg{&govv1.MsgSubmitProposal{}},
				msgTypeURLs:   []string{"dummymsg", "dummymsg", "dummymsg", "/cosmos.gov.v1.MsgSubmitProposal"},
				// These shouldn't change.
				used:   map[string]storetypes.Gas{"Read": 5555, "Write": 1239},
				counts: map[string]uint64{"Read": 7, "Write": 13},
			},
			ffk:    NewMockFlatFeesKeeper().WithExpandMsgs(nil, "powazaki"),
			msgs:   []sdk.Msg{&banktypes.MsgSend{}},
			expErr: "powazaki",
			expGM: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"Read": 5555, "Write": 1239},
				counts: map[string]uint64{"Read": 7, "Write": 13},
			},
		},
		{
			name: "error calculating msg cost.",
			gm: &FlatFeeGasMeter{
				// All of these should be reset
				upFrontCost:   cz("1upfrontcost"),
				onSuccessCost: cz("1onsuccesscost"),
				extraMsgsCost: cz("1extramsgscost"),
				addedFees:     cz("1addedfees"),
				knownMsgs:     map[string]int{"dummymsg": 3},
				extraMsgs:     []sdk.Msg{&govv1.MsgSubmitProposal{}},
				msgTypeURLs:   []string{"dummymsg", "dummymsg", "dummymsg", "/cosmos.gov.v1.MsgSubmitProposal"},
				// These shouldn't change.
				used:   map[string]storetypes.Gas{"Read": 5555, "Write": 1239},
				counts: map[string]uint64{"Read": 7, "Write": 13},
			},
			ffk:    NewMockFlatFeesKeeper().WithCalculateMsgCost(nil, nil, "oofda"),
			msgs:   []sdk.Msg{&banktypes.MsgSend{}},
			expErr: "oofda",
			expGM: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"Read": 5555, "Write": 1239},
				counts: map[string]uint64{"Read": 7, "Write": 13},
			},
			expCalcArgs: []sdk.Msg{&banktypes.MsgSend{}},
		},
		{
			name: "one msg on empty gas meter",
			gm:   &FlatFeeGasMeter{},
			// This also tests that the results from ExpandMsgs are what's given to CalculateMsgCost.
			ffk: NewMockFlatFeesKeeper().
				WithExpandMsgs([]sdk.Msg{&govv1.MsgVote{}}, "").
				WithCalculateMsgCost(cz("1upfront"), cz("2onsuccess"), ""),
			msgs: []sdk.Msg{&banktypes.MsgSend{}}, // Replaced by mocked ExpandMsgs.
			expGM: &FlatFeeGasMeter{
				upFrontCost:   cz("1upfront"),
				onSuccessCost: cz("2onsuccess"),
				msgTypeURLs:   []string{"/cosmos.gov.v1.MsgVote"},
				knownMsgs:     map[string]int{"/cosmos.gov.v1.MsgVote": 1},
			},
			expCalcArgs: []sdk.Msg{&govv1.MsgVote{}},
		},
		{
			name: "one msg on previously used gas meter",
			gm: &FlatFeeGasMeter{
				// All of these should be reset
				upFrontCost:   cz("1upfrontcost"),
				onSuccessCost: cz("1onsuccesscost"),
				extraMsgsCost: cz("1extramsgscost"),
				addedFees:     cz("1addedfees"),
				knownMsgs:     map[string]int{"dummymsg": 3},
				extraMsgs:     []sdk.Msg{&govv1.MsgSubmitProposal{}},
				msgTypeURLs:   []string{"dummymsg", "dummymsg", "dummymsg", "/cosmos.gov.v1.MsgSubmitProposal"},
				// These shouldn't change.
				used:   map[string]storetypes.Gas{"Read": 5555, "Write": 1239},
				counts: map[string]uint64{"Read": 7, "Write": 13},
			},
			ffk: NewMockFlatFeesKeeper().
				WithExpandMsgs([]sdk.Msg{&govv1.MsgVote{}}, "").
				WithCalculateMsgCost(cz("1upfront"), cz("2onsuccess"), ""),
			msgs: []sdk.Msg{&banktypes.MsgSend{}}, // Replaced by mocked ExpandMsgs.
			expGM: &FlatFeeGasMeter{
				upFrontCost:   cz("1upfront"),
				onSuccessCost: cz("2onsuccess"),
				msgTypeURLs:   []string{"/cosmos.gov.v1.MsgVote"},
				knownMsgs:     map[string]int{"/cosmos.gov.v1.MsgVote": 1},
				used:          map[string]storetypes.Gas{"Read": 5555, "Write": 1239},
				counts:        map[string]uint64{"Read": 7, "Write": 13},
			},
			expCalcArgs: []sdk.Msg{&govv1.MsgVote{}},
		},
		{
			name: "five msgs",
			gm:   &FlatFeeGasMeter{},
			ffk:  NewMockFlatFeesKeeper().WithCalculateMsgCost(cz("14nhash"), cz("83nhash"), ""),
			msgs: []sdk.Msg{
				&banktypes.MsgSend{}, &banktypes.MsgSend{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgVote{},
				&banktypes.MsgSend{},
			},
			expGM: &FlatFeeGasMeter{
				upFrontCost:   cz("14nhash"),
				onSuccessCost: cz("83nhash"),
				msgTypeURLs: []string{
					"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgSubmitProposal", "/cosmos.gov.v1.MsgVote",
					"/cosmos.bank.v1beta1.MsgSend",
				},
				knownMsgs: map[string]int{
					"/cosmos.bank.v1beta1.MsgSend":     3,
					"/cosmos.gov.v1.MsgSubmitProposal": 1,
					"/cosmos.gov.v1.MsgVote":           1,
				},
			},
			expCalcArgs: []sdk.Msg{
				&banktypes.MsgSend{}, &banktypes.MsgSend{},
				&govv1.MsgSubmitProposal{}, &govv1.MsgVote{},
				&banktypes.MsgSend{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.gm.ffk = tc.ffk
			if tc.expGM.knownMsgs == nil {
				tc.expGM.knownMsgs = make(map[string]int)
			}

			ctx := sdk.Context{}
			var err error
			testFunc := func() {
				err = tc.gm.SetCosts(ctx, tc.msgs)
			}
			require.NotPanics(t, testFunc, "SetCosts")
			assertions.AssertErrorValue(t, err, tc.expErr, "SetCosts error")
			AssertEqualFlatFeeGasMeters(t, tc.expGM, tc.gm)
			tc.ffk.AssertExpandMsgsCall(t, tc.msgs)
			tc.ffk.AssertCalculateMsgCostCall(t, tc.expCalcArgs)
		})
	}
}

func TestFlatFeeGasMeter_Finalize(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name        string
		gm          *FlatFeeGasMeter
		ffk         *MockFlatFeesKeeper // Will be set in gm automatically.
		expGM       *FlatFeeGasMeter
		expErr      string
		expCalcArgs bool
	}{
		{
			name: "nil extra msgs",
			gm: &FlatFeeGasMeter{
				extraMsgs:     nil,
				upFrontCost:   cz("16upfrontcost"),
				onSuccessCost: cz("63onsuccesscost"),
				extraMsgsCost: cz("47extramsgscost"),
				msgTypeURLs:   []string{"dummyMsg"},
				knownMsgs:     map[string]int{"dummyMsg": 0},
			},
			ffk: NewMockFlatFeesKeeper().WithCalculateMsgCost(nil, nil, "should-not-be-seen"),
			expGM: &FlatFeeGasMeter{
				extraMsgs:     nil,
				upFrontCost:   cz("16upfrontcost"),
				onSuccessCost: cz("63onsuccesscost"),
				extraMsgsCost: nil,
				msgTypeURLs:   []string{"dummyMsg"},
				knownMsgs:     map[string]int{"dummyMsg": 0},
			},
			expCalcArgs: false, // call shouldn't be made.
		},
		{
			name: "empty extra msgs",
			gm: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{},
				upFrontCost:   cz("12upfrontcost"),
				onSuccessCost: cz("49onsuccesscost"),
				extraMsgsCost: cz("23extramsgscost"),
				msgTypeURLs:   []string{"dummyMsg"},
				knownMsgs:     map[string]int{"dummyMsg": 0},
			},
			ffk: NewMockFlatFeesKeeper().WithCalculateMsgCost(nil, nil, "should-not-be-seen"),
			expGM: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{},
				upFrontCost:   cz("12upfrontcost"),
				onSuccessCost: cz("49onsuccesscost"),
				extraMsgsCost: nil,
				msgTypeURLs:   []string{"dummyMsg"},
				knownMsgs:     map[string]int{"dummyMsg": 0},
			},
			expCalcArgs: false, // call shouldn't be made.
		},
		{
			name: "one extra msg",
			gm: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				upFrontCost:   cz("59upfrontcost"),
				onSuccessCost: cz("81onsuccesscost"),
				msgTypeURLs:   []string{"dummyMsg", "/cosmos.bank.v1beta1.MsgSend"},
				knownMsgs:     map[string]int{"dummyMsg": 0},
			},
			ffk: NewMockFlatFeesKeeper().WithCalculateMsgCost(cz("6extraupfrontcost"), cz("7extraonsuccesscost"), ""),
			expGM: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				upFrontCost:   cz("59upfrontcost"),
				onSuccessCost: cz("81onsuccesscost"),
				msgTypeURLs:   []string{"dummyMsg", "/cosmos.bank.v1beta1.MsgSend"},
				knownMsgs:     map[string]int{"dummyMsg": 0},
				extraMsgsCost: cz("7extraonsuccesscost,6extraupfrontcost"),
			},
			expCalcArgs: true,
		},
		{
			name: "one free msg",
			gm: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				extraMsgsCost: cz("84extramsgscost"),
			},
			ffk: NewMockFlatFeesKeeper().WithCalculateMsgCost(nil, nil, ""),
			expGM: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				extraMsgsCost: nil,
			},
			expCalcArgs: true,
		},
		{
			name: "one msg with only up-front cost",
			gm: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				extraMsgsCost: cz("84extramsgscost"),
			},
			ffk: NewMockFlatFeesKeeper().WithCalculateMsgCost(cz("12nhash"), nil, ""),
			expGM: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				extraMsgsCost: cz("12nhash"),
			},
			expCalcArgs: true,
		},
		{
			name: "one msg with only on success cost",
			gm: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				extraMsgsCost: cz("84extramsgscost"),
			},
			ffk: NewMockFlatFeesKeeper().WithCalculateMsgCost(nil, cz("63nhash"), ""),
			expGM: &FlatFeeGasMeter{
				extraMsgs:     []sdk.Msg{&banktypes.MsgSend{}},
				extraMsgsCost: cz("63nhash"),
			},
			expCalcArgs: true,
		},
		{
			name: "five extra msgs",
			gm: &FlatFeeGasMeter{
				extraMsgs: []sdk.Msg{
					&banktypes.MsgSend{}, &banktypes.MsgSend{},
					&govv1.MsgSubmitProposal{}, &govv1.MsgVote{},
					&banktypes.MsgSend{},
				},
				upFrontCost:   cz("42upfrontcost"),
				onSuccessCost: cz("58onsuccesscost"),
				extraMsgsCost: cz("77extramsgscost"), // should get stomped on.
				addedFees:     cz("15addedfees"),
				msgTypeURLs: []string{
					"dummyMsg", "dummyMsg", "otherMsg",
					"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgSubmitProposal", "/cosmos.gov.v1.MsgVote",
					"/cosmos.bank.v1beta1.MsgSend",
				},
				knownMsgs: map[string]int{"dummyMsg": 0, "otherMsg": 0},
			},
			ffk: NewMockFlatFeesKeeper().WithCalculateMsgCost(cz("3apple,1banana"), cz("11apple,5cherry"), ""),
			expGM: &FlatFeeGasMeter{
				extraMsgs: []sdk.Msg{
					&banktypes.MsgSend{}, &banktypes.MsgSend{},
					&govv1.MsgSubmitProposal{}, &govv1.MsgVote{},
					&banktypes.MsgSend{},
				},
				upFrontCost:   cz("42upfrontcost"),
				onSuccessCost: cz("58onsuccesscost"),
				extraMsgsCost: cz("14apple,1banana,5cherry"),
				addedFees:     cz("15addedfees"),
				msgTypeURLs: []string{
					"dummyMsg", "dummyMsg", "otherMsg",
					"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgSubmitProposal", "/cosmos.gov.v1.MsgVote",
					"/cosmos.bank.v1beta1.MsgSend",
				},
				knownMsgs: map[string]int{"dummyMsg": 0, "otherMsg": 0},
			},
			expCalcArgs: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expCalcArgs []sdk.Msg
			if tc.expCalcArgs {
				expCalcArgs = tc.gm.extraMsgs
			}
			tc.gm.ffk = tc.ffk

			ctx := sdk.Context{}
			var err error
			testFunc := func() {
				err = tc.gm.Finalize(ctx)
			}
			require.NotPanics(t, testFunc, "Finalize")
			assertions.AssertErrorValue(t, err, tc.expErr, "Finalize error")
			AssertEqualFlatFeeGasMeters(t, tc.expGM, tc.gm)
			tc.ffk.AssertCalculateMsgCostCall(t, expCalcArgs)
		})
	}
}

func TestFlatFeeGasMeter_GetUpFrontCost(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  sdk.Coins
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  nil,
		},
		{
			name: "nil value",
			gm:   &FlatFeeGasMeter{upFrontCost: nil},
			exp:  nil,
		},
		{
			name: "empty",
			gm:   &FlatFeeGasMeter{upFrontCost: sdk.Coins{}},
			exp:  sdk.Coins{},
		},
		{
			name: "one coin",
			gm:   &FlatFeeGasMeter{upFrontCost: cz("5nhash")},
			exp:  cz("5nhash"),
		},
		{
			name: "three coins",
			gm:   &FlatFeeGasMeter{upFrontCost: cz("9apple,14banana,64cherry")},
			exp:  cz("9apple,14banana,64cherry"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = tc.gm.GetUpFrontCost()
			}
			require.NotPanics(t, testFunc, "GetUpFrontCost")
			assert.Equal(t, tc.exp.String(), act.String(), "GetUpFrontCost value")
		})
	}
}

func TestFlatFeeGasMeter_GetOnSuccessCost(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  sdk.Coins
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  nil,
		},
		{
			name: "nil value",
			gm:   &FlatFeeGasMeter{onSuccessCost: nil},
			exp:  nil,
		},
		{
			name: "empty",
			gm:   &FlatFeeGasMeter{onSuccessCost: sdk.Coins{}},
			exp:  sdk.Coins{},
		},
		{
			name: "one coin",
			gm:   &FlatFeeGasMeter{onSuccessCost: cz("5nhash")},
			exp:  cz("5nhash"),
		},
		{
			name: "three coins",
			gm:   &FlatFeeGasMeter{onSuccessCost: cz("9apple,14banana,64cherry")},
			exp:  cz("9apple,14banana,64cherry"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = tc.gm.GetOnSuccessCost()
			}
			require.NotPanics(t, testFunc, "GetOnSuccessCost")
			assert.Equal(t, tc.exp.String(), act.String(), "GetOnSuccessCost value")
		})
	}
}

func TestFlatFeeGasMeter_GetExtraMsgsCost(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  sdk.Coins
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  nil,
		},
		{
			name: "nil value",
			gm:   &FlatFeeGasMeter{extraMsgsCost: nil},
			exp:  nil,
		},
		{
			name: "empty",
			gm:   &FlatFeeGasMeter{extraMsgsCost: sdk.Coins{}},
			exp:  sdk.Coins{},
		},
		{
			name: "one coin",
			gm:   &FlatFeeGasMeter{extraMsgsCost: cz("5nhash")},
			exp:  cz("5nhash"),
		},
		{
			name: "three coins",
			gm:   &FlatFeeGasMeter{extraMsgsCost: cz("9apple,14banana,64cherry")},
			exp:  cz("9apple,14banana,64cherry"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = tc.gm.GetExtraMsgsCost()
			}
			require.NotPanics(t, testFunc, "GetExtraMsgsCost")
			assert.Equal(t, tc.exp.String(), act.String(), "GetExtraMsgsCost value")
		})
	}
}

func TestFlatFeeGasMeter_GetAddedFees(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  sdk.Coins
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  nil,
		},
		{
			name: "nil value",
			gm:   &FlatFeeGasMeter{addedFees: nil},
			exp:  nil,
		},
		{
			name: "empty",
			gm:   &FlatFeeGasMeter{addedFees: sdk.Coins{}},
			exp:  sdk.Coins{},
		},
		{
			name: "one coin",
			gm:   &FlatFeeGasMeter{addedFees: cz("5nhash")},
			exp:  cz("5nhash"),
		},
		{
			name: "three coins",
			gm:   &FlatFeeGasMeter{addedFees: cz("9apple,14banana,64cherry")},
			exp:  cz("9apple,14banana,64cherry"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = tc.gm.GetAddedFees()
			}
			require.NotPanics(t, testFunc, "GetAddedFees")
			assert.Equal(t, tc.exp.String(), act.String(), "GetAddedFees value")
		})
	}
}

func TestFlatFeeGasMeter_GetRequiredFee(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  sdk.Coins
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  nil,
		},
		{
			name: "empty gas meter",
			gm:   &FlatFeeGasMeter{},
			exp:  nil,
		},
		{
			name: "only up-front cost",
			gm:   &FlatFeeGasMeter{upFrontCost: cz("22upfrontcoin")},
			exp:  cz("22upfrontcoin"),
		},
		{
			name: "only on-success cost",
			gm:   &FlatFeeGasMeter{onSuccessCost: cz("54onsuccesscoin")},
			exp:  cz("54onsuccesscoin"),
		},
		{
			name: "only extra msgs cost",
			gm:   &FlatFeeGasMeter{extraMsgsCost: cz("72extramsgcoin")},
			exp:  cz("72extramsgcoin"),
		},
		{
			name: "only added fees",
			gm:   &FlatFeeGasMeter{addedFees: cz("18addedfeecoin")},
			exp:  cz("18addedfeecoin"),
		},
		{
			name: "each has different denoms",
			gm: &FlatFeeGasMeter{
				upFrontCost:   cz("68upfrontcoin"),
				onSuccessCost: cz("19onsuccesscoin"),
				extraMsgsCost: cz("57extramsgcoin"),
				addedFees:     cz("43addedfeecoin"),
			},
			exp: cz("43addedfeecoin,57extramsgcoin,19onsuccesscoin,68upfrontcoin"),
		},
		{
			// 329, 658, 147, 229, 303, 626, 773, 951
			name: "each has same denom",
			gm: &FlatFeeGasMeter{
				upFrontCost:   cz("50001gascoin"),
				onSuccessCost: cz("6002gascoin"),
				extraMsgsCost: cz("703gascoin"),
				addedFees:     cz("84gascoin"),
			},
			exp: cz("56790gascoin"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act sdk.Coins
			testFunc := func() {
				act = tc.gm.GetRequiredFee()
			}
			require.NotPanics(t, testFunc, "GetRequiredFee")
			assert.Equal(t, tc.exp.String(), act.String(), "GetRequiredFee value")
		})
	}
}

func TestFlatFeeGasMeter_String(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	t.Run("nil gas meter", func(t *testing.T) {
		exp := "<nil>"
		var gm *FlatFeeGasMeter
		var act string
		testFunc := func() {
			act = gm.String()
		}
		require.NotPanics(t, testFunc, "FlatFeeGasMeter.String()")
		assert.Equal(t, exp, act, "FlatFeeGasMeter.String()")
	})

	t.Run("gas meter with info", func(t *testing.T) {
		gm := &FlatFeeGasMeter{
			GasMeter:      NewMockGasMeter().WithString("mock-gas-meter-string\n"),
			upFrontCost:   cz("83upfrontcost"),
			onSuccessCost: cz("27onsuccesscost"),
			extraMsgsCost: cz("76extramsgscost"),
			addedFees:     cz("49addedfees"),
		}
		expLines := []string{
			"FlatFeeGasMeter:",
			"    msg type urls: <none>",
			"    up-front cost: 83upfrontcost",
			"  on-success cost: 27onsuccesscost",
			"  extra msgs cost: 76extramsgscost",
			"       added fees: 49addedfees",
			"   gas meter type: mock-gas-meter-string",
		}

		var act string
		testFunc := func() {
			act = gm.String()
		}
		require.NotPanics(t, testFunc, "FlatFeeGasMeter.String()")
		require.NotEmpty(t, act, "FlatFeeGasMeter.String()")
		actLines := strings.Split(act, "\n")
		for _, exp := range expLines {
			assert.Contains(t, actLines, exp, "Expected: %q")
		}
	})
}

func TestFlatFeeGasMeter_MsgCountsString(t *testing.T) {
	newGM := func(msgTypeURLs ...string) *FlatFeeGasMeter {
		return &FlatFeeGasMeter{msgTypeURLs: msgTypeURLs}
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  string
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  "<nil>",
		},
		{
			name: "no msgs",
			gm:   newGM(),
			exp:  "<none>",
		},
		{
			name: "one msg",
			gm:   newGM("somemsg"),
			exp:  "somemsg",
		},
		{
			name: "two msgs: different",
			gm:   newGM("/just.some.MsgWork", "/the.fun.MsgDance"),
			exp:  "/just.some.MsgWork, /the.fun.MsgDance",
		},
		{
			name: "two msgs: different, opposite order",
			gm:   newGM("/the.fun.MsgDance", "/just.some.MsgWork"),
			exp:  "/the.fun.MsgDance, /just.some.MsgWork",
		},
		{
			name: "two msgs: same",
			gm:   newGM("/whatever", "/whatever"),
			exp:  "2x/whatever",
		},
		{
			name: "multiple msgs, some duplicated",
			gm: newGM(
				"/fun.smile.v1.MsgDance",
				"/boring.goodbye.v4.MsgPlanet",
				"/boring.hello.v8.MsgWorld",
				"/fun.smile.v1.MsgDance",
				"/boring.hello.v8.MsgWorld",
				"/fun.smile.v1.MsgParties",
				"/fun.smile.v1.MsgDance",
			),
			exp: "3x/fun.smile.v1.MsgDance, /boring.goodbye.v4.MsgPlanet, 2x/boring.hello.v8.MsgWorld, /fun.smile.v1.MsgParties",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.gm.MsgCountsString()
			}
			require.NotPanics(t, testFunc, "MsgCountsString()")
			assert.Equal(t, tc.exp, act, "MsgCountsString() error")
		})
	}
}

func TestFlatFeeGasMeter_GasUseString(t *testing.T) {
	lb := "------------------------------\n"

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  string
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  "<nil>",
		},
		{
			name: "empty gas meter",
			gm:   &FlatFeeGasMeter{},
			exp: "\n" +
				lb +
				"         0 = Total gas",
		},
		{
			name: "one entry",
			gm: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"stuff": 350},
				counts: map[string]uint64{"stuff": 7},
			},
			exp: "       350 =   7x stuff\n" +
				lb +
				"       350 = Total gas",
		},
		{
			name: "five entries",
			gm: &FlatFeeGasMeter{
				used: map[string]storetypes.Gas{
					"txSize":    3150,
					"Has":       12001,
					"ReadFlat":  4003,
					"WriteFlat": 10007,
					"Delete":    3015,
				},
				counts: map[string]uint64{
					"Has":       12,
					"WriteFlat": 5,
					"txSize":    1,
					"ReadFlat":  4,
					"Delete":    3,
				},
			},
			exp: "      3015 =   3x Delete\n" +
				"     12001 =  12x Has\n" +
				"      4003 =   4x ReadFlat\n" +
				"     10007 =   5x WriteFlat\n" +
				"      3150 =   1x txSize\n" +
				lb +
				"     32176 = Total gas",
		},
		{
			name: "twenty entries",
			gm: &FlatFeeGasMeter{
				// All these numbers were generated randomly and don't have any special meaning.
				// The order of the entries was also randomized so that they're not alphabetical, in
				// the hopes that it makes it more likely to produce nondeterministic behavior if it's there.
				used: map[string]storetypes.Gas{
					"f-thing-2": 28087, "F-thing-2": 78151, "d-thing": 32055, "D-thing": 61279,
					"g-thing": 57088, "G-thing": 60531, "h-thing": 82971, "H-thing": 52613,
					"i-thing": 44098, "I-thing": 53594, "c-thing": 74348, "C-thing": 27077,
					"b-thing": 52120, "B-thing": 457, "a-thing": 54488, "A-thing": 42096,
					"f-thing-1": 78889, "F-thing-1": 1492, "e-thing": 19768, "E-thing": 47905,
				},
				counts: map[string]uint64{
					"f-thing-2": 10, "F-thing-2": 27, "d-thing": 21, "D-thing": 23,
					"g-thing": 17, "G-thing": 16, "h-thing": 7, "H-thing": 19,
					"i-thing": 29, "I-thing": 28, "c-thing": 15, "C-thing": 6,
					"b-thing": 20, "B-thing": 9, "a-thing": 26, "A-thing": 3,
					"f-thing-1": 18, "F-thing-1": 22, "e-thing": 1, "E-thing": 2,
				},
			},
			exp: "     42096 =   3x A-thing\n" +
				"       457 =   9x B-thing\n" +
				"     27077 =   6x C-thing\n" +
				"     61279 =  23x D-thing\n" +
				"     47905 =   2x E-thing\n" +
				"      1492 =  22x F-thing-1\n" +
				"     78151 =  27x F-thing-2\n" +
				"     60531 =  16x G-thing\n" +
				"     52613 =  19x H-thing\n" +
				"     53594 =  28x I-thing\n" +
				"     54488 =  26x a-thing\n" +
				"     52120 =  20x b-thing\n" +
				"     74348 =  15x c-thing\n" +
				"     32055 =  21x d-thing\n" +
				"     19768 =   1x e-thing\n" +
				"     78889 =  18x f-thing-1\n" +
				"     28087 =  10x f-thing-2\n" +
				"     57088 =  17x g-thing\n" +
				"     82971 =   7x h-thing\n" +
				"     44098 =  29x i-thing\n" +
				lb +
				"    949107 = Total gas",
		},
	}

	// Since this involves some maps and iteration, we also test that it's
	// deterministic by running each test case multiple times.
	count := 100
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.gm.GasUseString()
			}
			for i := 1; i <= count; i++ {
				require.NotPanics(t, testFunc, "[%d/%d]: GasUseString()", i, count)
				// Do the comparison on a slice of the lines because that's easier to read in failure messages.
				expLines := strings.Split(tc.exp, "\n")
				actLines := strings.Split(act, "\n")
				if !assert.Equal(t, expLines, actLines, "[%d/%d]: GasUseString() result split into lines", i, count) && i == 1 {
					// If it fails on the first one, stop now since that probably means they'll all fail.
					// If it's not the first, keep going so it's easier to see that it's not deterministic.
					break
				}
			}
		})
	}
}

func TestFlatFeeGasMeter_RequiredFeeString(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
		exp  string
	}{
		{
			name: "nil gas meter",
			gm:   nil,
			exp:  "<nil>",
		},
		{
			name: "empty gas meter",
			gm:   &FlatFeeGasMeter{},
			exp:  "",
		},
		{
			name: "one type: up-front",
			gm:   &FlatFeeGasMeter{upFrontCost: cz("3acorn")},
			exp:  "3acorn",
		},
		{
			name: "one type: on success",
			gm:   &FlatFeeGasMeter{onSuccessCost: cz("12banana")},
			exp:  "12banana",
		},
		{
			name: "one type: extra msg",
			gm:   &FlatFeeGasMeter{extraMsgsCost: cz("7cherry,8durian")},
			exp:  "7cherry,8durian",
		},
		{
			name: "one type: added fees",
			gm:   &FlatFeeGasMeter{addedFees: cz("41elderberry")},
			exp:  "41elderberry",
		},
		{
			name: "two types: up-front, on success",
			gm:   &FlatFeeGasMeter{upFrontCost: cz("15plum"), onSuccessCost: cz("6plum")},
			exp:  "21plum = 15plum (up-front) + 6plum (on success)",
		},
		{
			name: "two types: up-front, extra msg",
			gm:   &FlatFeeGasMeter{upFrontCost: cz("7cherry"), extraMsgsCost: cz("4apple")},
			exp:  "4apple,7cherry = 7cherry (up-front) + 4apple (extra msgs)",
		},
		{
			name: "two types: up-front, added fees",
			gm:   &FlatFeeGasMeter{upFrontCost: cz("12apple,3banana"), addedFees: cz("96banana,12cherry")},
			exp:  "12apple,99banana,12cherry = 12apple,3banana (up-front) + 96banana,12cherry (added fees)",
		},
		{
			name: "two types: on success, extra msg",
			gm:   &FlatFeeGasMeter{onSuccessCost: cz("6orange"), extraMsgsCost: cz("40orange")},
			exp:  "46orange = 6orange (on success) + 40orange (extra msgs)",
		},
		{
			name: "two types: on success, added fees",
			gm:   &FlatFeeGasMeter{onSuccessCost: cz("14grape"), addedFees: cz("3cantaloupe")},
			exp:  "3cantaloupe,14grape = 14grape (on success) + 3cantaloupe (added fees)",
		},
		{
			name: "two types: extra msg, added fees",
			gm:   &FlatFeeGasMeter{extraMsgsCost: cz("4strawberry"), addedFees: cz("77strawberry")},
			exp:  "81strawberry = 4strawberry (extra msgs) + 77strawberry (added fees)",
		},
		{
			name: "three types: up-front, on success, extra msg",
			gm: &FlatFeeGasMeter{
				upFrontCost:   cz("15pineapple"),
				onSuccessCost: cz("3pineapple"),
				extraMsgsCost: cz("870pineapple"),
			},
			exp: "888pineapple = 15pineapple (up-front) + 3pineapple (on success) + 870pineapple (extra msgs)",
		},
		{
			name: "three types: up-front, on success, added fees",
			gm: &FlatFeeGasMeter{
				upFrontCost:   cz("6apple"),
				onSuccessCost: cz("4apple,3banana"),
				addedFees:     cz("9banana,2cherry"),
			},
			exp: "10apple,12banana,2cherry = 6apple (up-front) + 4apple,3banana (on success) + 9banana,2cherry (added fees)",
		},
		{
			name: "three types: up-front, extra msg, added fees",
			gm: &FlatFeeGasMeter{
				upFrontCost:   cz("14pear"),
				extraMsgsCost: cz("27pear"),
				addedFees:     cz("43pear"),
			},
			exp: "84pear = 14pear (up-front) + 27pear (extra msgs) + 43pear (added fees)",
		},
		{
			name: "three types: on success, extra msg, added fees",
			gm: &FlatFeeGasMeter{
				onSuccessCost: cz("1success"),
				extraMsgsCost: cz("2extra"),
				addedFees:     cz("3added"),
			},
			exp: "3added,2extra,1success = 1success (on success) + 2extra (extra msgs) + 3added (added fees)",
		},
		{
			name: "four types: same denom",
			gm: &FlatFeeGasMeter{
				upFrontCost:   cz("10nhash"),
				onSuccessCost: cz("35nhash"),
				extraMsgsCost: cz("500nhash"),
				addedFees:     cz("2nhash"),
			},
			exp: "547nhash = 10nhash (up-front) + 35nhash (on success) + 500nhash (extra msgs) + 2nhash (added fees)",
		},
		{
			name: "four types: mixed denoms",
			// unique to field: upfront, success, extra, added
			// in all fields: shared
			// In exactly two fields:
			// 	* upsuc: up-front and on success
			// 	* sucex: on success and extra msgs
			// 	* exadd: extra msgs and added fees
			// 	* upadd: up-front and added fees.
			gm: &FlatFeeGasMeter{
				upFrontCost:   cz("7upfront,401shared,12upsuc,6upadd"),
				onSuccessCost: cz("4success,58shared,30upsuc,8sucex"),
				extraMsgsCost: cz("5extra,111shared,66exadd,144sucex"),
				addedFees:     cz("2added,6shared,25exadd,39upadd"),
			},
			exp: "2added,91exadd,5extra,576shared,4success,152sucex,45upadd,7upfront,42upsuc" +
				" = 401shared,6upadd,7upfront,12upsuc (up-front)" +
				" + 58shared,4success,8sucex,30upsuc (on success)" +
				" + 66exadd,5extra,111shared,144sucex (extra msgs)" +
				" + 2added,25exadd,6shared,39upadd (added fees)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var act string
			testFunc := func() {
				act = tc.gm.RequiredFeeString()
			}
			require.NotPanics(t, testFunc, "FlatFeeGasMeter.RequiredFeeString()")
			assert.Equal(t, tc.exp, act, "FlatFeeGasMeter.RequiredFeeString() result")
		})
	}
}

func TestFlatFeeGasMeter_DetailsString(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		gm   *FlatFeeGasMeter
	}{
		{
			name: "nil gas meter",
			gm:   nil,
		},
		{
			name: "empty gas meter",
			gm:   &FlatFeeGasMeter{},
		},
		{
			name: "with a little bit of everything",
			gm: &FlatFeeGasMeter{
				GasMeter:      NewMockGasMeter().WithString("mocked-gas-meter-string-result"),
				upFrontCost:   cz("1nhash"),
				onSuccessCost: cz("3nhash"),
				extraMsgsCost: cz("7nhash"),
				addedFees:     cz("15nhash"),
				msgTypeURLs:   []string{"/abc.xyz.MsgOne", "/abc.xyz.MsgTwo", "/abc.xyz.MsgOne", "/abc.xyz.MsgThree"},
				used:          map[string]storetypes.Gas{"Read": 91644, "Write": 7923, "Has": 33537, "Delete": 49735},
				counts:        map[string]uint64{"Read": 3, "Write": 10, "Has": 7, "Delete": 13},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expPrefix, expMsgs, expCost, expGas, expNil string
			if tc.gm != nil {
				setExp := func() {
					expPrefix = "FlatFeeGasMeter:\n"
					expMsgs = "  Msgs: " + tc.gm.MsgCountsString()
					expCost = "  Cost: " + tc.gm.RequiredFeeString()
					expGas = "  Gas:\n" + tc.gm.GasUseString()
				}
				require.NotPanics(t, setExp, "Setup: Defining expected strings.")
			} else {
				expNil = "<nil>"
			}

			var act string
			testFunc := func() {
				act = tc.gm.DetailsString()
			}
			require.NotPanics(t, testFunc, "FlatFeeGasMeter.DetailsString()")

			if tc.gm != nil {
				assert.True(t, strings.HasPrefix(act, expPrefix), "strings.HasPrefix(actual, %q):\nactual:\n%s", expPrefix, act)
				assert.Contains(t, act, expMsgs, "Msgs section")
				assert.Contains(t, act, expCost, "Cost section")
				assert.Contains(t, act, expGas, "Gas section")
			} else {
				assert.Equal(t, expNil, act, "FlatFeeGasMeter.DetailsString() result")
			}
		})
	}
}

func TestFlatFeeGasMeter_ConsumeGas(t *testing.T) {
	tests := []struct {
		name       string
		gm         *FlatFeeGasMeter
		amount     storetypes.Gas
		descriptor string
		expGM      *FlatFeeGasMeter
	}{
		{
			name: "empty gas meter",
			gm: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{},
				counts: map[string]uint64{},
			},
			amount:     7,
			descriptor: "pants",
			expGM: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"pants": 7},
				counts: map[string]uint64{"pants": 1},
			},
		},
		{
			name: "new descriptor",
			gm: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"pants": 98},
				counts: map[string]uint64{"pants": 4},
			},
			amount:     55,
			descriptor: "shirt",
			expGM: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"pants": 98, "shirt": 55},
				counts: map[string]uint64{"pants": 4, "shirt": 1},
			},
		},
		{
			name: "existing descriptor",
			gm: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"pants": 411, "shirt": 71, "shoes": 9},
				counts: map[string]uint64{"pants": 11, "shirt": 6, "shoes": 1},
			},
			amount:     14,
			descriptor: "shirt",
			expGM: &FlatFeeGasMeter{
				used:   map[string]storetypes.Gas{"pants": 411, "shirt": 85, "shoes": 9},
				counts: map[string]uint64{"pants": 11, "shirt": 7, "shoes": 1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseGM := NewMockGasMeter()
			tc.gm.GasMeter = baseGM
			testFunc := func() {
				tc.gm.ConsumeGas(tc.amount, tc.descriptor)
			}
			require.NotPanics(t, testFunc, "ConsumeGas(%d, %q)", tc.amount, tc.descriptor)
			AssertEqualFlatFeeGasMeters(t, tc.expGM, tc.gm)
			baseGM.AssertConsumeGasCall(t, NewConsumeGasArgs(tc.amount, tc.descriptor))
		})
	}
}

func TestFlatFeeGasMeter_GasConsumed(t *testing.T) {
	expGasConsumed := storetypes.Gas(4181777)
	expInLog := []string{"FlatFeeGasMeter:", "Msgs:", "Cost:", "Gas:"}

	baseGM := NewMockGasMeter().WithGasConsumed(expGasConsumed)
	var buffer bytes.Buffer
	logger := internal.NewBufferedInfoLogger(&buffer)

	gasMeter := NewFlatFeeGasMeter(baseGM, logger, nil)

	var actGasConsumed storetypes.Gas
	testFunc := func() {
		actGasConsumed = gasMeter.GasConsumed()
	}
	require.NotPanics(t, testFunc, "GasConsumed()")
	AssertEqualGas(t, expGasConsumed, actGasConsumed, "GasConsumed() return value")

	logged := buffer.String()
	require.NotEmpty(t, logged, "logs written")
	assert.True(t, strings.HasPrefix(logged, "INF "), "strings.HasPrefix(logged, \"INF \")\nlogged: %q", logged)
	for _, exp := range expInLog {
		assert.Contains(t, logged, exp, "logged  : %q\nexpected: %q", logged, exp)
	}
}

func TestFlatFeeGasMeter_ConsumeMsg(t *testing.T) {
	tests := []struct {
		name  string
		gm    *FlatFeeGasMeter
		msg   sdk.Msg
		expGM *FlatFeeGasMeter
	}{
		{
			name: "msg expected out of one",
			gm: &FlatFeeGasMeter{
				knownMsgs:   map[string]int{"/cosmos.bank.v1beta1.MsgSend": 1},
				msgTypeURLs: []string{"/cosmos.bank.v1beta1.MsgSend"},
			},
			msg: &banktypes.MsgSend{},
			expGM: &FlatFeeGasMeter{
				knownMsgs:   map[string]int{"/cosmos.bank.v1beta1.MsgSend": 0},
				msgTypeURLs: []string{"/cosmos.bank.v1beta1.MsgSend"},
			},
		},
		{
			name: "msg expected out of many",
			gm: &FlatFeeGasMeter{
				knownMsgs: map[string]int{
					"/cosmos.bank.v1beta1.MsgSend": 3,
					"/cosmos.gov.v1.MsgVote":       1,
				},
				msgTypeURLs: []string{
					"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgVote", "/cosmos.bank.v1beta1.MsgSend",
				},
			},
			msg: &banktypes.MsgSend{},
			expGM: &FlatFeeGasMeter{
				knownMsgs: map[string]int{
					"/cosmos.bank.v1beta1.MsgSend": 2,
					"/cosmos.gov.v1.MsgVote":       1,
				},
				msgTypeURLs: []string{
					"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgVote", "/cosmos.bank.v1beta1.MsgSend",
				},
			},
		},
		{
			name: "msg type not in known msgs map",
			gm: &FlatFeeGasMeter{
				knownMsgs: map[string]int{
					"/cosmos.bank.v1beta1.MsgSend": 3,
					"/cosmos.gov.v1.MsgVote":       1,
				},
				msgTypeURLs: []string{
					"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgVote", "/cosmos.bank.v1beta1.MsgSend",
				},
			},
			msg: &govv1.MsgDeposit{},
			expGM: &FlatFeeGasMeter{
				knownMsgs: map[string]int{
					"/cosmos.bank.v1beta1.MsgSend": 3,
					"/cosmos.gov.v1.MsgVote":       1,
				},
				extraMsgs: []sdk.Msg{&govv1.MsgDeposit{}},
				msgTypeURLs: []string{
					"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgVote", "/cosmos.bank.v1beta1.MsgSend",
					"/cosmos.gov.v1.MsgDeposit",
				},
			},
		},
		{
			name: "msg type at zero in known msgs map",
			gm: &FlatFeeGasMeter{
				knownMsgs:   map[string]int{"/cosmos.bank.v1beta1.MsgSend": 0},
				msgTypeURLs: []string{"/cosmos.bank.v1beta1.MsgSend"},
			},
			msg: &banktypes.MsgSend{},
			expGM: &FlatFeeGasMeter{
				knownMsgs:   map[string]int{"/cosmos.bank.v1beta1.MsgSend": 0},
				extraMsgs:   []sdk.Msg{&banktypes.MsgSend{}},
				msgTypeURLs: []string{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend"},
			},
		},
		{
			name: "msg type below zero in known msgs map",
			gm: &FlatFeeGasMeter{
				knownMsgs:   map[string]int{"/cosmos.bank.v1beta1.MsgSend": -1},
				msgTypeURLs: []string{"/cosmos.bank.v1beta1.MsgSend"},
			},
			msg: &banktypes.MsgSend{},
			expGM: &FlatFeeGasMeter{
				knownMsgs:   map[string]int{"/cosmos.bank.v1beta1.MsgSend": -1},
				extraMsgs:   []sdk.Msg{&banktypes.MsgSend{}},
				msgTypeURLs: []string{"/cosmos.bank.v1beta1.MsgSend", "/cosmos.bank.v1beta1.MsgSend"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testFunc := func() {
				tc.gm.ConsumeMsg(tc.msg)
			}
			require.NotPanics(t, testFunc, "ConsumeMsg(%s)", sdk.MsgTypeURL(tc.msg))
			AssertEqualFlatFeeGasMeters(t, tc.expGM, tc.gm)
		})
	}
}

func TestFlatFeeGasMeter_ConsumeAddedFee(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name  string
		gm    *FlatFeeGasMeter
		fee   sdk.Coins
		expGM *FlatFeeGasMeter
	}{
		{
			name:  "nil fee",
			gm:    &FlatFeeGasMeter{},
			fee:   nil,
			expGM: &FlatFeeGasMeter{},
		},
		{
			name:  "empty fee",
			gm:    &FlatFeeGasMeter{},
			fee:   sdk.Coins{},
			expGM: &FlatFeeGasMeter{},
		},
		{
			name:  "no previously added fees",
			gm:    &FlatFeeGasMeter{},
			fee:   cz("15plum"),
			expGM: &FlatFeeGasMeter{addedFees: cz("15plum")},
		},
		{
			name:  "previously added fees with same denom",
			gm:    &FlatFeeGasMeter{addedFees: cz("17cherry")},
			fee:   cz("101cherry"),
			expGM: &FlatFeeGasMeter{addedFees: cz("118cherry")},
		},
		{
			name:  "previously added fees with different denoms",
			gm:    &FlatFeeGasMeter{addedFees: cz("3apple,99banana,17cherry")},
			fee:   cz("8apple,6plum"),
			expGM: &FlatFeeGasMeter{addedFees: cz("11apple,99banana,17cherry,6plum")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testFunc := func() {
				tc.gm.ConsumeAddedFee(tc.fee)
			}
			require.NotPanics(t, testFunc, "ConsumeAddedFee")
			AssertEqualFlatFeeGasMeters(t, tc.expGM, tc.gm)
		})
	}
}

func TestFlatFeeGasMeter_adjustCostsForUnitTests(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name     string
		gm       *FlatFeeGasMeter
		chainID  string
		fee      sdk.Coins
		expGM    *FlatFeeGasMeter
		expInLog []string
	}{
		{
			name:     "empty chainID",
			gm:       &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			chainID:  "",
			fee:      cz("5new"),
			expGM:    &FlatFeeGasMeter{upFrontCost: cz("5new"), onSuccessCost: nil},
			expInLog: []string{"Using provided fee for test tx cost.", "fee_provided=5new"},
		},
		{
			name:     "simapp-unit-testing", // SimAppChainID
			gm:       &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			chainID:  "simapp-unit-testing",
			fee:      cz("6new"),
			expGM:    &FlatFeeGasMeter{upFrontCost: nil, onSuccessCost: nil},
			expInLog: []string{"Using zero for test tx cost."},
		},
		{
			name:     "simulation-app", // pioconfig.SimAppChainID,
			gm:       &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			chainID:  "simulation-app",
			fee:      cz("7new"),
			expGM:    &FlatFeeGasMeter{upFrontCost: nil, onSuccessCost: nil},
			expInLog: []string{"Using zero for test tx cost."},
		},
		{
			name:     "testchain-83",
			gm:       &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			chainID:  "testchain-83",
			fee:      cz("8new"),
			expGM:    &FlatFeeGasMeter{upFrontCost: nil, onSuccessCost: nil},
			expInLog: []string{"Using zero for test tx cost."},
		},
		{
			name:     "mainnet",
			gm:       &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			chainID:  "pio-mainnet-1",
			fee:      cz("9new"),
			expGM:    &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			expInLog: []string{"Not a unit test. Not adjusting tx cost."},
		},
		{
			name:     "testnet",
			gm:       &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			chainID:  "pio-testnet-1",
			fee:      cz("3new"),
			expGM:    &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			expInLog: []string{"Not a unit test. Not adjusting tx cost."},
		},
		{
			name:     "other chain id",
			gm:       &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			chainID:  "c98u-j3ec",
			fee:      cz("41new"),
			expGM:    &FlatFeeGasMeter{upFrontCost: cz("12orig"), onSuccessCost: cz("77orig")},
			expInLog: []string{"Not a unit test. Not adjusting tx cost."},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.expInLog = append(tc.expInLog, "adjustCostsForUnitTests")

			var buffer bytes.Buffer
			tc.gm.logger = internal.NewBufferedDebugLogger(&buffer)

			testFunc := func() {
				tc.gm.adjustCostsForUnitTests(tc.chainID, tc.fee)
			}
			require.NotPanics(t, testFunc, "adjustCostsForUnitTests(%q, %q)", tc.chainID, tc.fee)
			AssertEqualFlatFeeGasMeters(t, tc.expGM, tc.gm)

			logged := buffer.String()
			require.NotEmpty(t, logged, "logs written")
			assert.True(t, strings.HasPrefix(logged, "DBG "), "strings.HasPrefix(logged, \"DBG \")\nlogged: %q", logged)
			for _, exp := range tc.expInLog {
				assert.Contains(t, logged, exp, "logged  : %q\nexpected: %q", logged, exp)
			}
		})
	}
}

func TestGetFlatFeeGasMeter(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name   string
		ctxGM  storetypes.GasMeter
		expGM  *FlatFeeGasMeter
		expErr string
	}{
		{
			name:   "nil gas meter",
			ctxGM:  nil,
			expErr: "gas meter is not a FlatFeeGasMeter: <nil>: internal logic error",
		},
		{
			name:   "not a flat-fee gas meter",
			ctxGM:  NewMockGasMeter(),
			expErr: "gas meter is not a FlatFeeGasMeter: *antewrapper.MockGasMeter: internal logic error",
		},
		{
			name: "flat-fee gas meter",
			ctxGM: &FlatFeeGasMeter{
				upFrontCost:   cz("12nhash"),
				onSuccessCost: cz("48nhash"),
				addedFees:     cz("12usdc"),
				msgTypeURLs:   []string{"abc", "xyz"},
			},
			expGM: &FlatFeeGasMeter{
				upFrontCost:   cz("12nhash"),
				onSuccessCost: cz("48nhash"),
				addedFees:     cz("12usdc"),
				msgTypeURLs:   []string{"abc", "xyz"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithGasMeter(tc.ctxGM)

			var actGM *FlatFeeGasMeter
			var err error
			testFunc := func() {
				actGM, err = GetFlatFeeGasMeter(ctx)
			}
			require.NotPanics(t, testFunc, "GetFlatFeeGasMeter")
			assertions.AssertErrorValue(t, err, tc.expErr, "GetFlatFeeGasMeter error")
			AssertEqualFlatFeeGasMeters(t, tc.expGM, actGM)
		})
	}
}

func TestConsumeAdditionalFee(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name  string
		ctxGM storetypes.GasMeter
		fee   sdk.Coins
		expGM *FlatFeeGasMeter
	}{
		{
			name:  "nil fee",
			ctxGM: nil,
			fee:   nil,
			expGM: nil,
		},
		{
			name:  "empty fee",
			ctxGM: nil,
			fee:   sdk.Coins{},
			expGM: nil,
		},
		{
			name:  "zero fee",
			ctxGM: nil,
			fee:   sdk.Coins{sdk.Coin{Denom: "broccoli", Amount: sdkmath.ZeroInt()}},
			expGM: nil,
		},
		{
			name:  "no previously consumed added fee",
			ctxGM: &FlatFeeGasMeter{addedFees: nil},
			fee:   cz("3cherry"),
			expGM: &FlatFeeGasMeter{addedFees: cz("3cherry")},
		},
		{
			name:  "added to previously consumed added fee",
			ctxGM: &FlatFeeGasMeter{addedFees: cz("3apple,4cherry")},
			fee:   cz("99banana,18cherry"),
			expGM: &FlatFeeGasMeter{addedFees: cz("3apple,99banana,22cherry")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithGasMeter(tc.ctxGM)
			testFunc := func() {
				ConsumeAdditionalFee(ctx, tc.fee)
			}
			require.NotPanics(t, testFunc, "ConsumeAdditionalFee")
			actGM, _ := GetFlatFeeGasMeter(ctx)
			AssertEqualFlatFeeGasMeters(t, tc.expGM, actGM)
		})
	}
}
