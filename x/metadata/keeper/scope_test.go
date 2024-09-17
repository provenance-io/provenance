package keeper_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type ScopeKeeperTestSuite struct {
	suite.Suite

	app         *simapp.App
	queryClient types.QueryClient

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	pubkey3   cryptotypes.PubKey
	user3     string
	user3Addr sdk.AccAddress

	scUser     string
	scUserAddr sdk.AccAddress

	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	scopeSpecUUID uuid.UUID
	scopeSpecID   types.MetadataAddress
}

func (s *ScopeKeeperTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	ctx := s.FreshCtx()
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.MetadataKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	user1Acc := s.app.AccountKeeper.NewAccountWithAddress(ctx, s.user1Addr)
	s.Require().NoError(user1Acc.SetPubKey(s.pubkey1), "SetPubKey user1")
	s.app.AccountKeeper.SetAccount(ctx, user1Acc)

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubkey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubkey3.Address())
	s.user3 = s.user3Addr.String()

	s.scUserAddr = sdk.AccAddress("smart_contract_addr_")
	s.scUser = s.scUserAddr.String()
	s.app.AccountKeeper.SetAccount(ctx, s.app.AccountKeeper.NewAccount(ctx, authtypes.NewBaseAccount(s.scUserAddr, nil, 0, 0)))

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)

	s.scopeSpecUUID = uuid.New()
	s.scopeSpecID = types.ScopeSpecMetadataAddress(s.scopeSpecUUID)
}

func (s *ScopeKeeperTestSuite) FreshCtx() sdk.Context {
	return keeper.AddAuthzCacheToContext(s.app.BaseApp.NewContext(false))
}

// AssertErrorValue asserts that:
//   - If errorString is empty, theError must be nil
//   - If errorString is not empty, theError must equal the errorString.
func (s *ScopeKeeperTestSuite) AssertErrorValue(theError error, errorString string, msgAndArgs ...interface{}) bool {
	s.T().Helper()
	return AssertErrorValue(s.T(), theError, errorString, msgAndArgs...)
}

// SwapBankKeeper will set the bank keeper (in the metadata keeper) to the one provided
// and return a function that will set it back to its original value.
// Standard usage: defer s.SwapBankKeeper(tc.bk)()
// That will execute this method to set the bank keeper, then defer the resulting func (to put it back at the end).
func (s *ScopeKeeperTestSuite) SwapBankKeeper(bk keeper.BankKeeper) func() {
	orig := s.app.MetadataKeeper.SetBankKeeper(bk)
	return func() {
		s.app.MetadataKeeper.SetBankKeeper(orig)
	}
}

// SwapAuthzKeeper will set the authz keeper (in the metadata keeper) to the one provided
// and return a function that will set it back to its original value.
// Standard usage: defer s.SwapAuthzKeeper(tc.bk)()
// That will execute this method to set the authz keeper, then defer the resulting func (to put it back at the end).
func (s *ScopeKeeperTestSuite) SwapAuthzKeeper(ak keeper.AuthzKeeper) func() {
	orig := s.app.MetadataKeeper.SetAuthzKeeper(ak)
	return func() {
		s.app.MetadataKeeper.SetAuthzKeeper(orig)
	}
}

// SwapMarkerKeeper will set the marker keeper (in the metadata keeper) to the one provided
// and return a function that will set it back to its original value.
// Standard usage: defer s.SwapMarkerKeeper(tc.bk)()
// That will execute this method to set the marker keeper, then defer the resulting func (to put it back at the end).
func (s *ScopeKeeperTestSuite) SwapMarkerKeeper(mk keeper.MarkerKeeper) func() {
	orig := s.app.MetadataKeeper.SetMarkerKeeper(mk)
	return func() {
		s.app.MetadataKeeper.SetMarkerKeeper(orig)
	}
}

// WriteTempScope will call SetScope on the provided scope and return a func that will call RemoveScope for it.
// Standard usage: defer WriteTempScope(s.T(), s.app.MetadataKeeper, ctx, scope)()
// That will execute the SetScope and defer the call to RemoveScope.
func WriteTempScope(t *testing.T, mdKeeper keeper.Keeper, ctx sdk.Context, scope types.Scope) func() {
	assertions.RequireNotPanicsNoError(t, func() error {
		return mdKeeper.SetScope(ctx, scope)
	}, "SetScope")
	return func() {
		assertions.RequireNotPanicsNoError(t, func() error {
			return mdKeeper.RemoveScope(ctx, scope.ScopeId)
		}, "RemoveScope")
	}
}

func TestScopeKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ScopeKeeperTestSuite))
}

// func ownerPartyList defined in keeper_test.go

type testUser struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Addr    sdk.AccAddress
	Bech32  string
}

func randomUser() testUser {
	rv := testUser{}
	rv.PrivKey = secp256k1.GenPrivKey()
	rv.PubKey = rv.PrivKey.PubKey()
	rv.Addr = sdk.AccAddress(rv.PubKey.Address())
	rv.Bech32 = rv.Addr.String()
	return rv
}

func (s *ScopeKeeperTestSuite) TestMetadataScopeGetSet() {
	ctx := s.FreshCtx()
	theScope := *types.NewScope(s.scopeID, s.scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName).String() // cosmos1g4z8k7hm6hj5fa7s780slnxjvq2dnpgpj2jy0e
	eventCoinReceived := func(receiver string, amount sdk.Coin) sdk.Event {
		return sdk.NewEvent("coin_received",
			sdk.NewAttribute("receiver", receiver),
			sdk.NewAttribute("amount", amount.String()),
		)
	}
	eventCoinSpent := func(spender string, amount sdk.Coin) sdk.Event {
		return sdk.NewEvent("coin_spent",
			sdk.NewAttribute("spender", spender),
			sdk.NewAttribute("amount", amount.String()),
		)
	}
	eventTransfer := func(sender, recipient string, amount sdk.Coin) sdk.Event {
		return sdk.NewEvent("transfer",
			sdk.NewAttribute("recipient", recipient),
			sdk.NewAttribute("sender", sender),
			sdk.NewAttribute("amount", amount.String()),
		)
	}

	tests := []struct {
		name   string
		runner func()
	}{
		{
			name: "before setting scope",
			runner: func() {
				expScope := types.Scope{}
				actScope, found := s.app.MetadataKeeper.GetScope(ctx, theScope.ScopeId)
				s.Assert().False(found, "GetScope found")
				s.Assert().Equal(expScope, actScope, "GetScope result")
			},
		},
		{
			name: "set scope",
			runner: func() {
				// Note: Management of index entries during SetScope is tested in TestScopeIndexing.
				ctx = ctx.WithEventManager(sdk.NewEventManager())
				expEvent, err := sdk.TypedEventToEvent(types.NewEventScopeCreated(theScope.ScopeId))
				s.Require().NoError(err, "TypedEventToEvent NewEventScopeCreated")
				amt := theScope.ScopeId.Coin()
				expEvents := sdk.Events{
					eventCoinReceived(moduleAddr, amt),
					sdk.NewEvent("coinbase",
						sdk.NewAttribute("minter", moduleAddr),
						sdk.NewAttribute("amount", amt.String()),
					),
					eventCoinSpent(moduleAddr, amt),
					eventCoinReceived(theScope.ValueOwnerAddress, amt),
					eventTransfer(moduleAddr, theScope.ValueOwnerAddress, amt),
					sdk.NewEvent("message", sdk.NewAttribute("sender", moduleAddr)),
					expEvent,
				}

				err = s.app.MetadataKeeper.SetScope(ctx, theScope)
				s.Require().NoError(err, "SetScope")
				actEvents := ctx.EventManager().Events()
				assertions.AssertEqualEvents(s.T(), expEvents, actEvents, "events emitted during SetScope new")
			},
		},
		{
			name: "after setting it",
			runner: func() {
				expScope := theScope
				expScope.ValueOwnerAddress = ""
				actScope, found := s.app.MetadataKeeper.GetScope(ctx, theScope.ScopeId)
				s.Assert().True(found, "GetScope found")
				s.Assert().Equal(expScope, actScope, "GetScope result")

				actValueOwner, err := s.app.MetadataKeeper.GetScopeValueOwner(ctx, theScope.ScopeId)
				s.Require().NoError(err, "GetScopeValueOwner error")
				s.Assert().Equal(theScope.ValueOwnerAddress, actValueOwner.String(), "GetScopeValueOwner result")
			},
		},
		{
			name: "update scope",
			runner: func() {
				// Note: Management of index entries during SetScope is tested in TestScopeIndexing.
				ctx = ctx.WithEventManager(sdk.NewEventManager())
				theScope.DataAccess = append(theScope.DataAccess, s.user2)
				expEvent, err := sdk.TypedEventToEvent(types.NewEventScopeUpdated(theScope.ScopeId))
				s.Require().NoError(err, "TypedEventToEvent NewEventScopeUpdated")
				expEvents := sdk.Events{expEvent}

				err = s.app.MetadataKeeper.SetScope(ctx, theScope)
				s.Require().NoError(err, "SetScope")
				actEvents := ctx.EventManager().Events()
				assertions.AssertEqualEvents(s.T(), expEvents, actEvents, "events emitted during SetScope update")
			},
		},
		{
			name: "after update",
			runner: func() {
				expScope := theScope
				expScope.ValueOwnerAddress = ""
				actScope, found := s.app.MetadataKeeper.GetScope(ctx, theScope.ScopeId)
				s.Assert().True(found, "GetScope found")
				s.Assert().Equal(expScope, actScope, "GetScope result")
			},
		},
		{
			name: "update scope value owner",
			runner: func() {
				ctx = ctx.WithEventManager(sdk.NewEventManager())
				origOwner := theScope.ValueOwnerAddress
				theScope.ValueOwnerAddress = s.user2
				expEvent, err := sdk.TypedEventToEvent(types.NewEventScopeUpdated(theScope.ScopeId))
				s.Require().NoError(err, "TypedEventToEvent NewEventScopeUpdated")
				amt := theScope.ScopeId.Coin()
				expEvents := sdk.Events{
					eventCoinSpent(origOwner, amt),
					eventCoinReceived(theScope.ValueOwnerAddress, amt),
					eventTransfer(origOwner, theScope.ValueOwnerAddress, amt),
					sdk.NewEvent("message", sdk.NewAttribute("sender", origOwner)),
					expEvent,
				}

				err = s.app.MetadataKeeper.SetScope(ctx, theScope)
				s.Require().NoError(err, "SetScope")
				actEvents := ctx.EventManager().Events()
				assertions.AssertEqualEvents(s.T(), expEvents, actEvents, "events emitted during SetScope update")
			},
		},
		{
			name: "scope value owner after updating it",
			runner: func() {
				expScope := theScope
				actScope, found := s.app.MetadataKeeper.GetScopeWithValueOwner(ctx, theScope.ScopeId)
				s.Assert().True(found, "GetScope found")
				s.Assert().Equal(expScope, actScope, "GetScope result")
			},
		},
		{
			name: "remove scope",
			runner: func() {
				// Note: Management of index entries during RemoveScope is tested in TestScopeIndexing.
				// More detailed tests of RemoveScope is done in various other tests.
				ctx = ctx.WithEventManager(sdk.NewEventManager())
				expEvent, err := sdk.TypedEventToEvent(types.NewEventScopeDeleted(theScope.ScopeId))
				s.Require().NoError(err, "TypedEventToEvent NewEventScopeDeleted")
				amt := theScope.ScopeId.Coin()
				expEvents := sdk.Events{
					eventCoinSpent(theScope.ValueOwnerAddress, amt),
					eventCoinReceived(moduleAddr, amt),
					eventTransfer(theScope.ValueOwnerAddress, moduleAddr, amt),
					sdk.NewEvent("message", sdk.NewAttribute("sender", theScope.ValueOwnerAddress)),
					eventCoinSpent(moduleAddr, amt),
					sdk.NewEvent("burn",
						sdk.NewAttribute("burner", moduleAddr),
						sdk.NewAttribute("amount", amt.String()),
					),
					expEvent,
				}

				err = s.app.MetadataKeeper.RemoveScope(ctx, theScope.ScopeId)
				s.Require().NoError(err, "RemoveScope")
				actEvents := ctx.EventManager().Events()
				assertions.AssertEqualEvents(s.T(), expEvents, actEvents, "events emitted during RemoveScope")
			},
		},
		{
			name: "after remove scope",
			runner: func() {
				expScope := types.Scope{}
				actScope, found := s.app.MetadataKeeper.GetScope(ctx, theScope.ScopeId)
				s.Assert().False(found, "GetScope found")
				s.Assert().Equal(expScope, actScope, "GetScope result")
			},
		},
	}

	ok := true
	for _, tc := range tests {
		ok = s.Run(tc.name, func() {
			if !ok {
				s.T().Skip("Skipping due to previous failure.")
			}
			s.Require().NotPanics(tc.runner)
		}) && ok
	}
}

