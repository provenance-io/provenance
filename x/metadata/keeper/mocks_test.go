package keeper_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// This file houses mock stuff for use in unit tests.
// It has _test so that it's only available to unit tests, but it doesn't actually
// house any unit tests.

// ensure that the MockAuthzKeeper implements keeper.AuthzKeeper.
var _ keeper.AuthzKeeper = (*MockAuthzKeeper)(nil)

// MockAuthzKeeper is a mocked keeper.MockAuthzKeeper.
type MockAuthzKeeper struct {
	GetAuthorizationResults map[string]GetAuthorizationResult
	GetAuthorizationCalls   []*GetAuthorizationCall
	DeleteGrantResults      map[string]error
	DeleteGrantCalls        []*DeleteGrantCall
	SaveGrantResults        map[string]error
	SaveGrantCalls          []*SaveGrantCall
}

// NewMockAuthzKeeper creates a new empty MockAuthzKeeper.
// Usually followed by calls to WithGetAuthorizationResults, WithDeleteGrantResults, and/or WithSaveGrantResults.
func NewMockAuthzKeeper() *MockAuthzKeeper {
	return &MockAuthzKeeper{
		GetAuthorizationResults: make(map[string]GetAuthorizationResult),
		GetAuthorizationCalls:   nil,
		DeleteGrantResults:      make(map[string]error),
		DeleteGrantCalls:        nil,
		SaveGrantResults:        make(map[string]error),
		SaveGrantCalls:          nil,
	}
}

// WithGetAccountResults defines results to return from GetAuthorization in this MockAuthzKeeper and also returns it.
func (k *MockAuthzKeeper) WithGetAuthorizationResults(entries ...GetAuthorizationCall) *MockAuthzKeeper {
	for _, entry := range entries {
		k.GetAuthorizationResults[entry.Key()] = entry.Result
	}
	return k
}

// WithGetAccountResults defines results to return from DeleteGrant in this MockAuthzKeeper and also returns it.
func (k *MockAuthzKeeper) WithDeleteGrantResults(entries ...DeleteGrantCall) *MockAuthzKeeper {
	for _, entry := range entries {
		k.DeleteGrantResults[entry.Key()] = entry.Result
	}
	return k
}

// WithGetAccountResults defines results to return from SaveGrant in this MockAuthzKeeper and also returns it.
func (k *MockAuthzKeeper) WithSaveGrantResults(entries ...SaveGrantCall) *MockAuthzKeeper {
	for _, entry := range entries {
		k.SaveGrantResults[entry.Key()] = entry.Result
	}
	return k
}

// ClearResults clears previously recorded calls but leaves the desired results intact.
func (k *MockAuthzKeeper) ClearResults() {
	k.GetAuthorizationCalls = nil
	k.DeleteGrantCalls = nil
	k.SaveGrantCalls = nil
}

// GrantInfo is a common set of parameters provided to the authz keeper functions we care about.
type GrantInfo struct {
	Grantee sdk.AccAddress
	Granter sdk.AccAddress
	MsgType string
}

// Key gets the string to use as a map key for this info.
func (c GrantInfo) Key() string {
	return string(c.Grantee) + " " + string(c.Granter) + " " + c.MsgType
}

// GetAuthorizationResult is the stuff returned from a GetAuthorization call.
type GetAuthorizationResult struct {
	Auth authz.Authorization
	Exp  *time.Time
}

// GetAuthorizationCall has the inputs of GetAuthorization and the result associated with that input.
type GetAuthorizationCall struct {
	GrantInfo
	Result GetAuthorizationResult
}

// NewAcceptedGetAuthorizationCall returns a new GetAuthorizationCall that be accepted.
func NewAcceptedGetAuthorizationCall(grantee, granter sdk.AccAddress, msgTypeURL, authName string) GetAuthorizationCall {
	return GetAuthorizationCall{
		GrantInfo: GrantInfo{
			Grantee: grantee,
			Granter: granter,
			MsgType: msgTypeURL,
		},
		Result: GetAuthorizationResult{
			Auth: &MockAuthorization{
				Name:              authName,
				AcceptResponse:    authz.AcceptResponse{Accept: true},
				AcceptResponseErr: nil,
				AcceptCalls:       nil,
			},
			Exp: nil,
		},
	}
}

func NewNotFoundGetAuthorizationCall(grantee, granter sdk.AccAddress, msgTypeURL string) *GetAuthorizationCall {
	return &GetAuthorizationCall{
		GrantInfo: GrantInfo{
			Grantee: grantee,
			Granter: granter,
			MsgType: msgTypeURL,
		},
		Result: GetAuthorizationResult{
			Auth: nil,
			Exp:  nil,
		},
	}
}

