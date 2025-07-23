package cli_test

import (
	"encoding/base64"
	"fmt"
	cmtcli "github.com/cometbft/cometbft/libs/cli"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/gogoproto/proto"
	testcli "github.com/provenance-io/provenance/testutil/cli"
	"github.com/provenance-io/provenance/x/smartaccounts/client/cli"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
	"github.com/provenance-io/provenance/x/smartaccounts/utils"
	"github.com/spf13/cobra"
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	keyring          keyring.Keyring
	keyringEntries   []testutil.TestKeyringEntry
	accountAddresses []sdk.AccAddress

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey
}

func (s *IntegrationTestSuite) GenerateAccountsWithKeyrings(number int) {
	s.keyringEntries, s.keyring = testutil.GenerateTestKeyring(s.T(), number, s.cfg.Codec)
	s.accountAddresses = testutil.GetKeyringEntryAddresses(s.keyringEntries)
}
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	var addrErr error
	s.accountAddr, addrErr = sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(addrErr)

	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 1
	s.cfg.ChainID = antewrapper.SimAppChainID
	s.GenerateAccountsWithKeyrings(4)

	testutil.MutateGenesisState(s.T(), &s.cfg, banktypes.ModuleName, &banktypes.GenesisState{}, func(bankGenState *banktypes.GenesisState) *banktypes.GenesisState {
		for i := range s.accountAddresses {
			bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{Address: s.accountAddresses[i].String(), Coins: sdk.NewCoins(
				sdk.NewInt64Coin("nhash", 100_000_000), sdk.NewInt64Coin(s.cfg.BondDenom, 100_000_000),
			).Sort()})
		}
		return bankGenState
	})

	testutil.MutateGenesisState(s.T(), &s.cfg, authtypes.ModuleName, &authtypes.GenesisState{}, func(authData *authtypes.GenesisState) *authtypes.GenesisState {
		var genAccounts []authtypes.GenesisAccount
		genAccounts = append(genAccounts,
			authtypes.NewBaseAccount(s.accountAddresses[0], nil, 1, 0),
			authtypes.NewBaseAccount(s.accountAddresses[1], nil, 2, 0),
			authtypes.NewBaseAccount(s.accountAddresses[2], nil, 3, 0),
			authtypes.NewBaseAccount(s.accountAddresses[3], nil, 4, 0),
		)
		accounts, err := authtypes.PackAccounts(genAccounts)
		s.Require().NoError(err, "should be able to pack accounts for genesis state when setting up suite")
		authData.Accounts = accounts
		return authData
	})

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	s.network.Validators[0].ClientCtx = s.network.Validators[0].ClientCtx.WithKeyring(s.keyring)

	_, err = testutil.WaitForHeight(s.network, 6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.Cleanup(s.network, s.T())
}

func (s *IntegrationTestSuite) TestSmartAccountCredentialTxCommands() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"add a credential to a smart account",
			cli.GetCmdAddFido2Credentials(),
			[]string{
				fmt.Sprintf("--%s=%s", cli.FlagAttestation, base64.RawURLEncoding.EncodeToString([]byte(utils.TestCredentialRequestResponses["success"]))),
				fmt.Sprintf("--%s=%s", cli.FlagSender, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=%s", cli.FlagUserIdentifier, "foo@bar.com"),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
			},
			false,
			&sdk.TxResponse{},
			0,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			testcli.NewTxExecutor(tc.cmd, tc.args).
				WithExpErr(tc.expectErr).
				WithExpCode(tc.expectedCode).
				Execute(s.T(), s.network)
		})
	}
}

func (s *IntegrationTestSuite) TestSmartAccountQueryFlow() {
	// First register a FIDO2 credential to create a new smart account
	registerArgs := []string{
		fmt.Sprintf("--%s=%s", cli.FlagAttestation, base64.RawURLEncoding.EncodeToString([]byte(utils.TestCredentialRequestResponses["success"]))),
		fmt.Sprintf("--%s=%s", cli.FlagSender, s.accountAddresses[1].String()),
		fmt.Sprintf("--%s=%s", cli.FlagUserIdentifier, "test@example.com"),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[1].String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 10)).String()),
	}

	registerCmd := cli.GetCmdAddFido2Credentials()
	testcli.NewTxExecutor(registerCmd, registerArgs).
		WithExpErr(false).
		WithExpCode(0).
		Execute(s.T(), s.network)

	// Wait for the next block to ensure the account is created
	heightRequired, err := s.network.LatestHeight()
	s.Require().NoError(err)
	_, err = testutil.WaitForHeight(s.network, heightRequired+1)
	s.Require().NoError(err)

	// Now query the smart account
	queryArgs := []string{
		s.accountAddresses[1].String(),
		fmt.Sprintf("--%s=json", cmtcli.OutputFlag),
	}

	queryCmd := cli.QueryAccountByAddressCmd()
	clientCtx := s.network.Validators[0].ClientCtx
	responseFromExec, err := clitestutil.ExecTestCLICmd(clientCtx, queryCmd, queryArgs)
	s.Require().NoError(err)

	// Parse the response and verify the account exists
	var response types.SmartAccountQueryResponse
	err = s.cfg.Codec.UnmarshalJSON(responseFromExec.Bytes(), &response)
	s.Require().NoError(err)

	// Verify account data
	s.Require().Equal(s.accountAddresses[1].String(), response.Provenanceaccount.Address)
	s.Require().NotEmpty(response.Provenanceaccount.Credentials)
	s.Require().Equal(1, len(response.Provenanceaccount.Credentials))
	s.Require().Equal(types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN, response.Provenanceaccount.Credentials[0].BaseCredential.Variant)
}
