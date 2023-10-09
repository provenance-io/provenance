package keeper_test

import (
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/exchange"
)

// toStrings converts a slice to indexed strings using the provided stringer func.
func toStrings[T any](vals []T, stringer func(T) string) []string {
	if vals == nil {
		return nil
	}
	rv := make([]string, len(vals))
	for i, val := range vals {
		rv[i] = fmt.Sprintf("[%d]:%s", i, stringer(val))
	}
	return rv
}

// assertEqualSlice asserts that expected = actual and returns true if so.
// If not, returns false and the stringer is applied to each entry and the comparison
// is redone on the strings in the hopes that it helps identify the problem.
func assertEqualSlice[T any](s *TestSuite, expected, actual []T, stringer func(T) string, msg string, args ...interface{}) bool {
	if s.Assert().Equalf(expected, actual, msg, args...) {
		return true
	}
	// compare each as strings in the hopes that makes it easier to identify the problem.
	expStrs := toStrings(expected, stringer)
	actStrs := toStrings(actual, stringer)
	s.Assert().Equalf(expStrs, actStrs, "strings: "+msg, args...)
	return false
}

// MockAccountKeeper satisfies the exchange.AccountKeeper interface but just records the calls and allows dictation of results.
type MockAccountKeeper struct {
	GetAccountCalls      []sdk.AccAddress
	SetAccountCalls      []authtypes.AccountI
	HasAccountCalls      []sdk.AccAddress
	NewAccountCalls      []authtypes.AccountI
	GetAccountResultsMap map[string]authtypes.AccountI
	HasAccountResultsMap map[string]bool
	NewAccountResultsMap map[string]authtypes.AccountI
}

var _ exchange.AccountKeeper = (*MockAccountKeeper)(nil)

// NewMockAccountKeeper creates a new empty MockAccountKeeper.
// Follow it up with WithGetAccountResult, WithHasAccountResult,
// and/or WithNewAccountResult to dictate results.
func NewMockAccountKeeper() *MockAccountKeeper {
	return &MockAccountKeeper{
		GetAccountResultsMap: make(map[string]authtypes.AccountI),
		HasAccountResultsMap: make(map[string]bool),
		NewAccountResultsMap: make(map[string]authtypes.AccountI),
	}
}

// WithGetAccountResult associates the provided address and result for use with calls to GetAccount.
// When GetAccount is called, if the address provided has an entry here, that is returned, otherwise, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockAccountKeeper) WithGetAccountResult(addr sdk.AccAddress, result authtypes.AccountI) *MockAccountKeeper {
	k.GetAccountResultsMap[string(addr)] = result
	return k
}

// WithGetAccountResult associates the provided address and result for use with calls to HasAccount.
// When HasAccount is called, if the address provided has an entry here, that is returned, otherwise, false is returned.
// This method both updates the receiver and returns it.
func (k *MockAccountKeeper) WithHasAccountResult(addr sdk.AccAddress, result bool) *MockAccountKeeper {
	k.HasAccountResultsMap[string(addr)] = result
	return k
}

// WithGetAccountResult associates the provided address and result for use with calls to NewAccount.
// When NewAccount is called, if the address provided has an entry here, that is returned,
// otherwise, the provided AccountI is returned.
// This method both updates the receiver and returns it.
func (k *MockAccountKeeper) WithNewAccountResult(result authtypes.AccountI) *MockAccountKeeper {
	k.NewAccountResultsMap[string(result.GetAddress())] = result
	return k
}

func (k *MockAccountKeeper) GetAccount(_ sdk.Context, addr sdk.AccAddress) authtypes.AccountI {
	k.GetAccountCalls = append(k.GetAccountCalls, addr)
	if rv, found := k.GetAccountResultsMap[string(addr)]; found {
		return rv
	}
	return nil
}

func (k *MockAccountKeeper) SetAccount(_ sdk.Context, acc authtypes.AccountI) {
	k.SetAccountCalls = append(k.SetAccountCalls, acc)
}

func (k *MockAccountKeeper) HasAccount(_ sdk.Context, addr sdk.AccAddress) bool {
	k.HasAccountCalls = append(k.HasAccountCalls, addr)
	if rv, found := k.HasAccountResultsMap[string(addr)]; found {
		return rv
	}
	return false
}

