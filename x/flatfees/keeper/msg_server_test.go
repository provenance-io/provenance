package keeper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/flatfees/keeper"
	"github.com/provenance-io/provenance/x/flatfees/types"
)

func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

type MsgServerTestSuite struct {
	suite.Suite

	app *simapp.App
	ctx sdk.Context
}

func (s *MsgServerTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(true)
}

const authority = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" // should equal s.app.FlatFeesKeeper.GetAuthority().

// MockKeeper is a fake x/flatfees Keeper for use in the MsgServer.
type MockKeeper struct {
	ValidateAuthorityErrs []string
	ValidateAuthorityExp  []string
	ValidateAuthorityArgs []string

	SetParamsErrs []string
	SetParamsExp  []types.Params
	SetParamsArgs []types.Params

	SetMsgFeeErrs []string
	SetMsgFeeExp  []*types.MsgFee
	SetMsgFeeArgs []*types.MsgFee

	RemoveMsgFeeErrs []string
	RemoveMsgFeeExp  []string
	RemoveMsgFeeArgs []string

	SetConversionFactorErrs []string
	SetConversionFactorExp  []types.ConversionFactor
	SetConversionFactorArgs []types.ConversionFactor
}

var _ keeper.MsgKeeper = (*MockKeeper)(nil)

// NewMockKeeper creates a new (flatfees) MockKeeper for use in the msg-server tests.
func NewMockKeeper() *MockKeeper {
	return &MockKeeper{}
}

// WithValidateAuthorityErrs adds the provided errs to be returned from ValidateAuthority.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithValidateAuthorityErrs(errs ...string) *MockKeeper {
	k.ValidateAuthorityErrs = append(k.ValidateAuthorityErrs, errs...)
	return k
}

// WithSetParamsErrs adds the provided errs to be returned from SetParams.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithSetParamsErrs(errs ...string) *MockKeeper {
	k.SetParamsErrs = append(k.SetParamsErrs, errs...)
	return k
}

// WithSetMsgFeeErrs adds the provided errs to be returned from SetMsgFee.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithSetMsgFeeErrs(errs ...string) *MockKeeper {
	k.SetMsgFeeErrs = append(k.SetMsgFeeErrs, errs...)
	return k
}

// WithRemoveMsgFeeErrs adds the provided errs to be returned from RemoveMsgFee.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithRemoveMsgFeeErrs(errs ...string) *MockKeeper {
	k.RemoveMsgFeeErrs = append(k.RemoveMsgFeeErrs, errs...)
	return k
}

// WithSetConversionFactorErrs adds the provided errs to be returned from SetConversionFactor.
// An empty string indicates no error. This method both updates the receiver and returns it.
func (k *MockKeeper) WithSetConversionFactorErrs(errs ...string) *MockKeeper {
	k.SetConversionFactorErrs = append(k.SetConversionFactorErrs, errs...)
	return k
}

// WithExpValidateAuthority adds the provided authorities to the list of expected calls to ValidateAuthority.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpValidateAuthority(authorities ...string) *MockKeeper {
	k.ValidateAuthorityExp = append(k.ValidateAuthorityExp, authorities...)
	return k
}

// WithExpSetParams adds the provided params to the list of expected calls to SetParams.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpSetParams(params ...types.Params) *MockKeeper {
	k.SetParamsExp = append(k.SetParamsExp, params...)
	return k
}

// WithExpSetMsgFee adds the provided msgFees to the list of expected calls to SetMsgFee.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpSetMsgFee(msgFees ...*types.MsgFee) *MockKeeper {
	k.SetMsgFeeExp = append(k.SetMsgFeeExp, msgFees...)
	return k
}

// WithExpRemoveMsgFee adds the provided msgTypeURLs to the list of expected calls to RemoveMsgFee.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpRemoveMsgFee(msgTypeURLs ...string) *MockKeeper {
	k.RemoveMsgFeeExp = append(k.RemoveMsgFeeExp, msgTypeURLs...)
	return k
}

// WithExpSetConversionFactorE adds the provided conversion factors to the list of expected calls to SetConversionFactor.
// This method both updates the receiver and returns it.
func (k *MockKeeper) WithExpSetConversionFactor(conversionFactors ...types.ConversionFactor) *MockKeeper {
	k.SetConversionFactorExp = append(k.SetConversionFactorExp, conversionFactors...)
	return k
}

