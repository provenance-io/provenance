package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

func TestRemoveLeaveGroupMsgFee(t *testing.T) {
	var app *App
	require.NotPanics(t, func() {
		app = Setup(t)
	}, "Setup")
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	typeURL := sdk.MsgTypeURL(&group.MsgLeaveGroup{})

	tests := []struct {
		name  string
		setup func(t *testing.T)
	}{
		{
			name: "fee does not exist",
			setup: func(_ *testing.T) {
				_ = app.MsgFeesKeeper.RemoveMsgFee(ctx, typeURL)
			},
		},
		{
			name: "fee exists",
			setup: func(t *testing.T) {
				msgFee := msgfeetypes.MsgFee{
					MsgTypeUrl:           typeURL,
					AdditionalFee:        sdk.NewInt64Coin("feecoin", 8),
					Recipient:            "",
					RecipientBasisPoints: 0,
				}
				err := app.MsgFeesKeeper.SetMsgFee(ctx, msgFee)
				require.NoError(t, err, "SetMsgFee error")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(t)
			var err error
			testFunc := func() {
				err = RemoveLeaveGroupMsgFee(ctx, app)
			}
			require.NotPanics(t, testFunc, "RemoveLeaveGroupMsgFee")
			require.NoError(t, err, "RemoveLeaveGroupMsgFee error")
			msgFee, err := app.MsgFeesKeeper.GetMsgFee(ctx, typeURL)
			assert.NoError(t, err, "GetMsgFee error")
			assert.Nil(t, msgFee, "GetMsgFee value")
		})
	}
}