func (k *MockAccountKeeper) NewAccount(_ sdk.Context, acc authtypes.AccountI) authtypes.AccountI {
	k.NewAccountCalls = append(k.NewAccountCalls, acc)
	if rv, found := k.NewAccountResultsMap[string(acc.GetAddress())]; found {
		return rv
	}
	return acc
}

// assertEqualGetAccountCalls asserts that a mock keeper's GetAccountCalls match the provided expected calls.
func (s *TestSuite) assertEqualGetAccountCalls(mk *MockAccountKeeper, expected []sdk.AccAddress, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.GetAccountCalls, s.getAddrName,
		msg+" GetAccount calls", args...)
}

// assertEqualSetAccountCalls asserts that a mock keeper's SetAccountCalls match the provided expected calls.
func (s *TestSuite) assertEqualSetAccountCalls(mk *MockAccountKeeper, expected []authtypes.AccountI, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.SetAccountCalls, authtypes.AccountI.String,
		msg+" SetAccount calls", args...)
}

// assertEqualHasAccountCalls asserts that a mock keeper's HasAccountCalls match the provided expected calls.
func (s *TestSuite) assertEqualHasAccountCalls(mk *MockAccountKeeper, expected []sdk.AccAddress, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.HasAccountCalls, s.getAddrName,
		msg+" HasAccount calls", args...)
}

// assertEqualNewAccountCalls asserts that a mock keeper's NewAccountCalls match the provided expected calls.
func (s *TestSuite) assertEqualNewAccountCalls(mk *MockAccountKeeper, expected []authtypes.AccountI, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.NewAccountCalls, authtypes.AccountI.String,
		msg+" NewAccount calls", args...)
}

type MockAttributeKeeper struct {
	GetAllAttributesAddrCalls      [][]byte
	GetAllAttributesAddrResultsMap map[string]*GetAllAttributesAddrResult
}

var _ exchange.AttributeKeeper = (*MockAttributeKeeper)(nil)

func NewMockAttributeKeeper() *MockAttributeKeeper {
	return &MockAttributeKeeper{
		GetAllAttributesAddrResultsMap: make(map[string]*GetAllAttributesAddrResult),
	}
}

// WithGetAllAttributesAddrResult sets up the provided address to return the given attrs
// and error from calls to GetAllAttributesAddr. An empty string means no error.
// This method both updates the receiver and returns it.
func (k *MockAttributeKeeper) WithGetAllAttributesAddrResult(addr []byte, attrs []attrtypes.Attribute, errStr string) *MockAttributeKeeper {
	k.GetAllAttributesAddrResultsMap[string(addr)] = NewGetAllAttributesAddrResult(attrs, errStr)
	return k
}

func (k *MockAttributeKeeper) GetAllAttributesAddr(_ sdk.Context, addr []byte) ([]attrtypes.Attribute, error) {
	k.GetAllAttributesAddrCalls = append(k.GetAllAttributesAddrCalls, addr)
	if rv, found := k.GetAllAttributesAddrResultsMap[string(addr)]; found {
		return rv.attrs, rv.err
	}
	return nil, nil
}

// assertEqualGetAllAttributesAddrCalls asserts that a mock keeper's GetAllAttributesAddrCalls match the provided expected calls.
func (s *TestSuite) assertEqualGetAllAttributesAddrCalls(mk *MockAttributeKeeper, expected [][]byte, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.GetAllAttributesAddrCalls,
		func(addr []byte) string {
			return s.getAddrName(addr)
		},
		msg+" NewAccount calls", args...)
}

// GetAllAttributesAddrResult contains the result args to return for a GetAllAttributesAddr call.
type GetAllAttributesAddrResult struct {
	attrs []attrtypes.Attribute
	err   error
}

// NewGetAllAttributesAddrResult creates a new GetAllAttributesAddrResult from the provided stuff.
func NewGetAllAttributesAddrResult(attrs []attrtypes.Attribute, errStr string) *GetAllAttributesAddrResult {
	rv := &GetAllAttributesAddrResult{attrs: attrs}
	if len(errStr) > 0 {
		rv.err = errors.New(errStr)
	}
	return rv
}

// MockBankKeeper satisfies the exchange.BankKeeper interface but just records the calls and allows dictation of results.
type MockBankKeeper struct {
	SendCoinsCalls                           []*SendCoinsArgs
	SendCoinsResultsQueue                    []string
	SendCoinsFromAccountToModuleCalls        []*SendCoinsFromAccountToModuleArgs
	SendCoinsFromAccountToModuleResultsQueue []string
	InputOutputCoinsCalls                    []*InputOutputCoinsArgs
	InputOutputCoinsResultsQueue             []string
}

