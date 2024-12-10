package keeper_test

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/nav"

	. "github.com/provenance-io/provenance/x/nav/keeper"
)

type TestSuite struct {
	suite.Suite

	app       *app.App
	ctx       sdk.Context
	navKeeper *Keeper
}

func (s *TestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
	s.navKeeper = s.app.NAVKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// coinRx is a regular expression that matches an sdk.Coin string, capturing the amount and denom parts.
var coinRx = regexp.MustCompile(`^(-?\d+)(.+)$`)

// coin converts the provided string into a Coin, but is more forgiving than ParseCoinNormalized (and has a shorter name).
func (s *TestSuite) coin(coinStr string) sdk.Coin {
	if len(coinStr) == 0 {
		return sdk.Coin{}
	}
	matches := coinRx.FindStringSubmatch(coinStr)
	s.Require().Len(matches, 3, "coinRx matches of %q", coinStr)
	amt, ok := sdkmath.NewIntFromString(matches[1])
	s.Require().True(ok, "NewIntFromString(%q) ok bool", matches[1])
	return sdk.Coin{Amount: amt, Denom: matches[2]}
}

// newNAV creates a new NetAssetValue by converting the provided strings into coins.
func (s *TestSuite) newNAV(assets, price string) *nav.NetAssetValue {
	return &nav.NetAssetValue{Assets: s.coin(assets), Price: s.coin(price)}
}

// newNAVRec creates a new NetAssetValueRecord by converting the provided strings into coins.
func (s *TestSuite) newNAVRec(assets, price string, height uint64, source string) *nav.NetAssetValueRecord {
	return &nav.NetAssetValueRecord{Assets: s.coin(assets), Price: s.coin(price), Height: height, Source: source}
}

// newNAVEvent creates the event emitted when a new NAV is stored.
func (s *TestSuite) newNAVEvent(assets, price, source string) sdk.Event {
	return s.untypeEvent(&nav.EventSetNetAssetValue{Assets: assets, Price: price, Source: source})
}

// untypeEvent converts the provided typed event to an untyped one, requiring it to not have an error.
func (s *TestSuite) untypeEvent(tev proto.Message) sdk.Event {
	rv, err := sdk.TypedEventToEvent(tev)
	s.Require().NoError(err, "TypedEventToEvent(%#v)", tev)
	return rv
}

// storeNAVs will store the provided navs in the navKeeper's navs collection.
// It doesn't do any validation, and no events are emitted.
func (s *TestSuite) storeNAVs(ctx sdk.Context, navs nav.NAVRecords) {
	if len(navs) == 0 {
		return
	}
	navCol := s.navKeeper.NAVs()
	for i, navr := range navs {
		key := navr.Key()
		err := navCol.Set(ctx, key, *navr)
		s.Require().NoError(err, "[%d]: setting %#v in navs collection", i, navr)
	}
}

// compareNAVRecords returns 0 if r1 == r2, -1 if r1 < r2, or 1 if r1 > r2.
func compareNAVRecords(r1, r2 *nav.NetAssetValueRecord) int {
	if r1 == r2 {
		return 0
	}
	if r1 == nil {
		return -1
	}
	if r2 == nil {
		return 1
	}

	if rv := strings.Compare(r1.Assets.Denom, r2.Assets.Denom); rv != 0 {
		return rv
	}
	return strings.Compare(r1.Price.Denom, r2.Price.Denom)
}

// sliceStringer calls .String() and adds an index number to each entry.
func sliceStringer[S ~[]E, E fmt.Stringer](s S) []string {
	rv := make([]string, len(s))
	for i, v := range s {
		rv[i] = fmt.Sprintf("[%d]: ", i) + v.String()
	}
	return rv
}

// assertEqualEvents is a wrapper on assertions.AssertEqualEvents.
func (s *TestSuite) assertEqualEvents(expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertEqualEvents(s.T(), expected, actual, msgAndArgs...)
}

// requirePanicEquals is a wrapper on assertions.RequirePanicEquals.
func (s *TestSuite) requirePanicEquals(f assertions.PanicTestFunc, expected string, msgAndArgs ...interface{}) {
	s.T().Helper()
	assertions.RequirePanicEquals(s.T(), f, expected, msgAndArgs...)
}

// assertEqualNAVRecords will assert that the provided NAVRecords are equal.
func (s *TestSuite) assertEqualNAVRecords(expected, actual nav.NAVRecords, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	// Do the comparison first as strings since that should have easier failure messages.
	expStrs := sliceStringer(expected)
	actStrs := sliceStringer(actual)
	if !s.Assert().Equal(expStrs, actStrs, msgAndArgs...) {
		return false
	}
	// And now do a deep comparison, just in case the strings are missing something.
	return s.Assert().Equal(expected, actual, msgAndArgs...)
}

func (s *TestSuite) getQueryServer() nav.QueryServer {
	return NewQueryServer(*s.navKeeper)
}

func (s *TestSuite) TestKeeper_SetNAVs() {
	tests := []struct {
		name      string
		height    int64
		source    string
		navs      []*nav.NetAssetValue
		expErr    string
		expEvents sdk.Events
	}{
		{
			name:   "negative height",
			height: -3,
			source: "yellow",
			navs:   []*nav.NetAssetValue{s.newNAV("1red", "2blue")},
			expErr: "context has height -3: cannot be less than zero",
		},
		{
			name:   "bad source",
			height: 73,
			source: strings.Repeat("bad", 34),
			navs:   []*nav.NetAssetValue{s.newNAV("1red", "2blue")},
			expErr: "invalid source \"badbadb...badbad\": length 102 exceeds max 100",
		},
		{
			name:      "no navs",
			height:    12,
			source:    "yellow",
			navs:      nil,
			expErr:    "",
			expEvents: nil,
		},
		{
			name:      "one nav: good",
			height:    43,
			source:    "yellow",
			navs:      []*nav.NetAssetValue{s.newNAV("1red", "2blue")},
			expEvents: sdk.Events{s.newNAVEvent("1red", "2blue", "yellow")},
		},
		{
			name:   "one nav: bad",
			height: 77,
			source: "green",
			navs:   []*nav.NetAssetValue{s.newNAV("1fish", "2fish")},
			expErr: "0: nav assets \"1fish\" and price \"2fish\" must have different denoms",
		},
		{
			name:   "three navs: all good",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("35purple", "81gold"),
			},
			expEvents: sdk.Events{
				s.newNAVEvent("52maroon", "12silver", "montana"),
				s.newNAVEvent("18blue", "72gold", "montana"),
				s.newNAVEvent("35purple", "81gold", "montana"),
			},
		},
		{
			name:   "three navs: first bad",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("-52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("35purple", "81gold"),
			},
			expErr: "0: invalid assets \"-52maroon\": negative coin amount: -52",
		},
		{
			name:   "three navs: second bad",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "-72gold"),
				s.newNAV("35purple", "81gold"),
			},
			expErr: "1: invalid price \"-72gold\": negative coin amount: -72",
		},
		{
			name:   "three navs: third bad",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("35purple", "81purple"),
			},
			expErr: "2: nav assets \"35purple\" and price \"81purple\" must have different denoms",
		},
		{
			name:   "three navs: first and last have same denoms",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("8maroon", "77silver"),
			},
			expErr: "cannot have multiple (2) navs with the same asset (\"maroon\") and price (\"silver\") denoms",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx, _ := s.ctx.CacheContext()
			ctx = ctx.WithBlockHeight(tc.height).WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.navKeeper.SetNAVs(ctx, tc.source, tc.navs...)
			}
			s.Require().NotPanics(testFunc, "SetNAVs")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "SetNAVs error")

			actEvents := em.Events()
			s.assertEqualEvents(tc.expEvents, actEvents, "events emitted during SetNAVs")

			// Only run the rest of the checks if no error was expected and no error was encountered.
			if len(tc.expErr) > 0 || err != nil {
				return
			}

			navsCol := s.navKeeper.NAVs()
			for i := range tc.navs {
				exp := tc.navs[i].AsRecord(uint64(tc.height), tc.source)
				act, err2 := navsCol.Get(ctx, collections.Join(exp.Assets.Denom, exp.Price.Denom))
				if s.Assert().NoError(err2, "[%d]: trying to get %s from the navs collection", i, exp) {
					s.Assert().Equal(*exp, act, "[%d]: nav read from the collection")
				}
			}
		})
	}
}

