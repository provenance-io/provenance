package keeper_test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/provutils"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/exchange"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	"github.com/provenance-io/provenance/x/quarantine"
)

// #############################################################################
// #############################                   #############################
// ###########################   MockAccountKeeper   ###########################
// #############################                   #############################
// #############################################################################

var _ exchange.AccountKeeper = (*MockAccountKeeper)(nil)

// MockAccountKeeper satisfies the exchange.AccountKeeper interface but just records the calls and allows dictation of results.
type MockAccountKeeper struct {
	Calls                 AccountCalls
	GetAccountResultsMap  map[string]sdk.AccountI
	HasAccountResultsMap  map[string]bool
	NewAccountModifierMap map[string]AccountModifier
}

// AccountCalls contains all the calls that the mock account keeper makes.
type AccountCalls struct {
	GetAccount []sdk.AccAddress
	SetAccount []sdk.AccountI
	HasAccount []sdk.AccAddress
	NewAccount []sdk.AccountI
}

// AccountModifier is a function that can alter an account.
type AccountModifier func(sdk.AccountI) sdk.AccountI

// NoopAccMod is a no-op AccountModifier.
func NoopAccMod(a sdk.AccountI) sdk.AccountI {
	return a
}

var _ AccountModifier = NoopAccMod

// NewMockAccountKeeper creates a new empty MockAccountKeeper.
// Follow it up with WithGetAccountResult, WithHasAccountResult,
// and/or WithNewAccountResult to dictate results.
func NewMockAccountKeeper() *MockAccountKeeper {
	return &MockAccountKeeper{
		GetAccountResultsMap:  make(map[string]sdk.AccountI),
		HasAccountResultsMap:  make(map[string]bool),
		NewAccountModifierMap: make(map[string]AccountModifier),
	}
}

// WithGetAccountResult associates the provided address and result for use with calls to GetAccount.
// When GetAccount is called, if the address provided has an entry here, that is returned, otherwise, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockAccountKeeper) WithGetAccountResult(addr sdk.AccAddress, result sdk.AccountI) *MockAccountKeeper {
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
func (k *MockAccountKeeper) WithNewAccountModifier(addr sdk.AccAddress, modifier AccountModifier) *MockAccountKeeper {
	k.NewAccountModifierMap[string(addr)] = modifier
	return k
}

func (k *MockAccountKeeper) GetAccount(_ context.Context, addr sdk.AccAddress) sdk.AccountI {
	k.Calls.GetAccount = append(k.Calls.GetAccount, addr)
	if rv, found := k.GetAccountResultsMap[string(addr)]; found {
		return rv
	}
	return nil
}

func (k *MockAccountKeeper) SetAccount(_ context.Context, acc sdk.AccountI) {
	k.Calls.SetAccount = append(k.Calls.SetAccount, acc)
	k.WithGetAccountResult(acc.GetAddress(), acc)
}

func (k *MockAccountKeeper) HasAccount(_ context.Context, addr sdk.AccAddress) bool {
	k.Calls.HasAccount = append(k.Calls.HasAccount, addr)
	if rv, found := k.HasAccountResultsMap[string(addr)]; found {
		return rv
	}
	return false
}

func (k *MockAccountKeeper) NewAccount(_ context.Context, acc sdk.AccountI) sdk.AccountI {
	k.Calls.NewAccount = append(k.Calls.NewAccount, acc)
	if modifier, found := k.NewAccountModifierMap[string(acc.GetAddress())]; found {
		return modifier(acc)
	}
	return acc
}

// assertGetAccountCalls asserts that a mock keeper's Calls.GetAccount match the provided expected calls.
func (s *TestSuite) assertGetAccountCalls(mk *MockAccountKeeper, expected []sdk.AccAddress, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.GetAccount, s.getAddrName,
		msg+" GetAccount calls", args...)
}

// assertSetAccountCalls asserts that a mock keeper's Calls.SetAccount match the provided expected calls.
func (s *TestSuite) assertSetAccountCalls(mk *MockAccountKeeper, expected []sdk.AccountI, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.SetAccount, sdk.AccountI.String,
		msg+" SetAccount calls", args...)
}

// assertHasAccountCalls asserts that a mock keeper's Calls.HasAccount match the provided expected calls.
func (s *TestSuite) assertHasAccountCalls(mk *MockAccountKeeper, expected []sdk.AccAddress, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.HasAccount, s.getAddrName,
		msg+" HasAccount calls", args...)
}

// assertNewAccountCalls asserts that a mock keeper's Calls.NewAccount match the provided expected calls.
func (s *TestSuite) assertNewAccountCalls(mk *MockAccountKeeper, expected []sdk.AccountI, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.NewAccount, sdk.AccountI.String,
		msg+" NewAccount calls", args...)
}

// assertAccountKeeperCalls asserts that all the calls made to a mock account keeper match the provided expected calls.
func (s *TestSuite) assertAccountKeeperCalls(mk *MockAccountKeeper, expected AccountCalls, msg string, args ...interface{}) bool {
	s.T().Helper()
	rv := s.assertGetAccountCalls(mk, expected.GetAccount, msg, args...)
	rv = s.assertSetAccountCalls(mk, expected.SetAccount, msg, args...) && rv
	rv = s.assertHasAccountCalls(mk, expected.HasAccount, msg, args...) && rv
	return s.assertNewAccountCalls(mk, expected.NewAccount, msg, args...) && rv
}