func (s *ScopeKeeperTestSuite) TestGetScopeWithValueOwner() {
	noScopeUUID, err := uuid.FromBytes([]byte("1111111111111111"))
	s.Require().NoError(err, "uuid.FromBytes([]byte(\"1111111111111111\"))")
	noScopeID := types.ScopeMetadataAddress(noScopeUUID)

	okScopeUUID, err := uuid.FromBytes([]byte("2222222222222222"))
	s.Require().NoError(err, "uuid.FromBytes([]byte(\"2222222222222222\"))")
	okScopeID := types.ScopeMetadataAddress(okScopeUUID)

	okScope := types.Scope{
		ScopeId:            okScopeID,
		SpecificationId:    s.scopeSpecID,
		Owners:             ownerPartyList(s.user1),
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}

	s.app.MetadataKeeper.SetScope(s.FreshCtx(), okScope)

	tests := []struct {
		name     string
		bk       *MockBankKeeper
		id       types.MetadataAddress
		expScope types.Scope
		expFound bool
	}{
		{
			name:     "no such scope",
			id:       noScopeID,
			expScope: types.Scope{},
			expFound: false,
		},
		{
			name: "no such scope but has a value owner on record",
			bk:   NewMockBankKeeper().WithDenomOwnerResult(noScopeID, s.user3Addr),
			id:   noScopeID,
			// This is testing that the ValueOwnerAddress field is not populated in this case.
			expScope: types.Scope{},
			expFound: false,
		},
		{
			name:     "scope without value owner",
			id:       okScopeID,
			expScope: okScope,
			expFound: true,
		},
		{
			name: "scope with value owner",
			bk:   NewMockBankKeeper().WithDenomOwnerResult(okScopeID, s.user3Addr),
			id:   okScopeID,
			expScope: types.Scope{
				ScopeId:            okScope.ScopeId,
				SpecificationId:    okScope.SpecificationId,
				Owners:             okScope.Owners,
				DataAccess:         okScope.DataAccess,
				ValueOwnerAddress:  s.user3,
				RequirePartyRollup: okScope.RequirePartyRollup,
			},
			expFound: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}
			defer s.SwapBankKeeper(tc.bk)()

			ctx := s.FreshCtx()
			var actScope types.Scope
			var actFound bool
			testFunc := func() {
				actScope, actFound = s.app.MetadataKeeper.GetScopeWithValueOwner(ctx, tc.id)
			}
			s.Require().NotPanics(testFunc, "GetScopeWithValueOwner")
			s.Assert().Equal(tc.expScope, actScope, "GetScopeWithValueOwner scope")
			s.Assert().Equal(tc.expFound, actFound, "GetScopeWithValueOwner found")
		})
	}
}

func (s *ScopeKeeperTestSuite) TestPopulateScopeValueOwner() {
	tests := []struct {
		name  string
		bk    *MockBankKeeper
		scope types.Scope
		expVO string
	}{
		{
			name:  "error getting value owner",
			bk:    NewMockBankKeeper().WithDenomOwnerError(s.scopeID, "oops go boom"),
			scope: types.Scope{ScopeId: s.scopeID, ValueOwnerAddress: "initialvo"},
			expVO: "",
		},
		{
			name:  "no value owner",
			scope: types.Scope{ScopeId: s.scopeID, ValueOwnerAddress: "initialvo"},
			expVO: "",
		},
		{
			name:  "has value owner",
			bk:    NewMockBankKeeper().WithDenomOwnerResult(s.scopeID, s.user2Addr),
			scope: types.Scope{ScopeId: s.scopeID, ValueOwnerAddress: "initialvo"},
			expVO: s.user2,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}
			defer s.SwapBankKeeper(tc.bk)()

			ctx := s.FreshCtx()
			testFunc := func() {
				s.app.MetadataKeeper.PopulateScopeValueOwner(ctx, &tc.scope)
			}
			s.Require().NotPanics(testFunc, "PopulateScopeValueOwner")
			actVO := tc.scope.ValueOwnerAddress
			s.Assert().Equal(tc.expVO, actVO, "ValueOwnerAddress after PopulateScopeValueOwner")
		})
	}
}

func (s *ScopeKeeperTestSuite) TestGetScopeValueOwner() {
	nonScopeErr := func(id string) string {
		return "cannot get value owner for non-scope metadata address \"" + id + "\""
	}

	tests := []struct {
		name      string
		bk        *MockBankKeeper
		id        types.MetadataAddress
		expAddr   sdk.AccAddress
		expErr    string
		expBKCall bool
	}{
		{
			name:   "nil id",
			id:     nil,
			expErr: nonScopeErr(""),
		},
		{
			name:   "empty id",
			id:     types.MetadataAddress{},
			expErr: nonScopeErr(""),
		},
		{
			name:   "session id",
			id:     types.SessionMetadataAddress(s.scopeUUID, s.scopeSpecUUID),
			expErr: nonScopeErr(types.SessionMetadataAddress(s.scopeUUID, s.scopeSpecUUID).String()),
		},
		{
			name:   "record id",
			id:     types.RecordMetadataAddress(s.scopeUUID, "justsomerecord"),
			expErr: nonScopeErr(types.RecordMetadataAddress(s.scopeUUID, "justsomerecord").String()),
		},
		{
			name:   "scope spec id",
			id:     s.scopeSpecID,
			expErr: nonScopeErr(s.scopeSpecID.String()),
		},
		{
			name:   "contract spec id",
			id:     types.ContractSpecMetadataAddress(s.scopeUUID),
			expErr: nonScopeErr(types.ContractSpecMetadataAddress(s.scopeUUID).String()),
		},
		{
			name:   "record spec id",
			id:     types.RecordSpecMetadataAddress(s.scopeUUID, "justsomerecord"),
			expErr: nonScopeErr(types.RecordSpecMetadataAddress(s.scopeUUID, "justsomerecord").String()),
		},
		{
			name:      "scope id without owner",
			id:        s.scopeID,
			expAddr:   nil,
			expErr:    "",
			expBKCall: true,
		},
		{
			name:      "scope id with lookup error",
			bk:        NewMockBankKeeper().WithDenomOwnerError(s.scopeID, "this error was injected"),
			id:        s.scopeID,
			expErr:    "this error was injected",
			expBKCall: true,
		},
		{
			name:      "scope id with owner",
			bk:        NewMockBankKeeper().WithDenomOwnerResult(s.scopeID, s.user1Addr),
			id:        s.scopeID,
			expAddr:   s.user1Addr,
			expBKCall: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}
			defer s.SwapBankKeeper(tc.bk)()

			var expBKCalls BankKeeperCalls
			if tc.expBKCall {
				expBKCalls.DenomOwner = append(expBKCalls.DenomOwner, tc.id.Denom())
			}

			ctx := s.FreshCtx()
			var actAddr sdk.AccAddress
			var actErr error
			testFunc := func() {
				actAddr, actErr = s.app.MetadataKeeper.GetScopeValueOwner(ctx, tc.id)
			}
			s.Require().NotPanics(testFunc, "GetScopeValueOwner")
			s.AssertErrorValue(actErr, tc.expErr, "GetScopeValueOwner error")
			s.Assert().Equal(tc.expAddr, actAddr, "GetScopeValueOwner address")
			tc.bk.AssertCalls(s.T(), expBKCalls)
		})
	}
}

func (s *ScopeKeeperTestSuite) TestGetScopeValueOwners() {
	nonScopeErr := func(id string) string {
		return "cannot get value owner for non-scope metadata address \"" + id + "\""
	}
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}
	uuids := make([]uuid.UUID, 10)
	scopeIDs := make([]types.MetadataAddress, len(uuids))
	for i := range uuids {
		bz := []byte(fmt.Sprintf("uuids[%d]________", i))
		var err error
		uuids[i], err = uuid.FromBytes(bz)
		s.Require().NoError(err, "uuid.FromBytes(%q)", string(bz))
		scopeIDs[i] = types.ScopeMetadataAddress(uuids[i])
	}

	tests := []struct {
		name       string
		bk         *MockBankKeeper
		ids        []types.MetadataAddress
		expLinks   types.AccMDLinks
		expErr     string
		expDOCalls []string // Expected calls made to DenomOwner.
	}{
		{
			name:     "nil ids",
			ids:      nil,
			expLinks: types.AccMDLinks{},
		},
		{
			name:     "empty ids",
			ids:      []types.MetadataAddress{},
			expLinks: types.AccMDLinks{},
		},
		{
			name:     "one nil id",
			ids:      []types.MetadataAddress{nil},
			expLinks: types.AccMDLinks{},
			expErr:   nonScopeErr(""),
		},
		{
			name:     "one empty id",
			ids:      []types.MetadataAddress{{}},
			expLinks: types.AccMDLinks{},
			expErr:   nonScopeErr(""),
		},
		{
			name: "one of each non-scope id",
			ids: []types.MetadataAddress{
				types.SessionMetadataAddress(uuids[0], uuids[1]),
				types.RecordMetadataAddress(uuids[2], "somerecord"),
				types.ScopeSpecMetadataAddress(uuids[3]),
				types.ContractSpecMetadataAddress(uuids[4]),
				types.RecordSpecMetadataAddress(uuids[5], "somerecord"),
			},
			expLinks: types.AccMDLinks{},
			expErr: joinErrs(
				nonScopeErr(types.SessionMetadataAddress(uuids[0], uuids[1]).String()),
				nonScopeErr(types.RecordMetadataAddress(uuids[2], "somerecord").String()),
				nonScopeErr(types.ScopeSpecMetadataAddress(uuids[3]).String()),
				nonScopeErr(types.ContractSpecMetadataAddress(uuids[4]).String()),
				nonScopeErr(types.RecordSpecMetadataAddress(uuids[5], "somerecord").String()),
			),
		},
		{
			name:       "one id: no owner",
			ids:        []types.MetadataAddress{scopeIDs[0]},
			expLinks:   types.AccMDLinks{types.NewAccMDLink(nil, scopeIDs[0])},
			expDOCalls: []string{scopeIDs[0].Denom()},
		},
		{
			name:       "one id: DenomOwner error",
			bk:         NewMockBankKeeper().WithDenomOwnerError(scopeIDs[1], "something broke yo"),
			ids:        []types.MetadataAddress{scopeIDs[1]},
			expLinks:   types.AccMDLinks{},
			expErr:     "something broke yo",
			expDOCalls: []string{scopeIDs[1].Denom()},
		},
		{
			name: "three ids: errors for all",
			bk: NewMockBankKeeper().
				WithDenomOwnerError(scopeIDs[3], "its now on fire").
				WithDenomOwnerError(scopeIDs[6], "small thing go big boom").
				WithDenomOwnerError(scopeIDs[7], "something broke yo"),
			ids:        []types.MetadataAddress{scopeIDs[7], scopeIDs[3], scopeIDs[6]},
			expLinks:   types.AccMDLinks{},
			expErr:     joinErrs("something broke yo", "its now on fire", "small thing go big boom"),
			expDOCalls: []string{scopeIDs[7].Denom(), scopeIDs[3].Denom(), scopeIDs[6].Denom()},
		},
		{
			name: "three ids: same owner",
			bk: NewMockBankKeeper().
				WithDenomOwnerResult(scopeIDs[4], s.user1Addr).
				WithDenomOwnerResult(scopeIDs[5], s.user1Addr).
				WithDenomOwnerResult(scopeIDs[6], s.user1Addr),
			ids: []types.MetadataAddress{scopeIDs[4], scopeIDs[5], scopeIDs[6]},
			expLinks: types.AccMDLinks{
				types.NewAccMDLink(s.user1Addr, scopeIDs[4]),
				types.NewAccMDLink(s.user1Addr, scopeIDs[5]),
				types.NewAccMDLink(s.user1Addr, scopeIDs[6]),
			},
			expDOCalls: []string{scopeIDs[4].Denom(), scopeIDs[5].Denom(), scopeIDs[6].Denom()},
		},
		{
			name: "three ids: different owners",
			bk: NewMockBankKeeper().
				WithDenomOwnerResult(scopeIDs[0], s.user1Addr).
				WithDenomOwnerResult(scopeIDs[9], s.user2Addr).
				WithDenomOwnerResult(scopeIDs[4], s.user3Addr),
			ids: []types.MetadataAddress{scopeIDs[0], scopeIDs[9], scopeIDs[4]},
			expLinks: types.AccMDLinks{
				types.NewAccMDLink(s.user1Addr, scopeIDs[0]),
				types.NewAccMDLink(s.user2Addr, scopeIDs[9]),
				types.NewAccMDLink(s.user3Addr, scopeIDs[4]),
			},
			expDOCalls: []string{scopeIDs[0].Denom(), scopeIDs[9].Denom(), scopeIDs[4].Denom()},
		},
		{
			name: "four ids: one non-scope, error from one, one found, one not found",
			bk: NewMockBankKeeper().
				WithDenomOwnerResult(scopeIDs[1], s.user3Addr).
				WithDenomOwnerError(scopeIDs[2], "oopsie daisy: no worky"),
			ids: []types.MetadataAddress{
				types.ContractSpecMetadataAddress(uuids[7]),
				scopeIDs[1], scopeIDs[2], scopeIDs[8],
			},
			expLinks: types.AccMDLinks{
				types.NewAccMDLink(s.user3Addr, scopeIDs[1]),
				types.NewAccMDLink(nil, scopeIDs[8]),
			},
			expErr: joinErrs(
				nonScopeErr(types.ContractSpecMetadataAddress(uuids[7]).String()),
				"oopsie daisy: no worky",
			),
			expDOCalls: []string{scopeIDs[1].Denom(), scopeIDs[2].Denom(), scopeIDs[8].Denom()},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}
			defer s.SwapBankKeeper(tc.bk)()

			expBKCalls := BankKeeperCalls{
				DenomOwner: tc.expDOCalls,
			}

			ctx := s.FreshCtx()
			var actLinks types.AccMDLinks
			var actErr error
			testFunc := func() {
				actLinks, actErr = s.app.MetadataKeeper.GetScopeValueOwners(ctx, tc.ids)
			}
			s.Require().NotPanics(testFunc, "GetScopeValueOwners")
			s.AssertErrorValue(actErr, tc.expErr, "GetScopeValueOwners error")
			s.Assert().Equal(tc.expLinks, actLinks, "GetScopeValueOwners address")
			tc.bk.AssertCalls(s.T(), expBKCalls)
		})
	}
}

