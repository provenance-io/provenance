package keeper_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func TestBankTestSuite(t *testing.T) {
	suite.Run(t, new(BankTestSuite))
}

type BankTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
	bk  *keeper.MDBankKeeper

	logBuffer bytes.Buffer
}

func (s *BankTestSuite) SetupTest() {
	// Swap in the buffered logger maker so it's used in app.Setup, but then put it back (since that's a global thing).
	defer app.SetLoggerMaker(app.SetLoggerMaker(app.BufferedInfoLoggerMaker(&s.logBuffer)))
	s.app = app.Setup(s.T())
	s.logBuffer.Reset()
	s.ctx = s.app.NewContext(false)
	s.bk = keeper.NewMDBankKeeper(s.app.BankKeeper)
}

// getLogOutput gets the log buffer contents and logs that to the test.
// The returned value is the contents split on newline with empty lines removed from the end.
// This (probably) also clears out the log buffer.
func (s *BankTestSuite) getLogOutput(msg string, args ...interface{}) []string {
	logOutput := s.logBuffer.String()
	if len(strings.TrimSpace(logOutput)) == 0 {
		s.T().Logf(msg+" log output: <none>", args...)
		return nil
	}
	s.T().Logf(msg+" log output:\n%s", append(args, logOutput)...)

	rv := strings.Split(logOutput, "\n")
	for len(rv) > 0 && len(rv[len(rv)-1]) == 0 {
		rv = rv[:len(rv)-1]
	}
	if len(rv) == 0 {
		return nil
	}
	return rv
}

func (s *BankTestSuite) AssertErrorValue(theError error, errorString string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return assertions.AssertErrorValue(s.T(), theError, errorString, msgAndArgs...)
}

type balance struct {
	addr  sdk.AccAddress
	denom string
	amt   int64 // Defaults to 1 if not provided.
}

// setBalances adds the provided balances data to the bank keeper's Balances collection.
func (s *BankTestSuite) setBalances(ctx sdk.Context, balances []balance) {
	for _, bal := range balances {
		amt := sdkmath.OneInt()
		if bal.amt > 0 {
			amt = sdkmath.NewInt(bal.amt)
		}
		s.Require().NoError(s.bk.Balances.Set(ctx, collections.Join(bal.addr, bal.denom), amt),
			"s.bk.Balances.Set(ctx, collections.Join(%q, %q), %s)",
			bal.addr.String(), bal.denom, amt)
	}
}

// parseUUID parses the provided id into a UUID, requiring it to be successful.
func (s *BankTestSuite) parseUUID(uid string) uuid.UUID {
	rv, err := uuid.Parse(uid)
	s.Require().NoError(err, "uuid.Parse(%q)", uid)
	return rv
}

// scopeID creates a scope id from the provided uuid string.
func (s *BankTestSuite) scopeID(uid string) types.MetadataAddress {
	id := s.parseUUID(uid)
	return types.ScopeMetadataAddress(id)
}

// scopeSpecID creates a scope spec id from the provided uuid string.
func (s *BankTestSuite) scopeSpecID(uid string) types.MetadataAddress {
	id := s.parseUUID(uid)
	return types.ScopeSpecMetadataAddress(id)
}

