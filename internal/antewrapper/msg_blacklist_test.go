package antewrapper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/provenance-io/provenance/internal/antewrapper"
)

type TestTx struct {
	Msgs []sdk.Msg
}

func NewTestTx(msgs ...sdk.Msg) TestTx {
	return TestTx{Msgs: msgs}
}

var (
	_ sdk.Tx     = (*TestTx)(nil)
	_ ante.GasTx = (*TestTx)(nil)
)

// GetMsgs satisfies sdk.Tx interface.
func (t TestTx) GetMsgs() []sdk.Msg {
	return t.Msgs
}

// ValidateBasic satisfies sdk.Tx interface.
func (t TestTx) ValidateBasic() error {
	return nil
}

// GetGas satisfies ante.GasTx interface.
func (t TestTx) GetGas() uint64 {
	return 1_000_000_000
}

const AllGood = "terminator called"

func terminator(ctx sdk.Context, _ sdk.Tx, _ bool) (newCtx sdk.Context, err error) {
	return ctx, errors.New(AllGood)
}

func badMsgErr(msg sdk.Msg) string {
	return sdk.MsgTypeURL(msg) + " messages are not allowed: invalid request"
}

func TestMsgTypeBlacklistContextDecorator_AnteHandle(t *testing.T) {
	goodMsg := &banktypes.MsgSend{}

	perVMsg := &vestingtypes.MsgCreatePeriodicVestingAccount{}
	perVMsgErr := badMsgErr(perVMsg)
	vMsg := &vestingtypes.MsgCreateVestingAccount{}
	vMsgErr := badMsgErr(vMsg)
	permLVMsg := &vestingtypes.MsgCreatePermanentLockedAccount{}
	permLVMsgErr := badMsgErr(permLVMsg)

	tests := []struct {
		name string
		tx   sdk.Tx
		exp  string
	}{
		{
			name: "good",
			tx:   NewTestTx(goodMsg),
			exp:  AllGood,
		},
		{
			name: "periodic vesting",
			tx:   NewTestTx(perVMsg),
			exp:  perVMsgErr,
		},
		{
			name: "standard vesting",
			tx:   NewTestTx(vMsg),
			exp:  vMsgErr,
		},
		{
			name: "permanent locked",
			tx:   NewTestTx(permLVMsg),
			exp:  permLVMsgErr,
		},
		{
			name: "good good",
			tx:   NewTestTx(goodMsg, goodMsg),
			exp:  AllGood,
		},
		{
			name: "bad good",
			tx:   NewTestTx(perVMsg, goodMsg),
			exp:  perVMsgErr,
		},
		{
			name: "good bad",
			tx:   NewTestTx(goodMsg, perVMsg),
			exp:  perVMsgErr,
		},
		{
			name: "bad bad",
			tx:   NewTestTx(permLVMsg, vMsg),
			exp:  permLVMsgErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bl := antewrapper.NewMsgTypeBlacklistContextDecorator()
			_, err := bl.AnteHandle(sdk.Context{}, tc.tx, false, terminator)
			assert.EqualError(t, err, tc.exp, "MsgTypeBlacklistContextDecorator.AnteHandle")
		})
	}
}

func (s *AnteTestSuite) TestBlacklistedMsgs() {
	s.SetupTest(true)
	badMsgs := []sdk.Msg{
		&vestingtypes.MsgCreatePeriodicVestingAccount{},
		&vestingtypes.MsgCreateVestingAccount{},
		&vestingtypes.MsgCreatePeriodicVestingAccount{},
	}

	for _, msg := range badMsgs {
		name := sdk.MsgTypeURL(msg)
		exp := badMsgErr(msg)
		s.Run(name, func() {
			_, err := s.anteHandler(s.ctx, NewTestTx(msg), true)
			s.Assert().EqualError(err, exp, "anteHandler")
		})
	}
}
