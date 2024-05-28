package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtcli "github.com/cometbft/cometbft/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/testutil"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/testutil/queries"
	"github.com/provenance-io/provenance/x/sanction"
	client "github.com/provenance-io/provenance/x/sanction/client/cli"
)

const blocksPerVotingPeriod = 8

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) TestSanctionValidatorImmediateUsingGovCmds() {
	// Wait 2 blocks to start this. That way, hopefully the query tests are done.
	// In between the two, create all the stuff to send.
	s.waitForNextBlock("wait for next block 1")
	authority := s.getAuthority()
	proposerValI := 0
	sanctionValI := 4
	sanctMsg := &sanction.MsgSanction{
		Addresses: []string{s.network.Validators[sanctionValI].Address.String()},
		Authority: authority,
	}
	sanctMsgAny, err := codectypes.NewAnyWithValue(sanctMsg)
	depAmt := s.sanctionGenesis.Params.ImmediateSanctionMinDeposit
	feeAmt := s.bondCoins(10)
	s.Require().NoError(err, "NewAnyWithValue(MsgSanction)")
	// Thankfully, the struct used to unmarshal the proposal json (in NewCmdSubmitProposal), is private.
	// And to be really helpful, it's not the same as MsgSubmitProposal.
	// Specifically, the command wants keys "messages", "metadata", and "deposit",
	// but MsgSubmitProposal ends up with "messages", "metadata", "initial_deposit", and "proposer".
	// So I'm going to marshal MsgSubmitProposal to json, then unmarshal it to a map[string]json.RawMessage,
	// tweak the map, then marshal that into what ends up in the file.
	propMsg := &govv1.MsgSubmitProposal{
		Messages:       []*codectypes.Any{sanctMsgAny},
		InitialDeposit: depAmt,
		Proposer:       s.network.Validators[proposerValI].Address.String(),
		Metadata:       "",
		Title:          "Sanction an address",
		Summary:        "Sanction an address",
	}
	propMsgBzStep1, err := s.cfg.Codec.MarshalJSON(propMsg)
	s.Require().NoError(err, "MarshalJSON(MsgSubmitProposal)")
	var propMsgJSON map[string]json.RawMessage
	err = json.Unmarshal(propMsgBzStep1, &propMsgJSON)
	s.Require().NoError(err, "Unmarshal MsgSubmitProposal")
	propMsgJSON["deposit"] = []byte(fmt.Sprintf("%q", depAmt.String()))
	delete(propMsgJSON, "initial_deposit")
	delete(propMsgJSON, "proposer")
	propMsgBz, err := json.Marshal(propMsgJSON)
	s.Require().NoError(err, "Marshal propMsgJSON")
	propFile := filepath.Join(s.T().TempDir(), "gov-prop-sanction.json")
	err = os.WriteFile(propFile, propMsgBz, 0644)
	s.Require().NoError(err, "WriteFile %s", propFile)

	// Usage: simd tx gov submit-proposal [path/to/proposal.json] [flags]
	propCmd := govcli.NewCmdSubmitProposal()
	propArgs := []string{
		propFile,
		"--" + flags.FlagKeyringBackend, keyring.BackendTest,
		"--" + flags.FlagFrom, propMsg.Proposer,
		"--" + flags.FlagFees, feeAmt.String(),
		"--" + flags.FlagSkipConfirmation,
		"--" + flags.FlagBroadcastMode, flags.BroadcastSync,
		"--" + cmtcli.OutputFlag, "json",
	}

	// Usage: simd tx gov vote [proposal-id] [option] [flags]
	voteCmd := govcli.NewCmdVote()
	allVoteArgs := make([][]string, len(s.network.Validators))
	for i, val := range s.network.Validators {
		// Note: The no_with_veto vote from the validator being sanctioned should fail because
		// enough deposit is provided to make the sanction immediate, so they won't be able to pay fees.
		// The command won't return an error though. That failure will happen when the block is being processed.
		// Failure of that tx will be reflected in the final tally of the proposal later on though,
		// i.e. it won't have any recorded no-with-veto votes.
		option := "yes"
		if i == sanctionValI {
			option = "no_with_veto"
		}
		// Note: arg[0] will be updated with the gov prop once it's known.
		allVoteArgs[i] = []string{
			"0", option,
			"--" + flags.FlagKeyringDir, filepath.Join(val.Dir, "simcli"),
			"--" + flags.FlagKeyringBackend, keyring.BackendTest,
			"--" + flags.FlagFrom, val.Address.String(),
			"--" + flags.FlagFees, feeAmt.String(),
			"--" + flags.FlagBroadcastMode, flags.BroadcastAsync,
			"--" + flags.FlagSkipConfirmation,
			"--" + cmtcli.OutputFlag, "json",
		}
	}

	// Usage: simd query sanction is-sanctioned <address> [flags]
	isSanctCmd := client.QueryIsSanctionedCmd()
	isSanctArgs := []string{
		s.network.Validators[sanctionValI].Address.String(),
		"--" + cmtcli.OutputFlag, "json",
	}

	// Finally, wait for the next block.
	s.waitForNextBlock("wait for next block 2")

	// Submit the proposal.
	s.T().Logf("Proposal: %s\n%s", propFile, propMsgBz)
	testcli.NewTxExecutor(propCmd, propArgs).Execute(s.T(), s.network)
	propHeight := s.logHeight()

	// Find the last proposal (assuming it's the one just submitted above).
	lastProp := queries.GetLastGovProp(s.T(), s.network)
	propID := fmt.Sprintf("%d", lastProp.Id)
	s.T().Logf("Proposal id to vote on: %s", propID)

	// Verify that the validator is sanctioned
	isSanctOutBW1, err := cli.ExecTestCLICmd(s.clientCtx, isSanctCmd, isSanctArgs)
	s.Require().NoError(err, "ExecTestCLICmd query sanction is-sanctioned (first time)")
	isSanctOutBz1 := isSanctOutBW1.Bytes()
	s.T().Logf("query sanction is-sanctioned output (first time):\n%s", isSanctOutBz1)
	var isSanctOut1 sanction.QueryIsSanctionedResponse
	err = json.Unmarshal(isSanctOutBz1, &isSanctOut1)
	s.Require().NoError(err, "Unmarshal QueryIsSanctionedResponse (first time)")
	s.Assert().True(isSanctOut1.IsSanctioned, "is sanctioned (first time)")

	// Cast votes on it. We don't use a TxExecutor because we can't wait for a new
	// block after each vote. We'll check all of them manually once they're submitted.
	voteOutBzs := make([][]byte, len(allVoteArgs))
	for i, voteArgs := range allVoteArgs {
		s.logHeight()
		voteArgs[0] = propID
		voteOutBW, err := cli.ExecTestCLICmd(s.clientCtx, voteCmd, voteArgs)
		s.Require().NoError(err, "[%d]: ExecTestCLICmd tx gov vote", i)
		voteOutBzs[i] = voteOutBW.Bytes()
		s.T().Logf("[%d]: tx gov vote output:\n%s", i, voteOutBzs[i])
	}
	s.logHeight()
	// And now, we check that the votes happened as expected.
	for i, voteOutBz := range voteOutBzs {
		txResp := queries.GetTxFromResponse(s.T(), s.network, voteOutBz)
		if i != sanctionValI {
			s.Assert().Equal(0, int(txResp.Code), "vote[%d] response code", i)
		} else {
			s.Assert().Equal(5, int(txResp.Code), "vote[%d] response code", i)
			s.Assert().Contains(txResp.RawLog, "cannot send from "+s.network.Validators[i].Address.String(), "vote[%d] Raw Log")
			s.Assert().Contains(txResp.RawLog, "account is sanctioned", "vote[%d] Raw Log")
			s.Assert().Contains(txResp.RawLog, "insufficient funds", "vote[%d] Raw Log")
		}
	}

	// Wait for the proposal to pass.
	s.logHeight()
	s.T().Log("waiting for voting period to end")
	s.waitForHeight(propHeight + blocksPerVotingPeriod)

	// Check that the proposal passed.
	finalProp := queries.GetGovProp(s.T(), s.network, propID)
	s.Assert().Equal(govv1.StatusPassed, finalProp.Status, "proposal status")

	// Check that that validator is still sanctioned.
	isSanctOutBW2, err := cli.ExecTestCLICmd(s.clientCtx, isSanctCmd, isSanctArgs)
	s.Require().NoError(err, "ExecTestCLICmd query sanction is-sanctioned (second time)")
	isSanctOutBz2 := isSanctOutBW2.Bytes()
	s.T().Logf("query sanction is-sanctioned output (second time):\n%s", isSanctOutBz2)
	var isSanctOut2 sanction.QueryIsSanctionedResponse
	err = json.Unmarshal(isSanctOutBz2, &isSanctOut2)
	s.Require().NoError(err, "Unmarshal QueryIsSanctionedResponse (second time)")
	s.Assert().True(isSanctOut2.IsSanctioned, "is sanctioned (second time)")

	// Wait 5 more blocks to make sure nothing unravels.
	lastHeight := s.logHeight()
	s.T().Log("waiting 5 blocks before final checks")
	_, err = testutil.WaitForHeightWithTimeout(s.network, lastHeight+5, 30*time.Second)
	s.Require().NoError(err, "waiting for block %d (or 30 seconds)", lastHeight+5)
	s.logHeight()

	// Check that that validator is still sanctioned one last time.
	isSanctOutBW3, err := cli.ExecTestCLICmd(s.clientCtx, isSanctCmd, isSanctArgs)
	s.Require().NoError(err, "ExecTestCLICmd query sanction is-sanctioned (third time)")
	isSanctOutBz3 := isSanctOutBW3.Bytes()
	s.T().Logf("query sanction is-sanctioned output (third time):\n%s", isSanctOutBz3)
	var isSanctOut3 sanction.QueryIsSanctionedResponse
	err = json.Unmarshal(isSanctOutBz3, &isSanctOut3)
	s.Require().NoError(err, "Unmarshal QueryIsSanctionedResponse (third time)")
	s.Assert().True(isSanctOut3.IsSanctioned, "is sanctioned (third time)")

	s.T().Log("done")
}
