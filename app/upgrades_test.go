package app

import (
	"bytes"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"

	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func TestAddGovV1SubmitFee(t *testing.T) {
	v1TypeURL := "/cosmos.gov.v1.MsgSubmitProposal"
	v1B1TypeURL := "/cosmos.gov.v1beta1.MsgSubmitProposal"

	startingMsg := `Creating message fee for "` + v1TypeURL + `" if it doesn't already exist.`
	successMsg := func(amt string) string {
		return `Successfully set fee for "` + v1TypeURL + `" with amount "` + amt + `".`
	}

	coin := func(denom string, amt int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amt)
		return &rv
	}

	tests := []struct {
		name     string
		v1Amt    *sdk.Coin
		v1B1Amt  *sdk.Coin
		expInLog []string
		expAmt   sdk.Coin
	}{
		{
			name:    "v1 fee already exists",
			v1Amt:   coin("foocoin", 88),
			v1B1Amt: coin("betacoin", 99),
			expInLog: []string{
				startingMsg,
				`Message fee for "` + v1TypeURL + `" already exists with amount "88foocoin". Nothing to do.`,
			},
			expAmt: *coin("foocoin", 88),
		},
		{
			name:    "v1beta1 exists",
			v1B1Amt: coin("betacoin", 99),
			expInLog: []string{
				startingMsg,
				`Copying "` + v1B1TypeURL + `" fee to "` + v1TypeURL + `".`,
				successMsg("99betacoin"),
			},
			expAmt: *coin("betacoin", 99),
		},
		{
			name: "brand new",
			expInLog: []string{
				startingMsg,
				`Creating "` + v1TypeURL + `" fee.`,
				successMsg("100000000000nhash"),
			},
			expAmt: *coin("nhash", 100_000_000_000),
		},
	}

	// Create a loggerMaker that captures info+ log output.
	var logBuffer bytes.Buffer
	bufferedLoggerMaker := func() log.Logger {
		lw := zerolog.ConsoleWriter{Out: &logBuffer}
		logger := zerolog.New(lw).Level(zerolog.InfoLevel).With().Timestamp().Logger()
		return server.ZeroLogWrapper{Logger: logger}
	}
	defer SetLoggerMaker(SetLoggerMaker(bufferedLoggerMaker))

	app := Setup(t)
	ctx := app.NewContext(false, tmproto.Header{})

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set/unset the v1 fee.
			if tc.v1Amt != nil {
				fee := msgfeetypes.NewMsgFee(v1TypeURL, *tc.v1Amt, "", 0)
				require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, fee), "SetMsgFee v1")
			} else {
				err := app.MsgFeesKeeper.RemoveMsgFee(ctx, v1TypeURL)
				if err != nil && !errors.Is(err, msgfeetypes.ErrMsgFeeDoesNotExist) {
					require.NoError(t, err, "RemoveMsgFee v1")
				}
			}

			// Set/unset the v1beta1 fee.
			if tc.v1B1Amt != nil {
				fee := msgfeetypes.NewMsgFee(v1B1TypeURL, *tc.v1B1Amt, "", 0)
				require.NoError(t, app.MsgFeesKeeper.SetMsgFee(ctx, fee), "SetMsgFee v1beta1")
			} else {
				err := app.MsgFeesKeeper.RemoveMsgFee(ctx, v1B1TypeURL)
				if err != nil && !errors.Is(err, msgfeetypes.ErrMsgFeeDoesNotExist) {
					require.NoError(t, err, "RemoveMsgFee v1")
				}
			}

			// Reset the log buffer to clear out unrelated entries.
			logBuffer.Reset()
			// Call AddGovV1SubmitFee and relog its output (to help if things fail).
			testFunc := func() {
				AddGovV1SubmitFee(ctx, app)
			}
			require.NotPanics(t, testFunc, "AddGovV1SubmitFee")
			logOutput := logBuffer.String()
			t.Logf("AddGovV1SubmitFee log output:\n%s", logOutput)

			// Make sure the log has the expected lines.
			for _, exp := range tc.expInLog {
				assert.Contains(t, logOutput, exp, "AddGovV1SubmitFee log output")
			}

			// Get the fee and make sure it's now as expected.
			fee, err := app.MsgFeesKeeper.GetMsgFee(ctx, v1TypeURL)
			require.NoError(t, err, "GetMsgFee(%q) error", v1TypeURL)
			require.NotNil(t, fee, "GetMsgFee(%q) value", v1TypeURL)
			actFeeAmt := fee.AdditionalFee
			assert.Equal(t, tc.expAmt.String(), actFeeAmt.String(), "final %s fee amount", v1TypeURL)
		})
	}
}