func (s *TestSuite) TestKeeper_SetNAVsAtHeight() {
	tests := []struct {
		name      string
		height    uint64
		source    string
		navs      []*nav.NetAssetValue
		expErr    string
		expEvents sdk.Events
	}{
		{
			name:   "bad source",
			height: 73,
			source: strings.Repeat("bad", 34),
			navs:   []*nav.NetAssetValue{s.newNAV("1red", "2blue")},
			expErr: "invalid source \"badbadb...badbad\": length 102 exceeds max 100",
		},
		{
			name:      "no navs",
			height:    12,
			source:    "yellow",
			navs:      nil,
			expErr:    "",
			expEvents: nil,
		},
		{
			name:      "one nav: good",
			height:    43,
			source:    "yellow",
			navs:      []*nav.NetAssetValue{s.newNAV("1red", "2blue")},
			expEvents: sdk.Events{s.newNAVEvent("1red", "2blue", "yellow")},
		},
		{
			name:   "one nav: bad",
			height: 77,
			source: "green",
			navs:   []*nav.NetAssetValue{s.newNAV("1fish", "2fish")},
			expErr: "0: nav assets \"1fish\" and price \"2fish\" must have different denoms",
		},
		{
			name:   "three navs: all good",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("35purple", "81gold"),
			},
			expEvents: sdk.Events{
				s.newNAVEvent("52maroon", "12silver", "montana"),
				s.newNAVEvent("18blue", "72gold", "montana"),
				s.newNAVEvent("35purple", "81gold", "montana"),
			},
		},
		{
			name:   "three navs: first bad",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("-52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("35purple", "81gold"),
			},
			expErr: "0: invalid assets \"-52maroon\": negative coin amount: -52",
		},
		{
			name:   "three navs: second bad",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "-72gold"),
				s.newNAV("35purple", "81gold"),
			},
			expErr: "1: invalid price \"-72gold\": negative coin amount: -72",
		},
		{
			name:   "three navs: third bad",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("35purple", "81purple"),
			},
			expErr: "2: nav assets \"35purple\" and price \"81purple\" must have different denoms",
		},
		{
			name:   "three navs: first and last have same denoms",
			height: 406,
			source: "montana",
			navs: []*nav.NetAssetValue{
				s.newNAV("52maroon", "12silver"),
				s.newNAV("18blue", "72gold"),
				s.newNAV("8maroon", "77silver"),
			},
			expErr: "cannot have multiple (2) navs with the same asset (\"maroon\") and price (\"silver\") denoms",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx, _ := s.ctx.CacheContext()
			ctx = ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.navKeeper.SetNAVsAtHeight(ctx, tc.source, tc.height, tc.navs...)
			}
			s.Require().NotPanics(testFunc, "SetNAVsAtHeight")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "SetNAVsAtHeight error")

			actEvents := em.Events()
			s.assertEqualEvents(tc.expEvents, actEvents, "events emitted during SetNAVsAtHeight")

			// Only run the rest of the checks if no error was expected and no error was encountered.
			if len(tc.expErr) > 0 || err != nil {
				return
			}

			navsCol := s.navKeeper.NAVs()
			for i := range tc.navs {
				exp := tc.navs[i].AsRecord(tc.height, tc.source)
				act, err2 := navsCol.Get(ctx, collections.Join(exp.Assets.Denom, exp.Price.Denom))
				if s.Assert().NoError(err2, "[%d]: trying to get %s from the navs collection", i, exp) {
					s.Assert().Equal(*exp, act, "[%d]: nav read from the collection")
				}
			}
		})
	}
}