// WithAcceptCalls updates the Result.Auth to expect Accept calls for the provided msgs.
// Panics if the Result.Auth is not a MockAuthorization.
func (c GetAuthorizationCall) WithAcceptCalls(msgs ...sdk.Msg) *GetAuthorizationCall {
	c.Result.Auth.(*MockAuthorization).WithAcceptCalls(msgs...)
	return &c
}

// DeleteGrantCall has the inputs of DeleteGrant and the result associated with that input.
type DeleteGrantCall struct {
	GrantInfo
	Result error
}

// SaveGrantCall has the inputs of SaveGrant and the result associated with that input.
type SaveGrantCall struct {
	Grantee sdk.AccAddress
	Granter sdk.AccAddress
	Auth    authz.Authorization
	Exp     *time.Time
	Result  error
}

// Key gets the string to use as a map key for these calls.
func (c SaveGrantCall) Key() string {
	args := GrantInfo{
		Grantee: c.Grantee,
		Granter: c.Granter,
		MsgType: "",
	}
	if c.Auth != nil {
		args.MsgType = c.Auth.MsgTypeURL()
	}
	return args.Key()
}

// GetAuthorization records that a GetAuthorization call has been made and returns the pre-defined value or nil.
func (k *MockAuthzKeeper) GetAuthorization(_ context.Context, grantee, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time) {
	call := &GetAuthorizationCall{
		GrantInfo: GrantInfo{
			Grantee: grantee,
			Granter: granter,
			MsgType: msgType,
		},
	}
	call.Result = k.GetAuthorizationResults[call.Key()]
	k.GetAuthorizationCalls = append(k.GetAuthorizationCalls, call)
	return call.Result.Auth, call.Result.Exp
}

// DeleteGrant records that a DeleteGrant call has been made and returns the pre-defined value or nil.
func (k *MockAuthzKeeper) DeleteGrant(_ context.Context, grantee, granter sdk.AccAddress, msgType string) error {
	call := &DeleteGrantCall{
		GrantInfo: GrantInfo{
			Grantee: grantee,
			Granter: granter,
			MsgType: msgType,
		},
	}
	call.Result = k.DeleteGrantResults[call.Key()]
	k.DeleteGrantCalls = append(k.DeleteGrantCalls, call)
	return call.Result
}