func (s *ScopeKeeperTestSuite) TestSetScopeValueOwner() {
	decodeID := func(id string) types.MetadataAddress {
		rv, err := types.MetadataAddressFromBech32(id)
		s.Require().NoError(err, "types.MetadataAddressFromBech32(%q)", id)
		return rv
	}
	scopeIDStr := "scope1qpz0e5p8py55wa9mhckh3qg5qsasjwvmh2" // generated via CLI.
	scopeID := decodeID(scopeIDStr)
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName) // cosmos1g4z8k7hm6hj5fa7s780slnxjvq2dnpgpj2jy0e
	addr1 := sdk.AccAddress("1addr_______________")            // cosmos1x9skgerjta047h6lta047h6lta047h6l4429yc
	addr2 := sdk.AccAddress("2addr_______________")            // cosmos1xfskgerjta047h6lta047h6lta047h6lh0rr9a

	tests := []struct {
		name          string
		bk            *MockBankKeeper
		curOwner      sdk.AccAddress
		scopeID       types.MetadataAddress
		newValueOwner string
		expErr        string
		expCallBA     sdk.AccAddress // BA = BlockedAddr
		expCallDO     bool           // DO = Denom Owner
		expCallMint   bool
		expCallSend   *SendCoinsCall
		expCallBurn   bool
	}{
		{
			name:    "nil scope id",
			scopeID: nil,
			expErr:  "invalid scope metadata address MetadataAddress(nil): address is empty",
		},
		{
			name:    "empty scope id",
			scopeID: types.MetadataAddress{},
			expErr:  "invalid scope metadata address MetadataAddress{}: address is empty",
		},
		{
			name:    "invalid scope id",
			scopeID: types.MetadataAddress{types.ScopeKeyPrefix[0], 0x1, 0x2},
			expErr:  "invalid scope metadata address MetadataAddress{0x0, 0x1, 0x2}: incorrect address length (expected: 17, actual: 3)",
		},
		{
			name:    "session",
			scopeID: decodeID("session1q98duk50zlfyhpv3q7f88uzygyzdfw8hwdk2x3z8s4r009lk5nl6syhyghk"),
			expErr:  "invalid scope id \"session1q98duk50zlfyhpv3q7f88uzygyzdfw8hwdk2x3z8s4r009lk5nl6syhyghk\": wrong type",
		},
		{
			name:    "record",
			scopeID: decodeID("record1q26mxxwwvw2524dt3dpgf95gnhefy9ndhhsmphsxfntx7c8f52vpklgcn7v"),
			expErr:  "invalid scope id \"record1q26mxxwwvw2524dt3dpgf95gnhefy9ndhhsmphsxfntx7c8f52vpklgcn7v\": wrong type",
		},
		{
			name:    "scope spec",
			scopeID: decodeID("scopespec1qnna3wa2v4hy2l9jlklkvvtxjxes7wjq86"),
			expErr:  "invalid scope id \"scopespec1qnna3wa2v4hy2l9jlklkvvtxjxes7wjq86\": wrong type",
		},
		{
			name:    "contract spec",
			scopeID: decodeID("contractspec1qdwlarvm04p5cl4sca0vmzudksss654dk2"),
			expErr:  "invalid scope id \"contractspec1qdwlarvm04p5cl4sca0vmzudksss654dk2\": wrong type",
		},
		{
			name:    "record spec",
			scopeID: decodeID("recspec1qkrgw9lwe3k5gm5rh24kh0nkkkqujayqx92qrkvsezr6dvvyv4jmcw7t5tc"),
			expErr:  "invalid scope id \"recspec1qkrgw9lwe3k5gm5rh24kh0nkkkqujayqx92qrkvsezr6dvvyv4jmcw7t5tc\": wrong type",
		},
		{
			name:          "invalid new value owner",
			scopeID:       scopeID,
			newValueOwner: "bill",
			expErr:        "invalid new value owner address \"bill\": decoding bech32 failed: invalid bech32 string length 4",
		},
		{
			name:          "blocked new value owner",
			bk:            NewMockBankKeeper().WithBlockedAddr(addr1),
			scopeID:       scopeID,
			newValueOwner: addr1.String(),
			expErr:        fmt.Sprintf("new value owner %q is not allowed to receive funds: unauthorized", addr1.String()),
			expCallBA:     addr1,
		},
		{
			name:      "error getting current owner",
			bk:        NewMockBankKeeper().WithDenomOwnerError(scopeID, "not now clark"),
			scopeID:   scopeID,
			expErr:    fmt.Sprintf("could not get current value owner of %q: not now clark", scopeIDStr),
			expCallDO: true,
		},
		{
			name:          "no current owner to empty new owner",
			scopeID:       scopeID,
			newValueOwner: "",
			expErr:        "",
			expCallDO:     true,
		},
		{
			name:          "no current owner to new owner: error minting",
			bk:            NewMockBankKeeper().WithMintCoinsErrors("not so fresh"),
			scopeID:       scopeID,
			newValueOwner: addr1.String(),
			expErr:        fmt.Sprintf("could not mint scope coin \"1nft/%s\": not so fresh", scopeIDStr),
			expCallBA:     addr1,
			expCallDO:     true,
			expCallMint:   true,
		},
		{
			name:          "no current owner to new owner: error sending",
			bk:            NewMockBankKeeper().WithSendCoinsError(moduleAddr, "it is mine now"),
			scopeID:       scopeID,
			newValueOwner: addr1.String(),
			expErr: fmt.Sprintf("could not send scope coin \"1nft/%s\" from %s to %s: it is mine now",
				scopeIDStr, moduleAddr.String(), addr1.String()),
			expCallBA:   addr1,
			expCallDO:   true,
			expCallMint: true,
			expCallSend: NewSendCoinsCall(moduleAddr, addr1, sdk.Coins{scopeID.Coin()}),
		},
		{
			name:          "no current owner to new owner: okay",
			scopeID:       scopeID,
			newValueOwner: addr1.String(),
			expCallBA:     addr1,
			expCallDO:     true,
			expCallMint:   true,
			expCallSend:   NewSendCoinsCall(moduleAddr, addr1, sdk.Coins{scopeID.Coin()}),
		},
		{
			name:          "current owner to self",
			curOwner:      addr1,
			scopeID:       scopeID,
			newValueOwner: addr1.String(),
			expErr:        "",
			expCallBA:     addr1,
			expCallDO:     true,
		},
		{
			name:          "current owner to new owner: error sending",
			bk:            NewMockBankKeeper().WithSendCoinsError(addr1, "gonna keep this one"),
			curOwner:      addr1,
			scopeID:       scopeID,
			newValueOwner: addr2.String(),
			expErr: fmt.Sprintf("could not send scope coin \"1nft/%s\" from %s to %s: gonna keep this one",
				scopeIDStr, addr1.String(), addr2.String()),
			expCallBA:   addr2,
			expCallDO:   true,
			expCallSend: NewSendCoinsCall(addr1, addr2, sdk.Coins{scopeID.Coin()}),
		},
		{
			name:          "current owner to new owner: okay",
			curOwner:      addr1,
			scopeID:       scopeID,
			newValueOwner: addr2.String(),
			expCallBA:     addr2,
			expCallDO:     true,
			expCallSend:   NewSendCoinsCall(addr1, addr2, sdk.Coins{scopeID.Coin()}),
		},
		{
			name:          "current owner to empty new owner: error sending",
			bk:            NewMockBankKeeper().WithSendCoinsError(addr1, "finders keepers"),
			curOwner:      addr1,
			scopeID:       scopeID,
			newValueOwner: "",
			expErr: fmt.Sprintf("could not send scope coin \"1nft/%s\" from %s to %s: finders keepers",
				scopeIDStr, addr1.String(), moduleAddr.String()),
			expCallDO:   true,
			expCallSend: NewSendCoinsCall(addr1, moduleAddr, sdk.Coins{scopeID.Coin()}),
		},
		{
			name:          "current owner to empty new owner: error burning",
			bk:            NewMockBankKeeper().WithBurnCoinsErrors("too wet"),
			curOwner:      addr1,
			scopeID:       scopeID,
			newValueOwner: "",
			expErr:        fmt.Sprintf("could not burn scope coin \"1nft/%s\": too wet", scopeIDStr),
			expCallDO:     true,
			expCallSend:   NewSendCoinsCall(addr1, moduleAddr, sdk.Coins{scopeID.Coin()}),
			expCallBurn:   true,
		},
		{
			name:          "current owner to empty new owner: okay",
			curOwner:      addr1,
			scopeID:       scopeID,
			newValueOwner: "",
			expCallDO:     true,
			expCallSend:   NewSendCoinsCall(addr1, moduleAddr, sdk.Coins{scopeID.Coin()}),
			expCallBurn:   true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Set up expected bank keeper calls.
			expBKCalls := BankKeeperCalls{}
			if len(tc.expCallBA) > 0 {
				expBKCalls.BlockedAddr = append(expBKCalls.BlockedAddr, tc.expCallBA)
			}
			if tc.expCallMint {
				expBKCalls.MintCoins = append(expBKCalls.MintCoins, NewMintBurnCall(types.ModuleName, sdk.Coins{tc.scopeID.Coin()}))
			}
			if tc.expCallBurn {
				expBKCalls.BurnCoins = append(expBKCalls.BurnCoins, NewMintBurnCall(types.ModuleName, sdk.Coins{tc.scopeID.Coin()}))
			}
			if tc.expCallSend != nil {
				expBKCalls.SendCoins = append(expBKCalls.SendCoins, tc.expCallSend)
			}
			if tc.expCallDO {
				expBKCalls.DenomOwner = append(expBKCalls.DenomOwner, tc.scopeID.Denom())
			}

			// Set up the mock bank keeper.
			if tc.bk == nil {
				tc.bk = NewMockBankKeeper()
			}
			if len(tc.curOwner) > 0 {
				tc.bk = tc.bk.WithDenomOwnerResult(tc.scopeID, tc.curOwner)
			}
			defer s.SwapBankKeeper(tc.bk)()

			ctx := s.FreshCtx()
			var err error
			testFunc := func() {
				err = s.app.MetadataKeeper.SetScopeValueOwner(ctx, tc.scopeID, tc.newValueOwner)
			}
			s.Require().NotPanics(testFunc, "SetScopeValueOwner(%q, %q)", tc.scopeID, tc.newValueOwner)
			s.AssertErrorValue(err, tc.expErr, "error from SetScopeValueOwner(%q, %q)", tc.scopeID, tc.newValueOwner)
			tc.bk.AssertCalls(s.T(), expBKCalls)
		})
	}
}