func (s *BankTestSuite) TestDenomOwner() {
	addr1 := sdk.AccAddress("1_addr______________") // cosmos1x90kzerywf047h6lta047h6lta047h6l258ny6
	addr2 := sdk.AccAddress("2_addr______________") // cosmos1xf0kzerywf047h6lta047h6lta047h6lgww49l
	addr3 := sdk.AccAddress("3_addr______________") // cosmos1xd0kzerywf047h6lta047h6lta047h6l3lfhau
	addr4 := sdk.AccAddress("4_addr______________") // cosmos1x30kzerywf047h6lta047h6lta047h6lvnue84
	logNamedValues(s.T(), "addresses", []namedValue{
		{name: "addr1", value: addr1.String()},
		{name: "addr2", value: addr2.String()},
		{name: "addr3", value: addr3.String()},
		{name: "addr4", value: addr4.String()},
	})

	// subOne reduces the last character by 1 (ignoring overflow).
	subOne := func(val string) string {
		return val[:len(val)-1] + string(val[len(val)-1]-1)
	}
	// addOne increases the last character by 1 (ignoring overflow).
	addOne := func(val string) string {
		return val[:len(val)-1] + string(val[len(val)-1]+1)
	}

	scopeID := s.scopeID("69012AF4-2FA4-44DA-BAE4-1C13480362C9") // scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghw
	scopeDenom := scopeID.Denom()                                // nft/scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghw
	scopeDenomBefore := subOne(scopeDenom)                       // nft/scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghv
	scopeDenomAfter := addOne(scopeDenom)                        // nft/scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghx
	logNamedValues(s.T(), "ids and denoms", []namedValue{
		{name: "scopeID", value: scopeID.String()},
		{name: "scopeDenom", value: scopeDenom},
		{name: "scopeDenomBefore", value: scopeDenomBefore},
		{name: "scopeDenomAfter", value: scopeDenomAfter},
	})

	tests := []struct {
		name     string
		balances []balance
		denom    string
		expAddr  sdk.AccAddress
		expErr   string
	}{
		{
			name: "no owner",
			balances: []balance{
				{addr: addr1, denom: scopeDenomBefore},
				{addr: addr3, denom: scopeDenomAfter},
			},
			denom:   scopeDenom,
			expAddr: nil,
			expErr:  "",
		},
		{
			name: "one owner",
			balances: []balance{
				{addr: addr1, denom: scopeDenomBefore},
				{addr: addr2, denom: scopeDenom},
				{addr: addr3, denom: scopeDenomAfter},
			},
			denom:   scopeDenom,
			expAddr: addr2,
			expErr:  "",
		},
		{
			name: "two owners",
			balances: []balance{
				{addr: addr1, denom: scopeDenomBefore},
				{addr: addr2, denom: scopeDenom},
				{addr: addr3, denom: scopeDenom},
				{addr: addr4, denom: scopeDenomAfter},
			},
			denom:   scopeDenom,
			expAddr: nil,
			expErr:  "denom \"" + scopeDenom + "\" has more than one owner",
		},
		{
			name: "three owners",
			balances: []balance{
				{addr: addr1, denom: scopeDenom},
				{addr: addr2, denom: scopeDenom},
				{addr: addr3, denom: scopeDenom},
			},
			denom:   scopeDenom,
			expAddr: nil,
			expErr:  "denom \"" + scopeDenom + "\" has more than one owner",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Use a cache context for each test so that the setup doesn't persist between tests.
			ctx, _ := s.ctx.CacheContext()
			s.setBalances(ctx, tc.balances)

			var addr sdk.AccAddress
			var err error
			testFunc := func() {
				addr, err = s.bk.DenomOwner(ctx, tc.denom)
			}
			s.Require().NotPanics(testFunc, "DenomOwner(%q)", tc.denom)
			s.AssertErrorValue(err, tc.expErr, "error returned by DenomOwner(%q)", tc.denom)
			s.Assert().Equal(tc.expAddr, addr, "AccAddress returned by DenomOwner(%q)", tc.denom)
		})
	}
}