// #############################################################################
// ############################                     ############################
// ##########################   MockAttributeKeeper   ##########################
// ############################                     ############################
// #############################################################################

var _ exchange.AttributeKeeper = (*MockAttributeKeeper)(nil)

// MockAttributeKeeper satisfies the exchange.AttributeKeeper interface but just records the calls and allows dictation of results.
type MockAttributeKeeper struct {
	Calls                          AttributeCalls
	GetAllAttributesAddrResultsMap map[string]*GetAllAttributesAddrResult
}

// AttributeCalls contains all the calls that the mock attribute keeper makes.
type AttributeCalls struct {
	GetAllAttributesAddr [][]byte
}

// GetAllAttributesAddrResult contains the result args to return for a GetAllAttributesAddr call.
type GetAllAttributesAddrResult struct {
	attrs []attrtypes.Attribute
	err   error
}

// NewMockAttributeKeeper creates a new empty MockAttributeKeeper.
// Follow it up with WithGetAllAttributesAddrResult to dictate results.
func NewMockAttributeKeeper() *MockAttributeKeeper {
	return &MockAttributeKeeper{
		GetAllAttributesAddrResultsMap: make(map[string]*GetAllAttributesAddrResult),
	}
}

// WithGetAllAttributesAddrResult sets up the provided address to return the given attrs
// and error from calls to GetAllAttributesAddr. An empty string means no error.
// This method both updates the receiver and returns it.
func (k *MockAttributeKeeper) WithGetAllAttributesAddrResult(addr []byte, attrNames []string, errStr string) *MockAttributeKeeper {
	var attrs []attrtypes.Attribute
	if attrNames != nil {
		attrs = make([]attrtypes.Attribute, len(attrNames))
		for i, name := range attrNames {
			attrs[i] = attrtypes.Attribute{
				Name:          name,
				Value:         []byte("this is the " + name + " value"),
				AttributeType: attrtypes.AttributeType_String,
				Address:       sdk.AccAddress(addr).String(),
			}
		}
	}
	k.GetAllAttributesAddrResultsMap[string(addr)] = NewGetAllAttributesAddrResult(attrs, errStr)
	return k
}

func (k *MockAttributeKeeper) GetAllAttributesAddr(_ sdk.Context, addr []byte) ([]attrtypes.Attribute, error) {
	k.Calls.GetAllAttributesAddr = append(k.Calls.GetAllAttributesAddr, addr)
	if rv, found := k.GetAllAttributesAddrResultsMap[string(addr)]; found {
		return rv.attrs, rv.err
	}
	return nil, nil
}

// assertGetAllAttributesAddrCalls asserts that a mock keeper's Calls.GetAllAttributesAddr match the provided expected calls.
func (s *TestSuite) assertGetAllAttributesAddrCalls(mk *MockAttributeKeeper, expected [][]byte, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.GetAllAttributesAddr,
		func(addr []byte) string {
			return s.getAddrName(addr)
		},
		msg+" GetAllAttributesAddr calls", args...)
}

// assertAttributeKeeperCalls asserts that all the calls made to a mock account keeper match the provided expected calls.
func (s *TestSuite) assertAttributeKeeperCalls(mk *MockAttributeKeeper, expected AttributeCalls, msg string, args ...interface{}) bool {
	s.T().Helper()
	return s.assertGetAllAttributesAddrCalls(mk, expected.GetAllAttributesAddr, msg, args...)
}

// NewGetAllAttributesAddrResult creates a new GetAllAttributesAddrResult from the provided stuff.
func NewGetAllAttributesAddrResult(attrs []attrtypes.Attribute, errStr string) *GetAllAttributesAddrResult {
	rv := &GetAllAttributesAddrResult{attrs: attrs}
	if len(errStr) > 0 {
		rv.err = errors.New(errStr)
	}
	return rv
}

// #############################################################################
// ##############################                ###############################
// ############################   MockBankKeeper   #############################
// ##############################                ###############################
// #############################################################################

var _ exchange.BankKeeper = (*MockBankKeeper)(nil)

// MockBankKeeper satisfies the exchange.BankKeeper interface but just records the calls and allows dictation of results.
type MockBankKeeper struct {
	Calls                                    BankCalls
	SendCoinsResultsQueue                    []string
	SendCoinsFromAccountToModuleResultsQueue []string
	InputOutputCoinsResultsQueue             []string
	BlockedAddrQueue                         []bool
}

// BankCalls contains all the calls that the mock bank keeper makes.
type BankCalls struct {
	SendCoins                    []*SendCoinsArgs
	SendCoinsFromAccountToModule []*SendCoinsFromAccountToModuleArgs
	InputOutputCoins             []*InputOutputCoinsArgs
	BlockedAddr                  []sdk.AccAddress
}

// SendCoinsArgs is a record of a call that is made to SendCoins.
type SendCoinsArgs struct {
	ctxHasQuarantineBypass bool
	ctxTransferAgent       sdk.AccAddress
	fromAddr               sdk.AccAddress
	toAddr                 sdk.AccAddress
	amt                    sdk.Coins
}

// SendCoinsFromAccountToModuleArgs is a record of a call that is made to SendCoinsFromAccountToModule.
type SendCoinsFromAccountToModuleArgs struct {
	ctxHasQuarantineBypass bool
	ctxTransferAgent       sdk.AccAddress
	senderAddr             sdk.AccAddress
	recipientModule        string
	amt                    sdk.Coins
}