func (s *TestSuite) TestKeeper_SetNAVRecords() {
	tests := []struct {
		name      string
		navs      nav.NAVRecords
		expErr    string
		expEvents sdk.Events
	}{
		{
			name:      "nil navs",
			navs:      nil,
			expErr:    "",
			expEvents: nil,
		},
		{
			name:      "empty navs",
			navs:      nav.NAVRecords{},
			expErr:    "",
			expEvents: nil,
		},
		{
			name:      "one nav: good",
			navs:      nav.NAVRecords{s.newNAVRec("5pink", "3yellow", 15, "red")},
			expEvents: sdk.Events{s.newNAVEvent("5pink", "3yellow", "red")},
		},
		{
			name:   "one nav: bad",
			navs:   nav.NAVRecords{s.newNAVRec("5pink", "3pink", 15, "red")},
			expErr: "0: nav assets \"5pink\" and price \"3pink\" must have different denoms",
		},
		{
			name: "three navs: all good",
			navs: nav.NAVRecords{
				s.newNAVRec("5pink", "3yellow", 15, "red"),
				s.newNAVRec("12pink", "7blue", 17, "redder"),
				s.newNAVRec("87pink", "96green", 22, "reddest"),
			},
			expEvents: sdk.Events{
				s.newNAVEvent("5pink", "3yellow", "red"),
				s.newNAVEvent("12pink", "7blue", "redder"),
				s.newNAVEvent("87pink", "96green", "reddest"),
			},
		},
		{
			name: "three navs: second bad",
			navs: nav.NAVRecords{
				s.newNAVRec("5pink", "3yellow", 15, "red"),
				s.newNAVRec("12pink", "-7blue", 17, "redder"),
				s.newNAVRec("87pink", "96green", 22, "reddest"),
			},
			expErr: "1: invalid price \"-7blue\": negative coin amount: -7",
		},
		{
			name: "three navs: first and last have same denoms",
			navs: nav.NAVRecords{
				s.newNAVRec("5pink", "3yellow", 15, "red"),
				s.newNAVRec("12pink", "7blue", 17, "redder"),
				s.newNAVRec("87pink", "96yellow", 22, "reddest"),
			},
			expErr: "cannot have multiple (2) navs with the same asset (\"pink\") and price (\"yellow\") denoms",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx, _ := s.ctx.CacheContext()
			ctx = ctx.WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.navKeeper.SetNAVRecords(ctx, tc.navs)
			}
			s.Require().NotPanics(testFunc, "SetNAVs")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "SetNAVs error")

			actEvents := em.Events()
			s.assertEqualEvents(tc.expEvents, actEvents, "events emitted during SetNAVs")

			// Only run the rest of the checks if no error was expected and no error was encountered.
			if len(tc.expErr) > 0 || err != nil {
				return
			}

			navsCol := s.navKeeper.NAVs()
			for i, exp := range tc.navs {
				act, err2 := navsCol.Get(ctx, collections.Join(exp.Assets.Denom, exp.Price.Denom))
				if s.Assert().NoError(err2, "[%d]: trying to get %s from the navs collection", i, exp) {
					s.Assert().Equal(*exp, act, "[%d]: nav read from the collection")
				}
			}
		})
	}
}

