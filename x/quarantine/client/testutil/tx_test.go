package testutil

import (
	"fmt"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	"github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"

	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/testutil/queries"
	"github.com/provenance-io/provenance/x/quarantine"
	client "github.com/provenance-io/provenance/x/quarantine/client/cli"
)

func (s *IntegrationTestSuite) TestTxOptInCmd() {
	addr0 := s.createAndFundAccount(2000)

	tests := []struct {
		name    string
		args    []string
		expErr  []string
		expCode uint32
	}{
		{
			name:   "empty addr",
			args:   []string{""},
			expErr: []string{"no to_name_or_address provided"},
		},
		{
			name:   "bad addr",
			args:   []string{"somethingelse"},
			expErr: []string{"somethingelse.info: key not found"},
		},
		{
			name:    "good addr",
			args:    []string{addr0},
			expCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(client.TxOptInCmd(), s.appendCommonFlagsTo(tc.args...)).
				WithExpInErrMsg(tc.expErr).
				WithExpCode(tc.expCode).
				Execute(s.T(), s.network)
		})
	}
}

func (s *IntegrationTestSuite) TestTxOptOutCmd() {
	addr0 := s.createAndFundAccount(2000)

	tests := []struct {
		name    string
		args    []string
		expErr  []string
		expCode uint32
	}{
		{
			name:   "empty addr",
			args:   []string{""},
			expErr: []string{"no to_name_or_address provided"},
		},
		{
			name:   "bad addr",
			args:   []string{"somethingelse"},
			expErr: []string{"somethingelse.info: key not found"},
		},
		{
			name:    "good addr",
			args:    []string{addr0},
			expCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(client.TxOptOutCmd(), s.appendCommonFlagsTo(tc.args...)).
				WithExpInErrMsg(tc.expErr).
				WithExpCode(tc.expCode).
				Execute(s.T(), s.network)
		})
	}
}

func (s *IntegrationTestSuite) TestTxAcceptCmd() {
	addrs := s.createAndFundAccounts(4, 2000)
	addr0 := addrs[0]
	addr1 := addrs[1]
	addr2 := addrs[2]
	addr3 := addrs[3]

	permFlag := "--" + client.FlagPermanent
	tests := []struct {
		name    string
		args    []string
		expErr  []string
		expCode uint32
	}{
		{
			name:   "empty to address",
			args:   []string{"", addr1},
			expErr: []string{"no to_name_or_address provided"},
		},
		{
			name:   "bad to address",
			args:   []string{"notgood", addr1},
			expErr: []string{"notgood.info: key not found"},
		},
		{
			name:   "empty from address 1",
			args:   []string{addr0, ""},
			expErr: []string{"invalid from_address 1", "invalid address", "empty address string is not allowed"},
		},
		{
			name:   "bad from address 1",
			args:   []string{addr0, "stillbad"},
			expErr: []string{"invalid from_address 1", "invalid address", "decoding bech32 failed"},
		},
		{
			name:   "empty from address 3",
			args:   []string{addr0, addr1, addr2, ""},
			expErr: []string{"invalid from_address 3", "invalid address", "empty address string is not allowed"},
		},
		{
			name:   "bad from address 3",
			args:   []string{addr0, addr1, addr2, "stillbad"},
			expErr: []string{"invalid from_address 3", "invalid address", "decoding bech32 failed"},
		},
		{
			name:    "one from address",
			args:    []string{addr0, addr1},
			expCode: 0,
		},
		{
			name:    "two from addresses",
			args:    []string{addr0, addr1, addr2},
			expCode: 0,
		},
		{
			name:    "three from addresses",
			args:    []string{addr0, addr1, addr2, addr3},
			expCode: 0,
		},
		{
			name:    "one from address and perm",
			args:    []string{addr0, addr1, permFlag},
			expCode: 0,
		},
		{
			name:    "two from addresses and perm",
			args:    []string{addr0, addr1, addr2, permFlag},
			expCode: 0,
		},
		{
			name:    "three from addresses and perm",
			args:    []string{addr0, addr1, addr2, addr3, permFlag},
			expCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(client.TxAcceptCmd(), s.appendCommonFlagsTo(tc.args...)).
				WithExpInErrMsg(tc.expErr).
				WithExpCode(tc.expCode).
				Execute(s.T(), s.network)
		})
	}
}

