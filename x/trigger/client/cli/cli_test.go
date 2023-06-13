package cli_test

import (
	"fmt"
	"testing"
	"time"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	triggercli "github.com/provenance-io/provenance/x/trigger/client/cli"

	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/trigger/types"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg        network.Config
	network    *network.Network
	keyring    keyring.Keyring
	keyringDir string

	accountAddr      sdk.AccAddress
	accountKey       *secp256k1.PrivKey
	accountAddresses []sdk.AccAddress

	startingTriggerID  triggertypes.TriggerID
	startingQueueIndex uint64
	triggers           []triggertypes.Trigger
	gasLimits          []triggertypes.GasLimit
	queuedTriggers     []triggertypes.QueuedTrigger
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.cfg = testutil.DefaultTestNetworkConfig()
	genesisState := s.cfg.GenesisState

	s.cfg.NumValidators = 1
	s.GenerateAccountsWithKeyrings(2)

	var genBalances []banktypes.Balance
	for i := range s.accountAddresses {
		genBalances = append(genBalances, banktypes.Balance{Address: s.accountAddresses[i].String(), Coins: sdk.NewCoins(
			sdk.NewCoin("nhash", sdk.NewInt(100000000)), sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(100000000)),
		).Sort()})
	}
	genBalances = append(genBalances, banktypes.Balance{Address: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", Coins: sdk.NewCoins(
		sdk.NewCoin("nhash", sdk.NewInt(100000000)), sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(100000000))).Sort()})
	var bankGenState banktypes.GenesisState
	bankGenState.Params = banktypes.DefaultParams()
	bankGenState.Balances = genBalances
	bankDataBz, err := s.cfg.Codec.MarshalJSON(&bankGenState)
	s.Require().NoError(err, "should be able to marshal bank genesis state when setting up suite")
	genesisState[banktypes.ModuleName] = bankDataBz

	var authData authtypes.GenesisState
	var genAccounts []authtypes.GenesisAccount
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[1], nil, 4, 0))
	accounts, err := authtypes.PackAccounts(genAccounts)
	s.Require().NoError(err, "should be able to pack accounts for genesis state when setting up suite")
	authData.Accounts = accounts
	authDataBz, err := s.cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err, "should be able to marshal auth genesis state when setting up suite")
	genesisState[authtypes.ModuleName] = authDataBz

	now := time.Now().UTC()
	s.startingTriggerID = uint64(7)
	s.startingQueueIndex = uint64(1)
	s.triggers = []triggertypes.Trigger{
		s.CreateTrigger(1, s.accountAddresses[0].String(), &triggertypes.BlockHeightEvent{BlockHeight: 100}, &triggertypes.MsgDestroyTriggerRequest{Id: 3, Authority: s.accountAddresses[0].String()}),
		s.CreateTrigger(2, s.accountAddresses[1].String(), &triggertypes.BlockHeightEvent{BlockHeight: 100}, &triggertypes.MsgDestroyTriggerRequest{Id: 4, Authority: s.accountAddresses[1].String()}),
		s.CreateTrigger(3, s.accountAddresses[0].String(), &triggertypes.BlockHeightEvent{BlockHeight: 1000}, &triggertypes.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
		s.CreateTrigger(4, s.accountAddresses[1].String(), &triggertypes.BlockHeightEvent{BlockHeight: 1000}, &triggertypes.MsgDestroyTriggerRequest{Id: 2, Authority: s.accountAddresses[1].String()}),
		s.CreateTrigger(7, s.accountAddresses[0].String(), &triggertypes.BlockHeightEvent{BlockHeight: 1000}, &triggertypes.MsgDestroyTriggerRequest{Id: 2, Authority: s.accountAddresses[1].String()}),
	}
	s.gasLimits = []triggertypes.GasLimit{
		{
			TriggerId: 1,
			Amount:    20000,
		},
		{
			TriggerId: 2,
			Amount:    20000,
		},
		{
			TriggerId: 3,
			Amount:    20000,
		},
		{
			TriggerId: 4,
			Amount:    20000,
		},
		{
			TriggerId: 5,
			Amount:    20000,
		},
		{
			TriggerId: 6,
			Amount:    20000,
		},
		{
			TriggerId: 7,
			Amount:    20000,
		},
	}
	s.queuedTriggers = []triggertypes.QueuedTrigger{
		{
			BlockHeight: 5,
			Time:        now,
			Trigger:     s.CreateTrigger(5, s.accountAddresses[0].String(), &triggertypes.BlockHeightEvent{BlockHeight: 5}, &triggertypes.MsgDestroyTriggerRequest{Id: 3, Authority: s.accountAddresses[0].String()}),
		},
		{
			BlockHeight: 5,
			Time:        now,
			Trigger:     s.CreateTrigger(6, s.accountAddresses[1].String(), &triggertypes.BlockHeightEvent{BlockHeight: 5}, &triggertypes.MsgDestroyTriggerRequest{Id: 4, Authority: s.accountAddresses[1].String()}),
		},
	}

	triggerData := triggertypes.NewGenesisState(
		s.startingTriggerID,
		s.startingQueueIndex,
		s.triggers,
		s.gasLimits,
		s.queuedTriggers,
	)

	triggerDataBz, err := s.cfg.Codec.MarshalJSON(triggerData)
	s.Require().NoError(err, "should be able to marshal trigger genesis state when setting up suite")
	genesisState[triggertypes.ModuleName] = triggerDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	_, err = s.network.WaitForHeight(6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.Require().NoError(s.network.WaitForNextBlock(), "WaitForNextBlock")
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) GenerateAccountsWithKeyrings(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err, "Keyring.New")
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err, "Keyring.NewMneomonic")
		addr, err := info.GetAddress()
		if err != nil {
			panic(err)
		}
		s.accountAddresses = append(s.accountAddresses, addr)
	}
}

