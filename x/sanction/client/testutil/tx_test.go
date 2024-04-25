package testutil

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/testutil/queries"
	"github.com/provenance-io/provenance/x/sanction"
	client "github.com/provenance-io/provenance/x/sanction/client/cli"
)

// assertGovPropMsg gets a gov prop and makes sure it has one specific message.
func (s *IntegrationTestSuite) assertGovPropMsg(propID string, msg sdk.Msg) bool {
	s.T().Helper()
	if msg == nil {
		return true
	}

	if !s.Assert().NotEmpty(propID, "proposal id") {
		return false
	}
	expPropMsgAny, err := codectypes.NewAnyWithValue(msg)
	if !s.Assert().NoError(err, "NewAnyWithValue on %T", msg) {
		return false
	}

	prop := queries.GetGovProp(s.T(), s.network, propID)
	if !s.Assert().Len(prop.Messages, 1, "number of messages in proposal") {
		return false
	}
	if !s.Assert().Equal(expPropMsgAny, prop.Messages[0], "the message in the proposal") {
		return false
	}

	return true
}

// findProposalID looks through the provided response to find a governance proposal id.
// If one is found, it's returned (as a string). Otherwise, an empty string is returned.
func (s *IntegrationTestSuite) findProposalID(resp *sdk.TxResponse) string {
	if resp == nil {
		return ""
	}
	for _, event := range resp.Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "proposal_id" {
					return string(attr.Value)
				}
			}
		}
	}
	return ""
}

