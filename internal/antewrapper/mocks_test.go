package antewrapper

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// WithGetMsgs sets the GetMsgs() return value.
func (t *MockFeeTx) WithGetMsgs(msgs []sdk.Msg) *MockFeeTx {
	t.Msgs = msgs
	return t
}

// WithGetMsgsV2 sets the GetMsgsV2() return values.
func (t *MockFeeTx) WithGetMsgsV2(msgsV2 []protov2.Message, errStr string) *MockFeeTx {
	t.MsgsV2 = msgsV2
	if len(errStr) > 0 {
		t.MsgsV2Err = errors.New(errStr)
	} else {
		t.MsgsV2Err = nil
	}
	return t
}

// WithGas sets the GetGas() return value.
func (t *MockFeeTx) WithGas(gas uint64) *MockFeeTx {
	t.Gas = gas
	return t
}

// WithFee sets the GetFee() return value.
func (t *MockFeeTx) WithFee(fee sdk.Coins) *MockFeeTx {
	t.Fee = fee
	return t
}

// WithFeeStr sets the GetFee() return value requiring it to be valid.
func (t *MockFeeTx) WithFeeStr(tt *testing.T, feeStr string) *MockFeeTx {
	var err error
	t.Fee, err = sdk.ParseCoinsNormalized(feeStr)
	require.NoError(tt, err, "ParseCoinsNormalized(%q)", feeStr)
	return t
}

// WithFeePayer sets the FeePayer() return value.
func (t *MockFeeTx) WithFeePayer(payer []byte) *MockFeeTx {
	t.Payer = payer
	return t
}

// WithFeeGranter sets the FeeGranter() return value.
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

type MockFeegrantKeeper struct {
	UseGrantedFeesErr  error
	UseGrantedFeesCall *UseGrantedFeesArgs
}

var _ FeegrantKeeper = (*MockFeegrantKeeper)(nil)

func NewMockFeegrantKeeper() *MockFeegrantKeeper {
	return &MockFeegrantKeeper{}
}

func (k *MockFeegrantKeeper) WithUseGrantedFees(errStr string) *MockFeegrantKeeper {
	if len(errStr) > 0 {
		k.UseGrantedFeesErr = errors.New(errStr)
	} else {
		k.UseGrantedFeesErr = nil
	}
	return k
}

func (k *MockFeegrantKeeper) UseGrantedFees(_ context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
	k.UseGrantedFeesCall = NewUseGrantedFeesArgs(granter, grantee, fee, msgs)
	return k.UseGrantedFeesErr
}

func (k *MockFeegrantKeeper) AssertUseGrantedFeesCall(t *testing.T, expected *UseGrantedFeesArgs) bool {
	t.Helper()
	return k.UseGrantedFeesCall.AssertEqual(t, expected)
}

type UseGrantedFeesArgs struct {
	Granter sdk.AccAddress
	Grantee sdk.AccAddress
	Fee     sdk.Coins
	Msgs    []sdk.Msg
}

func NewUseGrantedFeesArgs(granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) *UseGrantedFeesArgs {
	return &UseGrantedFeesArgs{Granter: granter, Grantee: grantee, Fee: fee, Msgs: msgs}
}

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
	expMsgs := msgTypeURLs(expected.Msgs)
	actMsgs := msgTypeURLs(a.Msgs)
	rv = assert.Equal(t, expMsgs, actMsgs, "UseGrantedFeesArgs: msgs") && rv
	return rv
}

type MockBankKeeper struct {
	SendCoinsFromAccountToModuleErr  error
	SendCoinsFromAccountToModuleCall *SendCoinsFromAccountToModuleArgs

	Balances map[string]sdk.Coins
}

var _ BankKeeper = (*MockBankKeeper)(nil)

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{Balances: make(map[string]sdk.Coins)}
}

func (k *MockBankKeeper) WithSendCoinsFromAccountToModule(errStr string) *MockBankKeeper {
	if len(errStr) > 0 {
		k.SendCoinsFromAccountToModuleErr = errors.New(errStr)
	} else {
		k.SendCoinsFromAccountToModuleErr = nil
	}
	return k
}

func (k *MockBankKeeper) WithBalance(addr sdk.AccAddress, amount sdk.Coins) *MockBankKeeper {
	k.Balances[string(addr)] = amount
	return k
}

func (k *MockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	k.SendCoinsFromAccountToModuleCall = NewSendCoinsFromAccountToModuleArgs(senderAddr, recipientModule, amt)
	return k.SendCoinsFromAccountToModuleErr
}

func (k *MockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	amt, ok := k.Balances[string(addr)]
	if !ok {
		return sdk.Coin{Denom: denom, Amount: sdkmath.ZeroInt()}
	}
	_, rv := amt.Find(denom)
	return rv
}

func (k *MockBankKeeper) AssertSendCoinsFromAccountToModuleCall(t *testing.T, expected *SendCoinsFromAccountToModuleArgs) bool {
	t.Helper()
	return k.SendCoinsFromAccountToModuleCall.AssertEqual(t, expected)
}

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