// shiftErr removes the first entry from errs. If it's not an empty string, it's converted to an error and also returned.
// If errs is empty, or the first entry is an empty string, no error is returned (but the 1st entry is still removed).
func shiftErr(errs []string) ([]string, error) {
	var err error
	if len(errs) > 0 {
		errMsg := errs[0]
		errs = errs[1:]
		switch {
		case errMsg == "ErrMsgFeeDoesNotExist":
			err = types.ErrMsgFeeDoesNotExist
		case len(errMsg) > 0:
			err = errors.New(errMsg)
		}
	}
	return errs, err
}

func (k *MockKeeper) ValidateAuthority(authority string) error {
	k.ValidateAuthorityArgs = append(k.ValidateAuthorityArgs, authority)
	var err error
	k.ValidateAuthorityErrs, err = shiftErr(k.ValidateAuthorityErrs)
	return err
}

func (k *MockKeeper) SetParams(_ sdk.Context, params types.Params) error {
	k.SetParamsArgs = append(k.SetParamsArgs, params)
	var err error
	k.SetParamsErrs, err = shiftErr(k.SetParamsErrs)
	return err
}

func (k *MockKeeper) SetMsgFee(_ sdk.Context, msgFee types.MsgFee) error {
	k.SetMsgFeeArgs = append(k.SetMsgFeeArgs, &msgFee)
	var err error
	k.SetMsgFeeErrs, err = shiftErr(k.SetMsgFeeErrs)
	return err
}

func (k *MockKeeper) RemoveMsgFee(_ sdk.Context, msgType string) error {
	k.RemoveMsgFeeArgs = append(k.RemoveMsgFeeArgs, msgType)
	var err error
	k.RemoveMsgFeeErrs, err = shiftErr(k.RemoveMsgFeeErrs)
	return err
}

func (k *MockKeeper) SetConversionFactor(_ sdk.Context, conversionFactor types.ConversionFactor) error {
	k.SetConversionFactorArgs = append(k.SetConversionFactorArgs, conversionFactor)
	var err error
	k.SetConversionFactorErrs, err = shiftErr(k.SetConversionFactorErrs)
	return err
}

func (k *MockKeeper) AssertCalls(t testing.TB) bool {
	ok := assert.Equal(t, k.ValidateAuthorityExp, k.ValidateAuthorityArgs, "Calls to ValidateAuthority")
	if assert.Equal(t, len(k.SetParamsExp), len(k.SetParamsExp), "Number of calls to SetParams") {
		for i := range k.SetParamsExp {
			ok = assertEqualParams(t, k.SetParamsExp[i], k.SetParamsExp[i], "Call %d to SetParams", i+1) && ok
		}
	} else {
		ok = false
	}
	ok = assertEqualMsgFees(t, k.SetMsgFeeExp, k.SetMsgFeeArgs, "Calls to SetMsgFee") && ok
	ok = assert.Equal(t, k.RemoveMsgFeeExp, k.RemoveMsgFeeArgs, "Calls to RemoveMsgFee") && ok
	ok = assertEqualConversionFactors(t, k.SetConversionFactorExp, k.SetConversionFactorArgs, "Calls to SetConversionFactor") && ok
	return ok
}

