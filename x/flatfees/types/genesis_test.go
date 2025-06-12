package types_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/mocks"

	. "github.com/provenance-io/provenance/x/flatfees/types"
)

func TestGenesisState_Validate(t *testing.T) {
	pioconfig.SetProvConfig("pineapple")

	coin := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{
			Denom:  denom,
			Amount: sdkmath.NewInt(amt),
		}
	}
	cz := func(coins ...sdk.Coin) []sdk.Coin {
		return coins
	}
	joinErrs := func(errs ...string) string {
		return strings.Join(errs, "\n")
	}

	tests := []struct {
		name   string
		state  *GenesisState
		expErr string
	}{
		{
			name:   "nil",
			state:  nil,
			expErr: "flatfees genesis state cannot be nil",
		},
		{
			name: "invalid params",
			state: &GenesisState{
				Params: Params{
					DefaultCost: coin(100, "banana"),
					ConversionFactor: ConversionFactor{
						DefinitionAmount: coin(10, "apple"),
						ConvertedAmount:  coin(10, "banana"),
					},
				},
			},
			expErr: "invalid flatfees params: default cost denom \"banana\" does not equal conversion factor base amount denom \"apple\"",
		},
		{
			name: "invalid msg fee",
			state: &GenesisState{
				Params:  DefaultParams(),
				MsgFees: []*MsgFee{nil},
			},
			expErr: "invalid MsgFees[0]: nil MsgFee not allowed",
		},
		{
			name: "duplicate msg fee",
			state: &GenesisState{
				Params: DefaultParams(),
				MsgFees: []*MsgFee{
					{MsgTypeUrl: "thething", Cost: nil},
					{MsgTypeUrl: "thething", Cost: cz(coin(1, "banana"))},
				},
			},
			expErr: "duplicate MsgTypeUrl not allowed, \"thething\"",
		},
		{
			name: "multiple errors",
			state: &GenesisState{
				Params: Params{
					DefaultCost: coin(100, "banana"),
					ConversionFactor: ConversionFactor{
						DefinitionAmount: coin(10, "apple"),
						ConvertedAmount:  coin(10, "banana"),
					},
				},
				MsgFees: []*MsgFee{
					{MsgTypeUrl: "thingzero", Cost: cz(coin(99, "banana"))},
					{MsgTypeUrl: "thingone", Cost: nil},
					{MsgTypeUrl: "thingtwo", Cost: cz(coin(4, "x"))},
					{MsgTypeUrl: "thingzero", Cost: cz(coin(99, "banana"))},
					{MsgTypeUrl: "thingfour", Cost: cz(coin(21, "pear"))},
					{MsgTypeUrl: "thingzero", Cost: cz(coin(99, "banana"))},
					{MsgTypeUrl: "thingtwo", Cost: cz(coin(4, "x"))},
				},
			},
			expErr: joinErrs(
				"invalid flatfees params: default cost denom \"banana\" does not equal conversion factor base amount denom \"apple\"",
				"invalid MsgFees[2]: invalid thingtwo cost \"4x\": invalid denom: x",
				"duplicate MsgTypeUrl not allowed, \"thingzero\"",
				"duplicate MsgTypeUrl not allowed, \"thingtwo\"",
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.state.Validate()
			}
			require.NotPanics(t, testFunc, "GenesisState.Validate()")
			assertions.AssertErrorValue(t, err, tc.expErr, "GenesisState.Validate() error")
		})
	}
}

func TestDefaultGenesisState(t *testing.T) {
	pioconfig.SetProvConfig("pineapple")

	genState := DefaultGenesisState()
	require.NotNil(t, genState, "DefaultGenesisState()")
	assert.Equal(t, DefaultParams(), genState.Params, "genState.Params")
	assert.NotNil(t, genState.MsgFees, "genState.MsgFees")
	assert.Empty(t, genState.MsgFees, "genState.MsgFees")
}

func TestGetGenesisStateFromAppState(t *testing.T) {
	appCdc := app.MakeTestEncodingConfig(t).Marshaler
	// mocks.NewMockCodec(appCdc).WithUnmarshalJSONErrs("injected error message")
	asJSON := func(msg proto.Message) json.RawMessage {
		rv, err := appCdc.MarshalJSON(msg)
		require.NoError(t, err, "Marshal(%T)", msg)
		return rv
	}
	nonDefaultGenState := &GenesisState{
		Params: Params{
			DefaultCost: sdk.NewInt64Coin("apple", 500),
			ConversionFactor: ConversionFactor{
				DefinitionAmount: sdk.NewInt64Coin("apple", 25),
				ConvertedAmount:  sdk.NewInt64Coin("banana", 3),
			},
		},
		MsgFees: []*MsgFee{
			NewMsgFee("/msg.one", sdk.NewInt64Coin("apple", 400)),
			NewMsgFee("/msg.two", sdk.NewInt64Coin("apple", 600)),
			// NewMsgFee returns a nil cost when there aren't any coins, but unmarshalling
			// results in an empty slice. So this one needs to be made the hard way.
			{MsgTypeUrl: "/msg.free", Cost: []sdk.Coin{}},
			NewMsgFee("/msg.four", sdk.NewInt64Coin("apple", 1000), sdk.NewInt64Coin("plum", 3)),
		},
	}

	tests := []struct {
		name     string
		cdc      codec.Codec
		appState map[string]json.RawMessage
		expState *GenesisState
		expErr   string
	}{
		{
			name:     "empty app state",
			appState: make(map[string]json.RawMessage),
			expState: &GenesisState{},
		},
		{
			name: "empty flatfees state",
			appState: map[string]json.RawMessage{
				ModuleName: []byte(""),
			},
			expState: &GenesisState{},
		},
		{
			name: "unmarshal error",
			cdc:  mocks.NewMockCodec(appCdc).WithUnmarshalJSONErrs("injected 7 error message"),
			appState: map[string]json.RawMessage{
				ModuleName: asJSON(DefaultGenesisState()),
			},
			expState: &GenesisState{},
			expErr:   "could not unmarshal flatfees genesis state: injected 7 error message",
		},
		{
			name: "default",
			appState: map[string]json.RawMessage{
				ModuleName: asJSON(DefaultGenesisState()),
			},
			expState: DefaultGenesisState(),
		},
		{
			name: "populated",
			appState: map[string]json.RawMessage{
				ModuleName: asJSON(nonDefaultGenState),
			},
			expState: nonDefaultGenState,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cdc == nil {
				tc.cdc = appCdc
			}

			var actState *GenesisState
			var err error
			testFunc := func() {
				actState, err = GetGenesisStateFromAppState(tc.cdc, tc.appState)
			}
			require.NotPanics(t, testFunc, "GetGenesisStateFromAppState")
			assertions.AssertErrorValue(t, err, tc.expErr, "GetGenesisStateFromAppState error")
			assert.Equal(t, tc.expState, actState, "GetGenesisStateFromAppState result")
			if t.Failed() && tc.appState != nil {
				t.Logf("%s json:\n%s", ModuleName, string(tc.appState[ModuleName]))
			}
		})
	}
}
