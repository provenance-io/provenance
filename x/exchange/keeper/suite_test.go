package keeper_test

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

type TestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	k keeper.Keeper

	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	addr3 sdk.AccAddress
	addr4 sdk.AccAddress
	addr5 sdk.AccAddress

	marketAddr1 sdk.AccAddress
	marketAddr2 sdk.AccAddress
	marketAddr3 sdk.AccAddress

	feeCollector     string
	feeCollectorAddr sdk.AccAddress

	accKeeper *MockAccountKeeper

	logBuffer bytes.Buffer

	addrLookupMap map[string]string
}

func (s *TestSuite) SetupTest() {
	if s.addrLookupMap == nil {
		s.addrLookupMap = make(map[string]string)
	}

	// swap in the buffered logger maker so it's used in app.Setup, but then put it back (since that's a global thing).
	defer app.SetLoggerMaker(app.SetLoggerMaker(app.BufferedInfoLoggerMaker(&s.logBuffer)))

	s.app = app.Setup(s.T())
	s.logBuffer.Reset()
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.k = s.app.ExchangeKeeper

	addrs := app.AddTestAddrsIncremental(s.app, s.ctx, 5, sdk.NewInt(1_000_000_000))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
	s.addr4 = addrs[3]
	s.addr5 = addrs[4]
	s.addAddrLookup(s.addr1, "addr1")
	s.addAddrLookup(s.addr2, "addr2")
	s.addAddrLookup(s.addr3, "addr3")
	s.addAddrLookup(s.addr4, "addr4")
	s.addAddrLookup(s.addr5, "addr5")

	s.marketAddr1 = exchange.GetMarketAddress(1)
	s.marketAddr2 = exchange.GetMarketAddress(2)
	s.marketAddr3 = exchange.GetMarketAddress(3)
	s.addAddrLookup(s.marketAddr1, "marketAddr1")
	s.addAddrLookup(s.marketAddr2, "marketAddr2")
	s.addAddrLookup(s.marketAddr3, "marketAddr3")

	s.feeCollector = s.k.GetFeeCollectorName()
	s.feeCollectorAddr = authtypes.NewModuleAddress(s.feeCollector)
	s.addAddrLookup(s.feeCollectorAddr, "feeCollectorAddr")
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// sliceStrings converts each val into a string using the provided stringer, prefixing the slice index to each.
func sliceStrings[T any](vals []T, stringer func(T) string) []string {
	if vals == nil {
		return nil
	}
	strs := make([]string, len(vals))
	for i, v := range vals {
		strs[i] = fmt.Sprintf("[%d]:%s", i, stringer(v))
	}
	return strs
}

// sliceString converts each val into a string using the provided stringer with the index prefixed to it, and joins them with ", ".
func sliceString[T any](vals []T, stringer func(T) string) string {
	if vals == nil {
		return "<nil>"
	}
	return strings.Join(sliceStrings(vals, stringer), ", ")
}

// copySlice returns a copy of a slice using the provided copier for each value.
func copySlice[T any](vals []T, copier func(T) T) []T {
	if vals == nil {
		return nil
	}
	rv := make([]T, len(vals))
	for i, v := range vals {
		rv[i] = copier(v)
	}
	return rv
}

// noOpCopier is a passthrough "copier" function that just returns the exact same thing that was provided.
func noOpCopier[T any](val T) T {
	return val
}

// reverseSlice returns a new slice with the entries reversed.
func reverseSlice[T any](vals []T) []T {
	if vals == nil {
		return nil
	}
	rv := make([]T, len(vals))
	for i, val := range vals {
		rv[len(vals)-i-1] = val
	}
	return rv
}

// getLogOutput gets the log buffer contents. This (probably) also clears the log buffer.
func (s *TestSuite) getLogOutput(msg string, args ...interface{}) string {
	logOutput := s.logBuffer.String()
	s.T().Logf(msg+" log output:\n%s", append(args, logOutput)...)
	return logOutput
}

// splitOutputLog splits the given output log into its lines.
func (s *TestSuite) splitOutputLog(outputLog string) []string {
	if len(outputLog) == 0 {
		return nil
	}
	rv := strings.Split(outputLog, "\n")
	for len(rv) > 0 && len(rv[len(rv)-1]) == 0 {
		rv = rv[:len(rv)-1]
	}
	if len(rv) == 0 {
		return nil
	}
	return rv
}

// badKey creates a copy of the provided key, moves the last byte to the 2nd to last,
// then chops off the last byte (so the result is one byte shorter).
func (s *TestSuite) badKey(key []byte) []byte {
	rv := make([]byte, len(key)-1)
	copy(rv, key)
	rv[len(rv)-1] = key[len(key)-1]
	return rv
}

// coins creates a new sdk.Coins from a string, requiring it to work.
func (s *TestSuite) coins(coins string) sdk.Coins {
	s.T().Helper()
	rv, err := sdk.ParseCoinsNormalized(coins)
	s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
	return rv
}

// coin creates a new coin from a string, requiring it to work.
func (s *TestSuite) coin(coin string) sdk.Coin {
	rv, err := sdk.ParseCoinNormalized(coin)
	s.Require().NoError(err, "ParseCoinNormalized(%q)", coin)
	return rv
}

// coinP creates a reference to a new coin from a string, requiring it to work.
func (s *TestSuite) coinP(coin string) *sdk.Coin {
	rv := s.coin(coin)
	return &rv
}

// coinsString converts a slice of coin entries into a string.
// This is different from sdk.Coins.String because the entries aren't sorted in here.
func (s *TestSuite) coinsString(coins []sdk.Coin) string {
	return sliceString(coins, func(coin sdk.Coin) string {
		return fmt.Sprintf("%q", coin)
	})
}

// coinPString converts the provided coin to a quoted string, or "<nil>".
func (s *TestSuite) coinPString(coin *sdk.Coin) string {
	if coin == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%q", coin)
}

// ratio creates a FeeRatio from a "<price>:<fee>" string.
func (s *TestSuite) ratio(ratioStr string) exchange.FeeRatio {
	rv, err := exchange.ParseFeeRatio(ratioStr)
	s.Require().NoError(err, "ParseFeeRatio(%q)", ratioStr)
	return *rv
}

// ratios creates a slice of Fee ratio from a comma delimited list of "<price>:<fee>" entries in a string.
func (s *TestSuite) ratios(ratiosStr string) []exchange.FeeRatio {
	if len(ratiosStr) == 0 {
		return nil
	}

	ratios := strings.Split(ratiosStr, ",")
	rv := make([]exchange.FeeRatio, len(ratios))
	for i, r := range ratios {
		rv[i] = s.ratio(r)
	}
	return rv
}

// ratiosStrings converts the ratios into strings. It's because comparsions on sdk.Coin (or sdkmath.Int) are annoying.
func (s *TestSuite) ratiosStrings(ratios []exchange.FeeRatio) []string {
	return sliceStrings(ratios, exchange.FeeRatio.String)
}

// joinErrs joins the provided error strings into a single one to match what errors.Join does.
func (s *TestSuite) joinErrs(errs ...string) string {
	return strings.Join(errs, "\n")
}

// copyCoin creates a copy of a coin (as best as possible).
func (s *TestSuite) copyCoin(orig sdk.Coin) sdk.Coin {
	return sdk.NewCoin(orig.Denom, orig.Amount.AddRaw(0))
}

// copyCoinP copies a coin that's a reference.
func (s *TestSuite) copyCoinP(orig *sdk.Coin) *sdk.Coin {
	if orig == nil {
		return nil
	}
	rv := s.copyCoin(*orig)
	return &rv
}

// copyCoins creates a copy of coins (as best as possible).
func (s *TestSuite) copyCoins(orig []sdk.Coin) []sdk.Coin {
	return copySlice(orig, s.copyCoin)
}

// copyRatio creates a copy of a FeeRatio.
func (s *TestSuite) copyRatio(orig exchange.FeeRatio) exchange.FeeRatio {
	return exchange.FeeRatio{
		Price: s.copyCoin(orig.Price),
		Fee:   s.copyCoin(orig.Fee),
	}
}

// copyRatios creates a copy of a slice of FeeRatios.
func (s *TestSuite) copyRatios(orig []exchange.FeeRatio) []exchange.FeeRatio {
	return copySlice(orig, s.copyRatio)
}

// copyAccessGrant creates a copy of an AccessGrant.
func (s *TestSuite) copyAccessGrant(orig exchange.AccessGrant) exchange.AccessGrant {
	return exchange.AccessGrant{
		Address:     orig.Address,
		Permissions: copySlice(orig.Permissions, noOpCopier[exchange.Permission]),
	}
}

// copyAccessGrants creates a copy of a slice of AccessGrants.
func (s *TestSuite) copyAccessGrants(orig []exchange.AccessGrant) []exchange.AccessGrant {
	return copySlice(orig, s.copyAccessGrant)
}

// copyStrings creates a copy of a slice of strings.
func (s *TestSuite) copyStrings(orig []string) []string {
	return copySlice(orig, noOpCopier[string])
}

// copyMarket creates a deep copy of a market.
func (s *TestSuite) copyMarket(orig exchange.Market) exchange.Market {
	return exchange.Market{
		MarketId: orig.MarketId,
		MarketDetails: exchange.MarketDetails{
			Name:        orig.MarketDetails.Name,
			Description: orig.MarketDetails.Description,
			WebsiteUrl:  orig.MarketDetails.WebsiteUrl,
			IconUri:     orig.MarketDetails.IconUri,
		},
		FeeCreateAskFlat:          s.copyCoins(orig.FeeCreateAskFlat),
		FeeCreateBidFlat:          s.copyCoins(orig.FeeCreateBidFlat),
		FeeSellerSettlementFlat:   s.copyCoins(orig.FeeSellerSettlementFlat),
		FeeSellerSettlementRatios: s.copyRatios(orig.FeeSellerSettlementRatios),
		FeeBuyerSettlementFlat:    s.copyCoins(orig.FeeBuyerSettlementFlat),
		FeeBuyerSettlementRatios:  s.copyRatios(orig.FeeBuyerSettlementRatios),
		AcceptingOrders:           orig.AcceptingOrders,
		AllowUserSettlement:       orig.AllowUserSettlement,
		AccessGrants:              s.copyAccessGrants(orig.AccessGrants),
		ReqAttrCreateAsk:          s.copyStrings(orig.ReqAttrCreateAsk),
		ReqAttrCreateBid:          s.copyStrings(orig.ReqAttrCreateBid),
	}
}

// copyMarkets creates a copy of a slice of markets.
func (s *TestSuite) copyMarkets(orig []exchange.Market) []exchange.Market {
	return copySlice(orig, s.copyMarket)
}

// copyOrder creates a copy of an order.
func (s *TestSuite) copyOrder(orig exchange.Order) exchange.Order {
	rv := exchange.NewOrder(orig.OrderId)
	switch {
	case orig.IsAskOrder():
		rv.WithAsk(s.copyAskOrder(orig.GetAskOrder()))
	case orig.IsBidOrder():
		rv.WithBid(s.copyBidOrder(orig.GetBidOrder()))
	default:
		rv.Order = orig.Order
	}
	return *rv
}

// copyOrders creates a copy of a slice of orders.
func (s *TestSuite) copyOrders(orig []exchange.Order) []exchange.Order {
	return copySlice(orig, s.copyOrder)
}

// copyAskOrder creates a copy of an AskOrder.
func (s *TestSuite) copyAskOrder(orig *exchange.AskOrder) *exchange.AskOrder {
	if orig == nil {
		return nil
	}
	return &exchange.AskOrder{
		MarketId:                orig.MarketId,
		Seller:                  orig.Seller,
		Assets:                  s.copyCoin(orig.Assets),
		Price:                   s.copyCoin(orig.Price),
		SellerSettlementFlatFee: s.copyCoinP(orig.SellerSettlementFlatFee),
		AllowPartial:            orig.AllowPartial,
		ExternalId:              orig.ExternalId,
	}
}

// copyBidOrder creates a copy of a BidOrder.
func (s *TestSuite) copyBidOrder(orig *exchange.BidOrder) *exchange.BidOrder {
	if orig == nil {
		return nil
	}
	return &exchange.BidOrder{
		MarketId:            orig.MarketId,
		Buyer:               orig.Buyer,
		Assets:              s.copyCoin(orig.Assets),
		Price:               s.copyCoin(orig.Price),
		BuyerSettlementFees: s.copyCoins(orig.BuyerSettlementFees),
		AllowPartial:        orig.AllowPartial,
		ExternalId:          orig.ExternalId,
	}
}

// untypeEvent applies sdk.TypedEventToEvent(tev) requiring it to not error.
func (s *TestSuite) untypeEvent(tev proto.Message) sdk.Event {
	rv, err := sdk.TypedEventToEvent(tev)
	s.Require().NoError(err, "TypedEventToEvent(%T)", tev)
	return rv
}

// untypeEvents applies sdk.TypedEventToEvent(tev) to each of the provided things, requiring it to not error.
func untypeEvents[P proto.Message](s *TestSuite, tevs []P) sdk.Events {
	rv := make(sdk.Events, len(tevs))
	for i, tev := range tevs {
		event, err := sdk.TypedEventToEvent(tev)
		s.Require().NoError(err, "[%d]TypedEventToEvent(%T)", i, tev)
		rv[i] = event
	}
	return rv
}

// creates a copy of a DenomSplit.
func (s *TestSuite) copyDenomSplit(orig exchange.DenomSplit) exchange.DenomSplit {
	return exchange.DenomSplit{
		Denom: orig.Denom,
		Split: orig.Split,
	}
}

// copyDenomSplits creates a copy of a slice of DenomSplits.
func (s *TestSuite) copyDenomSplits(orig []exchange.DenomSplit) []exchange.DenomSplit {
	return copySlice(orig, s.copyDenomSplit)
}

// copyParams creates a copy of exchange Params.
func (s *TestSuite) copyParams(orig *exchange.Params) *exchange.Params {
	if orig == nil {
		return nil
	}
	return &exchange.Params{
		DefaultSplit: orig.DefaultSplit,
		DenomSplits:  s.copyDenomSplits(orig.DenomSplits),
	}
}

// copyGenState creates a copy of a GenesisState.
func (s *TestSuite) copyGenState(genState *exchange.GenesisState) *exchange.GenesisState {
	if genState == nil {
		return nil
	}
	return &exchange.GenesisState{
		Params:       s.copyParams(genState.Params),
		Markets:      s.copyMarkets(genState.Markets),
		Orders:       s.copyOrders(genState.Orders),
		LastMarketId: genState.LastMarketId,
		LastOrderId:  genState.LastOrderId,
	}
}

// sortMarket sorts all the fields in a market.
func (s *TestSuite) sortMarket(market *exchange.Market) *exchange.Market {
	if len(market.FeeSellerSettlementRatios) > 0 {
		sort.Slice(market.FeeSellerSettlementRatios, func(i, j int) bool {
			if market.FeeSellerSettlementRatios[i].Price.Denom < market.FeeSellerSettlementRatios[j].Price.Denom {
				return true
			}
			if market.FeeSellerSettlementRatios[i].Price.Denom > market.FeeSellerSettlementRatios[j].Price.Denom {
				return false
			}
			return market.FeeSellerSettlementRatios[i].Fee.Denom < market.FeeSellerSettlementRatios[j].Fee.Denom
		})
	}
	if len(market.FeeBuyerSettlementRatios) > 0 {
		sort.Slice(market.FeeBuyerSettlementRatios, func(i, j int) bool {
			if market.FeeBuyerSettlementRatios[i].Price.Denom < market.FeeBuyerSettlementRatios[j].Price.Denom {
				return true
			}
			if market.FeeBuyerSettlementRatios[i].Price.Denom > market.FeeBuyerSettlementRatios[j].Price.Denom {
				return false
			}
			return market.FeeBuyerSettlementRatios[i].Fee.Denom < market.FeeBuyerSettlementRatios[j].Fee.Denom
		})
	}
	if len(market.AccessGrants) > 0 {
		sort.Slice(market.AccessGrants, func(i, j int) bool {
			// Horribly inefficient. Not meant for production.
			addrI, err := sdk.AccAddressFromBech32(market.AccessGrants[i].Address)
			s.Require().NoError(err, "AccAddressFromBech32(%q)", market.AccessGrants[i].Address)
			addrJ, err := sdk.AccAddressFromBech32(market.AccessGrants[j].Address)
			s.Require().NoError(err, "AccAddressFromBech32(%q)", market.AccessGrants[j].Address)
			return bytes.Compare(addrI, addrJ) < 0
		})
		for _, ag := range market.AccessGrants {
			sort.Slice(ag.Permissions, func(i, j int) bool {
				return ag.Permissions[i] < ag.Permissions[j]
			})
		}
	}
	return market
}

// sortGenState sorts the contents of a GenesisState.
func (s *TestSuite) sortGenState(genState *exchange.GenesisState) *exchange.GenesisState {
	if genState == nil {
		return nil
	}
	if genState.Params != nil && len(genState.Params.DenomSplits) > 0 {
		sort.Slice(genState.Params.DenomSplits, func(i, j int) bool {
			return genState.Params.DenomSplits[i].Denom < genState.Params.DenomSplits[j].Denom
		})
	}
	if len(genState.Markets) > 0 {
		sort.Slice(genState.Markets, func(i, j int) bool {
			return genState.Markets[i].MarketId < genState.Markets[j].MarketId
		})
		for _, market := range genState.Markets {
			s.sortMarket(&market)
		}
	}
	if len(genState.Orders) > 0 {
		sort.Slice(genState.Orders, func(i, j int) bool {
			return genState.Orders[i].OrderId < genState.Orders[j].OrderId
		})
	}
	return genState
}

// getOrderIDStr gets a string of the given order's id.
func (s *TestSuite) getOrderIDStr(order *exchange.Order) string {
	if order == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", order.OrderId)
}

// agCanOnly creates an AccessGrant for the given address with only the provided permission.
func (s *TestSuite) agCanOnly(addr sdk.AccAddress, perm exchange.Permission) exchange.AccessGrant {
	return exchange.AccessGrant{
		Address:     addr.String(),
		Permissions: []exchange.Permission{perm},
	}
}

// agCanAllBut creates an AccessGrant for the given address with all permissions except the provided one.
func (s *TestSuite) agCanAllBut(addr sdk.AccAddress, perm exchange.Permission) exchange.AccessGrant {
	rv := exchange.AccessGrant{
		Address: addr.String(),
	}
	for _, p := range exchange.AllPermissions() {
		if p != perm {
			rv.Permissions = append(rv.Permissions, p)
		}
	}
	return rv
}

// agCanEverything creates an AccessGrant for the given address with all permissions available.
func (s *TestSuite) agCanEverything(addr sdk.AccAddress) exchange.AccessGrant {
	return exchange.AccessGrant{
		Address:     addr.String(),
		Permissions: exchange.AllPermissions(),
	}
}

// addAddrLookup adds an entry to the addrLookupMap (for use in getAddrName).
func (s *TestSuite) addAddrLookup(addr sdk.AccAddress, name string) {
	s.addrLookupMap[string(addr)] = name
}

// getAddrName returns the name of the variable in this TestSuite holding the provided address.
func (s *TestSuite) getAddrName(addr sdk.AccAddress) string {
	if s.addrLookupMap != nil {
		rv, found := s.addrLookupMap[string(addr)]
		if found {
			return rv
		}
	}
	return addr.String()
}

// getAddrStrName returns the name of the variable in this TestSuite holding the provided address.
func (s *TestSuite) getAddrStrName(addrStr string) string {
	addr, err := sdk.AccAddressFromBech32(addrStr)
	if err != nil {
		return addrStr
	}
	return s.getAddrName(addr)
}

// getStore gets the exchange store.
func (s *TestSuite) getStore() sdk.KVStore {
	return s.k.GetStore(s.ctx)
}

// clearExchangeState deletes everything from the exchange state store.
func (s *TestSuite) clearExchangeState() {
	keeper.DeleteAll(s.getStore(), nil)
	s.accKeeper = nil
}

// stateEntryString converts the provided key and value into a "<key>"="<value>" string.
func (s *TestSuite) stateEntryString(key, value []byte) string {
	return fmt.Sprintf("%q=%q", key, value)
}

// dumpExchangeState creates a string for each entry in the hold state store.
// Each entry has the format `"<key>"="<value>"`.
func (s *TestSuite) dumpExchangeState() []string {
	var rv []string
	keeper.Iterate(s.getStore(), nil, func(key, value []byte) bool {
		rv = append(rv, s.stateEntryString(key, value))
		return false
	})
	return rv
}

// requireSetOrderInStore calls SetOrderInStore making sure it doesn't panic or return an error.
func (s *TestSuite) requireSetOrderInStore(store sdk.KVStore, order *exchange.Order) {
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		return s.k.SetOrderInStore(store, *order)
	}, "SetOrderInStore(%d)", order.OrderId)
}