func (s *IntegrationTestSuite) TestTxSanctionCmd() {
	authority := s.getAuthority()
	addr1 := sdk.AccAddress("1_address_test_test_").String()
	addr2 := sdk.AccAddress("2_address_test_test_").String()

	tests := []struct {
		name       string
		args       []string
		expErr     []string
		expPropMsg *sanction.MsgSanction
	}{
		{
			name:   "no addresses given",
			args:   []string{},
			expErr: []string{"requires at least 1 arg(s), only received 0"},
		},
		{
			name: "one address good",
			args: []string{addr1},
			expPropMsg: &sanction.MsgSanction{
				Addresses: []string{addr1},
				Authority: authority,
			},
		},
		{
			name:   "one address bad",
			args:   []string{"thisis1addrthatisbad"},
			expErr: []string{"addresses[0]", `"thisis1addrthatisbad"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses first bad",
			args:   []string{"another1badaddr", addr2},
			expErr: []string{"addresses[0]", `"another1badaddr"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses second bad",
			args:   []string{addr1, "athird1badaddress"},
			expErr: []string{"addresses[1]", `"athird1badaddress"`, "decoding bech32 failed"},
		},
		{
			name: "two addresses good",
			args: []string{addr1, addr2},
			expPropMsg: &sanction.MsgSanction{
				Addresses: []string{addr1, addr2},
				Authority: authority,
			},
		},
		{
			name:   "bad authority",
			args:   []string{addr1, "--" + provcli.FlagAuthority, "bad1auth34sd2"},
			expErr: []string{"authority", `"bad1auth34sd2"`, "decoding bech32 failed"},
		},
		{
			name:   "bad deposit",
			args:   []string{addr1, "--" + govcli.FlagDeposit, "notcoins"},
			expErr: []string{"invalid deposit", "notcoins"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			args := s.appendCommonArgsTo(tc.args...)
			args = append(args, "--title", "TxSanctionCmd", "--summary", tc.name)
			txResp := testcli.NewCLITxExecutor(client.TxSanctionCmd(), args).
				WithExpInErrMsg(tc.expErr).
				Execute(s.T(), s.network)

			if tc.expPropMsg != nil {
				propID := s.findProposalID(txResp)
				s.assertGovPropMsg(propID, tc.expPropMsg)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUnsanctionCmd() {
	authority := s.getAuthority()
	addr1 := sdk.AccAddress("1_address_untest____").String()
	addr2 := sdk.AccAddress("2_address_untest____").String()

	tests := []struct {
		name       string
		args       []string
		expErr     []string
		expPropMsg *sanction.MsgUnsanction
	}{
		{
			name:   "no addresses given",
			args:   []string{},
			expErr: []string{"requires at least 1 arg(s), only received 0"},
		},
		{
			name: "one address good",
			args: []string{addr1},
			expPropMsg: &sanction.MsgUnsanction{
				Addresses: []string{addr1},
				Authority: authority,
			},
		},
		{
			name:   "one address bad",
			args:   []string{"thisis1addrthatisbad"},
			expErr: []string{"addresses[0]", `"thisis1addrthatisbad"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses first bad",
			args:   []string{"another1badaddr", addr2},
			expErr: []string{"addresses[0]", `"another1badaddr"`, "decoding bech32 failed"},
		},
		{
			name:   "two addresses second bad",
			args:   []string{addr1, "athird1badaddress"},
			expErr: []string{"addresses[1]", `"athird1badaddress"`, "decoding bech32 failed"},
		},
		{
			name: "two addresses good",
			args: []string{addr1, addr2},
			expPropMsg: &sanction.MsgUnsanction{
				Addresses: []string{addr1, addr2},
				Authority: authority,
			},
		},
		{
			name:   "bad authority",
			args:   []string{addr1, "--" + provcli.FlagAuthority, "bad1auth34sd2"},
			expErr: []string{"authority", `"bad1auth34sd2"`, "decoding bech32 failed"},
		},
		{
			name:   "bad deposit",
			args:   []string{addr1, "--" + govcli.FlagDeposit, "notcoins"},
			expErr: []string{"invalid deposit", "notcoins"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			args := s.appendCommonArgsTo(tc.args...)
			args = append(args, "--title", "TxUnsanctionCmd", "--summary", tc.name)
			txResp := testcli.NewCLITxExecutor(client.TxUnsanctionCmd(), args).
				WithExpInErrMsg(tc.expErr).
				Execute(s.T(), s.network)

			if tc.expPropMsg != nil {
				propID := s.findProposalID(txResp)
				s.assertGovPropMsg(propID, tc.expPropMsg)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateParamsCmd() {
	authority := s.getAuthority()

	tests := []struct {
		name       string
		args       []string
		expErr     []string
		expPropMsg *sanction.MsgUpdateParams
	}{
		{
			name:   "no args",
			args:   []string{},
			expErr: []string{"accepts 2 arg(s), received 0"},
		},
		{
			name:   "one arg",
			args:   []string{"arg1"},
			expErr: []string{"accepts 2 arg(s), received 1"},
		},
		{
			name:   "three args",
			args:   []string{"arg1", "arg2", "arg3"},
			expErr: []string{"accepts 2 arg(s), received 3"},
		},
		{
			name: "coins coins",
			args: []string{"1acoin", "2bcoin"},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("acoin", 1)),
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("bcoin", 2)),
				},
				Authority: authority,
			},
		},
		{
			name: "empty coins",
			args: []string{"", "3ccoin"},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("ccoin", 3)),
				},
				Authority: authority,
			},
		},
		{
			name: "coins empty",
			args: []string{"4dcoin", ""},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("dcoin", 4)),
					ImmediateUnsanctionMinDeposit: nil,
				},
				Authority: authority,
			},
		},
		{
			name: "empty empty",
			args: []string{"", ""},
			expPropMsg: &sanction.MsgUpdateParams{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   nil,
					ImmediateUnsanctionMinDeposit: nil,
				},
				Authority: authority,
			},
		},
		{
			name:   "bad good",
			args:   []string{"firscoinsbad", "5ecoin"},
			expErr: []string{"invalid immediate_sanction_min_deposit", `"firscoinsbad"`},
		},
		{
			name:   "good bad",
			args:   []string{"6fcoin", "secondcoinsbad"},
			expErr: []string{"invalid immediate_unsanction_min_deposit", `"secondcoinsbad"`},
		},
		{
			name:   "bad authority",
			args:   []string{"", "", "--" + provcli.FlagAuthority, "bad1auth34sd2"},
			expErr: []string{"authority", `"bad1auth34sd2"`, "decoding bech32 failed"},
		},
		{
			name:   "bad deposit",
			args:   []string{"", "", "--" + govcli.FlagDeposit, "notcoins"},
			expErr: []string{"invalid deposit", "notcoins"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			args := s.appendCommonArgsTo(tc.args...)
			args = append(args, "--title", "TxUpdateParamsCmd", "--summary", tc.name)
			txResp := testcli.NewCLITxExecutor(client.TxUpdateParamsCmd(), args).
				WithExpInErrMsg(tc.expErr).
				Execute(s.T(), s.network)

			if tc.expPropMsg != nil {
				propID := s.findProposalID(txResp)
				s.assertGovPropMsg(propID, tc.expPropMsg)
			}
		})
	}
}
