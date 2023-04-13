package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/provenance-io/provenance/x/metadata/keeper"
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
func (k *MockAuthzKeeper) GetAuthorization(_ sdk.Context, grantee, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time) {
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
func (k *MockAuthzKeeper) DeleteGrant(_ sdk.Context, grantee, granter sdk.AccAddress, msgType string) error {
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
func (k *MockAuthzKeeper) SaveGrant(_ sdk.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error {
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
	GetAccountResults map[string]authtypes.AccountI
	GetAccountCalls   []*GetAccountCall
}

// NewMockAuthKeeper creates a new empty MockAuthKeeper.
// Usually followed by calls to WithGetAccountResults.
func NewMockAuthKeeper() *MockAuthKeeper {
	return &MockAuthKeeper{
		GetAccountResults: make(map[string]authtypes.AccountI),
		GetAccountCalls:   nil,
	}
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
	Result authtypes.AccountI
}

func NewGetAccountCall(addr sdk.AccAddress, result authtypes.AccountI) *GetAccountCall {
	return &GetAccountCall{
		Addr:   addr,
		Result: result,
	}
}

// Key gets the string to use as a map key for these calls.
func (c GetAccountCall) Key() string {
	return string(c.Addr)
}

// GetAccount records that a GetAccount call has been made and returns the pre-defined value or nil.
func (k *MockAuthKeeper) GetAccount(_ sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
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
func (a *MockAuthorization) Accept(_ sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
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