// SaveGrant records that a SaveGrant call has been made and returns the pre-defined value or nil.
func (k *MockAuthzKeeper) SaveGrant(_ context.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error {
	call := &SaveGrantCall{
		Grantee: grantee,
		Granter: granter,
		Auth:    authorization,
		Exp:     expiration,
	}
	call.Result = k.SaveGrantResults[call.Key()]
	k.SaveGrantCalls = append(k.SaveGrantCalls, call)
	return call.Result
}

// ensure that the MockAuthKeeper implements keeper.AuthKeeper.
var _ keeper.AuthKeeper = (*MockAuthKeeper)(nil)

// MockAuthKeeper is a mocked keeper.AuthKeeper.
type MockAuthKeeper struct {
	GetAccountResults map[string]sdk.AccountI
	GetAccountCalls   []*GetAccountCall
}

// NewMockAuthKeeper creates a new empty MockAuthKeeper.
// Usually followed by calls to WithGetAccountResults.
func NewMockAuthKeeper() *MockAuthKeeper {
	return &MockAuthKeeper{
		GetAccountResults: make(map[string]sdk.AccountI),
		GetAccountCalls:   nil,
	}
}

// ClearResults clears previously recorded calls but leaves the desired results intact.
func (k *MockAuthKeeper) ClearResults() {
	k.GetAccountCalls = nil
}

// WithGetAccountResults defines results to return from GetAccount in this MockAuthKeeper and also returns it.
func (k *MockAuthKeeper) WithGetAccountResults(entries ...*GetAccountCall) *MockAuthKeeper {
	for _, entry := range entries {
		k.GetAccountResults[entry.Key()] = entry.Result
	}
	return k
}

// GetAccountCall has the inputs of GetAccount and the result associated with that input.
type GetAccountCall struct {
	Addr   sdk.AccAddress
	Result sdk.AccountI
}

func NewGetAccountCall(addr sdk.AccAddress, result sdk.AccountI) *GetAccountCall {
	return &GetAccountCall{
		Addr:   addr,
		Result: result,
	}
}

// NewWasmGetAccountCall returns a new GetAccountCall of an account that will return true from isWasmAccount.
func NewWasmGetAccountCall(addr sdk.AccAddress) *GetAccountCall {
	return NewGetAccountCall(addr, authtypes.NewBaseAccount(addr, nil, 0, 0))
}

// NewBaseGetAccountCall returns a new GetAccountCall for a base account that will return false from isWasmAccount.
func NewBaseGetAccountCall(addr sdk.AccAddress) *GetAccountCall {
	return NewGetAccountCall(addr, authtypes.NewBaseAccount(addr, nil, 0, 1))
}

// Key gets the string to use as a map key for these calls.
func (c GetAccountCall) Key() string {
	return string(c.Addr)
}

// GetAccount records that a GetAccount call has been made and returns the pre-defined value or nil.
func (k *MockAuthKeeper) GetAccount(_ context.Context, addr sdk.AccAddress) sdk.AccountI {
	call := &GetAccountCall{
		Addr:   addr,
		Result: nil,
	}
	call.Result = k.GetAccountResults[call.Key()]
	k.GetAccountCalls = append(k.GetAccountCalls, call)
	return call.Result
}

// ensure that the MockAuthorization implements authz.Authorization.
var _ authz.Authorization = (*MockAuthorization)(nil)

// MockAuthorization is a mocked authz.Authorization.
type MockAuthorization struct {
	Name              string
	AcceptResponse    authz.AcceptResponse
	AcceptResponseErr error
	AcceptCalls       []sdk.Msg
}

func NewMockAuthorization(name string, resp authz.AcceptResponse, err error) *MockAuthorization {
	return &MockAuthorization{
		Name:              name,
		AcceptResponse:    resp,
		AcceptResponseErr: err,
		AcceptCalls:       nil,
	}
}

func (a *MockAuthorization) WithAcceptCalls(calls ...sdk.Msg) *MockAuthorization {
	a.AcceptCalls = append(a.AcceptCalls, calls...)
	return a
}

// Accept records that an Accept call has been made and returns the pre-defined values.
func (a *MockAuthorization) Accept(_ context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	a.AcceptCalls = append(a.AcceptCalls, msg)
	return a.AcceptResponse, a.AcceptResponseErr
}

// MsgTypeURL returns this MockAuthorization's name and satisfies the authz.Authorization interface.
func (a *MockAuthorization) MsgTypeURL() string {
	return a.Name
}

// ValidateBasic panics but satisfies the authz.Authorization interface.
func (a *MockAuthorization) ValidateBasic() error {
	panic("MockAuthorization#ValidateBasic not implemented")
}

// Reset panics but satisfies the authz.Authorization interface.
func (a *MockAuthorization) Reset() {
	panic("MockAuthorization#Reset not implemented")
}

// String returns a string of this MockAuthorization and satisfies the authz.Authorization interface.
func (a *MockAuthorization) String() string {
	_ = MockAuthorization{
		Name:              "",
		AcceptResponse:    authz.AcceptResponse{},
		AcceptResponseErr: nil,
		AcceptCalls:       nil,
	}
	return fmt.Sprintf("MockAuthorization{%q, %v, %v, %d}", a.Name, a.AcceptResponse, a.AcceptResponseErr, a.AcceptCalls)
}

// ProtoMessage satisfies the authz.Authorization interface.
func (a *MockAuthorization) ProtoMessage() {}

// ensure that the MockBankKeeper implements keeper.BankKeeper.
var _ keeper.BankKeeper = (*MockBankKeeper)(nil)

// MockBankKeeper is a mocked keeper.BankKeeper.
type MockBankKeeper struct {
	BlockedAddrResults map[string]bool
	MintCoinsResults   []string
	BurnCoinsResults   []string
	SendCoinsResults   map[string]string
	DenomOwnerResults  map[string]DenomOwnerResult

	Calls BankKeeperCalls
}

// BankKeeperCalls contains records of calls made to the mock bank keeper.
type BankKeeperCalls struct {
	BlockedAddr []sdk.AccAddress
	MintCoins   []*MintBurnCall
	BurnCoins   []*MintBurnCall
	SendCoins   []*SendCoinsCall
	DenomOwner  []string
}

// NewMockBankKeeper creates a new MockBankKeeper.
// Usually followed by calls to WithBlockedAddr, WithMintCoinsErrors, WithBurnCoinsErrors,
// SendCoinsErrors, WithDenomOwnerResult, and/or WithDenomOwnerError.
func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		BlockedAddrResults: make(map[string]bool),
		SendCoinsResults:   make(map[string]string),
		DenomOwnerResults:  make(map[string]DenomOwnerResult),
	}
}

