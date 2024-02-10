package keeper_test

import (
	"errors"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/marker/types"
)

// WrappedBankKeeper wraps a BankKeeper such that some functions can be set up to return specific things instead.
type WrappedBankKeeper struct {
	types.BankKeeper
	SendCoinsErrs     []string
	ExtraBlockedAddrs []sdk.AccAddress
}

var _ types.BankKeeper = (*WrappedBankKeeper)(nil)

// NewWrappedBankKeeper creates a new WrappedBankKeeper.
// You'll need to call WithParent on the result before anything will work in here.
func NewWrappedBankKeeper() *WrappedBankKeeper {
	return &WrappedBankKeeper{}
}

// WithParent sets the parent bank keeper for this wrapping.
func (w *WrappedBankKeeper) WithParent(bankKeeper types.BankKeeper) *WrappedBankKeeper {
	w.BankKeeper = bankKeeper
	return w
}

// WithSendCoinsErrs adds the provided error strings to the list of errors that will be returned by SendCoins.
// A non-empty entry will be returned as an error (when its time comes).
// An empty entry (or if there aren't any entries left when SendCoins is called) will
// result in the parent's SendCoins function being called and returned.
func (w *WrappedBankKeeper) WithSendCoinsErrs(errs ...string) *WrappedBankKeeper {
	w.SendCoinsErrs = append(w.SendCoinsErrs, errs...)
	return w
}

// WithExtraBlockedAddrs adds the provided addresses to the list of addresses that this WrappedBankKeeper will return
// BlockedAddr = true for. These are on top of any that the parent bank keeper already has.
func (w *WrappedBankKeeper) WithExtraBlockedAddrs(addrs ...sdk.AccAddress) *WrappedBankKeeper {
	w.ExtraBlockedAddrs = append(w.ExtraBlockedAddrs, addrs...)
	return w
}

// SendCoins either returns a pre-defined error, or, if there isn't one, calls SendCoins on the parent.
func (w *WrappedBankKeeper) SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if len(w.SendCoinsErrs) > 0 {
		rv := w.SendCoinsErrs[0]
		w.SendCoinsErrs = w.SendCoinsErrs[1:]
		if len(rv) > 0 {
			return errors.New(rv)
		}
	}
	return w.BankKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
}

// BlockedAddr returns true if the address is in the list of extra blocked addresses.
// Otherwise, it calls BlockedAddr on the parent.
func (w *WrappedBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	for _, extra := range w.ExtraBlockedAddrs {
		if addr.Equals(extra) {
			return true
		}
	}
	return w.BankKeeper.BlockedAddr(addr)
}

// MockAuthzKeeper satisfies the AuthzKeeper interface but only returns what it's pre-defined to return.
type MockAuthzKeeper struct {
	GetAuthorizationResult *GetAuthorizationResult
	DeleteGrantResult      string
	SaveGrantResult        string
}

var _ types.AuthzKeeper = (*MockAuthzKeeper)(nil)

type GetAuthorizationResult struct {
	Authorization authz.Authorization
	Expiration    *time.Time
}

func NewMockAuthzKeeper() *MockAuthzKeeper {
	return &MockAuthzKeeper{}
}

// WithAuthzHandlerNoAuth sets up this MockAuthzKeeper such that MarkerKeeper.authzHandler will return
// without an error.
// This method also returns the updated MockAuthzKeeper.
func (m *MockAuthzKeeper) WithAuthzHandlerSuccess() *MockAuthzKeeper {
	m.GetAuthorizationResult = &GetAuthorizationResult{
		Authorization: newMockAuthorization().ToAccept(),
	}
	return m
}

// WithAuthzHandlerNoAuth sets up this MockAuthzKeeper such that MarkerKeeper.authzHandler will return
// an error that an authorization doesn't exist.
// This method also returns the updated MockAuthzKeeper.
func (m *MockAuthzKeeper) WithAuthzHandlerNoAuth() *MockAuthzKeeper {
	m.GetAuthorizationResult = nil
	return m
}

// WithAuthzHandlerNoAuth sets up this MockAuthzKeeper such that MarkerKeeper.authzHandler will return
// the provided error (as if the authorization.Accept method returned it).
// This method also returns the updated MockAuthzKeeper.
func (m *MockAuthzKeeper) WithAuthzHandlerAcceptError(err string) *MockAuthzKeeper {
	m.GetAuthorizationResult = &GetAuthorizationResult{
		Authorization: newMockAuthorization().WithError(err),
	}
	return m
}

// WithAuthzHandlerNoAuth sets up this MockAuthzKeeper such that MarkerKeeper.authzHandler will return
// an error that the authorization was not accepted.
// This method also returns the updated MockAuthzKeeper.
func (m *MockAuthzKeeper) WithAuthzHandlerNoAccept() *MockAuthzKeeper {
	m.GetAuthorizationResult = &GetAuthorizationResult{
		Authorization: newMockAuthorization(),
	}
	return m
}

// WithAuthzHandlerNoAuth sets up this MockAuthzKeeper such that MarkerKeeper.authzHandler will return
// the provided error (as if the DeleteGrant method returned it).
// This method also returns the updated MockAuthzKeeper.
func (m *MockAuthzKeeper) WithAuthzHandlerDeleteError(err string) *MockAuthzKeeper {
	m.GetAuthorizationResult = &GetAuthorizationResult{
		Authorization: newMockAuthorization().ToDelete(),
	}
	m.DeleteGrantResult = err
	return m
}