// InputOutputCoinsArgs is a record of a call that is made to InputOutputCoins.
type InputOutputCoinsArgs struct {
	ctxHasQuarantineBypass bool
	ctxTransferAgent       sdk.AccAddress
	inputs                 []banktypes.Input
	outputs                []banktypes.Output
}

// NewMockBankKeeper creates a new empty MockBankKeeper.
// Follow it up with WithSendCoinsResults, WithSendCoinsFromAccountToModuleResults,
// and/or WithInputOutputCoinsResults to dictate results.
func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{}
}

// WithSendCoinsResults queues up the provided error strings to be returned from SendCoins.
// An empty string means no result or error. Each entry is used only once. If entries run out, nil is returned.
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

// WithBlockedAddrResults queues up the provided bools to be returned from BlockedAddr.
// Each entry is used only once. If entries run out, false is returned.
// This method both updates the receiver and returns it.
func (k *MockBankKeeper) WithBlockedAddrResults(results ...bool) *MockBankKeeper {
	k.BlockedAddrQueue = append(k.BlockedAddrQueue, results...)
	return k
}

func (k *MockBankKeeper) SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	k.Calls.SendCoins = append(k.Calls.SendCoins, NewSendCoinsArgs(ctx, fromAddr, toAddr, amt))
	var err error
	if len(k.SendCoinsResultsQueue) > 0 {
		if len(k.SendCoinsResultsQueue[0]) > 0 {
			err = errors.New(k.SendCoinsResultsQueue[0])
		}
		k.SendCoinsResultsQueue = k.SendCoinsResultsQueue[1:]
	}
	return err
}

func (k *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	k.Calls.SendCoinsFromAccountToModule = append(k.Calls.SendCoinsFromAccountToModule,
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

func (k *MockBankKeeper) InputOutputCoinsProv(ctx context.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	k.Calls.InputOutputCoins = append(k.Calls.InputOutputCoins, NewInputOutputCoinsArgs(ctx, inputs, outputs))
	var err error
	if len(k.InputOutputCoinsResultsQueue) > 0 {
		if len(k.InputOutputCoinsResultsQueue[0]) > 0 {
			err = errors.New(k.InputOutputCoinsResultsQueue[0])
		}
		k.InputOutputCoinsResultsQueue = k.InputOutputCoinsResultsQueue[1:]
	}
	return err
}

func (k *MockBankKeeper) BlockedAddr(addr sdk.AccAddress) bool {
	k.Calls.BlockedAddr = append(k.Calls.BlockedAddr, addr)
	rv := false
	if len(k.BlockedAddrQueue) > 0 {
		rv = k.BlockedAddrQueue[0]
		k.BlockedAddrQueue = k.BlockedAddrQueue[1:]
	}
	return rv
}

// assertSendCoinsCalls asserts that a mock keeper's Calls.SendCoins match the provided expected calls.
func (s *TestSuite) assertSendCoinsCalls(mk *MockBankKeeper, expected []*SendCoinsArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.SendCoins, s.sendCoinsArgsString,
		msg+" SendCoins calls", args...)
}

// assertSendCoinsFromAccountToModuleCalls asserts that a mock keeper's
// Calls.SendCoinsFromAccountToModule match the provided expected calls.
func (s *TestSuite) assertSendCoinsFromAccountToModuleCalls(mk *MockBankKeeper, expected []*SendCoinsFromAccountToModuleArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.SendCoinsFromAccountToModule, s.sendCoinsFromAccountToModuleArgsString,
		msg+" SendCoinsFromAccountToModule calls", args...)
}

// assertInputOutputCoinsCalls asserts that a mock keeper's Calls.InputOutputCoins match the provided expected calls.
func (s *TestSuite) assertInputOutputCoinsCalls(mk *MockBankKeeper, expected []*InputOutputCoinsArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.InputOutputCoins, s.inputOutputCoinsArgsString,
		msg+" InputOutputCoins calls", args...)
}

// assertBlockedAddrCalls asserts that a mock keeper's Calls.BlockedAddr match the provided expected calls.
func (s *TestSuite) assertBlockedAddrCalls(mk *MockBankKeeper, expected []sdk.AccAddress, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.BlockedAddr, s.getAddrName,
		msg+" BlockedAddr calls", args...)
}

// assertBankKeeperCalls asserts that all the calls made to a mock bank keeper match the provided expected calls.
func (s *TestSuite) assertBankKeeperCalls(mk *MockBankKeeper, expected BankCalls, msg string, args ...interface{}) bool {
	s.T().Helper()
	rv := s.assertSendCoinsCalls(mk, expected.SendCoins, msg, args...)
	rv = s.assertInputOutputCoinsCalls(mk, expected.InputOutputCoins, msg, args...) && rv
	rv = s.assertSendCoinsFromAccountToModuleCalls(mk, expected.SendCoinsFromAccountToModule, msg, args...) && rv
	return s.assertBlockedAddrCalls(mk, expected.BlockedAddr, msg, args...) && rv
}

// NewSendCoinsArgs creates a new record of args provided to a call to SendCoins.
func NewSendCoinsArgs(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) *SendCoinsArgs {
	rv := &SendCoinsArgs{
		ctxHasQuarantineBypass: quarantine.HasBypass(ctx),
		fromAddr:               fromAddr,
		toAddr:                 toAddr,
		amt:                    amt,
	}
	xferAgents := markertypes.GetTransferAgents(ctx)
	if len(xferAgents) > 0 {
		rv.ctxTransferAgent = xferAgents[0]
	}
	return rv
}