// requireCreateMarket calls CreateMarket making sure it doesn't panic or return an error.
// It also uses the TestSuite.accKeeper for the market account.
func (s *TestSuite) requireCreateMarket(market exchange.Market) {
	if s.accKeeper == nil {
		s.accKeeper = NewMockAccountKeeper()
	}
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		_, err := s.k.WithAccountKeeper(s.accKeeper).CreateMarket(s.ctx, market)
		return err
	}, "CreateMarket(%d)", market.MarketId)
}

// requireCreateMarketUnmocked calls CreateMarket making sure it doesn't panic or return an error.
// This uses the normal account keeper (instead of a mocked one).
func (s *TestSuite) requireCreateMarketUnmocked(market exchange.Market) {
	assertions.RequireNotPanicsNoErrorf(s.T(), func() error {
		_, err := s.k.CreateMarket(s.ctx, market)
		return err
	}, "CreateMarket(%d)", market.MarketId)
}

// assertEqualSlice asserts that expected = actual and returns true if so.
// If not, returns false and the stringer is applied to each entry and the comparison
// is redone on the strings in the hopes that it helps identify the problem.
// If the strings are also equal, each individual entry is compared.
func assertEqualSlice[T any](s *TestSuite, expected, actual []T, stringer func(T) string, msg string, args ...interface{}) bool {
	s.T().Helper()
	if s.Assert().Equalf(expected, actual, msg, args...) {
		return true
	}
	// compare each as strings in the hopes that makes it easier to identify the problem.
	expStrs := sliceStrings(expected, stringer)
	actStrs := sliceStrings(actual, stringer)
	if !s.Assert().Equalf(expStrs, actStrs, "strings: "+msg, args...) {
		return false
	}
	// They're the same as strings, so compare each individually.
	for i := range expected {
		s.Assert().Equalf(expected[i], actual[i], msg+fmt.Sprintf("[%d]", i), args...)
	}
	return false
}