func (s *IntegrationTestSuite) CreateTrigger(id uint64, owner string, event types.TriggerEventI, action sdk.Msg) types.Trigger {
	actions, _ := sdktx.SetMsgs([]sdk.Msg{action})
	any, _ := codectypes.NewAnyWithValue(event)
	return types.NewTrigger(id, owner, any, actions)
}

func (s *IntegrationTestSuite) TestQueryTriggers() {
	testCases := []struct {
		name         string
		args         []string
		byId         bool
		expectErrMsg string
		expectedCode uint32
		expectedIds  []int
	}{
		{
			name: "query all triggers",
			args: []string{
				"all",
			},
			byId:         false,
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []int{1, 2, 8, 9},
		},
		{
			name: "query paginate with limit 1",
			args: []string{
				"all",
				"--limit",
				"1",
			},
			byId:         false,
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []int{1},
		},
		{
			name: "query paginate with excessive limit",
			args: []string{
				"all",
				"--limit",
				"100",
			},
			byId:         false,
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []int{1, 2, 8, 9},
		},
		{
			name: "query trigger by id",
			args: []string{
				"1",
			},
			byId:         true,
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []int{1},
		},
		{
			name: "query trigger by invalid id",
			args: []string{
				"1000",
			},
			byId:         true,
			expectErrMsg: "failed to query trigger 1000: rpc error: code = Unknown desc = trigger not found: unknown request",
			expectedCode: types.ErrTriggerNotFound.ABCICode(),
			expectedIds:  []int{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, triggercli.GetTriggersCmd(), append(tc.args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			if len(tc.expectErrMsg) > 0 {
				s.EqualError(err, tc.expectErrMsg, "should have correct error message for invalid QueryTriggers")
			} else if tc.byId {
				var response types.QueryTriggerByIDResponse
				s.NoError(err, "should have no error message for valid QueryTriggerByID")
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.NoError(err, "should have no error message when unmarshalling response to QueryTriggerByID")
				s.Equal(int(tc.expectedIds[0]), int(response.Trigger.Id), "should return correct trigger for QueryTriggerByID")
			} else {
				var response types.QueryTriggersResponse
				s.NoError(err, "should have no error message for valid QueryTriggers")
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.NoError(err, "should have no error message when unmarshalling response to QueryTriggers")
				var triggerIDs []int
				for _, rp := range response.Triggers {
					triggerIDs = append(triggerIDs, int(rp.Id))
				}
				s.ElementsMatch(tc.expectedIds, triggerIDs, "should have all triggers for QueryTriggers")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestAddBlockHeightTrigger() {
	testCases := []struct {
		name         string
		height       string
		fileContent  string
		expectErrMsg string
		expectedCode uint32
		expectedIds  []int
	}{
		{
			name:         "create block height trigger",
			height:       "900",
			fileContent:  "",
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []int{8},
		},
		{
			name:         "create invalid block height trigger for past block",
			height:       "1",
			fileContent:  "",
			expectErrMsg: "",
			expectedCode: types.ErrInvalidBlockHeight.ABCICode(),
			expectedIds:  []int{},
		},
		{
			name:         "invalid file format",
			height:       "1",
			fileContent:  "abc",
			expectErrMsg: "unable to parse message file: invalid character 'a' looking for beginning of value",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:         "bad height",
			height:       "abc",
			fileContent:  "",
			expectErrMsg: "invalid block height \"abc\": strconv.Atoi: parsing \"abc\": invalid syntax",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:         "invalid message format",
			height:       "1",
			fileContent:  "{}",
			expectErrMsg: "unable to parse message file: Any JSON doesn't have '@type'",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:   "unsupported action",
			height: "1000",
			fileContent: fmt.Sprintf(`
			{
					"@type": "/cosmos.bank.v1beta1.InvalidMessageSend",
					"from_address": "%s",
					"to_address": "%s",
					"amount": [
						{
							"denom": "nhash",
							"amount": "10"
						}
					]
			}`, s.accountAddresses[0].String(), s.accountAddresses[1].String()),
			expectErrMsg: "unable to parse message file: unable to resolve type URL /cosmos.bank.v1beta1.InvalidMessageSend",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:   "invalid internal message data",
			height: "1000",
			fileContent: fmt.Sprintf(`
			{
					"@type": "/cosmos.bank.v1beta1.InvalidMessageSend",
					"from_address": "%s",
					"to_address": "%s",
					"amount": [
						{
							"denom": "nhash",
							"amount": "10"
						}
					]
			}`, "abc", s.accountAddresses[1].String()),
			expectErrMsg: "unable to parse message file: unable to resolve type URL /cosmos.bank.v1beta1.InvalidMessageSend",
			expectedCode: 0,
			expectedIds:  []int{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			var message string
			if len(tc.fileContent) == 0 {
				message = fmt.Sprintf(`
				{
						"@type": "/cosmos.bank.v1beta1.MsgSend",
						"from_address": "%s",
						"to_address": "%s",
						"amount": [
							{
								"denom": "nhash",
								"amount": "10"
							}
						]
				}`, s.accountAddresses[0].String(), s.accountAddresses[1].String())
			} else {
				message = tc.fileContent
			}

			messageFile := sdktestutil.WriteToNewTempFile(s.T(), message)

			args := []string{
				tc.height,
				messageFile.Name(),
			}
			flags := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			args = append(args, flags...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, triggercli.GetCmdAddBlockHeightTrigger(), append(args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			var response sdk.TxResponse
			marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg, "should have correct error for invalid AddBlockHeightTrigger request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for invalid AddBlockHeightTrigger request")
			} else {
				s.Assert().NoError(err, "should have no error for valid AddBlockHeightTrigger request")
				s.Assert().NoError(marshalErr, out.String(), "should have no error for marshaling request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for AddBlockHeightTrigger request")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestAddTransactionTrigger() {
	testCases := []struct {
		name         string
		fileContent  string
		txEvent      string
		expectErrMsg string
		expectedCode uint32
		expectedIds  []int
	}{
		{
			name:         "create transaction trigger",
			fileContent:  "",
			txEvent:      "",
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []int{8},
		},
		{
			name:         "invalid tx event content",
			fileContent:  "",
			txEvent:      "abc",
			expectErrMsg: "unable to parse event file: invalid character 'a' looking for beginning of value",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:         "invalid file format",
			fileContent:  "abc",
			txEvent:      "",
			expectErrMsg: "unable to parse message file: invalid character 'a' looking for beginning of value",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:         "invalid message format",
			fileContent:  "{}",
			txEvent:      "",
			expectErrMsg: "unable to parse message file: Any JSON doesn't have '@type'",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:    "unsupported action",
			txEvent: "",
			fileContent: fmt.Sprintf(`
			{
					"@type": "/cosmos.bank.v1beta1.InvalidMessageSend",
					"from_address": "%s",
					"to_address": "%s",
					"amount": [
						{
							"denom": "nhash",
							"amount": "10"
						}
					]
			}`, s.accountAddresses[0].String(), s.accountAddresses[1].String()),
			expectErrMsg: "unable to parse message file: unable to resolve type URL /cosmos.bank.v1beta1.InvalidMessageSend",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:    "invalid internal message data",
			txEvent: "",
			fileContent: fmt.Sprintf(`
			{
					"@type": "/cosmos.bank.v1beta1.InvalidMessageSend",
					"from_address": "%s",
					"to_address": "%s",
					"amount": [
						{
							"denom": "nhash",
							"amount": "10"
						}
					]
			}`, "abc", s.accountAddresses[1].String()),
			expectErrMsg: "unable to parse message file: unable to resolve type URL /cosmos.bank.v1beta1.InvalidMessageSend",
			expectedCode: 0,
			expectedIds:  []int{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			var message string
			if len(tc.fileContent) == 0 {
				message = fmt.Sprintf(`
				{
						"@type": "/cosmos.bank.v1beta1.MsgSend",
						"from_address": "%s",
						"to_address": "%s",
						"amount": [
							{
								"denom": "nhash",
								"amount": "10"
							}
						]
				}`, s.accountAddresses[0].String(), s.accountAddresses[1].String())
			} else {
				message = tc.fileContent
			}
			messageFile := sdktestutil.WriteToNewTempFile(s.T(), message)

			var txEvent string
			if len(tc.txEvent) == 0 {
				txEvent = fmt.Sprintf(`
				{
					"name": "coin_received",
					"attributes": [
						{
							"name": "receiver",
							"value": "%s"
						},
						{
							"name": "amount",
							"value": "100nhash"
						}
					]
				}
				`, s.accountAddresses[0].String())
			} else {
				txEvent = tc.txEvent
			}
			txEventFile := sdktestutil.WriteToNewTempFile(s.T(), txEvent)

			args := []string{
				txEventFile.Name(),
				messageFile.Name(),
			}
			flags := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			args = append(args, flags...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, triggercli.GetCmdAddTransactionTrigger(), append(args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			var response sdk.TxResponse
			marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg, "should have correct error for invalid AddTransactionTrigger request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for invalid AddTransactionTrigger request")
			} else {
				s.Assert().NoError(err, "should have no error for valid AddTransactionTrigger request")
				s.Assert().NoError(marshalErr, out.String(), "should have no marshalling error for valid AddTransactionTrigger request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for valid AddTransactionTrigger request")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestAddBlockTimeTrigger() {
	testCases := []struct {
		name         string
		blockTime    string
		fileContent  string
		expectErrMsg string
		expectedCode uint32
		expectedIds  []int
	}{
		{
			name:         "create block time trigger",
			blockTime:    "2100-05-19T13:49:00-04:00",
			fileContent:  "",
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []int{8},
		},
		{
			name:         "create invalid block time trigger for past block",
			blockTime:    "2000-05-19T13:49:00-04:00",
			fileContent:  "",
			expectErrMsg: "",
			expectedCode: types.ErrInvalidBlockTime.ABCICode(),
			expectedIds:  []int{},
		},
		{
			name:         "invalid file format",
			blockTime:    "2100-05-19T13:49:00-04:00",
			fileContent:  "abc",
			expectErrMsg: "unable to parse message file: invalid character 'a' looking for beginning of value",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:         "invalid bad time",
			blockTime:    "abc",
			fileContent:  "",
			expectErrMsg: "unable to parse time (abc) required format is RFC3339 (2006-01-02T15:04:05Z07:00): parsing time \"abc\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"abc\" as \"2006\"",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:         "invalid message format",
			blockTime:    "2100-05-19T13:49:00-04:00",
			fileContent:  "{}",
			expectErrMsg: "unable to parse message file: Any JSON doesn't have '@type'",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:      "unsupported action",
			blockTime: "2100-05-19T13:49:00-04:00",
			fileContent: fmt.Sprintf(`
			{
					"@type": "/cosmos.bank.v1beta1.InvalidMessageSend",
					"from_address": "%s",
					"to_address": "%s",
					"amount": [
						{
							"denom": "nhash",
							"amount": "10"
						}
					]
			}`, s.accountAddresses[0].String(), s.accountAddresses[1].String()),
			expectErrMsg: "unable to parse message file: unable to resolve type URL /cosmos.bank.v1beta1.InvalidMessageSend",
			expectedCode: 0,
			expectedIds:  []int{},
		},
		{
			name:      "invalid internal message data",
			blockTime: "2100-05-19T13:49:00-04:00",
			fileContent: fmt.Sprintf(`
			{
					"@type": "/cosmos.bank.v1beta1.InvalidMessageSend",
					"from_address": "%s",
					"to_address": "%s",
					"amount": [
						{
							"denom": "nhash",
							"amount": "10"
						}
					]
			}`, "abc", s.accountAddresses[1].String()),
			expectErrMsg: "unable to parse message file: unable to resolve type URL /cosmos.bank.v1beta1.InvalidMessageSend",
			expectedCode: 0,
			expectedIds:  []int{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			var message string
			if len(tc.fileContent) == 0 {
				message = fmt.Sprintf(`
				{
						"@type": "/cosmos.bank.v1beta1.MsgSend",
						"from_address": "%s",
						"to_address": "%s",
						"amount": [
							{
								"denom": "nhash",
								"amount": "10"
							}
						]
				}`, s.accountAddresses[0].String(), s.accountAddresses[1].String())
			} else {
				message = tc.fileContent
			}

			messageFile := sdktestutil.WriteToNewTempFile(s.T(), message)

			args := []string{
				tc.blockTime,
				messageFile.Name(),
			}
			flags := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accountAddresses[0].String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			args = append(args, flags...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, triggercli.GetCmdAddBlockTimeTrigger(), append(args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			var response sdk.TxResponse
			marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg, "should have correct error for invalid AddBlockTimeTrigger request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for invalid AddBlockTimeTrigger request")
			} else {
				s.Assert().NoError(err, "should have no error for valid AddBlockTimeTrigger request")
				s.Assert().NoError(marshalErr, out.String(), "should have no marshal error for valid AddBlockTimeTrigger request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for valid AddBlockTimeTrigger request")
			}
		})
	}
}

func (s *IntegrationTestSuite) TestDestroyTrigger() {
	testCases := []struct {
		name         string
		triggerID    string
		expectErrMsg string
		expectedCode uint32
		signer       string
	}{
		{
			name:         "valid - destroy trigger",
			triggerID:    "7",
			expectErrMsg: "",
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "invalid - unable to destroy trigger created by someone else",
			triggerID:    "2",
			expectErrMsg: "",
			expectedCode: types.ErrInvalidTriggerAuthority.ABCICode(),
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "invalid - trigger id does not exist",
			triggerID:    "999",
			expectErrMsg: "",
			expectedCode: types.ErrTriggerNotFound.ABCICode(),
			signer:       s.accountAddresses[0].String(),
		},
		{
			name:         "invalid - trigger id format",
			triggerID:    "abc",
			expectErrMsg: "invalid trigger id \"abc\": strconv.Atoi: parsing \"abc\": invalid syntax",
			expectedCode: 0,
			signer:       s.accountAddresses[0].String(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {

			clientCtx := s.network.Validators[0].ClientCtx.WithKeyringDir(s.keyringDir).WithKeyring(s.keyring)

			args := []string{
				tc.triggerID,
			}
			flags := []string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, tc.signer),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			}
			args = append(args, flags...)

			out, err := clitestutil.ExecTestCLICmd(clientCtx, triggercli.GetCmdDestroyTrigger(), append(args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			var response sdk.TxResponse
			marshalErr := clientCtx.Codec.UnmarshalJSON(out.Bytes(), &response)
			if len(tc.expectErrMsg) > 0 {
				s.Assert().EqualError(err, tc.expectErrMsg, "should have correct error for invalid DestroyTrigger request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for invalid DestroyTrigger request")
			} else {
				s.Assert().NoError(err, "should have no error for valid DestroyTrigger request")
				s.Assert().NoError(marshalErr, out.String(), "should have no marshal error for valid DestroyTrigger request")
				s.Assert().Equal(tc.expectedCode, response.Code, "should have correct response code for valid DestroyTrigger request")
			}
		})
	}
}