// sendCoinsArgsString creates a string of a SendCoinsArgs
// substituting the address names as possible.
func (s *TestSuite) sendCoinsArgsString(a *SendCoinsArgs) string {
	return fmt.Sprintf("{q-bypass:%t, xfer-agent:%q, from:%s, to:%s, amt:%s}",
		a.ctxHasQuarantineBypass, s.getAddrName(a.ctxTransferAgent), s.getAddrName(a.fromAddr), s.getAddrName(a.toAddr), a.amt)
}

// NewSendCoinsFromAccountToModuleArgs creates a new record of args provided to a call to SendCoinsFromAccountToModule.
func NewSendCoinsFromAccountToModuleArgs(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) *SendCoinsFromAccountToModuleArgs {
	rv := &SendCoinsFromAccountToModuleArgs{
		ctxHasQuarantineBypass: quarantine.HasBypass(ctx),
		senderAddr:             senderAddr,
		recipientModule:        recipientModule,
		amt:                    amt,
	}
	xferAgents := markertypes.GetTransferAgents(ctx)
	if len(xferAgents) > 0 {
		rv.ctxTransferAgent = xferAgents[0]
	}
	return rv
}

// sendCoinsFromAccountToModuleArgsString creates a string of a SendCoinsFromAccountToModuleArgs
// substituting the address names as possible.
func (s *TestSuite) sendCoinsFromAccountToModuleArgsString(a *SendCoinsFromAccountToModuleArgs) string {
	return fmt.Sprintf("{q-bypass:%t, xfer-agent:%q, from:%s, to:%s, amt:%s}",
		a.ctxHasQuarantineBypass, s.getAddrName(a.ctxTransferAgent), s.getAddrName(a.senderAddr), a.recipientModule, a.amt)
}

// NewInputOutputCoinsArgs creates a new record of args provided to a call to InputOutputCoins.
func NewInputOutputCoinsArgs(ctx context.Context, inputs []banktypes.Input, outputs []banktypes.Output) *InputOutputCoinsArgs {
	rv := &InputOutputCoinsArgs{
		ctxHasQuarantineBypass: quarantine.HasBypass(ctx),
		inputs:                 inputs,
		outputs:                outputs,
	}
	xferAgents := markertypes.GetTransferAgents(ctx)
	if len(xferAgents) > 0 {
		rv.ctxTransferAgent = xferAgents[0]
	}
	return rv
}

// inputOutputCoinsArgsString creates a string of a InputOutputCoinsArgs substituting the address names as possible.
func (s *TestSuite) inputOutputCoinsArgsString(a *InputOutputCoinsArgs) string {
	return fmt.Sprintf("{q-bypass:%t, xfer-agent:%q, inputs:%s, outputs:%s}",
		a.ctxHasQuarantineBypass, s.getAddrName(a.ctxTransferAgent), s.inputsString(a.inputs), s.outputsString(a.outputs))
}

// inputString creates a string of a banktypes.Input substituting the address names as possible.
func (s *TestSuite) inputString(a banktypes.Input) string {
	return fmt.Sprintf("I{Address:%s, Coins:%s}", s.getAddrStrName(a.Address), a.Coins)
}

// inputsString creates a string of a slice of banktypes.Input substituting the address names as possible.
func (s *TestSuite) inputsString(vals []banktypes.Input) string {
	return fmt.Sprintf("{%s}", sliceString(vals, s.inputString))
}

// outputString creates a string of a banktypes.Output substituting the address names as possible.
func (s *TestSuite) outputString(a banktypes.Output) string {
	return fmt.Sprintf("O{Address:%s, Coins:%s}", s.getAddrStrName(a.Address), a.Coins)
}

// outputsString creates a string of a slice of banktypes.Output substituting the address names as possible.
func (s *TestSuite) outputsString(vals []banktypes.Output) string {
	return fmt.Sprintf("{%s}", sliceString(vals, s.outputString))
}

// #############################################################################
// ##############################                ###############################
// ############################   MockHoldKeeper   #############################
// ##############################                ###############################
// #############################################################################

var _ exchange.HoldKeeper = (*MockHoldKeeper)(nil)

// MockHoldKeeper satisfies the exchange.HoldKeeper interface but just records the calls and allows dictation of results.
type MockHoldKeeper struct {
	Calls                   HoldCalls
	AddHoldResultsQueue     []string
	ReleaseHoldResultsQueue []string
	GetHoldCoinResultsMap   map[string]map[string]*GetHoldCoinResults
}

// HoldCalls contains all the calls that the mock hold keeper makes.
type HoldCalls struct {
	AddHold     []*AddHoldArgs
	ReleaseHold []*ReleaseHoldArgs
	GetHoldCoin []*GetHoldCoinArgs
}

// AddHoldArgs is a record of a call that is made to AddHold.
type AddHoldArgs struct {
	addr   sdk.AccAddress
	funds  sdk.Coins
	reason string
}

// ReleaseHoldArgs is a record of a call that is made to ReleaseHold.
type ReleaseHoldArgs struct {
	addr  sdk.AccAddress
	funds sdk.Coins
}

// GetHoldCoinArgs is a record of a call that is made to GetHoldCoin.
type GetHoldCoinArgs struct {
	addr  sdk.AccAddress
	denom string
}