// assertEqualOrderID asserts that two uint64 values are equal, and if not, includes their decimal form in the log.
// This is nice because .Equal failures output uints in hex, which can make it difficult to identify what's going on.
func (s *TestSuite) assertEqualOrderID(expected, actual uint64, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	if s.Assert().Equal(expected, actual, msgAndArgs...) {
		return true
	}
	s.T().Logf("Expected order id: %d", expected)
	s.T().Logf("  Actual order id: %d", actual)
	return false
}

// assertEqualOrders asserts that the slices of orders are equal.
// If not, some further assertions are made to try to help try to clarify the differences.
func (s *TestSuite) assertEqualOrders(expected, actual []*exchange.Order, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, actual, s.getOrderIDStr, msg, args...)
}

// assertErrorValue is a wrapper for assertions.AssertErrorValue for this TestSuite.
func (s *TestSuite) assertErrorValue(theError error, expected string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertErrorValue(s.T(), theError, expected, msgAndArgs...)
}

// assertErrorContentsf is a wrapper for assertions.AssertErrorContentsf for this TestSuite.
func (s *TestSuite) assertErrorContentsf(theError error, contains []string, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertErrorContentsf(s.T(), theError, contains, msg, args...)
}

// assertEqualEvents is a wrapper for assertions.AssertEqualEvents for this TestSuite.
func (s *TestSuite) assertEqualEvents(expected, actual sdk.Events, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertEqualEvents(s.T(), expected, actual, msgAndArgs...)
}

