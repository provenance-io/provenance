package antewrapper

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// newErr returns the provided string as an error unless it's an empty string, then this returns nil.
func newErr(errStr string) error {
	if len(errStr) == 0 {
		return nil
	}
	return errors.New(errStr)
}

// assertEqualMsgTypes asserts that the provided Msg slices have the same type URLs.
// Returns true if not nil, false if the assertion fails.
func assertEqualMsgTypes(t *testing.T, expected, actual []sdk.Msg, msgAndArgs ...interface{}) bool {
	t.Helper()
	expStrs := msgTypeURLs(expected)
	actStrs := msgTypeURLs(actual)
	return assert.Equal(t, expStrs, actStrs, msgAndArgs...)
}

// assertNotNilInterface asserts that the provided actual value is not a nil interface value.
// Returns true if not nil, false if the assertion fails.
func assertNotNilInterface(t *testing.T, actual interface{}, msgAndArgs ...interface{}) bool {
	t.Helper()

	// Make sure it's not outright nil.
	if !assert.NotNil(t, actual, msgAndArgs...) {
		return false
	}

	// Make sure it's not a nil wrapped in a concrete type, e.g. (*MockFlatFeesKeeper)(nil).
	val := reflect.ValueOf(actual)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return assert.Fail(t, "Expected interface value not to be nil.", msgAndArgs...)
	}
	return true
}

// MockFeeTx satisfies both the sdk.Tx and sdk.FeeTx interfaces.
type MockFeeTx struct {
	ID string

	Msgs      []sdk.Msg
	MsgsV2    []protov2.Message
	MsgsV2Err error
	Gas       uint64
	Fee       sdk.Coins
	Payer     []byte
	Granter   []byte
}

var _ sdk.Tx = (*MockFeeTx)(nil)
var _ sdk.FeeTx = (*MockFeeTx)(nil)

func NewMockFeeTx(id string) *MockFeeTx {
	return &MockFeeTx{ID: id}
}

// WithGetMsgs sets the GetMsgs() return value in this, and returns this.
func (t *MockFeeTx) WithGetMsgs(msgs []sdk.Msg) *MockFeeTx {
	t.Msgs = msgs
	return t
}

// WithGetMsgsV2 sets the GetMsgsV2() return values in this, and returns this.
func (t *MockFeeTx) WithGetMsgsV2(msgsV2 []protov2.Message, errStr string) *MockFeeTx {
	t.MsgsV2 = msgsV2
	t.MsgsV2Err = newErr(errStr)
	return t
}

// WithGas sets the GetGas() return value in this, and returns this.
func (t *MockFeeTx) WithGas(gas uint64) *MockFeeTx {
	t.Gas = gas
	return t
}

// WithFee sets the GetFee() return value in this, and returns this.
func (t *MockFeeTx) WithFee(fee sdk.Coins) *MockFeeTx {
	t.Fee = fee
	return t
}

// WithFeeStr sets the GetFee() return value in this requiring it to be valid, and returns this.
func (t *MockFeeTx) WithFeeStr(tt *testing.T, feeStr string) *MockFeeTx {
	var err error
	t.Fee, err = sdk.ParseCoinsNormalized(feeStr)
	require.NoError(tt, err, "ParseCoinsNormalized(%q)", feeStr)
	return t
}

// WithFeePayer sets the FeePayer() return value in this, and returns this.
func (t *MockFeeTx) WithFeePayer(payer []byte) *MockFeeTx {
	t.Payer = payer
	return t
}

// WithFeeGranter sets the FeeGranter() return value in this, and returns this.
func (t *MockFeeTx) WithFeeGranter(granter []byte) *MockFeeTx {
	t.Granter = granter
	return t
}

func (t *MockFeeTx) GetMsgs() []sdk.Msg                    { return t.Msgs }
func (t *MockFeeTx) GetMsgsV2() ([]protov2.Message, error) { return t.MsgsV2, t.MsgsV2Err }
func (t *MockFeeTx) GetGas() uint64                        { return t.Gas }
func (t *MockFeeTx) GetFee() sdk.Coins                     { return t.Fee }
func (t *MockFeeTx) FeePayer() []byte                      { return t.Payer }
func (t *MockFeeTx) FeeGranter() []byte                    { return t.Granter }
func (t *MockFeeTx) String() string                        { return t.ID }

// NotFeeTx satisfies the sdk.Tx interface, but not the sdk.FeeTx interface.
type NotFeeTx struct {
	ID string
}

var _ sdk.Tx = (*NotFeeTx)(nil)

func NewNotFeeTx(id string) *NotFeeTx {
	return &NotFeeTx{ID: id}
}