// WithBlockedAddr makes the provided addr report as blocked.
func (k *MockBankKeeper) WithBlockedAddr(addr sdk.AccAddress) *MockBankKeeper {
	k.BlockedAddrResults[string(addr)] = true
	return k
}

// WithMintCoinsErrors queues up the provided strings as errors to return from MintCoins.
// An entry of "" means no error will be returned for that entry.
func (k *MockBankKeeper) WithMintCoinsErrors(errs ...string) *MockBankKeeper {
	k.MintCoinsResults = append(k.MintCoinsResults, errs...)
	return k
}

// WithMintCoinsErrors queues up the provided strings as errors to return from BurnCoins.
// An entry of "" means no error will be returned for that entry.
func (k *MockBankKeeper) WithBurnCoinsErrors(errs ...string) *MockBankKeeper {
	k.BurnCoinsResults = append(k.BurnCoinsResults, errs...)
	return k
}

// SendCoinsErrors makes the SendCoins return the provided err for the given fromAddr.
// An err of "" means no error will be returned for that fromAddr.
func (k *MockBankKeeper) WithSendCoinsError(fromAddr sdk.AccAddress, err string) *MockBankKeeper {
	k.SendCoinsResults[string(fromAddr)] = err
	return k
}

// WithDenomOwnerResult makes DenomOwner return the given accAddr for the given scope (with nil error).
func (k *MockBankKeeper) WithDenomOwnerResult(mdAddr types.MetadataAddress, accAddr sdk.AccAddress) *MockBankKeeper {
	k.DenomOwnerResults[mdAddr.Denom()] = DenomOwnerResult{Owner: accAddr}
	return k
}

// WithDenomOwnerError makes DenomOwner return the given err for the given scope (with nil AccAddress).
func (k *MockBankKeeper) WithDenomOwnerError(mdAddr types.MetadataAddress, err string) *MockBankKeeper {
	k.DenomOwnerResults[mdAddr.Denom()] = DenomOwnerResult{Err: err}
	return k
}

// AssertCalls asserts that all calls made using this bank keeper are equal to the provided expected calls.
func (k *MockBankKeeper) AssertCalls(t *testing.T, exp BankKeeperCalls) bool {
	t.Helper()
	rv := k.AssertBlockedAddrCalls(t, exp.BlockedAddr)
	rv = k.AssertMintCoinsCalls(t, exp.MintCoins) && rv
	rv = k.AssertBurnCoinsCalls(t, exp.BurnCoins) && rv
	rv = k.AssertSendCoinsCalls(t, exp.SendCoins) && rv
	rv = k.AssertDenomOwnerCalls(t, exp.DenomOwner) && rv
	return rv
}

// AssertBlockedAddrCalls asserts that calls made to BlockedAddr are as expected.
func (k *MockBankKeeper) AssertBlockedAddrCalls(t *testing.T, exp []sdk.AccAddress) bool {
	t.Helper()
	act := k.Calls.BlockedAddr
	if assert.Equal(t, exp, act, "Addrs provided to BlockedAddr") {
		return true
	}
	expStrs := addrsCastToStrings(exp)
	actStrs := addrsCastToStrings(act)
	assert.Equal(t, expStrs, actStrs, "Addrs (as strings) provided to BlockedAddr")
	return false
}

// AssertMintCoinsCalls asserts that calls made to MintCoins are as expected.
func (k *MockBankKeeper) AssertMintCoinsCalls(t *testing.T, exp []*MintBurnCall) bool {
	t.Helper()
	act := k.Calls.MintCoins
	if assert.Equal(t, exp, act, "Calls made to MintCoins") {
		return true
	}
	expStrs := mapToStrings(exp)
	actStrs := mapToStrings(act)
	assert.Equal(t, expStrs, actStrs, "Calls (as strings) made to MintCoins")
	return false
}

// AssertBurnCoinsCalls asserts that calls made to BurnCoins are as expected.
func (k *MockBankKeeper) AssertBurnCoinsCalls(t *testing.T, exp []*MintBurnCall) bool {
	t.Helper()
	act := k.Calls.BurnCoins
	if assert.Equal(t, exp, act, "Calls made to BurnCoins") {
		return true
	}
	expStrs := mapToStrings(exp)
	actStrs := mapToStrings(act)
	assert.Equal(t, expStrs, actStrs, "Calls (as strings) made to BurnCoins")
	return false
}