var _ exchange.BankKeeper = (*MockBankKeeper)(nil)

// NewMockBankKeeper creates a new empty MockBankKeeper.
// Follow it up with WithSendCoinsResults, WithSendCoinsFromAccountToModuleResults,
// and/or WithInputOutputCoinsResults to dictate results.
func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{}
}

// WithSendCoinsResults queues up the provided error strings to be returned from SendCoins.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockBankKeeper) WithSendCoinsResults(errs ...string) *MockBankKeeper {
	k.SendCoinsResultsQueue = append(k.SendCoinsResultsQueue, errs...)
	return k
}

// WithSendCoinsFromAccountToModuleResults queues up the provided error strings to be returned from SendCoinsFromAccountToModule.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockBankKeeper) WithSendCoinsFromAccountToModuleResults(errs ...string) *MockBankKeeper {
	k.SendCoinsFromAccountToModuleResultsQueue = append(k.SendCoinsFromAccountToModuleResultsQueue, errs...)
	return k
}

// WithInputOutputCoinsResults queues up the provided error strings to be returned from InputOutputCoins.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockBankKeeper) WithInputOutputCoinsResults(errs ...string) *MockBankKeeper {
	k.InputOutputCoinsResultsQueue = append(k.InputOutputCoinsResultsQueue, errs...)
	return k
}

func (k MockBankKeeper) SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	k.SendCoinsCalls = append(k.SendCoinsCalls, NewSendCoinsArgs(ctx, fromAddr, toAddr, amt))
	var err error
	if len(k.SendCoinsResultsQueue) > 0 {
		if len(k.SendCoinsResultsQueue[0]) > 0 {
			err = errors.New(k.SendCoinsResultsQueue[0])
		}
		k.SendCoinsResultsQueue = k.SendCoinsResultsQueue[1:]
	}
	return err
}

func (k MockBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	k.SendCoinsFromAccountToModuleCalls = append(k.SendCoinsFromAccountToModuleCalls,
		NewSendCoinsFromAccountToModuleArgs(ctx, senderAddr, recipientModule, amt))
	var err error
	if len(k.SendCoinsFromAccountToModuleResultsQueue) > 0 {
		if len(k.SendCoinsFromAccountToModuleResultsQueue[0]) > 0 {
			err = errors.New(k.SendCoinsFromAccountToModuleResultsQueue[0])
		}
		k.SendCoinsFromAccountToModuleResultsQueue = k.SendCoinsFromAccountToModuleResultsQueue[1:]
	}
	return err
}

func (k MockBankKeeper) InputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	k.InputOutputCoinsCalls = append(k.InputOutputCoinsCalls, NewInputOutputCoinsArgs(ctx, inputs, outputs))
	var err error
	if len(k.InputOutputCoinsResultsQueue) > 0 {
		if len(k.InputOutputCoinsResultsQueue[0]) > 0 {
			err = errors.New(k.InputOutputCoinsResultsQueue[0])
		}
		k.InputOutputCoinsResultsQueue = k.InputOutputCoinsResultsQueue[1:]
	}
	return err
}

// assertEqualSendCoinsCalls asserts that a mock keeper's SendCoinsCalls match the provided expected calls.
func (s *TestSuite) assertEqualSendCoinsCalls(mk *MockBankKeeper, expected []*SendCoinsArgs, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.SendCoinsCalls, s.sendCoinsArgsString,
		msg+" SendCoins calls", args...)
}

// assertEqualSendCoinsFromAccountToModuleCalls asserts that a mock keeper's
// SendCoinsFromAccountToModuleCalls match the provided expected calls.
func (s *TestSuite) assertEqualSendCoinsFromAccountToModuleCalls(mk *MockBankKeeper, expected []*SendCoinsFromAccountToModuleArgs, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.SendCoinsFromAccountToModuleCalls, s.sendCoinsFromAccountToModuleArgsString,
		msg+" SendCoinsFromAccountToModule calls", args...)
}

// assertEqualInputOutputCoinsCalls asserts that a mock keeper's InputOutputCoinsCalls match the provided expected calls.
func (s *TestSuite) assertEqualInputOutputCoinsCalls(mk *MockBankKeeper, expected []*InputOutputCoinsArgs, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.InputOutputCoinsCalls, s.inputOutputCoinsArgsString,
		msg+" InputOutputCoins calls", args...)
}