func (s *ScopeKeeperTestSuite) TestSetScopeValueOwners() {
	// TODO[2137]: Redo this test
	s.FailNow("this test needs to be overhauled")
	// Setup
	// Three scopes, each with different value owners.
	// 1st has the value owner also in owners.
	// 2nd has the value owner also in data access.
	// 3rd does not have the value owner in either data access or owners.
	// We will call SetScopeValueOwners once to update all three to a new value owner.
	// We will then do some state checking to make sure things are as expected.
	addrAlsoOwnerAcc := sdk.AccAddress("addrAlsoOwner_______")
	addrAlsoDataAccessAcc := sdk.AccAddress("addrAlsoDataAccess__")
	addrSoloAcc := sdk.AccAddress("addrSolo____________")
	addrAlsoOwner := addrAlsoOwnerAcc.String()
	addrAlsoDataAccess := addrAlsoDataAccessAcc.String()
	addrSolo := addrSoloAcc.String()
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeWOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   scopeSpecID,
		Owners:            []types.Party{{Address: addrAlsoOwner, Role: types.PartyType_PARTY_TYPE_OWNER}},
		DataAccess:        nil,
		ValueOwnerAddress: addrAlsoOwner,
	}
	scopeWDataAccess := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   scopeSpecID,
		Owners:            nil,
		DataAccess:        []string{addrAlsoDataAccess},
		ValueOwnerAddress: addrAlsoOwner,
	}
	scopeSolo := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   scopeSpecID,
		Owners:            nil,
		DataAccess:        nil,
		ValueOwnerAddress: addrSolo,
	}

	ctx := s.FreshCtx()
	mdKeeper := s.app.MetadataKeeper
	mdKeeper.SetScope(ctx, scopeWOwner)
	mdKeeper.SetScope(ctx, scopeWDataAccess)
	mdKeeper.SetScope(ctx, scopeSolo)

	// Get a fresh context without any events.
	ctx = s.FreshCtx()

	newUpdateEvent := func(scopeID types.MetadataAddress) sdk.Event {
		tev := types.NewEventScopeUpdated(scopeID)
		event, err := sdk.TypedEventToEvent(tev)
		if err != nil {
			panic(err)
		}
		return event
	}

	scopes := []*types.Scope{&scopeWOwner, &scopeWDataAccess, &scopeSolo}
	_ = scopes
	addrNewValueOwnerAcc := sdk.AccAddress("addrNewValueOwner___")
	addrNewValueOwner := addrNewValueOwnerAcc.String()
	testFunc := func() {
		// TODO[2137]: Provide the correct 2nd arg here, and pay attention to the error.
		mdKeeper.SetScopeValueOwners(ctx, nil, addrNewValueOwner)
	}
	s.Require().NotPanics(testFunc, "SetScopeValueOwners")

	s.Run("emitted events", func() {
		expectedEvents := sdk.Events{
			newUpdateEvent(scopeWOwner.ScopeId),
			newUpdateEvent(scopeWDataAccess.ScopeId),
			newUpdateEvent(scopeSolo.ScopeId),
		}
		events := ctx.EventManager().Events()
		s.Assert().Equal(expectedEvents, events, "events emitted during SetScopeValueOwners")
	})

	tests := []struct {
		name          string
		scope         *types.Scope
		expIndexes    [][]byte
		expRemIndexes [][]byte
	}{
		{
			name:  "scopeWOwner",
			scope: &scopeWOwner,
			expIndexes: [][]byte{
				types.GetAddressScopeCacheKey(addrAlsoOwnerAcc, scopeWOwner.ScopeId),
			},
		},
		{
			name:  "scopeWDataAccess",
			scope: &scopeWDataAccess,
			expIndexes: [][]byte{
				types.GetAddressScopeCacheKey(addrAlsoDataAccessAcc, scopeWDataAccess.ScopeId),
			},
		},
		{
			name:  "scopeSolo",
			scope: &scopeSolo,
			expRemIndexes: [][]byte{
				types.GetAddressScopeCacheKey(addrSoloAcc, scopeWDataAccess.ScopeId),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx = s.FreshCtx()
			newScope, found := mdKeeper.GetScope(ctx, tc.scope.ScopeId)
			if s.Assert().True(found, "GetScope found") {
				s.Assert().Equal(addrNewValueOwner, newScope.ValueOwnerAddress, "stored scope's value owner address")
			}
			s.Assert().NotEqual(addrNewValueOwner, tc.scope.ValueOwnerAddress, "old scope's value owner address")

			store := ctx.KVStore(mdKeeper.GetStoreKey())
			for i, exp := range tc.expIndexes {
				s.Assert().True(store.Has(exp), "expected index [%d]", i)
			}
			for i, notExp := range tc.expRemIndexes {
				s.Assert().False(store.Has(notExp), "expected index to be removed [%d]", i)
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestMetadataScopeIterator() {
	ctx := s.FreshCtx()
	for i := 1; i <= 10; i++ {
		valueOwner := ""
		if i == 5 {
			valueOwner = s.user2
		}
		ns := types.NewScope(types.ScopeMetadataAddress(uuid.New()), nil, ownerPartyList(s.user1), []string{s.user1}, valueOwner, false)
		s.app.MetadataKeeper.SetScope(ctx, *ns)
	}
	count := 0
	err := s.app.MetadataKeeper.IterateScopes(ctx, func(s types.Scope) (stop bool) {
		count++
		return false
	})
	s.Require().NoError(err, "IterateScopes")
	s.Assert().Equal(10, count, "number of scopes iterated")

	count = 0
	err = s.app.MetadataKeeper.IterateScopesForAddress(ctx, s.user1Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		s.True(scopeID.IsScopeAddress())
		return false
	})
	s.Require().NoError(err, "IterateScopesForAddress user1")
	s.Assert().Equal(10, count, "number of scope ids iterated for user1")

	count = 0
	err = s.app.MetadataKeeper.IterateScopesForAddress(ctx, s.user2Addr, func(scopeID types.MetadataAddress) (stop bool) {
		count++
		s.True(scopeID.IsScopeAddress())
		return false
	})
	s.Require().NoError(err, "IterateScopesForAddress user2")
	s.Assert().Equal(0, count, "number of scope ids iterated for user2")

	count = 0
	err = s.app.MetadataKeeper.IterateScopes(ctx, func(s types.Scope) (stop bool) {
		count++
		return count >= 5
	})
	s.Require().NoError(err, "IterateScopes with early stop")
	s.Assert().Equal(5, count, "number of scopes iterated with early stop")
}

func (s *ScopeKeeperTestSuite) TestValidateWriteScope() {
	s.FailNow("Need to refactor this to account for existing not being an arg anymore.")
	ns := func(scopeID, scopeSpecification types.MetadataAddress, owners []types.Party, dataAccess []string, valueOwner string) *types.Scope {
		return &types.Scope{
			ScopeId:           scopeID,
			SpecificationId:   scopeSpecification,
			Owners:            owners,
			DataAccess:        dataAccess,
			ValueOwnerAddress: valueOwner,
		}
	}
	rollupScope := func(scopeID, specID types.MetadataAddress, owners []types.Party, valueOwner string) *types.Scope {
		return &types.Scope{
			ScopeId:            scopeID,
			SpecificationId:    specID,
			Owners:             owners,
			DataAccess:         nil,
			ValueOwnerAddress:  valueOwner,
			RequirePartyRollup: true,
		}
	}
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	owner := types.PartyType_PARTY_TYPE_OWNER
	affiliate := types.PartyType_PARTY_TYPE_AFFILIATE
	provenance := types.PartyType_PARTY_TYPE_PROVENANCE

	ctx := s.FreshCtx()
	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := s.app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     s.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdkmath.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	s.Require().NoError(err, "AddMarkerAccount")

	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	s.app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpec)
	scopeSpecSC := types.NewScopeSpecification(types.ScopeSpecMetadataAddress(uuid.New()), nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_PROVENANCE}, []types.MetadataAddress{})
	s.app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpecSC)

	scopeID := types.ScopeMetadataAddress(uuid.New())
	scopeID2 := types.ScopeMetadataAddress(uuid.New())

	// Give user 3 authority to sign for user 1 for scope updates.
	a := authz.NewGenericAuthorization(types.TypeURLMsgWriteScopeRequest)
	s.Require().NoError(s.app.AuthzKeeper.SaveGrant(ctx, s.user3Addr, s.user1Addr, a, nil), "authz SaveGrant user1 to user3")

	otherAddr := sdk.AccAddress("other_address_______").String()

	cases := []struct {
		name     string
		existing *types.Scope
		proposed types.Scope
		signers  []string
		authzK   *MockAuthzKeeper
		errorMsg string
		expAddrs []sdk.AccAddress // TODO[2137]: Define this in each test case and make sure it's fully tested.
	}{
		{
			name:     "nil previous, proposed throws address error",
			existing: nil,
			proposed: types.Scope{},
			signers:  []string{s.user1},
			errorMsg: "address is empty",
		},
		{
			name:     "valid proposed with nil existing doesn't error",
			existing: nil,
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "can't change scope id in update",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID2, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("cannot update scope identifier. expected %s, got %s", scopeID.String(), scopeID2.String()),
		},
		{
			name:     "missing existing owner signer on update fails",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, ""),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "missing existing owner signer on update fails",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, ""),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "no error when update includes existing owner signer",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user1}, ""),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "no error when there are no updates regardless of signatures",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			signers:  []string{},
			errorMsg: "",
		},
		{
			name:     "setting value owner when unset does not error",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting value owner when unset requires current owner signature",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "setting value owner to user does not require their signature",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting value owner to new user does not require their signature",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "no change to value owner should not error",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting a new value owner should not error with withdraw permission",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, markerAddr),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "with rollup setting a new value owner should not error with withdraw permission",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), markerAddr),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), s.user1),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting a new value owner fails if missing withdraw permission",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, markerAddr),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to withdraw/remove it as scope value owner", markerAddr),
		},
		{
			name:     "with rollup setting a new value owner fails if missing withdraw permission",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), markerAddr),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to withdraw/remove it as scope value owner", markerAddr),
		},
		{
			name:     "setting a new value owner fails if missing deposit permission",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user2), []string{}, markerAddr),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to deposit/add it as scope value owner", markerAddr),
		},
		{
			name:     "with rollup setting a new value owner fails if missing deposit permission",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user2), markerAddr),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature for %s (testcoin) with authority to deposit/add it as scope value owner", markerAddr),
		},
		{
			name:     "setting a new value owner fails for scope owner when value owner signature is missing",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "with rollup setting a new value owner fails for scope owner when value owner signature is missing",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), s.user2),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "changing only value owner only requires value owner sig",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{}, otherAddr),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{}, s.user1),
			signers:  []string{otherAddr},
			errorMsg: "",
		},
		{
			name:     "with rollup changing only value owner only requires value owner sig",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), otherAddr),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), s.user1),
			signers:  []string{otherAddr},
			errorMsg: "",
		},
		{
			name:     "unsetting all fields on a scope should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: types.Scope{ScopeId: scopeID, SpecificationId: scopeSpecID, Owners: ownerPartyList(s.user1)},
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "setting specification id to nil should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, nil, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: "invalid specification id: address is empty",
		},
		{
			name:     "setting unknown specification id should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, types.ScopeSpecMetadataAddress(s.scopeUUID), ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("scope specification %s not found", types.ScopeSpecMetadataAddress(s.scopeUUID)),
		},
		{
			name:     "adding data access with authz grant should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{s.user2}, s.user1),
			signers:  []string{s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "multi owner adding data access with authz grant should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), []string{s.user2}, s.user1),
			signers:  []string{s.user2, s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "changing value owner with authz grant should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user3}, // user 1 has granted scope-write to user 3
			errorMsg: "",
		},
		{
			name:     "changing value owner by authz granter should be successful",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "changing value owner by non-authz grantee should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user1),
		},
		{
			name:     "changing value owner from non-authz granter with different signer should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user1),
			signers:  []string{s.user3},
			errorMsg: fmt.Sprintf("missing signature from existing value owner %s", s.user2),
		},
		{
			name:     "setting value owner from nothing to non-owner only signed by non-owner should fail",
			existing: ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, ""),
			proposed: *ns(scopeID, scopeSpecID, ownerPartyList(s.user1), []string{}, s.user2),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "with rollup scope missing req role",
			existing: nil,
			proposed: *rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, affiliate, false)), ""),
			signers:  nil,
			errorMsg: "missing roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup without existing but has req role and signer not involved in scope",
			existing: nil,
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), ""),
			signers:  []string{otherAddr},
			errorMsg: "",
		},
		{
			name:     "with rollup existing required owner is not signer",
			existing: rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ownerPartyList(s.user1, s.user2), ""),
			signers:  []string{s.user2},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:     "with rollup missing signer from required role",
			existing: rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, owner, true), pt(s.user2, affiliate, false)), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, owner, true), pt(s.user2, affiliate, false), pt(s.user2, types.PartyType_PARTY_TYPE_OWNER, true)), ""),
			signers:  []string{s.user2},
			errorMsg: "missing signers for roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup two optional owners one signs",
			existing: rollupScope(scopeID, scopeSpecID, ptz(pt(s.user1, owner, true), pt(s.user2, owner, true)), ""),
			proposed: *rollupScope(scopeID, scopeSpecID, ptz(pt(s.user2, owner, true)), ""),
			signers:  []string{s.user2},
			errorMsg: "",
		},
		{
			name:     "smart contract account is not PROVENANCE role",
			existing: nil,
			proposed: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpecID,
				Owners:             ptz(pt(s.scUser, owner, false)),
				RequirePartyRollup: false,
			},
			signers:  []string{s.scUser},
			errorMsg: `account "` + s.scUser + `" is a smart contract but does not have the PROVENANCE role`,
		},
		{
			name:     "with rollup smart contract account is not PROVENANCE role",
			existing: nil,
			proposed: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpecID,
				Owners:             ptz(pt(s.scUser, owner, false)),
				RequirePartyRollup: true,
			},
			signers:  []string{s.scUser},
			errorMsg: `account "` + s.scUser + `" is a smart contract but does not have the PROVENANCE role`,
		},
		{
			name:     "non-smart contract party has PROVENANCE role",
			existing: nil,
			proposed: types.Scope{
				ScopeId:         scopeID,
				SpecificationId: scopeSpecID,
				Owners:          ptz(pt(s.user1, owner, false), pt(s.user2, provenance, false)),
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "account \"" + s.user2 + "\" has role PROVENANCE but is not a smart contract",
		},
		{
			name:     "with rollup non-smart contract party has PROVENANCE role",
			existing: nil,
			proposed: types.Scope{
				ScopeId:            scopeID,
				SpecificationId:    scopeSpecID,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, provenance, true)),
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "account \"" + s.user2 + "\" has role PROVENANCE but is not a smart contract",
		},
		{
			name: "only change is value owner signed by smart contract",
			// Even though the smart contract owns this scope. it shouldn't be allowed to change that value owner.
			existing: &types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress: s.user1,
			},
			proposed: types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress: s.scUser,
			},
			signers:  []string{s.scUser, s.user1},
			errorMsg: "smart contract signer " + s.scUser + " is not authorized",
		},
		{
			name: "with rollup only change is value owner signed by smart contract",
			// Even though the smart contract owns this scope. it shouldn't be allowed to change that value owner.
			existing: &types.Scope{
				ScopeId:            scopeID,
				SpecificationId:    scopeSpecSC.SpecificationId,
				Owners:             ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress:  s.user1,
				RequirePartyRollup: true,
			},
			proposed: types.Scope{
				ScopeId:            scopeID,
				SpecificationId:    scopeSpecSC.SpecificationId,
				Owners:             ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress:  s.scUser,
				RequirePartyRollup: true,
			},
			signers:  []string{s.scUser, s.user1},
			errorMsg: "smart contract signer " + s.scUser + " is not authorized",
		},
		{
			name: "only change is value owner signed by smart contract and authorized",
			existing: &types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress: s.user1,
			},
			proposed: types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress: s.scUser,
			},
			signers: []string{s.scUser, s.user1},
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: GrantInfo{
						Granter: s.user1Addr,
						Grantee: s.scUserAddr,
						MsgType: types.TypeURLMsgWriteScopeRequest},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil,
					},
				},
			),
			errorMsg: "",
		},
		{
			name: "with rollup only change is value owner signed by smart contract",
			existing: &types.Scope{
				ScopeId:            scopeID,
				SpecificationId:    scopeSpecSC.SpecificationId,
				Owners:             ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress:  s.user1,
				RequirePartyRollup: true,
			},
			proposed: types.Scope{
				ScopeId:            scopeID,
				SpecificationId:    scopeSpecSC.SpecificationId,
				Owners:             ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress:  s.scUser,
				RequirePartyRollup: true,
			},
			signers: []string{s.scUser, s.user1},
			authzK: NewMockAuthzKeeper().WithGetAuthorizationResults(
				GetAuthorizationCall{
					GrantInfo: GrantInfo{
						Granter: s.user1Addr,
						Grantee: s.scUserAddr,
						MsgType: types.TypeURLMsgWriteScopeRequest},
					Result: GetAuthorizationResult{
						Auth: NewMockAuthorization("one", authz.AcceptResponse{Accept: true}, nil),
						Exp:  nil,
					},
				},
			),
			errorMsg: "",
		},
		{
			name: "only change is smart contract value owner signed by smart contract",
			existing: &types.Scope{
				ScopeId:            scopeID,
				SpecificationId:    scopeSpecSC.SpecificationId,
				Owners:             ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress:  s.scUser,
				RequirePartyRollup: true,
			},
			proposed: types.Scope{
				ScopeId:            scopeID,
				SpecificationId:    scopeSpecSC.SpecificationId,
				Owners:             ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress:  s.user1,
				RequirePartyRollup: true,
			},
			signers:  []string{s.scUser},
			errorMsg: "",
		},
		{
			name: "with rollup only change is smart contract value owner signed by smart contract",
			existing: &types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress: s.scUser,
			},
			proposed: types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.scUser, types.PartyType_PARTY_TYPE_PROVENANCE, false)),
				ValueOwnerAddress: s.user1,
			},
			signers:  []string{s.scUser},
			errorMsg: "",
		},
		{
			name: "only change is value owner roles not checked with spec",
			// The spec requires an owner, so this will fail if owners are checked against the spec.
			// But it shouldn't be checked because the only change is to the value owner.
			existing: &types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.user1, affiliate, false)),
				ValueOwnerAddress: s.user1,
			},
			proposed: types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.user1, affiliate, false)),
				ValueOwnerAddress: s.user2,
			},
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name: "only change is value owner provenance roles not checked",
			// The spec requires an owner, so we have one. But we also have a PROVENANCE party that isn't
			// a smart contract. That should fail if checked, but shouldn't be checked in this case.
			existing: &types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.user1, owner, false), pt(s.user1, provenance, false)),
				ValueOwnerAddress: s.user1,
			},
			proposed: types.Scope{
				ScopeId:           scopeID,
				SpecificationId:   scopeSpecSC.SpecificationId,
				Owners:            ptz(pt(s.user1, owner, false), pt(s.user1, provenance, false)),
				ValueOwnerAddress: s.user2,
			},
			signers:  []string{s.user1},
			errorMsg: "",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			if tc.authzK != nil {
				origAuthzK := s.app.MetadataKeeper.SetAuthzKeeper(tc.authzK)
				defer s.app.MetadataKeeper.SetAuthzKeeper(origAuthzK)
			}
			msg := &types.MsgWriteScopeRequest{
				Scope:   tc.proposed,
				Signers: tc.signers,
			}
			// TODO[2137]: Set tc.existing in state so it can be retrieved.
			addrs, err := s.app.MetadataKeeper.ValidateWriteScope(s.FreshCtx(), msg)
			s.AssertErrorValue(err, tc.errorMsg, "error from ValidateWriteScope")
			s.Assert().Equal(tc.expAddrs, addrs, "addrs from ValidateWriteScope")
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateDeleteScope() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	ctx := s.FreshCtx()
	markerDenom := "testcoins2"
	markerAddr := markertypes.MustGetMarkerAddress(markerDenom).String()
	err := s.app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 24,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     s.user1,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      markerDenom,
		Supply:     sdkmath.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	s.Require().NoError(err, "AddMarkerAccount")

	scopeNoValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user1, s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeNoValueOwner)

	scopeMarkerValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: markerAddr,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeMarkerValueOwner)

	scopeUserValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: s.user1,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeUserValueOwner)

	owner := types.PartyType_PARTY_TYPE_OWNER
	servicer := types.PartyType_PARTY_TYPE_SERVICER

	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "test scope spec", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, servicer},
		ContractSpecIds: []types.MetadataAddress{types.ContractSpecMetadataAddress(uuid.New())},
	}
	s.app.MetadataKeeper.SetScopeSpecification(ctx, scopeSpec)

	otherUser := sdk.AccAddress("some_other_user_____").String()

	// with rollup no scope spec req party not signed
	scopeRollupNoSpecReq := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, servicer, false), pt(otherUser, owner, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupNoSpecReq)

	// with rollup no scope spec all optional parties signer not involved
	scopeRollupNoSpecAllOpt := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:             ptz(pt(s.user1, owner, true), pt(s.user2, servicer, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupNoSpecAllOpt)

	// with rollup req scope owner not signed
	// with rollup req role not signed
	// with rollup req scope owner and req role signed.
	scopeRollup := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    scopeSpec.SpecificationId,
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, servicer, true), pt(otherUser, owner, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollup)

	// with rollup marker value owner no signer has withdraw
	scopeRollupMarkerValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: markerAddr,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupMarkerValueOwner)

	// with rollup value owner not signed
	scopeRollupUserValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: s.user1,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupUserValueOwner)

	dneScopeID := types.ScopeMetadataAddress(uuid.New())

	missing1Sig := func(addr string) string {
		return fmt.Sprintf("missing signature: %s", addr)
	}

	missing2Sigs := func(addr1, addr2 string) string {
		return fmt.Sprintf("missing signatures: %s, %s", addr1, addr2)
	}

	tests := []struct {
		name     string
		scope    types.Scope
		signers  []string
		expAddrs []sdk.AccAddress
		expErr   string
	}{
		{
			name:    "no value owner all signers",
			scope:   scopeNoValueOwner,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
		{
			name:    "no value owner all signers reversed",
			scope:   scopeNoValueOwner,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
		{
			name:    "no value owner extra signer",
			scope:   scopeNoValueOwner,
			signers: []string{s.user1, s.user2, s.user3},
			expErr:  "",
		},
		{
			name:    "no value owner missing signer 1",
			scope:   scopeNoValueOwner,
			signers: []string{s.user2},
			expErr:  missing1Sig(s.user1),
		},
		{
			name:    "no value owner missing signer 2",
			scope:   scopeNoValueOwner,
			signers: []string{s.user1},
			expErr:  missing1Sig(s.user2),
		},
		{
			name:    "no value owner no signers",
			scope:   scopeNoValueOwner,
			signers: []string{},
			expErr:  missing2Sigs(s.user1, s.user2),
		},
		{
			name:    "no value owner wrong signer",
			scope:   scopeNoValueOwner,
			signers: []string{s.user3},
			expErr:  missing2Sigs(s.user1, s.user2),
		},
		{
			name:    "marker value owner signed by owner and user with auth", // TODO[2137]: Figure out what to do with this case.
			scope:   scopeMarkerValueOwner,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
		{
			name:    "marker value owner signed by owner and user with auth reversed", // TODO[2137]: Figure out what to do with this case.
			scope:   scopeMarkerValueOwner,
			signers: []string{s.user2, s.user1},
			expErr:  "",
		},
		{
			name:    "marker value owner not signed by owner",
			scope:   scopeMarkerValueOwner,
			signers: []string{s.user1},
			expErr:  missing1Sig(s.user2),
		},
		{
			name:    "marker value owner not signed by user with auth", // TODO[2137]: Figure out what to do with this case.
			scope:   scopeMarkerValueOwner,
			signers: []string{s.user2},
			expErr:  fmt.Sprintf("missing signature for %s (testcoins2) with authority to withdraw/remove it as scope value owner", markerAddr),
		},
		{
			name:    "user value owner signed by owner and value owner",
			scope:   scopeUserValueOwner,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
		{
			name:    "user value owner signed by owner and value owner reversed",
			scope:   scopeUserValueOwner,
			signers: []string{s.user2, s.user1},
			expErr:  "",
		},
		{
			name:    "user value owner not signed by owner",
			scope:   scopeUserValueOwner,
			signers: []string{s.user1},
			expErr:  missing1Sig(s.user2),
		},
		{
			name:    "user value owner not signed by value owner",
			scope:   scopeUserValueOwner,
			signers: []string{s.user2},
			expErr:  fmt.Sprintf("missing signature from existing value owner %s", s.user1),
		},
		{
			name:    "scope does not exist",
			scope:   types.Scope{ScopeId: dneScopeID},
			signers: []string{},
			expErr:  fmt.Sprintf("scope not found with id %s", dneScopeID),
		},
		{
			name:    "with rollup no scope spec neither req party signed",
			scope:   scopeRollupNoSpecReq,
			signers: []string{otherUser},
			expErr:  "missing signatures: " + s.user1 + ", " + s.user2 + "",
		},
		{
			name:    "with rollup no scope spec req party 1 not signed",
			scope:   scopeRollupNoSpecReq,
			signers: []string{s.user2},
			expErr:  "missing signature: " + s.user1,
		},
		{
			name:    "with rollup no scope spec req party 2 not signed",
			scope:   scopeRollupNoSpecReq,
			signers: []string{s.user1},
			expErr:  "missing signature: " + s.user2,
		},
		{
			name:    "with rollup no scope spec both req parties signed",
			scope:   scopeRollupNoSpecReq,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
		{
			name:    "with rollup no scope spec all optional parties signer not involved",
			scope:   scopeRollupNoSpecAllOpt,
			signers: []string{otherUser},
			expErr:  "",
		},
		{
			name:    "with rollup req scope owner not signed",
			scope:   scopeRollup,
			signers: []string{s.user2, otherUser},
			expErr:  "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:    "with rollup req role not signed",
			scope:   scopeRollup,
			signers: []string{s.user1},
			expErr:  "missing signers for roles required by spec: SERVICER need 1 have 0",
		},
		{
			name:    "with rollup req scope owner and req roles signed",
			scope:   scopeRollup,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
		{
			name:    "with rollup marker value owner no signer has withdraw", // TODO[2137]: Figure out what to do with this case.
			scope:   scopeRollupMarkerValueOwner,
			signers: []string{s.user2},
			expErr:  "missing signature for " + markerAddr + " (testcoins2) with authority to withdraw/remove it as scope value owner",
		},
		{
			name:    "with rollup marker value owner signer has withdraw", // TODO[2137]: Figure out what to do with this case.
			scope:   scopeRollupMarkerValueOwner,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
		{
			name:    "with rollup value owner not signed",
			scope:   scopeRollupUserValueOwner,
			signers: []string{s.user2},
			expErr:  "missing signature from existing value owner " + s.user1,
		},
		{
			name:    "with rollup value owner signed",
			scope:   scopeRollupUserValueOwner,
			signers: []string{s.user1, s.user2},
			expErr:  "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msg := &types.MsgDeleteScopeRequest{
				ScopeId: tc.scope.ScopeId,
				Signers: tc.signers,
			}
			// TODO[2137]: Define the expAddrs in these test cases.
			var addrs []sdk.AccAddress
			testFunc := func() {
				addrs, err = s.app.MetadataKeeper.ValidateDeleteScope(s.FreshCtx(), msg)
			}
			s.Require().NotPanics(testFunc, "ValidateDeleteScope")
			s.AssertErrorValue(err, tc.expErr, "error from ValidateDeleteScope")
			s.Assert().Equal(tc.expAddrs, addrs, "addresses from ValidateDeleteScope")
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateSetScopeAccountData() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	ctx := s.FreshCtx()

	scopeNoValueOwner := types.Scope{
		ScopeId:           types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:            ownerPartyList(s.user1, s.user2),
		DataAccess:        nil,
		ValueOwnerAddress: "",
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeNoValueOwner)

	owner := types.PartyType_PARTY_TYPE_OWNER
	servicer := types.PartyType_PARTY_TYPE_SERVICER

	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "test scope spec", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, servicer},
		ContractSpecIds: []types.MetadataAddress{types.ContractSpecMetadataAddress(uuid.New())},
	}
	s.app.MetadataKeeper.SetScopeSpecification(ctx, scopeSpec)

	otherUser := sdk.AccAddress("some_other_user_____").String()

	// with rollup no scope spec req party not signed
	scopeRollupNoSpecReq := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    types.ScopeSpecMetadataAddress(uuid.New()),
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, servicer, false), pt(otherUser, owner, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollupNoSpecReq)

	// with rollup req scope owner not signed
	// with rollup req role not signed
	// with rollup req scope owner and req role signed.
	scopeRollup := types.Scope{
		ScopeId:            types.ScopeMetadataAddress(uuid.New()),
		SpecificationId:    scopeSpec.SpecificationId,
		Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, servicer, true), pt(otherUser, owner, true)),
		DataAccess:         nil,
		ValueOwnerAddress:  "",
		RequirePartyRollup: true,
	}
	s.app.MetadataKeeper.SetScope(ctx, scopeRollup)

	dneScopeID := types.ScopeMetadataAddress(uuid.New())

	missing1Sig := func(addr string) string {
		return fmt.Sprintf("missing signature: %s", addr)
	}

	missing2Sigs := func(addr1, addr2 string) string {
		return fmt.Sprintf("missing signatures: %s, %s", addr1, addr2)
	}

	tests := []struct {
		name     string
		addr     types.MetadataAddress
		value    string
		signers  []string
		expected string
	}{
		{
			name:     "all signers",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "all signers reversed",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "extra signer",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{s.user1, s.user2, s.user3},
			expected: "",
		},
		{
			name:     "missing signer 1",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{s.user2},
			expected: missing1Sig(s.user1),
		},
		{
			name:     "missing signer 2",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{s.user1},
			expected: missing1Sig(s.user2),
		},
		{
			name:     "no signers",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{},
			expected: missing2Sigs(s.user1, s.user2),
		},
		{
			name:     "wrong signer",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{s.user3},
			expected: missing2Sigs(s.user1, s.user2),
		},
		{
			name:     "scope does not exist",
			addr:     dneScopeID,
			value:    "Some new value.",
			signers:  []string{},
			expected: fmt.Sprintf("scope not found with id %s", dneScopeID),
		},
		{
			name:     "scope does not exist but value is empty",
			addr:     dneScopeID,
			value:    "",
			signers:  []string{},
			expected: "",
		},
		{
			name:     "with rollup no scope spec",
			addr:     scopeRollupNoSpecReq.ScopeId,
			signers:  []string{otherUser},
			expected: fmt.Sprintf("scope specification %s not found for scope id %s", scopeRollupNoSpecReq.SpecificationId, scopeRollupNoSpecReq.ScopeId),
		},
		{
			name:     "with rollup req scope owner not signed",
			addr:     scopeRollup.ScopeId,
			signers:  []string{s.user2, otherUser},
			expected: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:     "with rollup req role not signed",
			addr:     scopeRollup.ScopeId,
			signers:  []string{s.user1},
			expected: "missing signers for roles required by spec: SERVICER need 1 have 0",
		},
		{
			name:     "with rollup req scope owner and req roles signed",
			addr:     scopeRollup.ScopeId,
			signers:  []string{s.user1, s.user2},
			expected: "",
		},
		{
			name:     "smart contract singer not involved",
			addr:     scopeNoValueOwner.ScopeId,
			signers:  []string{s.user1, s.user2, s.scUser},
			expected: "smart contract signer " + s.scUser + " cannot follow non-smart-contract signer",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			msg := &types.MsgSetAccountDataRequest{
				MetadataAddr: tc.addr,
				Value:        tc.value,
				Signers:      tc.signers,
			}
			actual := s.app.MetadataKeeper.ValidateSetScopeAccountData(s.FreshCtx(), msg)
			if len(tc.expected) > 0 {
				require.EqualError(t, actual, tc.expected)
			} else {
				require.NoError(t, actual)
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeAddDataAccess() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	scope := types.Scope{
		ScopeId:           s.scopeID,
		SpecificationId:   types.ScopeSpecMetadataAddress(s.scopeUUID),
		Owners:            ownerPartyList(s.user1),
		DataAccess:        []string{s.user1},
		ValueOwnerAddress: s.user1,
	}

	owner := types.PartyType_PARTY_TYPE_OWNER
	controller := types.PartyType_PARTY_TYPE_CONTROLLER

	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "test description", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, controller},
		ContractSpecIds: []types.MetadataAddress{types.ContractSpecMetadataAddress(uuid.New())},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.FreshCtx(), scopeSpec)

	dneSpecID := types.ScopeSpecMetadataAddress(uuid.New())

	otherAddr := sdk.AccAddress("blah_blah_blah_blah_").String()

	cases := []struct {
		name            string
		dataAccessAddrs []string
		existing        types.Scope
		signers         []string
		errorMsg        string
	}{
		{
			name:            "should fail to validate add scope data access, does not have any users",
			dataAccessAddrs: []string{},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "data access list cannot be empty",
		},
		{
			name:            "should fail to validate add scope data access, user is already on the data access list",
			dataAccessAddrs: []string{s.user1},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        fmt.Sprintf("address already exists for data access %s", s.user1),
		},
		{
			name:            "should fail to validate add scope data access, incorrect signer for scope",
			dataAccessAddrs: []string{s.user2},
			existing:        scope,
			signers:         []string{s.user2},
			errorMsg:        fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:            "should fail to validate add scope data access, incorrect address type",
			dataAccessAddrs: []string{"invalidaddr"},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "failed to decode data access address invalidaddr : decoding bech32 failed: invalid separator index -1",
		},
		{
			name:            "should successfully validate add scope data access",
			dataAccessAddrs: []string{s.user2},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 1",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 2",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user3},
			errorMsg: "",
		},
		{
			name:            "with rollup spec not found",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    dneSpecID,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "scope specification " + dneSpecID.String() + " not found",
		},
		{
			name:            "with rollup req party not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user2, s.user3},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:            "with rollup req role not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         nil,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1},
			errorMsg: "missing signers for roles required by spec: CONTROLLER need 1 have 0",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			msg := &types.MsgAddScopeDataAccessRequest{
				DataAccess: tc.dataAccessAddrs,
				Signers:    tc.signers,
			}
			err := s.app.MetadataKeeper.ValidateAddScopeDataAccess(s.FreshCtx(), tc.existing, msg)
			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateAddScopeDataAccess")
			} else {
				s.Assert().NoError(err, tc.errorMsg, "ValidateAddScopeDataAccess")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateScopeDeleteDataAccess() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{
			Address:  addr,
			Role:     role,
			Optional: opt,
		}
	}
	ptz := func(parties ...types.Party) []types.Party {
		rv := make([]types.Party, 0, len(parties))
		rv = append(rv, parties...)
		return rv
	}

	scope := types.Scope{
		ScopeId:           s.scopeID,
		SpecificationId:   types.ScopeSpecMetadataAddress(s.scopeUUID),
		Owners:            ownerPartyList(s.user1),
		DataAccess:        []string{s.user1, s.user2},
		ValueOwnerAddress: s.user1,
	}

	owner := types.PartyType_PARTY_TYPE_OWNER
	controller := types.PartyType_PARTY_TYPE_CONTROLLER

	scopeSpec := types.ScopeSpecification{
		SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
		Description:     types.NewDescription("tester", "test description", "", ""),
		OwnerAddresses:  []string{s.user1},
		PartiesInvolved: []types.PartyType{owner, controller},
		ContractSpecIds: []types.MetadataAddress{types.ContractSpecMetadataAddress(uuid.New())},
	}
	s.app.MetadataKeeper.SetScopeSpecification(s.FreshCtx(), scopeSpec)

	dneSpecID := types.ScopeSpecMetadataAddress(uuid.New())

	otherAddr := sdk.AccAddress("blah_blah_blah_blah_").String()

	cases := []struct {
		name            string
		dataAccessAddrs []string
		existing        types.Scope
		signers         []string
		errorMsg        string
	}{
		{
			name:            "should fail to validate delete scope data access, does not have any users",
			dataAccessAddrs: []string{},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "data access list cannot be empty",
		},
		{
			name:            "should fail to validate delete scope data access, address is not in data access list",
			dataAccessAddrs: []string{s.user2, s.user3},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        fmt.Sprintf("address does not exist in scope data access: %s", s.user3),
		},
		{
			name:            "should fail to validate delete scope data access, incorrect signer for scope",
			dataAccessAddrs: []string{s.user2},
			existing:        scope,
			signers:         []string{s.user2},
			errorMsg:        fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:            "should successfully validate delete scope data access",
			dataAccessAddrs: []string{s.user1, s.user2},
			existing:        scope,
			signers:         []string{s.user1},
			errorMsg:        "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 1",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:            "with rollup spec found signed correctly with opt addr 2",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    scopeSpec.SpecificationId,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user3},
			errorMsg: "",
		},
		{
			name:            "with rollup spec not found",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:            types.ScopeMetadataAddress(uuid.New()),
				SpecificationId:    dneSpecID,
				Owners:             ptz(pt(s.user1, owner, false), pt(s.user2, controller, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1, s.user2},
			errorMsg: "scope specification " + dneSpecID.String() + " not found",
		},
		{
			name:            "with rollup req party not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user2, s.user3},
			errorMsg: "missing required signature: " + s.user1 + " (OWNER)",
		},
		{
			name:            "with rollup req role not signed",
			dataAccessAddrs: []string{otherAddr},
			existing: types.Scope{
				ScopeId:         types.ScopeMetadataAddress(uuid.New()),
				SpecificationId: scopeSpec.SpecificationId,
				Owners: ptz(pt(s.user1, owner, false), pt(s.user2, controller, true),
					pt(s.user3, owner, true), pt(s.user3, controller, true)),
				DataAccess:         []string{otherAddr},
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1},
			errorMsg: "missing signers for roles required by spec: CONTROLLER need 1 have 0",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			msg := &types.MsgDeleteScopeDataAccessRequest{
				DataAccess: tc.dataAccessAddrs,
				Signers:    tc.signers,
			}
			err := s.app.MetadataKeeper.ValidateDeleteScopeDataAccess(s.FreshCtx(), tc.existing, msg)
			if len(tc.errorMsg) > 0 {
				s.Assert().EqualError(err, tc.errorMsg, "ValidateDeleteScopeDataAccess")
			} else {
				s.Assert().NoError(err, "ValidateDeleteScopeDataAccess")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestValidateUpdateScopeOwners() {
	pt := func(addr string, role types.PartyType, opt bool) types.Party {
		return types.Party{Address: addr, Role: role, Optional: opt}
	}
	ctx := s.FreshCtx()
	scopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())
	scopeSpec := types.NewScopeSpecification(scopeSpecID, nil, []string{s.user1}, []types.PartyType{types.PartyType_PARTY_TYPE_OWNER}, []types.MetadataAddress{})
	s.app.MetadataKeeper.SetScopeSpecification(ctx, *scopeSpec)

	scopeWithOwners := func(owners []types.Party) types.Scope {
		return types.Scope{
			ScopeId:           s.scopeID,
			SpecificationId:   scopeSpecID,
			Owners:            owners,
			DataAccess:        []string{s.user1},
			ValueOwnerAddress: s.user1,
		}
	}
	rollupScopeWithOwners := func(owners ...types.Party) types.Scope {
		return types.Scope{
			ScopeId:            s.scopeID,
			SpecificationId:    scopeSpecID,
			Owners:             owners,
			DataAccess:         []string{s.user1},
			ValueOwnerAddress:  s.user1,
			RequirePartyRollup: true,
		}
	}

	originalOwners := ownerPartyList(s.user1)
	dneScopeSpecID := types.ScopeSpecMetadataAddress(uuid.New())

	owner := types.PartyType_PARTY_TYPE_OWNER
	omnibus := types.PartyType_PARTY_TYPE_OMNIBUS

	testCases := []struct {
		name     string
		existing types.Scope
		proposed types.Scope
		signers  []string
		errorMsg string
	}{
		{
			name:     "should fail to validate update scope owners, fail to decode address",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: "shoulderror", Role: types.PartyType_PARTY_TYPE_AFFILIATE}}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("invalid scope owners: invalid party address [%s]: %s", "shoulderror", "decoding bech32 failed: invalid separator index -1"),
		},
		{
			name:     "should fail to validate update scope owners, role cannot be unspecified",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_UNSPECIFIED}}),
			signers:  []string{s.user1},
			errorMsg: fmt.Sprintf("invalid scope owners: invalid party type for party %s", s.user1),
		},
		{
			name:     "should fail to validate update scope owner, wrong signer new owner",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}),
			signers:  []string{s.user2},
			errorMsg: fmt.Sprintf("missing signature: %s", s.user1),
		},
		{
			name:     "should successfully validate update scope owner, same owner different role",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			proposed: scopeWithOwners([]types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_CUSTODIAN},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER},
			}),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:     "should successfully validate update scope owner, new owner",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER}}),
			signers:  []string{s.user1},
			errorMsg: "",
		},
		{
			name:     "should fail to validate update scope owner, missing role",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{{Address: s.user1, Role: types.PartyType_PARTY_TYPE_CUSTODIAN}}),
			signers:  []string{s.user1},
			errorMsg: "missing roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "should fail to validate update scope owner, empty list",
			existing: scopeWithOwners(originalOwners),
			proposed: scopeWithOwners([]types.Party{}),
			signers:  []string{s.user1},
			errorMsg: "invalid scope owners: at least one party is required",
		},
		{
			name:     "should successfully validate update scope owner, 1st owner removed",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			proposed: scopeWithOwners(ownerPartyList(s.user2)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:     "should successfully validate update scope owner, 2nd owner removed",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.user2)),
			proposed: scopeWithOwners(ownerPartyList(s.user1)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:     "should fail to add optional owner to a non-rollup scope",
			existing: scopeWithOwners(ownerPartyList(s.user1)),
			proposed: scopeWithOwners([]types.Party{
				{Address: s.user1, Role: types.PartyType_PARTY_TYPE_OWNER, Optional: false},
				{Address: s.user2, Role: types.PartyType_PARTY_TYPE_OWNER, Optional: true},
			}),
			signers:  []string{s.user1},
			errorMsg: "parties can only be optional when require_party_rollup = true",
		},
		{
			name: "scope spec not found",
			existing: types.Scope{
				ScopeId:           s.scopeID,
				SpecificationId:   dneScopeSpecID,
				Owners:            originalOwners,
				DataAccess:        []string{s.user1},
				ValueOwnerAddress: s.user1,
			},
			proposed: types.Scope{
				ScopeId:           s.scopeID,
				SpecificationId:   dneScopeSpecID,
				Owners:            ownerPartyList(s.user1, s.user2),
				DataAccess:        []string{s.user1},
				ValueOwnerAddress: s.user1,
			},
			signers:  []string{s.user1},
			errorMsg: "scope specification " + dneScopeSpecID.String() + " not found",
		},
		{
			name: "with rollup, scope spec not found",
			existing: types.Scope{
				ScopeId:            s.scopeID,
				SpecificationId:    dneScopeSpecID,
				Owners:             originalOwners,
				DataAccess:         []string{s.user1},
				ValueOwnerAddress:  s.user1,
				RequirePartyRollup: true,
			},
			proposed: types.Scope{
				ScopeId:            s.scopeID,
				SpecificationId:    dneScopeSpecID,
				Owners:             ownerPartyList(s.user1, s.user2),
				DataAccess:         []string{s.user1},
				ValueOwnerAddress:  s.user1,
				RequirePartyRollup: true,
			},
			signers:  []string{s.user1},
			errorMsg: "scope specification " + dneScopeSpecID.String() + " not found",
		},
		{
			name:     "with rollup new owners do not have required roles",
			existing: rollupScopeWithOwners(pt(s.user1, owner, false)),
			proposed: rollupScopeWithOwners(pt(s.user1, omnibus, false)),
			signers:  []string{s.user1},
			errorMsg: "missing roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup neither optional party signed for required role",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, owner, true)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, true)),
			signers:  []string{s.user3},
			errorMsg: "missing signers for roles required by spec: OWNER need 1 have 0",
		},
		{
			name:     "with rollup one optional party signed for required role",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, owner, true)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, true)),
			signers:  []string{s.user2},
			errorMsg: "",
		},
		{
			name:     "with rollup required party not signed",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, false)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user3, omnibus, false)),
			signers:  []string{s.user1},
			errorMsg: "missing required signature: " + s.user2 + " (OMNIBUS)",
		},
		{
			name:     "with rollup all good",
			existing: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user2, omnibus, false)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, true), pt(s.user3, omnibus, false)),
			signers:  []string{s.user1, s.user2},
			errorMsg: "",
		},
		{
			name:     "smart contract without provenance role added",
			existing: scopeWithOwners(ownerPartyList(s.user1)),
			proposed: scopeWithOwners(ownerPartyList(s.user1, s.scUser)),
			signers:  []string{s.user1},
			errorMsg: `account "` + s.scUser + `" is a smart contract but does not have the PROVENANCE role`,
		},
		{
			name:     "smart contract without provenance role removed",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.scUser)),
			proposed: scopeWithOwners(ownerPartyList(s.user1)),
			signers:  []string{s.scUser, s.user1},
			errorMsg: "",
		},
		{
			name:     "smart contract without provenance role removed but wrong signer order",
			existing: scopeWithOwners(ownerPartyList(s.user1, s.scUser)),
			proposed: scopeWithOwners(ownerPartyList(s.user1)),
			signers:  []string{s.user1, s.scUser},
			errorMsg: "smart contract signer " + s.scUser + " cannot follow non-smart-contract signer",
		},
		{
			name:     "with rollup smart contract without provenance role added",
			existing: rollupScopeWithOwners(pt(s.user1, owner, false)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, false), pt(s.scUser, owner, true)),
			signers:  []string{s.user1},
			errorMsg: `account "` + s.scUser + `" is a smart contract but does not have the PROVENANCE role`,
		},
		{
			name:     "with rollup smart contract without provenance role removed",
			existing: rollupScopeWithOwners(pt(s.user1, owner, false), pt(s.scUser, owner, true)),
			proposed: rollupScopeWithOwners(pt(s.user1, owner, false)),
			signers:  []string{s.user1},
			errorMsg: "",
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			var msg types.MetadataMsg
			if len(tc.proposed.Owners) > len(tc.existing.Owners) {
				msg = &types.MsgAddScopeOwnerRequest{Signers: tc.signers}
			} else {
				msg = &types.MsgDeleteScopeOwnerRequest{Signers: tc.signers}
			}
			err := s.app.MetadataKeeper.ValidateUpdateScopeOwners(s.FreshCtx(), tc.existing, tc.proposed, msg)
			if len(tc.errorMsg) > 0 {
				assert.EqualError(t, err, tc.errorMsg, "ValidateUpdateScopeOwners expected error")
			} else {
				assert.NoError(t, err, "ValidateUpdateScopeOwners unexpected error")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestScopeIndexing() {
	scopeID := types.ScopeMetadataAddress(uuid.New())

	specIDOrig := types.ScopeSpecMetadataAddress(uuid.New())
	specIDNew := types.ScopeSpecMetadataAddress(uuid.New())

	ownerConstant := randomUser()
	ownerToAdd := randomUser()
	ownerToRemove := randomUser()
	valueOwnerOrig := randomUser()
	valueOwnerNew := randomUser()

	scopeV1 := types.Scope{
		ScopeId:           scopeID,
		SpecificationId:   specIDOrig,
		Owners:            ownerPartyList(ownerConstant.Bech32, ownerToRemove.Bech32),
		DataAccess:        nil,
		ValueOwnerAddress: valueOwnerOrig.Bech32,
	}
	scopeV2 := types.Scope{
		ScopeId:           scopeID,
		SpecificationId:   specIDNew,
		Owners:            ownerPartyList(ownerConstant.Bech32, ownerToAdd.Bech32),
		DataAccess:        nil,
		ValueOwnerAddress: valueOwnerNew.Bech32,
	}

	ctx := s.FreshCtx()
	store := ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.Run("1 write new scope", func() {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
		}

		err := s.app.MetadataKeeper.SetScope(ctx, scopeV1)
		s.Require().NoError(err, "SetScope")

		for _, expected := range expectedIndexes {
			s.Assert().True(store.Has(expected.key), expected.name)
		}
	})

	s.Run("2 update scope", func() {
		expectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToAdd.Addr, scopeID), "ownerToAdd address index"},

			{types.GetScopeSpecScopeCacheKey(specIDNew, scopeID), "specIDNew spec index"},
		}
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
		}

		err := s.app.MetadataKeeper.SetScope(ctx, scopeV2)
		s.Require().NoError(err, "SetScope")

		for _, expected := range expectedIndexes {
			s.Assert().True(store.Has(expected.key), expected.name)
		}
		for _, unexpected := range unexpectedIndexes {
			s.Assert().False(store.Has(unexpected.key), unexpected.name)
		}
	})

	s.Run("3 delete scope", func() {
		unexpectedIndexes := []struct {
			key  []byte
			name string
		}{
			{types.GetAddressScopeCacheKey(ownerConstant.Addr, scopeID), "ownerConstant address index"},
			{types.GetAddressScopeCacheKey(ownerToRemove.Addr, scopeID), "ownerToRemove address index"},
			{types.GetAddressScopeCacheKey(ownerToAdd.Addr, scopeID), "ownerToAdd address index"},
			{types.GetAddressScopeCacheKey(valueOwnerOrig.Addr, scopeID), "valueOwnerOrig address index"},
			{types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeID), "valueOwnerNew address index"},

			{types.GetScopeSpecScopeCacheKey(specIDOrig, scopeID), "specIDOrig spec index"},
			{types.GetScopeSpecScopeCacheKey(specIDNew, scopeID), "specIDNew spec index"},
		}

		err := s.app.MetadataKeeper.RemoveScope(ctx, scopeID)
		s.Require().NoError(err, "RemoveScope")

		for _, unexpected := range unexpectedIndexes {
			s.Assert().False(store.Has(unexpected.key), unexpected.name)
		}
	})
}