// WithAuthzHandlerNoAuth sets up this MockAuthzKeeper such that MarkerKeeper.authzHandler will return
// the provided error (as if the SaveGrant method returned it).
// This method also returns the updated MockAuthzKeeper.
func (m *MockAuthzKeeper) WithAuthzHandlerUpdateError(err string) *MockAuthzKeeper {
	m.GetAuthorizationResult = &GetAuthorizationResult{
		Authorization: newMockAuthorization().ToUpdate(),
	}
	m.SaveGrantResult = err
	return m
}

func (m *MockAuthzKeeper) GetAuthorization(_ sdk.Context, _ sdk.AccAddress, _ sdk.AccAddress, _ string) (authz.Authorization, *time.Time) {
	if m.GetAuthorizationResult != nil {
		return m.GetAuthorizationResult.Authorization, m.GetAuthorizationResult.Expiration
	}
	return nil, nil
}

func (m *MockAuthzKeeper) DeleteGrant(_ sdk.Context, _ sdk.AccAddress, _ sdk.AccAddress, _ string) error {
	if len(m.DeleteGrantResult) > 0 {
		return errors.New(m.DeleteGrantResult)
	}
	return nil
}

func (m *MockAuthzKeeper) SaveGrant(_ sdk.Context, _, _ sdk.AccAddress, _ authz.Authorization, _ *time.Time) error {
	if len(m.SaveGrantResult) > 0 {
		return errors.New(m.SaveGrantResult)
	}
	return nil
}

// mockAuthorization satisfies the authz.Authorization interface and does what it's told.
type mockAuthorization struct {
	RespAccept  bool
	RespDelete  bool
	RespUpdated bool
	RespErr     string
}

var _ authz.Authorization = (*mockAuthorization)(nil)

// newMockAuthorization creates a new mock authorization.
func newMockAuthorization() *mockAuthorization {
	return &mockAuthorization{}
}

// ToAccept sets this mockAuthorization up to return Accept = true in the accept response.
func (a *mockAuthorization) ToAccept() *mockAuthorization {
	a.RespAccept = true
	return a
}

// ToAccept sets this mockAuthorization up to return Delete = true in the accept response.
func (a *mockAuthorization) ToDelete() *mockAuthorization {
	a.RespDelete = true
	return a
}

// ToAccept sets this mockAuthorization up to return itself in the accept response's Updated field.
func (a *mockAuthorization) ToUpdate() *mockAuthorization {
	a.RespUpdated = true
	return a
}

// ToAccept sets this mockAuthorization up to the provided error from Accept.
func (a *mockAuthorization) WithError(err string) *mockAuthorization {
	a.RespErr = err
	return a
}

// Accept just returns everything it was defined to return.
func (a *mockAuthorization) Accept(_ sdk.Context, _ sdk.Msg) (authz.AcceptResponse, error) {
	resp := authz.AcceptResponse{
		Accept: a.RespAccept,
		Delete: a.RespDelete,
	}
	if a.RespUpdated {
		resp.Updated = a
	}
	var err error
	if len(a.RespErr) > 0 {
		err = errors.New(a.RespErr)
	}
	return resp, err
}

// MsgTypeURL returns "mockAuthorization". Satisfies the authz.Authorization interface.
func (a *mockAuthorization) MsgTypeURL() string {
	return "mockAuthorization"
}

// ValidateBasic returns nil. Satisfies the authz.Authorization interface.
func (a *mockAuthorization) ValidateBasic() error {
	return nil
}

// Reset does nothing. Satisfies the authz.Authorization interface.
func (a *mockAuthorization) Reset() {}

// String returns a string representation of this mockAuthorization.
func (a *mockAuthorization) String() string {
	return fmt.Sprintf("mockAuthorization{Accept=%t,Delete=%t,Update=%t,Err=%q}",
		a.RespAccept, a.RespDelete, a.RespUpdated, a.RespErr)
}

// ProtoMessage does nothing. Satisfies the authz.Authorization interface.
func (a *mockAuthorization) ProtoMessage() {}

// WrappedAttrKeeper wraps an AttrKeeper such that some functions can be set up to return specific things instead.
type WrappedAttrKeeper struct {
	types.AttrKeeper
	GetAllAttributesAddrErrs []string
}

var _ types.AttrKeeper = (*WrappedAttrKeeper)(nil)

// NewWrappedBankKeeper creates a new WrappedBankKeeper.
// You'll need to call WithParent on the result before anything will work in here.
func NewWrappedAttrKeeper() *WrappedAttrKeeper {
	return &WrappedAttrKeeper{}
}

// WithParent sets the parent bank keeper for this wrapping.
func (w *WrappedAttrKeeper) WithParent(attrKeeper types.AttrKeeper) *WrappedAttrKeeper {
	w.AttrKeeper = attrKeeper
	return w
}

// WithGetAllAttributesAddrErrs adds the provided error strings to the list of errors that will be
// returned by GetAllAttributesAddr. A non-empty entry will be returned as an error (when its time comes).
// An empty entry (or if there aren't any entries left when SendCoins is called) will
// result in the parent's SendCoins function being called and returned.
func (w *WrappedAttrKeeper) WithGetAllAttributesAddrErrs(errs ...string) *WrappedAttrKeeper {
	w.GetAllAttributesAddrErrs = append(w.GetAllAttributesAddrErrs, errs...)
	return w
}

// GetAllAttributesAddr either returns a pre-defined error, or, if there isn't one, calls GetAllAttributesAddr on the parent.
func (w *WrappedAttrKeeper) GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]attrtypes.Attribute, error) {
	if len(w.GetAllAttributesAddrErrs) > 0 {
		rv := w.GetAllAttributesAddrErrs[0]
		w.GetAllAttributesAddrErrs = w.GetAllAttributesAddrErrs[1:]
		if len(rv) > 0 {
			return nil, errors.New(rv)
		}
	}
	return w.AttrKeeper.GetAllAttributesAddr(ctx, addr)
}