func (s *BankTestSuite) TestGetScopesForValueOwner() {
	addr1 := sdk.AccAddress("1_addr______________") // cosmos1x90kzerywf047h6lta047h6lta047h6l258ny6
	addr2 := sdk.AccAddress("2_addr______________") // cosmos1xf0kzerywf047h6lta047h6lta047h6lgww49l
	addr3 := sdk.AccAddress("3_addr______________") // cosmos1xd0kzerywf047h6lta047h6lta047h6l3lfhau
	logNamedValues(s.T(), "addresses", []namedValue{
		{name: "addr1", value: addr1.String()},
		{name: "addr2", value: addr2.String()},
		{name: "addr3", value: addr3.String()},
	})

	scopeID1 := s.scopeID("4CDFD0C4-F08C-403E-A8F7-EC723E7A0001") // scope1qpxdl5xy7zxyq04g7lk8y0n6qqqspk95yz
	scopeID2 := s.scopeID("4CDFD0C4-F08C-403E-A8F7-EC723E7A0002") // scope1qpxdl5xy7zxyq04g7lk8y0n6qqpqm2zk6h
	scopeID3 := s.scopeID("4CDFD0C4-F08C-403E-A8F7-EC723E7A0003") // scope1qpxdl5xy7zxyq04g7lk8y0n6qqpswt2w0y
	scopeID4 := s.scopeID("4CDFD0C4-F08C-403E-A8F7-EC723E7A0004") // scope1qpxdl5xy7zxyq04g7lk8y0n6qqzq2yn38a
	scopeID5 := s.scopeID("4CDFD0C4-F08C-403E-A8F7-EC723E7A0005") // scope1qpxdl5xy7zxyq04g7lk8y0n6qqzsl9mfjw
	// Note that when sorted by denom, they have this order: scopeID2, scopeID3, scopeID1, scopeID4, scopeID5.
	// I'm including tests involving a scope spec denom because scope denoms start with "nft/scope" and scope spec
	// denoms would start with "nft/scopespec". The prefix being used should include the "1" that separates the HRP
	// and bytes in a bech32 address string. But if the prefix does not have that "1", a scope spec entry would
	// end up being included in the results (which we don't want).
	scopeSpecID := s.scopeSpecID("4CDFD0C4-F08C-403E-A8F7-EC723E7A0001") // scopespec1q3xdl5xy7zxyq04g7lk8y0n6qqqs0su6rh
	logNamedValues(s.T(), "addresses", []namedValue{
		{name: "scopeID1", value: scopeID1.String()},
		{name: "scopeID2", value: scopeID2.String()},
		{name: "scopeID3", value: scopeID3.String()},
		{name: "scopeID4", value: scopeID4.String()},
		{name: "scopeID5", value: scopeID5.String()},
		{name: "scopeSpecID", value: scopeSpecID.String()},
	})

	logMsg := func(owner sdk.AccAddress, denom, err string) string {
		return "ERR invalid metadata balance entry for account \"" + owner.String() + "\": " +
			"invalid metadata address in denom \"" + denom + "\": " + err + " module=x/bank"
	}
	badChecksumErr := func(expected, actual string) string {
		return "decoding bech32 failed: invalid checksum (expected " + expected + " got " + actual + ")"
	}
	nextKey := func(nextScopeID types.MetadataAddress) []byte {
		// The prefix used for iteration is "nft/scope1".
		// So the NextKey will be the bytes and checksum portion of the scope id as a bech32 string.
		return []byte(strings.TrimPrefix(nextScopeID.String(), "scope1"))
	}

	tests := []struct {
		name        string
		balances    []balance
		valueOwner  sdk.AccAddress
		pageReq     *query.PageRequest
		expLinks    types.AccMDLinks
		expPageResp *query.PageResponse
		expErr      string
		expLogs     []string
	}{
		{
			name: "unpaginated: no scopes",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr3, denom: scopeID3.Denom()},
			},
			valueOwner: addr2,
			expLinks:   nil,
		},
		{
			name: "unpaginated: one scope",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr2, denom: scopeID2.Denom()},
				{addr: addr3, denom: scopeID3.Denom()},
			},
			valueOwner: addr2,
			expLinks:   types.AccMDLinks{{addr2, scopeID2}},
		},
		{
			name: "unpaginated: three scopes",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr2, denom: scopeID2.Denom()},
				{addr: addr2, denom: scopeID3.Denom()},
				{addr: addr2, denom: scopeID4.Denom()},
				{addr: addr3, denom: scopeID5.Denom()},
			},
			valueOwner: addr2,
			expLinks:   types.AccMDLinks{{addr2, scopeID2}, {addr2, scopeID3}, {addr2, scopeID4}},
		},
		{
			name: "unpaginated: four entries, three good, one bad",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr2, denom: scopeID2.Denom()},
				{addr: addr2, denom: scopeID3.Denom()},
				{addr: addr2, denom: scopeID3.Denom() + "x"},
				{addr: addr2, denom: scopeID4.Denom()},
				{addr: addr3, denom: scopeID5.Denom()},
			},
			valueOwner: addr2,
			expLinks:   types.AccMDLinks{{addr2, scopeID2}, {addr2, scopeID3}, {addr2, nil}, {addr2, scopeID4}},
			expLogs:    []string{logMsg(addr2, scopeID3.Denom()+"x", badChecksumErr("t2w09p", "t2w0yx"))},
		},
		{
			name: "unpaginated: four entries, all bad",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom() + "w"},
				{addr: addr1, denom: scopeID2.Denom() + "x"},
				{addr: addr1, denom: scopeID3.Denom() + "y"},
				{addr: addr1, denom: scopeID4.Denom() + "z"},
			},
			valueOwner: addr1,
			expLinks:   types.AccMDLinks{{addr1, nil}, {addr1, nil}, {addr1, nil}, {addr1, nil}},
			expLogs: []string{
				logMsg(addr1, scopeID2.Denom()+"x", badChecksumErr("2zk6kp", "2zk6hx")),
				logMsg(addr1, scopeID3.Denom()+"y", badChecksumErr("t2w09p", "t2w0yy")),
				logMsg(addr1, scopeID1.Denom()+"w", badChecksumErr("k95yrp", "k95yzw")),
				logMsg(addr1, scopeID4.Denom()+"z", badChecksumErr("yn38up", "yn38az")),
			},
		},
		{
			name:       "unpaginated: scope spec denom ignored",
			balances:   []balance{{addr: addr1, denom: scopeSpecID.Denom()}},
			valueOwner: addr1,
			expLinks:   nil,
		},
		{
			name: "unpaginated: scope spec denom ignored with scope results",
			balances: []balance{
				{addr: addr1, denom: scopeSpecID.Denom()},
				{addr: addr1, denom: scopeID3.Denom()},
				{addr: addr1, denom: scopeID5.Denom()},
			},
			valueOwner: addr1,
			expLinks:   types.AccMDLinks{{addr1, scopeID3}, {addr1, scopeID5}},
		},
		{
			name: "paginated: no scopes",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr3, denom: scopeID3.Denom()},
			},
			valueOwner:  addr2,
			pageReq:     &query.PageRequest{Limit: 50},
			expLinks:    nil,
			expPageResp: &query.PageResponse{},
		},
		{
			name: "paginated: one scope",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr2, denom: scopeID2.Denom()},
				{addr: addr3, denom: scopeID3.Denom()},
			},
			valueOwner:  addr2,
			pageReq:     &query.PageRequest{Limit: 50, CountTotal: true},
			expLinks:    types.AccMDLinks{{addr2, scopeID2}},
			expPageResp: &query.PageResponse{Total: 1},
		},
		{
			name: "paginated: three scopes, get all",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr2, denom: scopeID2.Denom()},
				{addr: addr2, denom: scopeID3.Denom()},
				{addr: addr2, denom: scopeID4.Denom()},
				{addr: addr3, denom: scopeID5.Denom()},
			},
			valueOwner:  addr2,
			pageReq:     &query.PageRequest{Limit: 3},
			expLinks:    types.AccMDLinks{{addr2, scopeID2}, {addr2, scopeID3}, {addr2, scopeID4}},
			expPageResp: &query.PageResponse{},
		},
		{
			name: "paginated: three scopes, get all reversed",
			balances: []balance{
				{addr: addr3, denom: scopeID1.Denom()},
				{addr: addr1, denom: scopeID2.Denom()},
				{addr: addr1, denom: scopeID3.Denom()},
				{addr: addr1, denom: scopeID4.Denom()},
				{addr: addr2, denom: scopeID5.Denom()},
			},
			valueOwner:  addr1,
			pageReq:     &query.PageRequest{Limit: 3, Reverse: true},
			expLinks:    types.AccMDLinks{{addr1, scopeID4}, {addr1, scopeID3}, {addr1, scopeID2}},
			expPageResp: &query.PageResponse{},
		},
		{
			name: "paginated: three scopes, get just first",
			balances: []balance{
				{addr: addr2, denom: scopeID1.Denom()},
				{addr: addr3, denom: scopeID2.Denom()},
				{addr: addr3, denom: scopeID3.Denom()},
				{addr: addr3, denom: scopeID4.Denom()},
				{addr: addr1, denom: scopeID5.Denom()},
			},
			valueOwner:  addr3,
			pageReq:     &query.PageRequest{Limit: 1},
			expLinks:    types.AccMDLinks{{addr3, scopeID2}},
			expPageResp: &query.PageResponse{NextKey: nextKey(scopeID3)},
		},
		{
			name: "paginated: three scopes, get just second using offset",
			balances: []balance{
				{addr: addr1, denom: scopeID1.Denom()},
				{addr: addr3, denom: scopeID2.Denom()},
				{addr: addr3, denom: scopeID3.Denom()},
				{addr: addr3, denom: scopeID4.Denom()},
				{addr: addr2, denom: scopeID5.Denom()},
			},
			valueOwner:  addr3,
			pageReq:     &query.PageRequest{Limit: 1, Offset: 1},
			expLinks:    types.AccMDLinks{{addr3, scopeID3}},
			expPageResp: &query.PageResponse{NextKey: nextKey(scopeID4)},
		},
		{
			name: "paginated: three scopes, get second using next key",
			balances: []balance{
				{addr: addr3, denom: scopeID1.Denom()},
				{addr: addr2, denom: scopeID2.Denom()},
				{addr: addr2, denom: scopeID3.Denom()},
				{addr: addr2, denom: scopeID4.Denom()},
				{addr: addr1, denom: scopeID5.Denom()},
			},
			valueOwner:  addr2,
			pageReq:     &query.PageRequest{Limit: 1, Key: nextKey(scopeID3)},
			expLinks:    types.AccMDLinks{{addr2, scopeID3}},
			expPageResp: &query.PageResponse{NextKey: nextKey(scopeID4)},
		},
		{
			name: "paginated: three scopes, get last by reversing with limit 1",
			balances: []balance{
				{addr: addr2, denom: scopeID1.Denom()},
				{addr: addr1, denom: scopeID2.Denom()},
				{addr: addr1, denom: scopeID3.Denom()},
				{addr: addr1, denom: scopeID4.Denom()},
				{addr: addr3, denom: scopeID5.Denom()},
			},
			valueOwner:  addr1,
			pageReq:     &query.PageRequest{Limit: 1, Reverse: true},
			expLinks:    types.AccMDLinks{{addr1, scopeID4}},
			expPageResp: &query.PageResponse{NextKey: nextKey(scopeID3)},
		},
		{
			name: "paginated: four entries, three good, one bad",
			balances: []balance{
				{addr: addr1, denom: scopeID2.Denom()},
				{addr: addr1, denom: scopeID3.Denom() + "x"},
				{addr: addr1, denom: scopeID4.Denom()},
				{addr: addr1, denom: scopeID5.Denom()},
			},
			valueOwner:  addr1,
			pageReq:     &query.PageRequest{Limit: 4},
			expLinks:    types.AccMDLinks{{addr1, scopeID2}, {addr1, nil}, {addr1, scopeID4}, {addr1, scopeID5}},
			expPageResp: &query.PageResponse{},
			expLogs: []string{
				logMsg(addr1, scopeID3.Denom()+"x", badChecksumErr("t2w09p", "t2w0yx")),
			},
		},
		{
			name: "paginated: four entries, all bad",
			balances: []balance{
				{addr: addr1, denom: scopeID2.Denom() + "w"},
				{addr: addr1, denom: scopeID3.Denom() + "x"},
				{addr: addr1, denom: scopeID4.Denom() + "y"},
				{addr: addr1, denom: scopeID5.Denom() + "z"},
			},
			valueOwner:  addr1,
			pageReq:     &query.PageRequest{Limit: 4},
			expLinks:    types.AccMDLinks{{addr1, nil}, {addr1, nil}, {addr1, nil}, {addr1, nil}},
			expPageResp: &query.PageResponse{},
			expLogs: []string{
				logMsg(addr1, scopeID2.Denom()+"w", badChecksumErr("2zk6kp", "2zk6hw")),
				logMsg(addr1, scopeID3.Denom()+"x", badChecksumErr("t2w09p", "t2w0yx")),
				logMsg(addr1, scopeID4.Denom()+"y", badChecksumErr("yn38up", "yn38ay")),
				logMsg(addr1, scopeID5.Denom()+"z", badChecksumErr("9mfj0p", "9mfjwz")),
			},
		},
		{
			name: "paginated: four entries, all bad, get middle two",
			balances: []balance{
				{addr: addr1, denom: scopeID2.Denom() + "w"},
				{addr: addr1, denom: scopeID3.Denom() + "x"},
				{addr: addr1, denom: scopeID4.Denom() + "y"},
				{addr: addr1, denom: scopeID5.Denom() + "z"},
			},
			valueOwner:  addr1,
			pageReq:     &query.PageRequest{Limit: 2, Offset: 1, CountTotal: true},
			expLinks:    types.AccMDLinks{{addr1, nil}, {addr1, nil}},
			expPageResp: &query.PageResponse{Total: 4, NextKey: append(nextKey(scopeID5), 'z')},
			expLogs: []string{
				logMsg(addr1, scopeID3.Denom()+"x", badChecksumErr("t2w09p", "t2w0yx")),
				logMsg(addr1, scopeID4.Denom()+"y", badChecksumErr("yn38up", "yn38ay")),
			},
		},
		{
			name: "paginated: scope spec denom ignored",
			balances: []balance{
				{addr: addr1, denom: scopeSpecID.Denom()},
			},
			valueOwner:  addr1,
			pageReq:     &query.PageRequest{Limit: 50},
			expLinks:    nil,
			expPageResp: &query.PageResponse{},
		},
		{
			name: "paginated: scope spec denom ignored with scope results",
			balances: []balance{
				{addr: addr1, denom: scopeSpecID.Denom()},
				{addr: addr1, denom: scopeID2.Denom()},
				{addr: addr1, denom: scopeID4.Denom()},
			},
			valueOwner:  addr1,
			pageReq:     &query.PageRequest{Limit: 50},
			expLinks:    types.AccMDLinks{{addr1, scopeID2}, {addr1, scopeID4}},
			expPageResp: &query.PageResponse{},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Use a cache context for each test so that the setup doesn't persist between tests.
			ctx, _ := s.ctx.CacheContext()
			s.setBalances(ctx, tc.balances)

			callDesc := fmt.Sprintf("GetScopesForValueOwner(%q, %#v)", tc.valueOwner.String(), tc.pageReq)
			s.logBuffer.Reset()
			var links types.AccMDLinks
			var pageResp *query.PageResponse
			var err error
			testFunc := func() {
				links, pageResp, err = s.bk.GetScopesForValueOwner(ctx, tc.valueOwner, tc.pageReq)
			}
			s.Require().NotPanics(testFunc, callDesc)
			logs := s.getLogOutput(callDesc)

			s.AssertErrorValue(err, tc.expErr, "error from %s", callDesc)
			if !s.Assert().Equal(tc.expLinks, links, "links from %s", callDesc) {
				expStrs := mapToStrings(tc.expLinks)
				actStrs := mapToStrings(links)
				s.Assert().Equal(expStrs, actStrs, "strings of the links")
			}

			if !s.Assert().Equal(tc.expPageResp, pageResp, "page response from %s", callDesc) && tc.expPageResp != nil && pageResp != nil {
				s.Assert().Equal(fmt.Sprintf("%q", string(tc.expPageResp.NextKey)), fmt.Sprintf("%q", string(pageResp.NextKey)), "quoted pageResp.NextKey")
				s.Assert().Equal(int(tc.expPageResp.Total), int(pageResp.Total), "pageResp.Total as int")
			}

			s.Assert().Equal(tc.expLogs, logs, "log output during %s", callDesc)
		})
	}
}