func (t *NotFeeTx) GetMsgs() []sdk.Msg                    { return nil }
func (t *NotFeeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (t *NotFeeTx) String() string                        { return t.ID }

// MockGasMeter satisfies the GasMeter interface and allows for mocking and recording some calls.
type MockGasMeter struct {
	GasConsumedReturn storetypes.Gas
	GasConsumedCalled bool

	ConsumeGasCall *ConsumeGasArgs

	StringReturn string
}

var _ storetypes.GasMeter = (*MockGasMeter)(nil)

func NewMockGasMeter() *MockGasMeter {
	return &MockGasMeter{}
}

// WithGasConsumed sets the GasConsumed() return value in this, and returns this.
func (g *MockGasMeter) WithGasConsumed(gas storetypes.Gas) *MockGasMeter {
	g.GasConsumedReturn = gas
	return g
}

// WithString sets the String() reeturn value in this, and returns this.
func (g *MockGasMeter) WithString(str string) *MockGasMeter {
	g.StringReturn = str
	return g
}

func (g *MockGasMeter) GasConsumed() storetypes.Gas {
	g.GasConsumedCalled = true
	return g.GasConsumedReturn
}

func (g *MockGasMeter) ConsumeGas(amount storetypes.Gas, descriptor string) {
	g.ConsumeGasCall = NewConsumeGasArgs(amount, descriptor)
}

func (g *MockGasMeter) String() string {
	return g.StringReturn
}

func (g *MockGasMeter) GasConsumedToLimit() storetypes.Gas   { panic("not implemented") }
func (g *MockGasMeter) GasRemaining() storetypes.Gas         { panic("not implemented") }
func (g *MockGasMeter) Limit() storetypes.Gas                { panic("not implemented") }
func (g *MockGasMeter) RefundGas(_ storetypes.Gas, _ string) { panic("not implemented") }
func (g *MockGasMeter) IsPastLimit() bool                    { panic("not implemented") }
func (g *MockGasMeter) IsOutOfGas() bool                     { panic("not implemented") }

// AssertGasConsumedCalled asserts that there was or wasn't a call made to GasConsumed.
// Returns true if as expected, false if the assertion fails.
func (g *MockGasMeter) AssertGasConsumedCalled(t *testing.T, expected bool) bool {
	t.Helper()
	return assert.Equal(t, expected, g.GasConsumedCalled, "Whether GasConsumed() was called.")
}

// AssertConsumeGasCall asserts that the call made to ConsumeGas equals the expected.
// Returns true if equal, false if the assertion fails.
func (g *MockGasMeter) AssertConsumeGasCall(t *testing.T, expected *ConsumeGasArgs) bool {
	t.Helper()
	return g.ConsumeGasCall.AssertEqual(t, expected)
}

// ConsumeGasArgs are the arguments of the GasMeter.ConsumeGas method.
type ConsumeGasArgs struct {
	Amount     storetypes.Gas
	Descriptor string
}

func NewConsumeGasArgs(amount storetypes.Gas, descriptor string) *ConsumeGasArgs {
	return &ConsumeGasArgs{
		Amount:     amount,
		Descriptor: descriptor,
	}
}

// AssertEqual asserts that this equals the provided expected values.
// Returns true if equal, false if the assertion fails.
func (a *ConsumeGasArgs) AssertEqual(t *testing.T, expected *ConsumeGasArgs) bool {
	t.Helper()
	if !assert.Equal(t, expected == nil, a == nil, "ConsumeGasArgs: whether expected and actual are nil") {
		return false
	}
	if expected == nil {
		return true
	}

	rv := AssertEqualGas(t, expected.Amount, a.Amount, "ConsumeGasArgs: amount")
	rv = assert.Equal(t, expected.Descriptor, a.Descriptor, "ConsumeGasArgs: descriptor") && rv
	return rv
}

// MockFeegrantKeeper satisfies the FeegrantKeeper interface and allows for mocking and recording calls.
type MockFeegrantKeeper struct {
	UseGrantedFeesReturnError error
	UseGrantedFeesCall        *UseGrantedFeesArgs
}

var _ FeegrantKeeper = (*MockFeegrantKeeper)(nil)

func NewMockFeegrantKeeper() *MockFeegrantKeeper {
	return &MockFeegrantKeeper{}
}

// WithUseGrantedFees sets the value to return from UseGrantedFees in this, and returns this.
func (k *MockFeegrantKeeper) WithUseGrantedFees(errStr string) *MockFeegrantKeeper {
	k.UseGrantedFeesReturnError = newErr(errStr)
	return k
}

func (k *MockFeegrantKeeper) UseGrantedFees(_ context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
	k.UseGrantedFeesCall = NewUseGrantedFeesArgs(granter, grantee, fee, msgs)
	return k.UseGrantedFeesReturnError
}

// AssertUseGrantedFeesCall asserts that the call made to UseGrantedFees equals the expected.
// Returns true if equal, false if the assertion fails.
func (k *MockFeegrantKeeper) AssertUseGrantedFeesCall(t *testing.T, expected *UseGrantedFeesArgs) bool {
	t.Helper()
	return k.UseGrantedFeesCall.AssertEqual(t, expected)
}

// UseGrantedFeesArgs are the arguments of the FeegrantKeeper.UseGrantedFees method.
type UseGrantedFeesArgs struct {
	Granter sdk.AccAddress
	Grantee sdk.AccAddress
	Fee     sdk.Coins
	Msgs    []sdk.Msg
}

func NewUseGrantedFeesArgs(granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) *UseGrantedFeesArgs {
	return &UseGrantedFeesArgs{Granter: granter, Grantee: grantee, Fee: fee, Msgs: msgs}
}

// AssertEqual asserts that this equals the provided expected values.
// Returns true if equal, false if the assertion fails.
func (a *UseGrantedFeesArgs) AssertEqual(t *testing.T, expected *UseGrantedFeesArgs) bool {
	t.Helper()
	if !assert.Equal(t, expected == nil, a == nil, "UseGrantedFeesArgs: whether expected and actual are nil") {
		return false
	}
	if expected == nil {
		return true
	}

	rv := assert.Equal(t, expected.Granter, a.Granter, "UseGrantedFeesArgs: granter")
	rv = assert.Equal(t, expected.Grantee, a.Grantee, "UseGrantedFeesArgs: grantee") && rv
	rv = assert.Equal(t, expected.Fee.String(), a.Fee.String(), "UseGrantedFeesArgs: fee") && rv
	rv = assertEqualMsgTypes(t, expected.Msgs, a.Msgs, "UseGrantedFeesArgs: msgs") && rv
	return rv
}

// MockBankKeeper satisfies the BankKeeper interface and allows for mocking and recording calls.
type MockBankKeeper struct {
	SendCoinsFromAccountToModuleReturnError error
	SendCoinsFromAccountToModuleReturnCall  *SendCoinsFromAccountToModuleArgs

	Balances map[string]sdk.Coins
}

var _ BankKeeper = (*MockBankKeeper)(nil)

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{Balances: make(map[string]sdk.Coins)}
}

