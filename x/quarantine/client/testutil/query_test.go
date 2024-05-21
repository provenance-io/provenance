package testutil

import (
	"fmt"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/types/query"

	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/x/quarantine"
	client "github.com/provenance-io/provenance/x/quarantine/client/cli"
)

// These tests are initiated by TestIntegrationTestSuite in cli_test.go

func (s *IntegrationTestSuite) TestQueryQuarantinedFundsCmd() {
	addrs := s.createAndFundAccounts(2, 2000)
	addr0 := addrs[0]
	addr1 := addrs[1]

	// Opt addr0 into quarantine.
	testcli.NewTxExecutor(client.TxOptInCmd(), s.appendCommonFlagsTo(addr0)).
		Execute(s.T(), s.network)

	quarantinedAmount := int64(50)
	// Send some funds from 1 to 0 so that there's some quarantined funds to find.
	s.execBankSend(addr1, addr0, s.bondCoins(quarantinedAmount).String())

	newQF := func(to, from string, amt int64) *quarantine.QuarantinedFunds {
		return &quarantine.QuarantinedFunds{
			ToAddress:               to,
			UnacceptedFromAddresses: []string{from},
			Coins:                   s.bondCoins(amt),
			Declined:                false,
		}
	}

	tests := []struct {
		name   string
		args   []string
		exp    *quarantine.QueryQuarantinedFundsResponse
		expErr []string
	}{
		{
			name: "to and from no funds found",
			args: []string{addr1, addr0},
			exp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: []*quarantine.QuarantinedFunds{},
				Pagination:       nil,
			},
		},
		{
			name: "to and from funds found",
			args: []string{addr0, addr1},
			exp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: []*quarantine.QuarantinedFunds{
					newQF(addr0, addr1, quarantinedAmount),
				},
				Pagination: nil,
			},
		},
		{
			name: "only to no funds found",
			args: []string{addr1},
			exp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: []*quarantine.QuarantinedFunds{},
				Pagination:       &query.PageResponse{NextKey: nil, Total: 0},
			},
		},
		{
			name: "only to funds found",
			args: []string{addr0},
			exp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: []*quarantine.QuarantinedFunds{
					newQF(addr0, addr1, quarantinedAmount),
				},
				Pagination: &query.PageResponse{NextKey: nil, Total: 0},
			},
		},
		{
			name: "no args",
			args: nil,
			exp: &quarantine.QueryQuarantinedFundsResponse{
				QuarantinedFunds: []*quarantine.QuarantinedFunds{
					newQF(addr0, addr1, quarantinedAmount),
				},
				Pagination: &query.PageResponse{NextKey: nil, Total: 0},
			},
		},
		{
			name:   "bad to address",
			args:   []string{"what?"},
			expErr: []string{"invalid to_address", "invalid address", "decoding bech32 failed"},
		},
		{
			name:   "bad from address",
			args:   []string{addr0, "nope"},
			expErr: []string{"invalid from_address", "invalid address", "decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.QueryQuarantinedFundsCmd()
			args := append(tc.args, fmt.Sprintf("--%s=json", cmtcli.OutputFlag))
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			outBz := outBW.Bytes()
			s.T().Logf("Output:\n%s", string(outBz))
			s.assertErrorContents(err, tc.expErr, "QueryQuarantinedFundsCmd error")
			for _, expErr := range tc.expErr {
				s.Assert().Contains(string(outBz), expErr, "QueryQuarantinedFundsCmd output with error")
			}
			if tc.exp != nil {
				act := &quarantine.QueryQuarantinedFundsResponse{}
				testFunc := func() {
					err = s.clientCtx.Codec.UnmarshalJSON(outBz, act)
				}
				if s.Assert().NotPanics(testFunc, "UnmarshalJSON on output") {
					if s.Assert().NoError(err, "UnmarshalJSON on output") {
						s.Assert().ElementsMatch(tc.exp.QuarantinedFunds, act.QuarantinedFunds, "QuarantinedFunds")
						s.Assert().Equal(tc.exp.Pagination, act.Pagination, "Pagination")
					}
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryIsQuarantinedCmd() {
	addrs := s.createAndFundAccounts(2, 2000)
	addr0 := addrs[0]
	addr1 := addrs[1]

	// Opt addr0 into quarantine.
	testcli.NewTxExecutor(client.TxOptInCmd(), s.appendCommonFlagsTo(addr0)).
		Execute(s.T(), s.network)

	tests := []struct {
		name   string
		args   []string
		exp    *quarantine.QueryIsQuarantinedResponse
		expErr []string
	}{
		{
			name: "quarantined addr",
			args: []string{addr0},
			exp:  &quarantine.QueryIsQuarantinedResponse{IsQuarantined: true},
		},
		{
			name: "not quarantined addr",
			args: []string{addr1},
			exp:  &quarantine.QueryIsQuarantinedResponse{IsQuarantined: false},
		},
		{
			name:   "bad addr",
			args:   []string{"invalid1addritisbadbad"},
			expErr: []string{"invalid to_address", "invalid address", "decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.QueryIsQuarantinedCmd()
			args := append(tc.args, fmt.Sprintf("--%s=json", cmtcli.OutputFlag))
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			out := outBW.String()
			s.T().Logf("Output:\n%s", out)
			s.assertErrorContents(err, tc.expErr, "QueryIsQuarantinedCmd error")
			for _, expErr := range tc.expErr {
				s.Assert().Contains(out, expErr, "QueryIsQuarantinedCmd output with error")
			}
			if tc.exp != nil {
				act := &quarantine.QueryIsQuarantinedResponse{}
				testFunc := func() {
					err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), act)
				}
				if s.Assert().NotPanics(testFunc, "UnmarshalJSON on output") {
					if s.Assert().NoError(err, "UnmarshalJSON on output") {
						s.Assert().Equal(tc.exp, act)
					}
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryAutoResponsesCmd() {
	addrs := s.createAndFundAccounts(3, 200)
	addr0 := addrs[0]
	addr1 := addrs[1]
	addr2 := addrs[2]

	// Set 0 <- 1 to auto-accept.
	// Set 0 <- 2 to auto-decline.
	testcli.NewTxExecutor(client.TxUpdateAutoResponsesCmd(), s.appendCommonFlagsTo(addr0, "accept", addr1, "decline", addr2)).
		Execute(s.T(), s.network)

	newARE := func(to, from string, response quarantine.AutoResponse) *quarantine.AutoResponseEntry {
		return &quarantine.AutoResponseEntry{
			ToAddress:   to,
			FromAddress: from,
			Response:    response,
		}
	}

	tests := []struct {
		name   string
		args   []string
		exp    *quarantine.QueryAutoResponsesResponse
		expErr []string
	}{
		{
			name:   "bad to addr",
			args:   []string{"badnotgood"},
			expErr: []string{"invalid to_address", "invalid address", "decoding bech32 failed"},
		},
		{
			name:   "bad from addr",
			args:   []string{addr0, "notgoodbad"},
			expErr: []string{"invalid from_address", "invalid address", "decoding bech32 failed"},
		},
		{
			name: "to and from accept",
			args: []string{addr0, addr1},
			exp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0, addr1, quarantine.AUTO_RESPONSE_ACCEPT),
				},
				Pagination: nil,
			},
		},
		{
			name: "to and from decline",
			args: []string{addr0, addr2},
			exp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0, addr2, quarantine.AUTO_RESPONSE_DECLINE),
				},
				Pagination: nil,
			},
		},
		{
			name: "to and from unspecified",
			args: []string{addr2, addr1},
			exp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr2, addr1, quarantine.AUTO_RESPONSE_UNSPECIFIED),
				},
				Pagination: nil,
			},
		},
		{
			name: "only to with results",
			args: []string{addr0},
			exp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{
					newARE(addr0, addr1, quarantine.AUTO_RESPONSE_ACCEPT),
					newARE(addr0, addr2, quarantine.AUTO_RESPONSE_DECLINE),
				},
				Pagination: &query.PageResponse{NextKey: nil, Total: 0},
			},
		},
		{
			name: "only to no results",
			args: []string{addr2},
			exp: &quarantine.QueryAutoResponsesResponse{
				AutoResponses: []*quarantine.AutoResponseEntry{},
				Pagination:    &query.PageResponse{NextKey: nil, Total: 0},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.QueryAutoResponsesCmd()
			args := append(tc.args, fmt.Sprintf("--%s=json", cmtcli.OutputFlag))
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			out := outBW.String()
			s.T().Logf("Output:\n%s", out)
			s.assertErrorContents(err, tc.expErr, "QueryAutoResponsesCmd error")
			for _, expErr := range tc.expErr {
				s.Assert().Contains(out, expErr, "QueryAutoResponsesCmd output with error")
			}
			if tc.exp != nil {
				act := &quarantine.QueryAutoResponsesResponse{}
				testFunc := func() {
					err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), act)
				}
				if s.Assert().NotPanics(testFunc, "UnmarshalJSON on output") {
					if s.Assert().NoError(err, "UnmarshalJSON on output") {
						s.Assert().ElementsMatch(tc.exp.AutoResponses, act.AutoResponses, "AutoResponses")
						s.Assert().Equal(tc.exp.Pagination, act.Pagination, "Pagination")
					}
				}
			}
		})
	}
}