func (s *TestSuite) TestKeeper_GetNAVRecord() {
	navs := nav.NAVRecords{
		s.newNAVRec("1white", "2purple", 1, "one"),
		s.newNAVRec("1white", "4blue", 1, "one"),
		s.newNAVRec("2white", "6green", 1, "one"),
		s.newNAVRec("2white", "8yellow", 2, "two"),
		s.newNAVRec("3white", "10orange", 3, "three"),
		s.newNAVRec("3white", "12red", 3, "three"),
		s.newNAVRec("8purple", "2gray", 4, "four"),
		s.newNAVRec("8blue", "4gray", 5, "five"),
		s.newNAVRec("8green", "6gray", 5, "five"),
		s.newNAVRec("8yellow", "8gray", 5, "five"),
		s.newNAVRec("8orange", "10gray", 5, "five"),
		s.newNAVRec("8red", "12gray", 5, "five"),
		s.newNAVRec("10brown", "13black", 6, "six"),
		s.newNAVRec("10black", "9brown", 6, "six"),
		s.newNAVRec("12indigo", "5gray", 7, "seven"),
	}

	ctx, _ := s.ctx.CacheContext()
	s.storeNAVs(ctx, navs)

	type testCase struct {
		name  string
		asset string
		price string
		exp   *nav.NetAssetValueRecord
	}

	tests := []testCase{
		{
			name:  "empty denoms",
			asset: "",
			price: "",
			exp:   nil,
		},
		{
			name:  "empty asset denom",
			asset: "",
			price: "gray",
			exp:   nil,
		},
		{
			name:  "empty price denom",
			asset: "white",
			price: "",
			exp:   nil,
		},
		{
			name:  "unknown denoms",
			asset: "silver",
			price: "gold",
			exp:   nil,
		},
		{
			name:  "known asset, unknown price",
			asset: "white",
			price: "gold",
			exp:   nil,
		},
		{
			name:  "unknown asset, known price",
			asset: "silver",
			price: "gray",
			exp:   nil,
		},
		{
			name:  "known denoms in wrong order",
			asset: navs[0].Price.Denom,
			price: navs[0].Assets.Denom,
			exp:   nil,
		},
		{
			name:  "known denoms, but from different navs",
			asset: "brown",
			price: "red",
			exp:   nil,
		},
	}

	for _, navr := range navs {
		tests = append(tests, testCase{
			name:  fmt.Sprintf("known: %s->%s", navr.Assets.Denom, navr.Price.Denom),
			asset: navr.Assets.Denom,
			price: navr.Price.Denom,
			exp:   navr,
		})
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act *nav.NetAssetValueRecord
			testFunc := func() {
				act = s.navKeeper.GetNAVRecord(ctx, tc.asset, tc.price)
			}
			s.Require().NotPanics(testFunc, "GetNAVRecord(ctx, %q, %q)", tc.asset, tc.price)
			s.Assert().Equal(tc.exp, act, "GetNAVRecord(ctx, %q, %q) result", tc.asset, tc.price)
		})
	}
}