// SendCoinsArgs is a record of a call that is made to SendCoins.
type SendCoinsArgs struct {
	ctxHasQuarantineBypass bool
	fromAddr               sdk.AccAddress
	toAddr                 sdk.AccAddress
	amt                    sdk.Coins
}

// NewSendCoinsArgs creates a new record of args provided to a call to SendCoins.
func NewSendCoinsArgs(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) *SendCoinsArgs {
	return &SendCoinsArgs{
		ctxHasQuarantineBypass: quarantine.HasBypass(ctx),
		fromAddr:               fromAddr,
		toAddr:                 toAddr,
		amt:                    amt,
	}
}

// sendCoinsArgsString creates a string of a SendCoinsArgs
// substituting the address names as possible.
func (s *TestSuite) sendCoinsArgsString(a *SendCoinsArgs) string {
	return fmt.Sprintf("{q-bypass:%t, from:%s, to:%s, amt:%s}",
		a.ctxHasQuarantineBypass, s.getAddrName(a.fromAddr), s.getAddrName(a.toAddr), a.amt)
}

// SendCoinsFromAccountToModuleArgs is a record of a call that is made to SendCoinsFromAccountToModule.
type SendCoinsFromAccountToModuleArgs struct {
	ctxHasQuarantineBypass bool
	senderAddr             sdk.AccAddress
	recipientModule        string
	amt                    sdk.Coins
}

// NewSendCoinsFromAccountToModuleArgs creates a new record of args provided to a call to SendCoinsFromAccountToModule.
func NewSendCoinsFromAccountToModuleArgs(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) *SendCoinsFromAccountToModuleArgs {
	return &SendCoinsFromAccountToModuleArgs{
		ctxHasQuarantineBypass: quarantine.HasBypass(ctx),
		senderAddr:             senderAddr,
		recipientModule:        recipientModule,
		amt:                    amt,
	}
}

// sendCoinsFromAccountToModuleArgsString creates a string of a SendCoinsFromAccountToModuleArgs
// substituting the address names as possible.
func (s *TestSuite) sendCoinsFromAccountToModuleArgsString(a *SendCoinsFromAccountToModuleArgs) string {
	return fmt.Sprintf("{q-bypass:%t, from:%s, to:%s, amt:%s}",
		a.ctxHasQuarantineBypass, s.getAddrName(a.senderAddr), a.recipientModule, a.amt)
}

// InputOutputCoinsArgs is a record of a call that is made to InputOutputCoins.
type InputOutputCoinsArgs struct {
	ctxHasQuarantineBypass bool
	inputs                 []banktypes.Input
	outputs                []banktypes.Output
}

// NewInputOutputCoinsArgs creates a new record of args provided to a call to InputOutputCoins.
func NewInputOutputCoinsArgs(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) *InputOutputCoinsArgs {
	return &InputOutputCoinsArgs{
		ctxHasQuarantineBypass: quarantine.HasBypass(ctx),
		inputs:                 inputs,
		outputs:                outputs,
	}
}

// inputOutputCoinsArgsString creates a string of a InputOutputCoinsArgs substituting the address names as possible.
func (s *TestSuite) inputOutputCoinsArgsString(a *InputOutputCoinsArgs) string {
	return fmt.Sprintf("{q-bypass:%t, inputs:%s, outputs:%s}",
		a.ctxHasQuarantineBypass, s.inputsString(a.inputs), s.outputsString(a.outputs))
}

// inputString creates a string of a banktypes.Input substituting the address names as possible.
func (s *TestSuite) inputString(a banktypes.Input) string {
	return fmt.Sprintf("I{Address:%s, Coins:%s}", s.getAddrStrName(a.Address), a.Coins)
}

// inputsString creates a string of a slice of banktypes.Input substituting the address names as possible.
func (s *TestSuite) inputsString(vals []banktypes.Input) string {
	strs := toStrings(vals, s.inputString)
	return fmt.Sprintf("{%s}", strings.Join(strs, ", "))
}

// outputString creates a string of a banktypes.Output substituting the address names as possible.
func (s *TestSuite) outputString(a banktypes.Output) string {
	return fmt.Sprintf("O{Address:%s, Coins:%s}", s.getAddrStrName(a.Address), a.Coins)
}