// GetHoldCoinResults contains the result args to return for a GetHoldCoin call.
type GetHoldCoinResults struct {
	amount sdkmath.Int
	err    error
}

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
	k.Calls.AddHold = append(k.Calls.AddHold, NewAddHoldArgs(addr, funds, reason))
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
	k.Calls.ReleaseHold = append(k.Calls.ReleaseHold, NewReleaseHoldArgs(addr, funds))
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
	k.Calls.GetHoldCoin = append(k.Calls.GetHoldCoin, NewGetHoldCoinArgs(addr, denom))
	if denomMap, aFound := k.GetHoldCoinResultsMap[string(addr)]; aFound {
		if rv, dFound := denomMap[denom]; dFound {
			return sdk.NewCoin(denom, rv.amount), rv.err
		}
	}
	return sdk.NewInt64Coin(denom, 0), nil
}

// assertAddHoldCalls asserts that a mock keeper's Calls.AddHold match the provided expected calls.
func (s *TestSuite) assertAddHoldCalls(mk *MockHoldKeeper, expected []*AddHoldArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.AddHold, s.addHoldArgsString,
		msg+" AddHold calls", args...)
}

// assertReleaseHoldCalls asserts that a mock keeper's Calls.ReleaseHold match the provided expected calls.
func (s *TestSuite) assertReleaseHoldCalls(mk *MockHoldKeeper, expected []*ReleaseHoldArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.ReleaseHold, s.releaseHoldArgsString,
		msg+" ReleaseHold calls", args...)
}

// assertGetHoldCoinCalls asserts that a mock keeper's Calls.GetHoldCoin match the provided expected calls.
func (s *TestSuite) assertGetHoldCoinCalls(mk *MockHoldKeeper, expected []*GetHoldCoinArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.GetHoldCoin, s.getHoldCoinArgsString,
		msg+" GetHoldCoin calls", args...)
}

// assertHoldKeeperCalls asserts that all the calls made to a mock hold keeper match the provided expected calls.
func (s *TestSuite) assertHoldKeeperCalls(mk *MockHoldKeeper, expected HoldCalls, msg string, args ...interface{}) bool {
	s.T().Helper()
	rv := s.assertAddHoldCalls(mk, expected.AddHold, msg, args...)
	rv = s.assertReleaseHoldCalls(mk, expected.ReleaseHold, msg, args...) && rv
	return s.assertGetHoldCoinCalls(mk, expected.GetHoldCoin, msg, args...) && rv
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

// #############################################################################
// #############################                  ##############################
// ###########################   MockMarkerKeeper   ############################
// #############################                  ##############################
// #############################################################################

var _ exchange.MarkerKeeper = (*MockMarkerKeeper)(nil)

// MockMarkerKeeper satisfies the exchange.MarkerKeeper interface but just records the calls and allows dictation of results.
type MockMarkerKeeper struct {
	Calls                            MarkerCalls
	GetMarkerResultsMap              map[string]*GetMarkerResult
	AddSetNetAssetValuesResultsQueue []string
	GetNetAssetValueMap              map[string]map[string]*GetNetAssetValueResult
}

// MarkerCalls contains all the calls that the mock marker keeper makes.
type MarkerCalls struct {
	GetMarker            []sdk.AccAddress
	AddSetNetAssetValues []*AddSetNetAssetValuesArgs
	GetNetAssetValue     []*GetNetAssetValueArgs
}

// AddSetNetAssetValuesArgs is a record of a call that is made to AddSetNetAssetValues.
type AddSetNetAssetValuesArgs struct {
	marker         markertypes.MarkerAccountI
	netAssetValues []markertypes.NetAssetValue
	source         string
}

// GetNetAssetValueArgs is a record of a call that is made to GetNetAssetValue.
type GetNetAssetValueArgs struct {
	markerDenom string
	priceDenom  string
}

// GetMarkerResult contains the result args to return for a GetMarker call.
type GetMarkerResult struct {
	account markertypes.MarkerAccountI
	err     error
}

// GetNetAssetValueResult contains the result args to return for a GetNetAssetValue call.
type GetNetAssetValueResult struct {
	nav *markertypes.NetAssetValue
	err error
}

// NewMockMarkerKeeper creates a new empty MockMarkerKeeper.
// Follow it up with WithGetMarkerErr, WithGetMarkerAccount,
// and/or WithAddSetNetAssetValuesResults to dictate results.
func NewMockMarkerKeeper() *MockMarkerKeeper {
	return &MockMarkerKeeper{
		GetMarkerResultsMap: make(map[string]*GetMarkerResult),
		GetNetAssetValueMap: make(map[string]map[string]*GetNetAssetValueResult),
	}
}

// WithGetMarkerErr sets up this mock keeper to return the provided error when GetMarker is called for the given address.
// This method both updates the receiver and returns it.
func (k *MockMarkerKeeper) WithGetMarkerErr(addr sdk.AccAddress, err string) *MockMarkerKeeper {
	k.GetMarkerResultsMap[string(addr)] = &GetMarkerResult{err: errors.New(err)}
	return k
}

// WithGetMarkerAccount sets up this mock keeper to return the provided marker account when GetMarker is called for the given address.
// This method both updates the receiver and returns it.
func (k *MockMarkerKeeper) WithGetMarkerAccount(account markertypes.MarkerAccountI) *MockMarkerKeeper {
	k.GetMarkerResultsMap[string(account.GetAddress())] = &GetMarkerResult{account: account}
	return k
}

// WithAddSetNetAssetValuesResults queues up the provided error strings to be returned from AddSetNetAssetValues.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockMarkerKeeper) WithAddSetNetAssetValuesResults(errs ...string) *MockMarkerKeeper {
	k.AddSetNetAssetValuesResultsQueue = append(k.AddSetNetAssetValuesResultsQueue, errs...)
	return k
}