func (s *MsgServerTestSuite) TestUpdateParams() {
	tests := []struct {
		name    string
		kpr     *MockKeeper
		req     *types.MsgUpdateParamsRequest
		expErr  string
		expCall bool // Automatically true if expErr is empty.
	}{
		{
			name: "invalid authority",
			kpr:  NewMockKeeper().WithValidateAuthorityErrs("injected validate authority error"),
			req: &types.MsgUpdateParamsRequest{
				Authority: "invalid",
				Params:    types.DefaultParams(),
			},
			expErr: "injected validate authority error",
		},
		{
			name: "error setting params",
			kpr:  NewMockKeeper().WithSetParamsErrs("just a fake error here"),
			req: &types.MsgUpdateParamsRequest{
				Authority: authority,
				Params:    types.DefaultParams(),
			},
			expErr:  "rpc error: code = InvalidArgument desc = just a fake error here",
			expCall: true,
		},
		{
			name: "okay: non-defaults",
			req: &types.MsgUpdateParamsRequest{
				Authority: authority,
				Params: types.Params{
					DefaultCost: sdk.NewInt64Coin("pink", 3_000),
					ConversionFactor: types.ConversionFactor{
						DefinitionAmount: sdk.NewInt64Coin("pink", 3),
						ConvertedAmount:  sdk.NewInt64Coin("orange", 1),
					},
				},
			},
		},
		{
			name: "okay: defaults",
			req: &types.MsgUpdateParamsRequest{
				Authority: authority,
				Params:    types.DefaultParams(),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.kpr == nil {
				tc.kpr = NewMockKeeper()
			}
			tc.kpr = tc.kpr.WithExpValidateAuthority(tc.req.Authority)

			var expResp, actResp *types.MsgUpdateParamsResponse
			if len(tc.expErr) == 0 {
				expResp = &types.MsgUpdateParamsResponse{}
				tc.expCall = true
			}

			if tc.expCall {
				tc.kpr = tc.kpr.WithExpSetParams(tc.req.Params)
			}

			msgServer := keeper.NewMsgServer(tc.kpr)

			var err error
			testFunc := func() {
				actResp, err = msgServer.UpdateParams(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "UpdateParams(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "UpdateParams(...) error")
			s.Assert().Equal(expResp, actResp, "UpdateParams(...) response")

			tc.kpr.AssertCalls(s.T())
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateConversionFactor() {
	tests := []struct {
		name    string
		kpr     *MockKeeper
		req     *types.MsgUpdateConversionFactorRequest
		expErr  string
		expCall bool // Automatically true if expErr is empty.
	}{
		{
			name: "incorrect authority",
			kpr:  NewMockKeeper().WithValidateAuthorityErrs("that is a naughty authority"),
			req: &types.MsgUpdateConversionFactorRequest{
				Authority: "whatever",
				ConversionFactor: types.ConversionFactor{
					DefinitionAmount: sdk.NewInt64Coin("green", 4),
					ConvertedAmount:  sdk.NewInt64Coin("orange", 16),
				},
			},
			expErr: "that is a naughty authority",
		},
		{
			name: "error setting conversion factor",
			kpr:  NewMockKeeper().WithSetConversionFactorErrs("notgonnaconvert"),
			req: &types.MsgUpdateConversionFactorRequest{
				Authority: sdk.AccAddress("whatever____________").String(),
				ConversionFactor: types.ConversionFactor{
					DefinitionAmount: sdk.NewInt64Coin("green", 4),
					ConvertedAmount:  sdk.NewInt64Coin("orange", 16),
				},
			},
			expErr:  "rpc error: code = InvalidArgument desc = notgonnaconvert",
			expCall: true,
		},
		{
			name: "okay: non-defaults",
			req: &types.MsgUpdateConversionFactorRequest{
				Authority: sdk.AccAddress("some_address________").String(),
				ConversionFactor: types.ConversionFactor{
					DefinitionAmount: sdk.NewInt64Coin("pink", 4),
					ConvertedAmount:  sdk.NewInt64Coin("fuchsia", 16),
				},
			},
			expCall: true,
		},
		{
			name: "okay: defaults",
			req: &types.MsgUpdateConversionFactorRequest{
				Authority:        sdk.AccAddress("some_address________").String(),
				ConversionFactor: types.DefaultParams().ConversionFactor,
			},
			expCall: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.kpr == nil {
				tc.kpr = NewMockKeeper()
			}
			tc.kpr = tc.kpr.WithExpValidateAuthority(tc.req.Authority)

			var expResp, actResp *types.MsgUpdateConversionFactorResponse
			if len(tc.expErr) == 0 {
				expResp = &types.MsgUpdateConversionFactorResponse{}
				tc.expCall = true
			}

			if tc.expCall {
				tc.kpr = tc.kpr.WithExpSetConversionFactor(tc.req.ConversionFactor)
			}

			msgServer := keeper.NewMsgServer(tc.kpr)

			var err error
			testFunc := func() {
				actResp, err = msgServer.UpdateConversionFactor(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "UpdateConversionFactor()")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "UpdateConversionFactor() error")
			s.Assert().Equal(expResp, actResp, "UpdateConversionFactor() response")

			tc.kpr.AssertCalls(s.T())
		})
	}
}

func (s *MsgServerTestSuite) TestUpdateMsgFees() {
	tests := []struct {
		name   string
		kpr    *MockKeeper
		req    *types.MsgUpdateMsgFeesRequest
		expErr string
	}{
		{
			name:   "invalid authority",
			kpr:    NewMockKeeper().WithValidateAuthorityErrs("just another unreal error"),
			req:    &types.MsgUpdateMsgFeesRequest{Authority: "invalid"},
			expErr: "just another unreal error",
		},
		{
			name: "one to set: okay",
			kpr:  NewMockKeeper().WithExpSetMsgFee(types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3))),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
				},
			},
		},
		{
			name: "one to set: error",
			kpr: NewMockKeeper().WithSetMsgFeeErrs("hot error injection").
				WithExpSetMsgFee(types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3))),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
				},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not set msg fee: hot error injection",
		},
		{
			name: "three to set: okay",
			kpr: NewMockKeeper().WithExpSetMsgFee(
				types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
				types.NewMsgFee("test2", sdk.NewInt64Coin("apple", 14)),
				types.NewMsgFee("test3", sdk.NewInt64Coin("cherry", 27)),
			),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
					types.NewMsgFee("test2", sdk.NewInt64Coin("apple", 14)),
					types.NewMsgFee("test3", sdk.NewInt64Coin("cherry", 27)),
				},
			},
		},
		{
			name: "three to set: error from first",
			kpr: NewMockKeeper().WithSetMsgFeeErrs("fake error").WithExpSetMsgFee(
				types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
			),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
					types.NewMsgFee("test2", sdk.NewInt64Coin("apple", 14)),
					types.NewMsgFee("test3", sdk.NewInt64Coin("cherry", 27)),
				},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not set msg fee: fake error",
		},
		{
			name: "three to set: error from second",
			kpr: NewMockKeeper().WithSetMsgFeeErrs("", "not really an error").WithExpSetMsgFee(
				types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
				types.NewMsgFee("test2", sdk.NewInt64Coin("apple", 14)),
			),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
					types.NewMsgFee("test2", sdk.NewInt64Coin("apple", 14)),
					types.NewMsgFee("test3", sdk.NewInt64Coin("cherry", 27)),
				},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not set msg fee: not really an error",
		},
		{
			name: "three to set: error from third",
			kpr: NewMockKeeper().WithSetMsgFeeErrs("", "", "another sham error").WithExpSetMsgFee(
				types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
				types.NewMsgFee("test2", sdk.NewInt64Coin("apple", 14)),
				types.NewMsgFee("test3", sdk.NewInt64Coin("cherry", 27)),
			),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test1", sdk.NewInt64Coin("banana", 3)),
					types.NewMsgFee("test2", sdk.NewInt64Coin("apple", 14)),
					types.NewMsgFee("test3", sdk.NewInt64Coin("cherry", 27)),
				},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not set msg fee: another sham error",
		},
		{
			name: "one to remove: okay",
			kpr:  NewMockKeeper().WithExpRemoveMsgFee("testurl"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToUnset:   []string{"testurl"},
			},
		},
		{
			name: "one to remove: error",
			kpr:  NewMockKeeper().WithRemoveMsgFeeErrs("phony error message").WithExpRemoveMsgFee("testurl"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToUnset:   []string{"testurl"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not remove msg fee: phony error message",
		},
		{
			name: "three to remove: okay",
			kpr:  NewMockKeeper().WithExpRemoveMsgFee("test1", "test2", "test3"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToUnset:   []string{"test1", "test2", "test3"},
			},
		},
		{
			name: "three to remove: error from first",
			kpr:  NewMockKeeper().WithRemoveMsgFeeErrs("this error is an imitation").WithExpRemoveMsgFee("test1"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToUnset:   []string{"test1", "test2", "test3"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not remove msg fee: this error is an imitation",
		},
		{
			name: "three to remove: error from second",
			kpr:  NewMockKeeper().WithRemoveMsgFeeErrs("", "another sham error").WithExpRemoveMsgFee("test1", "test2"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToUnset:   []string{"test1", "test2", "test3"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not remove msg fee: another sham error",
		},
		{
			name: "three to remove: error from third",
			kpr:  NewMockKeeper().WithRemoveMsgFeeErrs("", "", "knockoff error").WithExpRemoveMsgFee("test1", "test2", "test3"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToUnset:   []string{"test1", "test2", "test3"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not remove msg fee: knockoff error",
		},
		{
			name: "one of each: okay",
			kpr:  NewMockKeeper().WithExpSetMsgFee(types.NewMsgFee("test.a")).WithExpRemoveMsgFee("test.1"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet:     []*types.MsgFee{types.NewMsgFee("test.a")},
				ToUnset:   []string{"test.1"},
			},
		},
		{
			name: "one of each: error on set",
			kpr:  NewMockKeeper().WithExpSetMsgFee(types.NewMsgFee("test.a")).WithSetMsgFeeErrs("error for testing"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet:     []*types.MsgFee{types.NewMsgFee("test.a")},
				ToUnset:   []string{"test.1"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not set msg fee: error for testing",
		},
		{
			name: "one of each: error on remove",
			kpr: NewMockKeeper().WithExpSetMsgFee(types.NewMsgFee("test.a")).
				WithExpRemoveMsgFee("test.1").WithRemoveMsgFeeErrs("too yellow"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet:     []*types.MsgFee{types.NewMsgFee("test.a")},
				ToUnset:   []string{"test.1"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not remove msg fee: too yellow",
		},
		{
			name: "three of each: error from second set",
			kpr: NewMockKeeper().WithExpSetMsgFee(
				types.NewMsgFee("test.a", sdk.NewInt64Coin("apple", 4)),
				types.NewMsgFee("test.b", sdk.NewInt64Coin("banana", 8)),
			).WithSetMsgFeeErrs("", "this is crazy"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test.a", sdk.NewInt64Coin("apple", 4)),
					types.NewMsgFee("test.b", sdk.NewInt64Coin("banana", 8)),
					types.NewMsgFee("test.c", sdk.NewInt64Coin("cherry", 47)),
				},
				ToUnset: []string{"test.1", "test.2", "test.3"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not set msg fee: this is crazy",
		},
		{
			name: "three of each: error from second remove",
			kpr: NewMockKeeper().WithExpSetMsgFee(
				types.NewMsgFee("test.a", sdk.NewInt64Coin("apple", 4)),
				types.NewMsgFee("test.b", sdk.NewInt64Coin("banana", 8)),
				types.NewMsgFee("test.c", sdk.NewInt64Coin("cherry", 47)),
			).WithExpRemoveMsgFee("test.1", "test.2").WithRemoveMsgFeeErrs("", "not bananas enough"),
			req: &types.MsgUpdateMsgFeesRequest{
				Authority: authority,
				ToSet: []*types.MsgFee{
					types.NewMsgFee("test.a", sdk.NewInt64Coin("apple", 4)),
					types.NewMsgFee("test.b", sdk.NewInt64Coin("banana", 8)),
					types.NewMsgFee("test.c", sdk.NewInt64Coin("cherry", 47)),
				},
				ToUnset: []string{"test.1", "test.2", "test.3"},
			},
			expErr: "rpc error: code = InvalidArgument desc = could not remove msg fee: not bananas enough",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.kpr == nil {
				tc.kpr = NewMockKeeper()
			}
			tc.kpr = tc.kpr.WithExpValidateAuthority(tc.req.Authority)

			var expResp, actResp *types.MsgUpdateMsgFeesResponse
			if len(tc.expErr) == 0 {
				expResp = &types.MsgUpdateMsgFeesResponse{}
			}

			msgServer := keeper.NewMsgServer(tc.kpr)

			var err error
			testFunc := func() {
				actResp, err = msgServer.UpdateMsgFees(s.ctx, tc.req)
			}
			s.Require().NotPanics(testFunc, "UpdateMsgFees(...)")
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "UpdateMsgFees(...) error")
			s.Assert().Equal(expResp, actResp, "UpdateMsgFees(...) response")
			tc.kpr.AssertCalls(s.T())
		})
	}
}