func (s *ScopeKeeperTestSuite) TestValidateUpdateValueOwners() {
	newUUID := func(i string) uuid.UUID {
		str := strings.ReplaceAll("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", "x", i)
		rv, err := uuid.Parse(str)
		s.Require().NoError(err, "uuid.Parse(%q)", str)
		return rv
	}
	scopeID1 := types.ScopeMetadataAddress(newUUID("1"))
	scopeID2 := types.ScopeMetadataAddress(newUUID("2"))
	scopeID3 := types.ScopeMetadataAddress(newUUID("3"))
	scopeID4 := types.ScopeMetadataAddress(newUUID("4"))

	addr1 := sdk.AccAddress("1addr_______________") // cosmos1x9skgerjta047h6lta047h6lta047h6l4429yc
	addr2 := sdk.AccAddress("2addr_______________") // cosmos1xfskgerjta047h6lta047h6lta047h6lh0rr9a
	addr3 := sdk.AccAddress("3addr_______________") // cosmos1xdskgerjta047h6lta047h6lta047h6lw7ypa7
	addr4 := sdk.AccAddress("4addr_______________") // cosmos1x3skgerjta047h6lta047h6lta047h6lnj308h
	addr5 := sdk.AccAddress("5addr_______________") // cosmos1x4skgerjta047h6lta047h6lta047h6l2rkdl5

	accStrs := func(addrs []sdk.AccAddress) []string {
		if addrs == nil {
			return nil
		}
		rv := make([]string, len(addrs))
		for i, addr := range addrs {
			rv[i] = addr.String()
		}
		return rv
	}
	type msgMaker struct {
		name    string
		msgType string
		make    func(signers []sdk.AccAddress) types.MetadataMsg
	}
	msgMakerUpdate := msgMaker{
		name:    "update",
		msgType: types.TypeURLMsgUpdateValueOwnersRequest,
		make: func(signers []sdk.AccAddress) types.MetadataMsg {
			return &types.MsgUpdateValueOwnersRequest{Signers: accStrs(signers)}
		},
	}
	msgMakerMigrate := msgMaker{
		name:    "migrate",
		msgType: types.TypeURLMsgMigrateValueOwnerRequest,
		make: func(signers []sdk.AccAddress) types.MetadataMsg {
			return &types.MsgMigrateValueOwnerRequest{Signers: accStrs(signers)}
		},
	}
	allMsgMakers := []msgMaker{msgMakerUpdate, msgMakerMigrate}

	missingSig := func(addr sdk.AccAddress) string {
		return "missing signature from existing value owner \"" + addr.String() + "\""
	}

	tests := []struct {
		name      string
		wasmAddrs []sdk.AccAddress
		links     types.AccMDLinks
		proposed  string
		signers   []sdk.AccAddress
		expErr    string
		expAddrs  []sdk.AccAddress
	}{
		{
			name:   "nil links",
			links:  nil,
			expErr: "no scopes found",
		},
		{
			name:   "empty links",
			links:  types.AccMDLinks{},
			expErr: "no scopes found",
		},
		{
			name:   "nil entry in links",
			links:  types.AccMDLinks{{AccAddr: addr1, MDAddr: scopeID1}, nil, {AccAddr: addr2, MDAddr: scopeID2}},
			expErr: "nil entry not allowed",
		},
		{
			name:   "link without acc addr",
			links:  types.AccMDLinks{{AccAddr: nil, MDAddr: scopeID1}},
			expErr: "no account address associated with metadata address \"" + scopeID1.String() + "\"",
		},
		{
			name:   "link without md addr",
			links:  types.AccMDLinks{{AccAddr: addr1, MDAddr: nil}},
			expErr: "invalid scope metadata address MetadataAddress(nil): address is empty",
		},
		{
			name:   "duplicate md addr in links",
			links:  types.AccMDLinks{{AccAddr: addr1, MDAddr: scopeID1}, {AccAddr: addr1, MDAddr: scopeID1}},
			expErr: "duplicate metadata address \"" + scopeID1.String() + "\" not allowed",
		},
		{
			name:      "first signer is wasm: only first signer returned",
			wasmAddrs: []sdk.AccAddress{addr1},
			links:     types.AccMDLinks{{AccAddr: addr1, MDAddr: scopeID1}, {AccAddr: addr1, MDAddr: scopeID2}},
			proposed:  addr5.String(),
			signers:   []sdk.AccAddress{addr1, addr2},
			expAddrs:  []sdk.AccAddress{addr1},
		},
		{
			name:      "first signer is wasm: missing sig from second",
			wasmAddrs: []sdk.AccAddress{addr1},
			links:     types.AccMDLinks{{AccAddr: addr1, MDAddr: scopeID1}, {AccAddr: addr2, MDAddr: scopeID2}},
			proposed:  addr5.String(),
			signers:   []sdk.AccAddress{addr1, addr2},
			expErr:    missingSig(addr2),
		},
		{
			name:      "first signer is not wasm: all signers returned.",
			wasmAddrs: []sdk.AccAddress{addr2}, // second one, just to show it doesn't matter.
			links:     types.AccMDLinks{{AccAddr: addr1, MDAddr: scopeID1}, {AccAddr: addr2, MDAddr: scopeID2}},
			proposed:  addr5.String(),
			signers:   []sdk.AccAddress{addr1, addr2},
			expAddrs:  []sdk.AccAddress{addr1, addr2},
		},
		{
			name:      "missing signature",
			wasmAddrs: nil,
			links: types.AccMDLinks{
				{AccAddr: addr1, MDAddr: scopeID1}, {AccAddr: addr2, MDAddr: scopeID2},
				{AccAddr: addr3, MDAddr: scopeID3}, {AccAddr: addr4, MDAddr: scopeID4},
			},
			proposed: addr5.String(),
			signers:  []sdk.AccAddress{addr1, addr2, addr4},
			expErr:   missingSig(addr3),
		},
	}

	for _, tc := range tests {
		for _, maker := range allMsgMakers {
			s.Run(maker.name+": "+tc.name, func() {
				// Ignore authz and marker stuff for these tests and assume that tests on ValidateScopeValueOwnersSigners hit that.
				defer s.SwapAuthzKeeper(NewMockAuthzKeeper())()
				defer s.SwapMarkerKeeper(NewMockMarkerKeeper())()

				msg := maker.make(tc.signers)
				ctx := s.FreshCtx()
				if len(tc.wasmAddrs) > 0 {
					cache := keeper.GetAuthzCache(ctx)
					for _, addr := range tc.wasmAddrs {
						cache.SetIsWasm(addr, true)
					}
				}

				var addrs []sdk.AccAddress
				var err error
				testFunc := func() {
					addrs, err = s.app.MetadataKeeper.ValidateUpdateValueOwners(ctx, tc.links, tc.proposed, msg)
				}
				s.Require().NotPanics(testFunc, "ValidateUpdateValueOwners")
				s.AssertErrorValue(err, tc.expErr, "error from ValidateUpdateValueOwners")
				s.Assert().Equal(tc.expAddrs, addrs, "addrs from ValidateUpdateValueOwners")
			})
		}
	}
}