// WithGetNetAssetValueResult sets up this mock keepr to return the provided nav result when GetNetAssetValue is called for the given denoms.
// This method both updates the receiver and returns it.
func (k *MockMarkerKeeper) WithGetNetAssetValueResult(markerCoin, priceCoin sdk.Coin) *MockMarkerKeeper {
	if k.GetNetAssetValueMap[markerCoin.Denom] == nil {
		k.GetNetAssetValueMap[markerCoin.Denom] = make(map[string]*GetNetAssetValueResult)
	}
	k.GetNetAssetValueMap[markerCoin.Denom][priceCoin.Denom] = &GetNetAssetValueResult{
		nav: &markertypes.NetAssetValue{
			Price:  priceCoin,
			Volume: markerCoin.Amount.Uint64(),
		},
	}
	return k
}

// WithGetNetAssetValueError sets up this mock keepr to return the provided error when GetNetAssetValue is called for the given denoms.
// This method both updates the receiver and returns it.
func (k *MockMarkerKeeper) WithGetNetAssetValueError(markerDenom, priceDenom, errMsg string) *MockMarkerKeeper {
	if k.GetNetAssetValueMap[markerDenom] == nil {
		k.GetNetAssetValueMap[markerDenom] = make(map[string]*GetNetAssetValueResult)
	}
	k.GetNetAssetValueMap[markerDenom][priceDenom] = &GetNetAssetValueResult{err: errors.New(errMsg)}
	return k
}

func (k *MockMarkerKeeper) GetMarker(_ sdk.Context, address sdk.AccAddress) (markertypes.MarkerAccountI, error) {
	k.Calls.GetMarker = append(k.Calls.GetMarker, address)
	if rv, found := k.GetMarkerResultsMap[string(address)]; found {
		return rv.account, rv.err
	}
	return nil, nil
}

func (k *MockMarkerKeeper) AddSetNetAssetValues(_ sdk.Context, marker markertypes.MarkerAccountI, netAssetValues []markertypes.NetAssetValue, source string) error {
	k.Calls.AddSetNetAssetValues = append(k.Calls.AddSetNetAssetValues, NewAddSetNetAssetValuesArgs(marker, netAssetValues, source))
	var err error
	if len(k.AddSetNetAssetValuesResultsQueue) > 0 {
		if len(k.AddSetNetAssetValuesResultsQueue[0]) > 0 {
			err = errors.New(k.AddSetNetAssetValuesResultsQueue[0])
		}
		k.AddSetNetAssetValuesResultsQueue = k.AddSetNetAssetValuesResultsQueue[1:]
	}
	return err
}

func (k *MockMarkerKeeper) GetNetAssetValue(_ sdk.Context, markerDenom, priceDenom string) (*markertypes.NetAssetValue, error) {
	k.Calls.WithGetNetAssetValue(markerDenom, priceDenom)
	var nav *markertypes.NetAssetValue
	var err error
	if k.GetNetAssetValueMap != nil && k.GetNetAssetValueMap[markerDenom] != nil && k.GetNetAssetValueMap[markerDenom][priceDenom] != nil {
		nav = k.GetNetAssetValueMap[markerDenom][priceDenom].nav
		err = k.GetNetAssetValueMap[markerDenom][priceDenom].err
	}
	return nav, err
}

// assertGetMarkerCalls asserts that a mock keeper's Calls.GetMarker match the provided expected calls.
func (s *TestSuite) assertGetMarkerCalls(mk *MockMarkerKeeper, expected []sdk.AccAddress, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.GetMarker, s.getAddrName,
		msg+" GetMarker calls", args...)
}

// assertAddSetNetAssetValuesCalls asserts that a mock keeper's Calls.AddSetNetAssetValues match the provided expected calls.
func (s *TestSuite) assertAddSetNetAssetValuesCalls(mk *MockMarkerKeeper, expected []*AddSetNetAssetValuesArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.AddSetNetAssetValues, s.getAddSetNetAssetValuesArgsDenom,
		msg+" AddSetNetAssetValues calls", args...)
}

// assertGetNetAssetValueCalls asserts that a mock keeper's Calls.GetNetAssetValueArgs match the provided expected calls.
func (s *TestSuite) assertGetNetAssetValueCalls(mk *MockMarkerKeeper, expected []*GetNetAssetValueArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.GetNetAssetValue, s.getNetAssetValueArgsString,
		msg+" GetNetAssetValue calls", args...)
}

// assertMarkerKeeperCalls asserts that all the calls made to a mock marker keeper match the provided expected calls.
func (s *TestSuite) assertMarkerKeeperCalls(mk *MockMarkerKeeper, expected MarkerCalls, msg string, args ...interface{}) bool {
	s.T().Helper()
	rv := s.assertGetMarkerCalls(mk, expected.GetMarker, msg, args...)
	rv = s.assertAddSetNetAssetValuesCalls(mk, expected.AddSetNetAssetValues, msg, args...) && rv
	return s.assertGetNetAssetValueCalls(mk, expected.GetNetAssetValue, msg, args...) && rv
}

// WithGetNetAssetValue adds the provided args to the GetNetAssetValue list.
func (c *MarkerCalls) WithGetNetAssetValue(markerDenom, priceDenom string) {
	c.GetNetAssetValue = append(c.GetNetAssetValue, &GetNetAssetValueArgs{
		markerDenom: markerDenom,
		priceDenom:  priceDenom,
	})
}

