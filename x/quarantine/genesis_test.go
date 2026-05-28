package quarantine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/quarantine"
	"github.com/provenance-io/provenance/x/quarantine/testutil"
)

func TestGenesisState_Validate(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("gsv", 0).String()
	testAddr1 := testutil.MakeTestAddr("gsv", 1).String()
	badAddr := "this1addressisnaughty"

	goodAutoResponse := &quarantine.AutoResponseEntry{
		ToAddress:   testAddr0,
		FromAddress: testAddr1,
		Response:    quarantine.AUTO_RESPONSE_ACCEPT,
	}
	badAutoResponse := &quarantine.AutoResponseEntry{
		ToAddress:   testAddr0,
		FromAddress: testAddr1,
		Response:    -10,
	}

	goodQuarantinedFunds := &quarantine.QuarantinedFunds{
		ToAddress:               testAddr0,
		UnacceptedFromAddresses: []string{testAddr1},
		Coins:                   coinMakerOK(),
		Declined:                false,
	}
	badQuarantinedFunds := &quarantine.QuarantinedFunds{
		ToAddress:               testAddr0,
		UnacceptedFromAddresses: []string{testAddr1},
		Coins:                   coinMakerBad(),
		Declined:                false,
	}

	tests := []struct {
		name    string
		gs      *quarantine.GenesisState
		expErrs []string
	}{
		{
			name: "control",
			gs: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: nil,
		},
		{
			name:    "empty",
			gs:      &quarantine.GenesisState{},
			expErrs: nil,
		},
		{
			name: "bad first addr",
			gs: &quarantine.GenesisState{
				QuarantinedAddresses: []string{badAddr, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined address[0]"},
		},
		{
			name: "bad second addr",
			gs: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, badAddr},
				AutoResponses:        []*quarantine.AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined address[1]"},
		},
		{
			name: "bad first auto response",
			gs: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{badAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantine auto response entry[0]"},
		},
		{
			name: "bad second auto response",
			gs: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{goodAutoResponse, badAutoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{goodQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantine auto response entry[1]"},
		},
		{
			name: "bad first quarantined funds",
			gs: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{badQuarantinedFunds, goodQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined funds[0]"},
		},
		{
			name: "bad second quarantined funds",
			gs: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{goodAutoResponse, goodAutoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{goodQuarantinedFunds, badQuarantinedFunds},
			},
			expErrs: []string{"invalid quarantined funds[1]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			orig := testutil.MakeCopyOfGenesisState(tc.gs)
			var err error
			testFunc := func() {
				err = tc.gs.Validate()
			}
			assert.NotPanics(t, testFunc, "GenesisState.Validate()")
			assertions.AssertErrorContents(t, err, tc.expErrs, "Validate")
			assert.Equal(t, orig, tc.gs, "GenesisState before and after Validate")
		})
	}
}

func TestNewGenesisState(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("ngs", 0).String()
	testAddr1 := testutil.MakeTestAddr("ngs", 1).String()

	autoResponse := &quarantine.AutoResponseEntry{
		ToAddress:   testAddr0,
		FromAddress: testAddr1,
		Response:    quarantine.AUTO_RESPONSE_ACCEPT,
	}

	quarantinedFunds := &quarantine.QuarantinedFunds{
		ToAddress:               testAddr0,
		UnacceptedFromAddresses: []string{testAddr1},
		Coins:                   coinMakerOK(),
		Declined:                false,
	}

	tests := []struct {
		name  string
		addrs []string
		ars   []*quarantine.AutoResponseEntry
		qfs   []*quarantine.QuarantinedFunds
		exp   *quarantine.GenesisState
	}{
		{
			name:  "control",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "nil addrs",
			addrs: nil,
			ars:   []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: nil,
				AutoResponses:        []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "empty addrs",
			addrs: []string{},
			ars:   []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: []string{},
				AutoResponses:        []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "nil auto responses",
			addrs: []string{testAddr0, testAddr1},
			ars:   nil,
			qfs:   []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        nil,
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "empty auto responses",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*quarantine.AutoResponseEntry{},
			qfs:   []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{quarantinedFunds, quarantinedFunds},
			},
		},
		{
			name:  "nil quarantined funds",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
			qfs:   nil,
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     nil,
			},
		},
		{
			name:  "empty quarantined funds",
			addrs: []string{testAddr0, testAddr1},
			ars:   []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
			qfs:   []*quarantine.QuarantinedFunds{},
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: []string{testAddr0, testAddr1},
				AutoResponses:        []*quarantine.AutoResponseEntry{autoResponse, autoResponse},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{},
			},
		},
		{
			name:  "all empty",
			addrs: []string{},
			ars:   []*quarantine.AutoResponseEntry{},
			qfs:   []*quarantine.QuarantinedFunds{},
			exp: &quarantine.GenesisState{
				QuarantinedAddresses: []string{},
				AutoResponses:        []*quarantine.AutoResponseEntry{},
				QuarantinedFunds:     []*quarantine.QuarantinedFunds{},
			},
		},
		{
			name:  "DefaultGenesisState",
			addrs: nil,
			ars:   nil,
			qfs:   nil,
			exp:   quarantine.DefaultGenesisState(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := quarantine.NewGenesisState(tc.addrs, tc.ars, tc.qfs)
			assert.Equal(t, tc.exp, actual, "NewGenesisState")
		})
	}
}