// requirePanicEquals is a wrapper for assertions.RequirePanicEquals for this TestSuite.
func (s *TestSuite) requirePanicEquals(f assertions.PanicTestFunc, expected string, msgAndArgs ...interface{}) {
	s.T().Helper()
	assertions.RequirePanicEquals(s.T(), f, expected, msgAndArgs...)
}

// markerAddr gets the address of a marker account for the given denom.
func (s *TestSuite) markerAddr(denom string) sdk.AccAddress {
	markerAddr, err := markertypes.MarkerAddress(denom)
	s.Require().NoError(err, "MarkerAddress(%q)", denom)
	s.addAddrLookup(markerAddr, denom+"MarkerAddr")
	return markerAddr
}

// markerAccount returns a new marker account with the given supply.
func (s *TestSuite) markerAccount(supplyCoinStr string) markertypes.MarkerAccountI {
	supply := s.coin(supplyCoinStr)
	return &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{Address: s.markerAddr(supply.Denom).String()},
		Status:      markertypes.StatusActive,
		Denom:       supply.Denom,
		Supply:      supply.Amount,
		MarkerType:  markertypes.MarkerType_RestrictedCoin,
		SupplyFixed: true,
	}
}

// navSetEvent returns a new EventSetNetAssetValue converted to sdk.Event.
func (s *TestSuite) navSetEvent(assetsStr, priceStr string, marketID uint32) sdk.Event {
	assets := s.coin(assetsStr)
	event := &markertypes.EventSetNetAssetValue{
		Denom:  assets.Denom,
		Price:  priceStr,
		Volume: assets.Amount.String(),
		Source: fmt.Sprintf("x/exchange market %d", marketID),
	}
	return s.untypeEvent(event)
}
