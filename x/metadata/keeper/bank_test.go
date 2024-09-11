package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func TestBankTestSuite(t *testing.T) {
	suite.Run(t, new(BankTestSuite))
}

type BankTestSuite struct {
	suite.Suite

	app *simapp.App
	ctx sdk.Context
	bk  *keeper.MDBankKeeper
}

func (s *BankTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.NewContext(false)
	s.bk = keeper.NewMDBankKeeper(s.app.BankKeeper)
}

func (s *BankTestSuite) AssertErrorValue(theError error, errorString string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return AssertErrorValue(s.T(), theError, errorString, msgAndArgs...)
}

type balance struct {
	addr  sdk.AccAddress
	denom string
	amt   int64 // Defaults to 1 if not provided.
}

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

func (s *BankTestSuite) TestDenomOwner() {
	addr1 := sdk.AccAddress("1_addr______________") // cosmos1x90kzerywf047h6lta047h6lta047h6l258ny6
	addr2 := sdk.AccAddress("2_addr______________") // cosmos1xf0kzerywf047h6lta047h6lta047h6lgww49l
	addr3 := sdk.AccAddress("3_addr______________") // cosmos1xd0kzerywf047h6lta047h6lta047h6l3lfhau
	addr4 := sdk.AccAddress("4_addr______________") // cosmos1x30kzerywf047h6lta047h6lta047h6lvnue84

	// subOne creates a string that sorts immediately before the provided val.
	subOne := func(val string) string {
		return val[:len(val)-1] + string(val[len(val)-1]-1)
	}
	// addOne creates a string that sorts immediately after the provided val.
	addOne := func(val string) string {
		return val[:len(val)-1] + string(val[len(val)-1]+1)
	}

	scopeUUIDStr := "69012AF4-2FA4-44DA-BAE4-1C13480362C9"
	scopeUUID, scopeUUIDErr := uuid.Parse(scopeUUIDStr)
	s.Require().NoError(scopeUUIDErr, "uuid.Parse(%q)", scopeUUIDStr)
	scopeID := types.ScopeMetadataAddress(scopeUUID) // scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghw
	scopeDenom := scopeID.Denom()                    // nft/scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghw
	scopeDenomBefore := subOne(scopeDenom)           // nft/scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghx
	scopeDenomAfter := addOne(scopeDenom)            // nft/scope1qp5sz2h597jyfk46uswpxjqrvtys3y0ghv

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

// TODO[2137]: func (s *BankTestSuite) TestGetScopesForValueOwner()