// WithSendCoinsFromAccountToModule sets the value to return from SendCoinsFromAccountToModule in this, and returns this.
func (k *MockBankKeeper) WithSendCoinsFromAccountToModule(errStr string) *MockBankKeeper {
	k.SendCoinsFromAccountToModuleReturnError = newErr(errStr)
	return k
}

// WithBalance defines the balance of the provided address for use in any method that gets balance info for an address,
// (e.g. GetBalance) in this, and returns this.
func (k *MockBankKeeper) WithBalance(addr sdk.AccAddress, amount sdk.Coins) *MockBankKeeper {
	k.Balances[string(addr)] = amount
	return k
}

func (k *MockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	k.SendCoinsFromAccountToModuleReturnCall = NewSendCoinsFromAccountToModuleArgs(senderAddr, recipientModule, amt)
	return k.SendCoinsFromAccountToModuleReturnError
}

func (k *MockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	amt, ok := k.Balances[string(addr)]
	if !ok {
		return sdk.Coin{Denom: denom, Amount: sdkmath.ZeroInt()}
	}
	_, rv := amt.Find(denom)
	return rv
}

// AssertSendCoinsFromAccountToModuleCall asserts that the call made to SendCoinsFromAccountToModule equals the expected.
// Returns true if equal, false if the assertion fails.
func (k *MockBankKeeper) AssertSendCoinsFromAccountToModuleCall(t *testing.T, expected *SendCoinsFromAccountToModuleArgs) bool {
	t.Helper()
	return k.SendCoinsFromAccountToModuleReturnCall.AssertEqual(t, expected)
}

// SendCoinsFromAccountToModuleArgs are the args of the SendCoinsFromAccountToModule method.
type SendCoinsFromAccountToModuleArgs struct {
	SenderAddr      sdk.AccAddress
	RecipientModule string
	Amt             sdk.Coins
}