// outputsString creates a string of a slice of banktypes.Output substituting the address names as possible.
func (s *TestSuite) outputsString(vals []banktypes.Output) string {
	strs := toStrings(vals, s.outputString)
	return fmt.Sprintf("{%s}", strings.Join(strs, ", "))
}

// MockHoldKeeper satisfies the exchange.HoldKeeper interface but just records the calls and allows dictation of results.
type MockHoldKeeper struct {
	AddHoldCalls            []*AddHoldArgs
	ReleaseHoldCalls        []*ReleaseHoldArgs
	GetHoldCoinCalls        []*GetHoldCoinArgs
	AddHoldResultsQueue     []string
	ReleaseHoldResultsQueue []string
	GetHoldCoinResultsMap   map[string]map[string]*GetHoldCoinResults
}

var _ exchange.HoldKeeper = (*MockHoldKeeper)(nil)

// NewMockHoldKeeper creates a new empty MockHoldKeeper.
// Follow it up with WithAddHoldResults, WithReleaseHoldResults, WithGetHoldCoinResult
// and/or WithGetHoldCoinErrorResult to dictate results.
func NewMockHoldKeeper() *MockHoldKeeper {
	return &MockHoldKeeper{
		GetHoldCoinResultsMap: make(map[string]map[string]*GetHoldCoinResults),
	}
}

// WithAddHoldResults queues up the provided error strings to be returned from AddHold.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockHoldKeeper) WithAddHoldResults(errs ...string) *MockHoldKeeper {
	k.AddHoldResultsQueue = append(k.AddHoldResultsQueue, errs...)
	return k
}

// WithReleaseHoldResults queues up the provided error strings to be returned from ReleaseHold.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockHoldKeeper) WithReleaseHoldResults(errs ...string) *MockHoldKeeper {
	k.ReleaseHoldResultsQueue = append(k.ReleaseHoldResultsQueue, errs...)
	return k
}

// WithGetHoldCoinResult sets the results of GetHoldCoin for the provided address and coins.
// If there isn't an entry for a requested address and denom, a zero-coin and nil error will be returned.
// To cause an error to be returned, use WithGetHoldCoinErrorResult.
func (k *MockHoldKeeper) WithGetHoldCoinResult(addr sdk.AccAddress, coins ...sdk.Coin) *MockHoldKeeper {
	denomMap, found := k.GetHoldCoinResultsMap[string(addr)]
	if !found {
		denomMap = make(map[string]*GetHoldCoinResults)
		k.GetHoldCoinResultsMap[string(addr)] = denomMap
	}
	for _, coin := range coins {
		denomMap[coin.Denom] = &GetHoldCoinResults{amount: coin.Amount}
	}
	return k
}

// WithGetHoldCoinErrorResult sets the result of GetHoldCoin for the provided address and denom to be the provided error.
// An empty string means no error. A zero-coin is also returned with the result.
// To return a coin value without an error, use WithGetHoldCoinResult.
func (k *MockHoldKeeper) WithGetHoldCoinErrorResult(addr sdk.AccAddress, denom string, errStr string) *MockHoldKeeper {
	denomMap, found := k.GetHoldCoinResultsMap[string(addr)]
	if !found {
		denomMap = make(map[string]*GetHoldCoinResults)
		k.GetHoldCoinResultsMap[string(addr)] = denomMap
	}
	denomMap[denom] = &GetHoldCoinResults{amount: sdkmath.ZeroInt()}
	if len(errStr) > 0 {
		denomMap[denom].err = errors.New(errStr)
	}
	return k
}

func (k *MockHoldKeeper) AddHold(_ sdk.Context, addr sdk.AccAddress, funds sdk.Coins, reason string) error {
	k.AddHoldCalls = append(k.AddHoldCalls, NewAddHoldArgs(addr, funds, reason))
	var err error
	if len(k.AddHoldResultsQueue) > 0 {
		if len(k.AddHoldResultsQueue[0]) > 0 {
			err = errors.New(k.AddHoldResultsQueue[0])
		}
		k.AddHoldResultsQueue = k.AddHoldResultsQueue[1:]
	}
	return err
}

func (k *MockHoldKeeper) ReleaseHold(_ sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	k.ReleaseHoldCalls = append(k.ReleaseHoldCalls, NewReleaseHoldArgs(addr, funds))
	var err error
	if len(k.ReleaseHoldResultsQueue) > 0 {
		if len(k.ReleaseHoldResultsQueue[0]) > 0 {
			err = errors.New(k.ReleaseHoldResultsQueue[0])
		}
		k.ReleaseHoldResultsQueue = k.ReleaseHoldResultsQueue[1:]
	}
	return err
}