func (s *TestSuite) TestKeeper_GetNAVRecords() {
	navs := nav.NAVRecords{
		s.newNAVRec("1white", "2purple", 1, "one"),
		s.newNAVRec("1white", "4blue", 1, "one"),
		s.newNAVRec("2white", "6green", 1, "one"),
		s.newNAVRec("2white", "8yellow", 2, "two"),
		s.newNAVRec("3white", "10orange", 3, "three"),
		s.newNAVRec("3white", "12red", 3, "three"),
		s.newNAVRec("8purple", "2gray", 4, "four"),
		s.newNAVRec("8blue", "4gray", 5, "five"),
		s.newNAVRec("8green", "6gray", 5, "five"),
		s.newNAVRec("8yellow", "8gray", 5, "five"),
		s.newNAVRec("8orange", "10gray", 5, "five"),
		s.newNAVRec("8red", "12gray", 5, "five"),
		s.newNAVRec("10brown", "13black", 6, "six"),
		s.newNAVRec("10black", "9brown", 6, "six"),
		s.newNAVRec("10black", "4green", 6, "six"),
		s.newNAVRec("12indigo", "5gray", 7, "seven"),
	}

	ctx, _ := s.ctx.CacheContext()
	s.storeNAVs(ctx, navs)

	sortedNAVs := make(nav.NAVRecords, len(navs))
	copy(sortedNAVs, navs)
	slices.SortFunc(sortedNAVs, compareNAVRecords)

	tests := []struct {
		name  string
		asset string
		exp   nav.NAVRecords
	}{
		{
			name:  "no results: unknown denom",
			asset: "pink",
			exp:   nil,
		},
		{
			name:  "no results: denom that is only in a price",
			asset: "gray",
			exp:   nil,
		},
		{
			name:  "one result",
			asset: "indigo",
			exp:   nav.NAVRecords{s.newNAVRec("12indigo", "5gray", 7, "seven")},
		},
		{
			name:  "two results",
			asset: "black",
			exp: nav.NAVRecords{
				s.newNAVRec("10black", "9brown", 6, "six"),
				s.newNAVRec("10black", "4green", 6, "six"),
			},
		},
		{
			name:  "six results",
			asset: "white",
			exp: nav.NAVRecords{
				s.newNAVRec("1white", "4blue", 1, "one"),
				s.newNAVRec("2white", "6green", 1, "one"),
				s.newNAVRec("3white", "10orange", 3, "three"),
				s.newNAVRec("1white", "2purple", 1, "one"),
				s.newNAVRec("3white", "12red", 3, "three"),
				s.newNAVRec("2white", "8yellow", 2, "two"),
			},
		},
		{
			name:  "empty asset denom",
			asset: "",
			exp:   sortedNAVs,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act nav.NAVRecords
			testFunc := func() {
				act = s.navKeeper.GetNAVRecords(ctx, tc.asset)
			}
			s.Require().NotPanics(testFunc, "GetNAVRecords(ctx, %q)", tc.asset)
			s.assertEqualNAVRecords(tc.exp, act, "GetNAVRecords(ctx, %q)", tc.asset)
		})
	}
}