// AssertSendCoinsCalls asserts that calls made to SendCoins are as expected.
func (k *MockBankKeeper) AssertSendCoinsCalls(t *testing.T, exp []*SendCoinsCall) bool {
	t.Helper()
	act := k.Calls.SendCoins
	if assert.Equal(t, exp, act, "Calls made to SendCoins") {
		return true
	}
	expStrs := mapToStrings(exp)
	actStrs := mapToStrings(act)
	assert.Equal(t, expStrs, actStrs, "Calls (as strings) made to SendCoins")
	return false
}

// AssertDenomOwnerCalls asserts that calls made to DenomOwner are as expected.
func (k *MockBankKeeper) AssertDenomOwnerCalls(t *testing.T, exp []string) bool {
	t.Helper()
	return assert.Equal(t, exp, k.Calls.DenomOwner, "Calls made to DenomOwner")
}

// MintBurnCall is the args provided to either MintCoins or BurnCoins.
type MintBurnCall struct {
	ModuleName string
	Coins      sdk.Coins
}

func (c MintBurnCall) String() string {
	return c.ModuleName + "_" + c.Coins.String()
}

func NewMintBurnCall(moduleName string, coins sdk.Coins) *MintBurnCall {
	return &MintBurnCall{
		ModuleName: moduleName,
		Coins:      coins,
	}
}

// SendCoinsCall is the args provided to SendCoins.
type SendCoinsCall struct {
	FromAddr sdk.AccAddress
	ToAddr   sdk.AccAddress
	Amt      sdk.Coins
}

func (c SendCoinsCall) String() string {
	return string(c.FromAddr) + "->" + string(c.ToAddr) + ":" + c.Amt.String()
}

func NewSendCoinsCall(fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) *SendCoinsCall {
	return &SendCoinsCall{
		FromAddr: fromAddr,
		ToAddr:   toAddr,
		Amt:      amt,
	}
}

// DenomOwnerResult is the args returned by DenomOwner.
type DenomOwnerResult struct {
	Owner sdk.AccAddress
	Err   string
}

func (k *MockBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	k.Calls.BlockedAddr = append(k.Calls.BlockedAddr, addr)
	return k.BlockedAddrResults[string(addr)]
}

func (k *MockBankKeeper) MintCoins(_ context.Context, moduleName string, amt sdk.Coins) error {
	k.Calls.MintCoins = append(k.Calls.MintCoins, NewMintBurnCall(moduleName, amt))
	if len(k.MintCoinsResults) > 0 {
		err := k.MintCoinsResults[0]
		k.MintCoinsResults = k.MintCoinsResults[1:]
		if len(err) > 0 {
			return errors.New(err)
		}
	}
	return nil
}

func (k *MockBankKeeper) BurnCoins(_ context.Context, moduleName string, amt sdk.Coins) error {
	k.Calls.BurnCoins = append(k.Calls.BurnCoins, NewMintBurnCall(moduleName, amt))
	if len(k.BurnCoinsResults) > 0 {
		err := k.BurnCoinsResults[0]
		k.BurnCoinsResults = k.BurnCoinsResults[1:]
		if len(err) > 0 {
			return errors.New(err)
		}
	}
	return nil
}

func (k *MockBankKeeper) SendCoins(_ context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	k.Calls.SendCoins = append(k.Calls.SendCoins, NewSendCoinsCall(fromAddr, toAddr, amt))
	err := k.SendCoinsResults[string(fromAddr)]
	if len(err) > 0 {
		return errors.New(err)
	}
	return nil
}

func (k *MockBankKeeper) SpendableCoin(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	panic("not implemented")
}

func (k *MockBankKeeper) DenomOwner(_ context.Context, denom string) (sdk.AccAddress, error) {
	k.Calls.DenomOwner = append(k.Calls.DenomOwner, denom)
	result, found := k.DenomOwnerResults[denom]
	if found {
		if len(result.Err) > 0 {
			return nil, errors.New(result.Err)
		}
		return result.Owner, nil
	}
	return nil, nil
}

func (k *MockBankKeeper) GetBalancesCollection() *collections.IndexedMap[collections.Pair[sdk.AccAddress, string], sdkmath.Int, bankkeeper.BalancesIndexes] {
	panic("not implemented")
}

func addrsCastToStrings(addrs []sdk.AccAddress) []string {
	if addrs == nil {
		return nil
	}
	rv := make([]string, len(addrs))
	for i, v := range addrs {
		rv[i] = string(v)
	}
	return rv
}

func mapToStrings[S ~[]E, E fmt.Stringer](vals S) []string {
	if vals == nil {
		return nil
	}
	rv := make([]string, len(vals))
	for i, v := range vals {
		rv[i] = v.String()
	}
	return rv
}