func (k *MockHoldKeeper) GetHoldCoin(_ sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	k.GetHoldCoinCalls = append(k.GetHoldCoinCalls, NewGetHoldCoinArgs(addr, denom))
	if denomMap, aFound := k.GetHoldCoinResultsMap[string(addr)]; aFound {
		if rv, dFound := denomMap[denom]; dFound {
			return sdk.NewCoin(denom, rv.amount), rv.err
		}
	}
	return sdk.NewInt64Coin(denom, 0), nil
}

// assertEqualAddHoldCalls asserts that a mock keeper's AddHoldCalls match the provided expected calls.
func (s *TestSuite) assertEqualAddHoldCalls(mk *MockHoldKeeper, expected []*AddHoldArgs, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.AddHoldCalls, s.addHoldArgsString,
		msg+" AddHoldCalls calls", args...)
}

// assertEqualReleaseHoldCalls asserts that a mock keeper's ReleaseHoldCalls match the provided expected calls.
func (s *TestSuite) assertEqualReleaseHoldCalls(mk *MockHoldKeeper, expected []*ReleaseHoldArgs, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.ReleaseHoldCalls, s.releaseHoldArgsString,
		msg+" ReleaseHoldCalls calls", args...)
}

// assertEqualGetHoldCoinCalls asserts that a mock keeper's GetHoldCoinCalls match the provided expected calls.
func (s *TestSuite) assertEqualGetHoldCoinCalls(mk *MockHoldKeeper, expected []*GetHoldCoinArgs, msg string, args ...interface{}) bool {
	return assertEqualSlice(s, expected, mk.GetHoldCoinCalls, s.getHoldCoinArgsString,
		msg+" GetHoldCoinCalls calls", args...)
}

// AddHoldArgs is a record of a call that is made to AddHold.
type AddHoldArgs struct {
	addr   sdk.AccAddress
	funds  sdk.Coins
	reason string
}

// NewAddHoldArgs creates a new record of args provided to a call to AddHold.
func NewAddHoldArgs(addr sdk.AccAddress, funds sdk.Coins, reason string) *AddHoldArgs {
	return &AddHoldArgs{
		addr:   addr,
		funds:  funds,
		reason: reason,
	}
}

// addHoldArgsString creates a string of a AddHoldArgs substituting the address names as possible.
func (s *TestSuite) addHoldArgsString(a *AddHoldArgs) string {
	return fmt.Sprintf("{addr:%s, funds:%s, reason:%q}", s.getAddrName(a.addr), a.funds, a.reason)
}

// ReleaseHoldArgs is a record of a call that is made to ReleaseHold.
type ReleaseHoldArgs struct {
	addr  sdk.AccAddress
	funds sdk.Coins
}

// NewReleaseHoldArgs creates a new record of args provided to a call to ReleaseHold.
func NewReleaseHoldArgs(addr sdk.AccAddress, funds sdk.Coins) *ReleaseHoldArgs {
	return &ReleaseHoldArgs{
		addr:  addr,
		funds: funds,
	}
}

// releaseHoldArgsString creates a string of a ReleaseHoldArgs substituting the address names as possible.
func (s *TestSuite) releaseHoldArgsString(a *ReleaseHoldArgs) string {
	return fmt.Sprintf("{addr:%s, funds:%s}", s.getAddrName(a.addr), a.funds)
}

// GetHoldCoinArgs is a record of a call that is made to GetHoldCoin.
type GetHoldCoinArgs struct {
	addr  sdk.AccAddress
	denom string
}

// NewGetHoldCoinArgs creates a new record of args provided to a call to GetHoldCoin.
func NewGetHoldCoinArgs(addr sdk.AccAddress, denom string) *GetHoldCoinArgs {
	return &GetHoldCoinArgs{
		addr:  addr,
		denom: denom,
	}
}

// getHoldCoinArgsString creates a string of a GetHoldCoinArgs substituting the address names as possible.
func (s *TestSuite) getHoldCoinArgsString(a *GetHoldCoinArgs) string {
	return fmt.Sprintf("{addr:%s, denom:%s}", s.getAddrName(a.addr), a.denom)
}

// GetHoldCoinResults contains the result args to return for a GetHoldCoin call.
type GetHoldCoinResults struct {
	amount sdkmath.Int
	err    error
}