func (s *ScopeKeeperTestSuite) TestAddSetNetAssetValues() {
	markerDenom := "jackthecat"
	mAccount := authtypes.NewBaseAccount(markertypes.MustGetMarkerAddress(markerDenom), nil, 0, 0)
	s.app.MarkerKeeper.SetMarker(s.FreshCtx(), s.app.MarkerKeeper.NewMarker(s.FreshCtx(), markertypes.NewMarkerAccount(mAccount, sdk.NewInt64Coin(markerDenom, 1000), s.user1Addr, []markertypes.AccessGrant{{Address: s.user1, Permissions: []markertypes.Access{markertypes.Access_Transfer}}}, markertypes.StatusFinalized, markertypes.MarkerType_RestrictedCoin, true, true, false, []string{})))
	scopeID := types.ScopeMetadataAddress(uuid.New())
	tests := []struct {
		name           string
		scopeID        types.MetadataAddress
		netAssetValues []types.NetAssetValue
		source         string
		expErr         string
	}{
		{
			name:    "Invalid Denom",
			scopeID: scopeID,
			netAssetValues: []types.NetAssetValue{
				{
					Price: sdk.Coin{
						Denom:  "invalid",
						Amount: sdkmath.NewInt(1000),
					},
				},
			},
			source: "source",
			expErr: "net asset value denom does not exist",
		},
		{
			name:    "Valid Net Asset Values USD",
			scopeID: scopeID,
			netAssetValues: []types.NetAssetValue{
				{
					Price: sdk.Coin{
						Denom:  "usd",
						Amount: sdkmath.NewInt(1000),
					},
				},
			},
			source: "source",
		},
		{
			name:    "Valid Net Asset Values stake",
			scopeID: scopeID,
			netAssetValues: []types.NetAssetValue{
				{
					Price: sdk.Coin{
						Denom:  "jackthecat",
						Amount: sdkmath.NewInt(1000),
					},
				},
			},
			source: "source",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			err := s.app.MetadataKeeper.AddSetNetAssetValues(s.FreshCtx(), tc.scopeID, tc.netAssetValues, tc.source)
			if tc.expErr == "" {
				s.Assert().NoError(err, "AddSetNetAssetValues should have no error.")
			} else {
				s.Assert().Error(err, "AddSetNetAssetValues should have error.")
				s.Assert().Contains(err.Error(), tc.expErr, "AddSetNetAssetValues error message incorrect.")
			}
		})
	}
}