func (s *TestSuite) TestKeeper_GetAllNAVRecords() {
	navs := nav.NAVRecords{
		s.newNAVRec("1white", "2purple", 1, "one"),
		s.newNAVRec("1white", "4blue", 1, "one"),
		s.newNAVRec("2white", "6green", 1, "one"),
		s.newNAVRec("2white", "8yellow", 2, "two"),
		s.newNAVRec("3white", "10orange", 3, "three"),
		s.newNAVRec("3white", "12red", 3, "three"),
		s.newNAVRec("8purple", "2gray", 4, "four"),
		s.newNAVRec("8blue", "4gray", 5, "five"),
		s.newNAVRec("8green", "6gray", 5, "five"),
		s.newNAVRec("8yellow", "8gray", 5, "five"),
		s.newNAVRec("8orange", "10gray", 5, "five"),
		s.newNAVRec("8red", "12gray", 5, "five"),
		s.newNAVRec("10brown", "13black", 6, "six"),
		s.newNAVRec("10black", "9brown", 6, "six"),
		s.newNAVRec("10black", "4green", 6, "six"),
		s.newNAVRec("12indigo", "5gray", 7, "seven"),
	}
	sortedNAVs := make(nav.NAVRecords, len(navs))
	copy(sortedNAVs, navs)
	slices.SortFunc(sortedNAVs, compareNAVRecords)

	tests := []struct {
		name string
		ini  nav.NAVRecords
		exp  nav.NAVRecords
	}{
		{
			name: "empty state",
			ini:  nil,
			exp:  nil,
		},
		{
			name: "only one entry",
			ini:  nav.NAVRecords{s.newNAVRec("2gold", "18copper", 33, "threethree")},
			exp:  nav.NAVRecords{s.newNAVRec("2gold", "18copper", 33, "threethree")},
		},
		{
			name: "multiple entries",
			ini:  navs,
			exp:  sortedNAVs,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx, _ := s.ctx.CacheContext()
			s.storeNAVs(ctx, tc.ini)

			var act nav.NAVRecords
			testFunc := func() {
				act = s.navKeeper.GetAllNAVRecords(ctx)
			}
			s.Require().NotPanics(testFunc, "GetAllNAVRecords(ctx)")
			s.assertEqualNAVRecords(tc.exp, act, "GetAllNAVRecords(ctx)")
		})
	}
}