func NewSendCoinsFromAccountToModuleArgs(senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) *SendCoinsFromAccountToModuleArgs {
	return &SendCoinsFromAccountToModuleArgs{
		SenderAddr:      senderAddr,
		RecipientModule: recipientModule,
		Amt:             amt,
	}
}

// AssertEqual asserts that this equals the expected values.
// Returns true if equal, false if the assertion fails.
func (a *SendCoinsFromAccountToModuleArgs) AssertEqual(t *testing.T, expected *SendCoinsFromAccountToModuleArgs) bool {
	t.Helper()
	if !assert.Equal(t, expected == nil, a == nil, "SendCoinsFromAccountToModuleArgs: whether expected and actual are nil") {
		return false
	}
	if expected == nil {
		return true
	}

	rv := assert.Equal(t, expected.SenderAddr, a.SenderAddr, "SendCoinsFromAccountToModuleArgs: SenderAddr")
	rv = assert.Equal(t, expected.RecipientModule, a.RecipientModule, "SendCoinsFromAccountToModuleArgs: RecipientModule") && rv
	rv = assert.Equal(t, expected.Amt.String(), a.Amt.String(), "SendCoinsFromAccountToModuleArgs: Amt") && rv
	return rv
}

// MockFlatFeesKeeper satisfies the FlatFeesKeeper interface and allows for mocking and recording calls.
type MockFlatFeesKeeper struct {
	CalculateMsgCostReturnUpFront   sdk.Coins
	CalculateMsgCostReturnOnSuccess sdk.Coins
	CalculateMsgCostReturnError     error
	CalculateMsgCostCall            []sdk.Msg

	ExpandMsgsReturnMsgs  []sdk.Msg
	ExpandMsgsReturnError error
	ExpandMsgsCall        []sdk.Msg
}

var _ FlatFeesKeeper = (*MockFlatFeesKeeper)(nil)

func NewMockFlatFeesKeeper() *MockFlatFeesKeeper {
	return &MockFlatFeesKeeper{}
}

// WithCalculateMsgCost sets the values to return from CalculateMsgCost in this, and returns this.
func (k *MockFlatFeesKeeper) WithCalculateMsgCost(upFront, onSuccess sdk.Coins, errStr string) *MockFlatFeesKeeper {
	k.CalculateMsgCostReturnUpFront = upFront
	k.CalculateMsgCostReturnOnSuccess = onSuccess
	k.CalculateMsgCostReturnError = newErr(errStr)
	return k
}

// WithExpandMsgs sets the values to return from ExpandMsgs in this, and returns this.
// If nil, nil, is provided (or this isn't called, the mock ExpandMsgs method will return the msgs it receives.
// Otherwise, it will return what was provided here regardless of the msgs provided.
func (k *MockFlatFeesKeeper) WithExpandMsgs(msgs []sdk.Msg, errStr string) *MockFlatFeesKeeper {
	k.ExpandMsgsReturnMsgs = msgs
	k.ExpandMsgsReturnError = newErr(errStr)
	return k
}

func (k *MockFlatFeesKeeper) CalculateMsgCost(_ sdk.Context, msgs ...sdk.Msg) (upFront sdk.Coins, onSuccess sdk.Coins, err error) {
	k.CalculateMsgCostCall = msgs
	return k.CalculateMsgCostReturnUpFront, k.CalculateMsgCostReturnOnSuccess, k.CalculateMsgCostReturnError
}

func (k *MockFlatFeesKeeper) ExpandMsgs(msgs []sdk.Msg) ([]sdk.Msg, error) {
	k.ExpandMsgsCall = msgs
	// If nothing was configured to return here, return the provided msgs.
	if k.ExpandMsgsReturnMsgs == nil && k.ExpandMsgsReturnError == nil {
		return msgs, nil
	}
	return k.ExpandMsgsReturnMsgs, k.ExpandMsgsReturnError
}

// AssertCalculateMsgCostCall asserts that the call made to CalculateMsgCost equals the expected.
// Returns true if equal, false if the assertion fails.
func (k *MockFlatFeesKeeper) AssertCalculateMsgCostCall(t *testing.T, expected []sdk.Msg) bool {
	t.Helper()
	return assertEqualMsgTypes(t, expected, k.CalculateMsgCostCall, "CalculateMsgCost: msg types provided")
}

// AssertExpandMsgsCall asserts that the call made to ExpandMsgs equals the expected.
// Returns true if equal, false if the assertion fails.
func (k *MockFlatFeesKeeper) AssertExpandMsgsCall(t *testing.T, expected []sdk.Msg) bool {
	t.Helper()
	return assertEqualMsgTypes(t, expected, k.ExpandMsgsCall, "ExpandMsgs: msg types provided")
}