func (s *ScopeKeeperTestSuite) TestSetNetAssetValue() {
	scopeID := types.ScopeMetadataAddress(uuid.New())
	tests := []struct {
		name          string
		scopeID       types.MetadataAddress
		netAssetValue types.NetAssetValue
		source        string
		expErr        string
	}{
		{
			name:    "Valid Net Asset Value",
			scopeID: scopeID,
			netAssetValue: types.NetAssetValue{
				Price: sdk.Coin{
					Denom:  "usd",
					Amount: sdkmath.NewInt(1000),
				},
			},
			source: "test",
		},
		{
			name:    "Invalid Net Asset Value",
			scopeID: scopeID,
			netAssetValue: types.NetAssetValue{
				Price: sdk.Coin{
					Denom:  "",
					Amount: sdkmath.NewInt(1000),
				},
			},
			source: "source",
			expErr: "invalid denom: ",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var expErrs []string
			var expEvents sdk.Events
			if len(tc.expErr) == 0 {
				event := types.NewEventSetNetAssetValue(scopeID, tc.netAssetValue.Price, "test")
				eventU, err := sdk.TypedEventToEvent(event)
				s.Require().NoError(err, "TypedEventToEvent(NewEventSetNetAssetValue)")
				expEvents = sdk.Events{eventU}
			} else {
				expErrs = append(expErrs, tc.expErr)
			}

			em := sdk.NewEventManager()
			ctx := s.FreshCtx().WithEventManager(em)
			var err error
			testFunc := func() {
				err = s.app.MetadataKeeper.SetNetAssetValue(ctx, tc.scopeID, tc.netAssetValue, tc.source)
			}
			s.Require().NotPanics(testFunc, "SetNetAssetValue")
			assertions.AssertErrorContents(s.T(), err, expErrs, "SetNetAssetValue error")
			actEvents := em.Events()
			assertions.AssertEqualEvents(s.T(), expEvents, actEvents, "events emitted during SetNetAssetValue")
		})
	}
}