// NewAddSetNetAssetValuesArgs creates a new record of args provided to a call to AddSetNetAssetValues.
func NewAddSetNetAssetValuesArgs(marker markertypes.MarkerAccountI, netAssetValues []markertypes.NetAssetValue, source string) *AddSetNetAssetValuesArgs {
	return &AddSetNetAssetValuesArgs{
		marker:         marker,
		netAssetValues: netAssetValues,
		source:         source,
	}
}

// getAddSetNetAssetValuesArgsDenom returns the denom of the marker in the provided AddSetNetAssetValuesArgs.
func (s *TestSuite) getAddSetNetAssetValuesArgsDenom(args *AddSetNetAssetValuesArgs) string {
	if args != nil && args.marker != nil {
		return args.marker.GetDenom()
	}
	return ""
}

// getNetAssetValueArgsString returns a string representation of the given GetNetAssetValueArgs.
func (s *TestSuite) getNetAssetValueArgsString(args *GetNetAssetValueArgs) string {
	if args == nil {
		return "<nil>"
	}
	md := args.markerDenom
	if len(md) == 0 {
		md = `""`
	}
	pd := args.priceDenom
	if len(pd) == 0 {
		pd = `""`
	}
	return fmt.Sprintf("%q->%q", md, pd)
}

// #############################################################################
// ###########################                      ############################
// #########################    MockMetadataKeeper    ##########################
// ###########################                      ############################
// #############################################################################

var _ exchange.MetadataKeeper = (*MockMetadataKeeper)(nil)

// MockMetadataKeeper satisfies the exchange.MetadataKeeper interface but just records the calls and allows dictation of results.
type MockMetadataKeeper struct {
	Calls                            MetadataCalls
	AddSetNetAssetValuesResultsQueue []string
	GetNetAssetValueResultsQueue     []*MdGetNetAssetValueResult
}

// MetadataCalls contains all the calls that the mock metadata keeper makes.
type MetadataCalls struct {
	AddSetNetAssetValues []*MdAddSetNetAssetValuesArgs
	GetNetAssetValue     []*MdGetNetAssetValueArgs
}

// MdAddSetNetAssetValuesArgs is a record of a call that is made to AddSetNetAssetValues (in the metadata module).
type MdAddSetNetAssetValuesArgs struct {
	ScopeID metadatatypes.MetadataAddress
	NAVs    []metadatatypes.NetAssetValue
	Source  string
}

type MdGetNetAssetValueResult provutils.Pair[*metadatatypes.NetAssetValue, string]

// MdGetNetAssetValueArgs is a record of a call that is made to GetNetAssetValue (in the metadata module).
type MdGetNetAssetValueArgs provutils.Pair[string, string]

// NewMockMetadataKeeper creates a new empty MockMetadataKeeper.
// Follow it up with WithAddSetNetAssetValuesErrors, WithGetNetAssetValueErrors,
// and/or WithGetNetAssetValueResults to dictate results.
func NewMockMetadataKeeper() *MockMetadataKeeper {
	return &MockMetadataKeeper{}
}

// WithAddSetNetAssetValuesErrors queues up the provided error strings to be returned from AddSetNetAssetValues.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
func (k *MockMetadataKeeper) WithAddSetNetAssetValuesErrors(errs ...string) *MockMetadataKeeper {
	k.AddSetNetAssetValuesResultsQueue = append(k.AddSetNetAssetValuesResultsQueue, errs...)
	return k
}

// WithGetNetAssetValueErrors queues up the provided error strings to be returned from GetNetAssetValue.
// An empty string means no error. Each entry is used only once. If entries run out, nil is returned.
// This method both updates the receiver and returns it.
// See also: WithGetNetAssetValueResults.
func (k *MockMetadataKeeper) WithGetNetAssetValueErrors(errs ...string) *MockMetadataKeeper {
	for _, err := range errs {
		k.GetNetAssetValueResultsQueue = append(k.GetNetAssetValueResultsQueue, NewMdGetNetAssetValueResult(nil, err))
	}
	return k
}

func (k *MockMetadataKeeper) WithGetNetAssetValueResult(price sdk.Coin) *MockMetadataKeeper {
	k.GetNetAssetValueResultsQueue = append(k.GetNetAssetValueResultsQueue,
		NewMdGetNetAssetValueResult(&metadatatypes.NetAssetValue{Price: price}, ""))
	return k
}

func (k *MockMetadataKeeper) AddSetNetAssetValues(_ sdk.Context, scopeID metadatatypes.MetadataAddress, navs []metadatatypes.NetAssetValue, source string) error {
	k.Calls.WithAddSetNetAssetValues(scopeID, navs, source)
	if len(k.AddSetNetAssetValuesResultsQueue) > 0 {
		rv := k.AddSetNetAssetValuesResultsQueue[0]
		k.AddSetNetAssetValuesResultsQueue = k.AddSetNetAssetValuesResultsQueue[1:]
		if len(rv) > 0 {
			return errors.New(rv)
		}
	}
	return nil
}

func (k *MockMetadataKeeper) GetNetAssetValue(_ sdk.Context, metadataDenom, priceDenom string) (*metadatatypes.NetAssetValue, error) {
	k.Calls.WithGetNetAssetValue(metadataDenom, priceDenom)
	if len(k.GetNetAssetValueResultsQueue) > 0 {
		rv := k.GetNetAssetValueResultsQueue[0]
		k.GetNetAssetValueResultsQueue = k.GetNetAssetValueResultsQueue[1:]
		return rv.Nav(), rv.Err()
	}
	return nil, nil
}