func (s *IntegrationTestSuite) TestTxDeclineCmd() {
	addrs := s.createAndFundAccounts(4, 2000)
	addr0 := addrs[0]
	addr1 := addrs[1]
	addr2 := addrs[2]
	addr3 := addrs[3]

	permFlag := "--" + client.FlagPermanent
	tests := []struct {
		name    string
		args    []string
		expErr  []string
		expCode uint32
	}{
		{
			name:   "empty to address",
			args:   []string{"", addr1},
			expErr: []string{"no to_name_or_address provided"},
		},
		{
			name:   "bad to address",
			args:   []string{"notgood", addr1},
			expErr: []string{"notgood.info: key not found"},
		},
		{
			name:   "empty from address 1",
			args:   []string{addr0, ""},
			expErr: []string{"invalid from_address 1", "invalid address", "empty address string is not allowed"},
		},
		{
			name:   "bad from address 1",
			args:   []string{addr0, "stillbad"},
			expErr: []string{"invalid from_address 1", "invalid address", "decoding bech32 failed"},
		},
		{
			name:   "empty from address 3",
			args:   []string{addr0, addr1, addr2, ""},
			expErr: []string{"invalid from_address 3", "invalid address", "empty address string is not allowed"},
		},
		{
			name:   "bad from address 3",
			args:   []string{addr0, addr1, addr2, "stillbad"},
			expErr: []string{"invalid from_address 3", "invalid address", "decoding bech32 failed"},
		},
		{
			name:    "one from address",
			args:    []string{addr0, addr1},
			expCode: 0,
		},
		{
			name:    "two from addresses",
			args:    []string{addr0, addr1, addr2},
			expCode: 0,
		},
		{
			name:    "three from addresses",
			args:    []string{addr0, addr1, addr2, addr3},
			expCode: 0,
		},
		{
			name:    "one from address and perm",
			args:    []string{addr0, addr1, permFlag},
			expCode: 0,
		},
		{
			name:    "two from addresses and perm",
			args:    []string{addr0, addr1, addr2, permFlag},
			expCode: 0,
		},
		{
			name:    "three from addresses and perm",
			args:    []string{addr0, addr1, addr2, addr3, permFlag},
			expCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(client.TxDeclineCmd(), s.appendCommonFlagsTo(tc.args...)).
				WithExpInErrMsg(tc.expErr).
				WithExpCode(tc.expCode).
				Execute(s.T(), s.network)
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateAutoResponsesCmd() {
	addrs := s.createAndFundAccounts(4, 2000)
	addr0 := addrs[0]
	addr1 := addrs[1]
	addr2 := addrs[2]
	addr3 := addrs[3]

	tests := []struct {
		name    string
		args    []string
		expErr  []string
		expCode uint32
	}{
		{
			name:   "empty to address",
			args:   []string{"", "accept", addr1},
			expErr: []string{"no to_name_or_address provided"},
		},
		{
			name:   "bad to address",
			args:   []string{"naughty", "accept", addr1},
			expErr: []string{"naughty.info: key not found"},
		},
		{
			name: "bad from addr",
			args: []string{addr0, "accept", "notokay"},
			expErr: []string{
				`unknown arg 3 "notokay"`, `auto-response 1 "accept"`,
				"from_address 1", "invalid address", "decoding bech32 failed",
			},
		},
		{
			name:   "bad auto-response",
			args:   []string{addr0, "not-a-resp", addr1},
			expErr: []string{"invalid arg 2", "invalid auto-response", `"not-a-resp"`},
		},
		{
			name:    "simply good",
			args:    []string{addr0, "decline", addr1},
			expCode: 0,
		},
		{
			name:    "complexly good",
			args:    []string{addr0, "decline", addr2, addr3, "o", addr1},
			expCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(client.TxUpdateAutoResponsesCmd(), s.appendCommonFlagsTo(tc.args...)).
				WithExpInErrMsg(tc.expErr).
				WithExpCode(tc.expCode).
				Execute(s.T(), s.network)
		})
	}
}

func (s *IntegrationTestSuite) TestSendAndAcceptQuarantinedFunds() {
	addrs := s.createAndFundAccounts(3, 2000)
	toAddr := addrs[0]
	fromAddr1 := addrs[1]
	fromAddr2 := addrs[2]

	amt1 := int64(50)
	amt2 := int64(75)
	expToAddrAmt := 2000 + amt1 + amt2 - 20
	expFromAddr1Amt := 2000 - amt1 - 10
	expFromAddr2Amt := 2000 - amt2 - 10

	asJSONFlag := fmt.Sprintf("--%s=json", cmtcli.OutputFlag)

	s.Run("opt toAddr into quarantine", func() {
		testcli.NewTxExecutor(client.TxOptInCmd(), s.appendCommonFlagsTo(toAddr)).Execute(s.T(), s.network)

		outBW, err := cli.ExecTestCLICmd(s.clientCtx, client.QueryIsQuarantinedCmd(), []string{toAddr, asJSONFlag})
		out := outBW.String()
		s.T().Logf("QueryIsQuarantinedCmd Output:\n%s", out)
		s.Require().NoError(err, "QueryIsQuarantinedCmd error")
		resp := &quarantine.QueryIsQuarantinedResponse{}
		s.Require().NotPanics(func() {
			err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), resp)
		})
		s.Require().NoError(err, "UnmarshalJSON QueryIsQuarantinedResponse")
		s.Require().True(resp.IsQuarantined, "IsQuarantined")
	})

	s.stopIfFailed()

	s.Run("do two sends from different addresses", func() {
		s.execBankSend(fromAddr1, toAddr, s.bondCoins(amt1).String())
		s.execBankSend(fromAddr2, toAddr, s.bondCoins(amt2).String())

		expFunds := []*quarantine.QuarantinedFunds{
			{
				ToAddress:               toAddr,
				UnacceptedFromAddresses: []string{fromAddr1},
				Coins:                   s.bondCoins(amt1),
				Declined:                false,
			},
			{
				ToAddress:               toAddr,
				UnacceptedFromAddresses: []string{fromAddr2},
				Coins:                   s.bondCoins(amt2),
				Declined:                false,
			},
		}

		outBW, err := cli.ExecTestCLICmd(s.clientCtx, client.QueryQuarantinedFundsCmd(), []string{toAddr, asJSONFlag})
		out := outBW.String()
		s.T().Logf("QueryQuarantinedFundsCmd Output:\n%s", out)
		s.Require().NoError(err, "QueryQuarantinedFundsCmd error")

		resp := &quarantine.QueryQuarantinedFundsResponse{}
		s.Require().NotPanics(func() {
			err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), resp)
		})
		s.Require().NoError(err, "UnmarshalJSON QueryQuarantinedFundsResponse")
		s.Require().ElementsMatch(expFunds, resp.QuarantinedFunds, "QuarantinedFunds A: expected, B: actual")
	})

	s.stopIfFailed()

	s.Run("accept the quarantined funds", func() {
		testcli.NewTxExecutor(client.TxAcceptCmd(), s.appendCommonFlagsTo(toAddr, fromAddr2, fromAddr1)).
			Execute(s.T(), s.network)

		outBW, err := cli.ExecTestCLICmd(s.clientCtx, client.QueryQuarantinedFundsCmd(), []string{toAddr, asJSONFlag})
		out := outBW.String()
		s.T().Logf("QueryQuarantinedFundsCmd Output:\n%s", out)
		s.Require().NoError(err, "QueryQuarantinedFundsCmd error")
		resp := &quarantine.QueryQuarantinedFundsResponse{}
		s.Require().NotPanics(func() {
			err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), resp)
		})
		s.Require().NoError(err, "UnmarshalJSON QueryQuarantinedFundsResponse")
		s.Require().Empty(resp.QuarantinedFunds, "QuarantinedFunds")
	})

	s.stopIfFailed()

	tests := []struct {
		name string
		addr string
		exp  sdk.Coins
	}{
		{
			name: "final toAddr balance",
			addr: toAddr,
			exp:  s.bondCoins(expToAddrAmt),
		},
		{
			name: "final fromAddr1 balance",
			addr: fromAddr1,
			exp:  s.bondCoins(expFromAddr1Amt),
		},
		{
			name: "final fromAddr2 balance",
			addr: fromAddr2,
			exp:  s.bondCoins(expFromAddr2Amt),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			balances := queries.GetAllBalances(s.T(), s.network, tc.addr)
			s.Require().Equal(tc.exp, balances, "Balances")
		})
	}
}