func (s *ScopeKeeperTestSuite) TestRemoveNetAssetValues() {
	scopeID := types.ScopeMetadataAddress(uuid.New())
	tests := []struct {
		name          string
		scopeID       types.MetadataAddress
		netAssetValue types.NetAssetValue
		expErr        string
	}{
		{
			name:    "Valid Removal",
			scopeID: scopeID,
			netAssetValue: types.NetAssetValue{
				Price: sdk.Coin{
					Denom:  "usd",
					Amount: sdkmath.NewInt(1000),
				},
			},
			expErr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			ctx := s.FreshCtx()
			err := s.app.MetadataKeeper.SetNetAssetValue(ctx, tc.scopeID, tc.netAssetValue, "test")
			s.Require().NoError(err, "SetNetAssetValue err")
			netAssetValues := []types.NetAssetValue{}
			err = s.app.MetadataKeeper.IterateNetAssetValues(ctx, tc.scopeID, func(state types.NetAssetValue) (stop bool) {
				netAssetValues = append(netAssetValues, state)
				return false
			})
			s.Require().NoError(err, "IterateNetAssetValues err")
			s.Require().Len(netAssetValues, 1, "Should have added a NAV")
			s.Require().Equal(tc.netAssetValue, netAssetValues[0], "Should have added the test case nave.")
			s.app.MetadataKeeper.RemoveNetAssetValues(ctx, tc.scopeID)
			netAssetValues = []types.NetAssetValue{}
			err = s.app.MetadataKeeper.IterateNetAssetValues(ctx, tc.scopeID, func(state types.NetAssetValue) (stop bool) {
				netAssetValues = append(netAssetValues, state)
				return false
			})
			s.Require().NoError(err, "IterateNetAssetValues err")
			s.Require().Len(netAssetValues, 0, "Should have removed NAV")
		})
	}
}