// assertSendCoinsCalls asserts that a mock keeper's Calls.AddSetNetAssetValues match the provided expected calls.
func (s *TestSuite) assertAddSetNetAssetValues(mk *MockMetadataKeeper, expected []*MdAddSetNetAssetValuesArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.AddSetNetAssetValues, mdAddNAVArgsString,
		msg+" AddSetNetAsset calls", args...)
}

// assertMdGetNetAssetValueCalls asserts that a mock keeper's Calls.GetNetAssetValue match the provided expected calls.
func (s *TestSuite) assertMdGetNetAssetValueCalls(mk *MockMetadataKeeper, expected []*MdGetNetAssetValueArgs, msg string, args ...interface{}) bool {
	s.T().Helper()
	return assertEqualSlice(s, expected, mk.Calls.GetNetAssetValue, mdNAVString,
		msg+" GetNetAssetValue calls", args...)
}

// assertMetadataKeeperCalls asserts that all the calls made to a mock metadata keeper match the provided expected calls.
func (s *TestSuite) assertMetadataKeeperCalls(mk *MockMetadataKeeper, expected MetadataCalls, msg string, args ...interface{}) bool {
	s.T().Helper()
	rv := s.assertAddSetNetAssetValues(mk, expected.AddSetNetAssetValues, msg, args...)
	return s.assertMdGetNetAssetValueCalls(mk, expected.GetNetAssetValue, msg, args...) && rv
}

// WithAddSetNetAssetValues adds the provided args to the AddSetNetAssetValues list.
func (c *MetadataCalls) WithAddSetNetAssetValues(scopeID metadatatypes.MetadataAddress, navs []metadatatypes.NetAssetValue, source string) {
	c.AddSetNetAssetValues = append(c.AddSetNetAssetValues, NewMdAddSetNetAssetValuesArgs(scopeID, navs, source))
}

// WithGetNetAssetValue adds the provided args to the GetNetAssetValue list.
func (c *MetadataCalls) WithGetNetAssetValue(metadataDenom, priceDenom string) {
	c.GetNetAssetValue = append(c.GetNetAssetValue, NewMdGetNetAssetValueArgs(metadataDenom, priceDenom))
}

// NewMdAddSetNetAssetValuesArgs creates a new record of the args provided to a call to the metadata keeper's AddSetNetAssetValues method.
func NewMdAddSetNetAssetValuesArgs(scopeID metadatatypes.MetadataAddress, navs []metadatatypes.NetAssetValue, source string) *MdAddSetNetAssetValuesArgs {
	return &MdAddSetNetAssetValuesArgs{
		ScopeID: scopeID,
		NAVs:    navs,
		Source:  source,
	}
}

// String returns a string of this MdAddSetNetAssetValuesArgs.
func (a MdAddSetNetAssetValuesArgs) String() string {
	navStrs := sliceStrings(a.NAVs, func(nav metadatatypes.NetAssetValue) string {
		return nav.String()
	})
	return fmt.Sprintf("%s(%s):%s",
		a.ScopeID.String(),
		a.Source,
		strings.Join(navStrs, ","),
	)
}

// mdAddNAVArgsString is the same as MdAddSetNetAssetValuesArgs.String but with a pointer arg.
func mdAddNAVArgsString(p *MdAddSetNetAssetValuesArgs) string {
	return p.String()
}

// NewMdGetNetAssetValueArgs creates a new record of the args provided to a call to the metadata keeper's GetNetAssetValue method.
func NewMdGetNetAssetValueArgs(metadataDenom, priceDenom string) *MdGetNetAssetValueArgs {
	return (*MdGetNetAssetValueArgs)(provutils.NewPair(metadataDenom, priceDenom))
}

// MetadataDenom gets the metadataDenom string associated with the GetNetAssetValue method.
func (p MdGetNetAssetValueArgs) MetadataDenom() string {
	return p.A
}

// PriceDenom gets the priceDenom string associated with the GetNetAssetValue method.
func (p MdGetNetAssetValueArgs) PriceDenom() string {
	return p.B
}

// String returns a string of this MdGetNetAssetValueArgs.
func (p MdGetNetAssetValueArgs) String() string {
	return fmt.Sprintf("%s:%s", p.MetadataDenom(), p.PriceDenom())
}

// mdNAVString is the same as MdGetNetAssetValueArgs.String but with a pointer arg.
func mdNAVString(p *MdGetNetAssetValueArgs) string {
	return p.String()
}

// NewMdGetNetAssetValueResult creates a record of things associated with the GetNetAssetValue return values.
func NewMdGetNetAssetValueResult(nav *metadatatypes.NetAssetValue, err string) *MdGetNetAssetValueResult {
	return (*MdGetNetAssetValueResult)(provutils.NewPair(nav, err))
}

// Nav returns the Net Asset Value arg of this MdGetNetAssetValueResult.
func (p MdGetNetAssetValueResult) Nav() *metadatatypes.NetAssetValue {
	return p.A
}

// Err returns the error arg of this MdGetNetAssetValueResult by converting it's string (B) into either an error or nil.
func (p MdGetNetAssetValueResult) Err() error {
	if len(p.B) == 0 {
		return nil
	}
	return errors.New(p.B)
}
