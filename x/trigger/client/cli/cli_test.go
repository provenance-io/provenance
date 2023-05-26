package cli_test

import (
	"fmt"
	"testing"
	"time"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
	s.Require().NoError(err)
	genesisState[banktypes.ModuleName] = bankDataBz

	var authData authtypes.GenesisState
	var genAccounts []authtypes.GenesisAccount
	authData.Params = authtypes.DefaultParams()
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[0], nil, 3, 0))
	genAccounts = append(genAccounts, authtypes.NewBaseAccount(s.accountAddresses[1], nil, 4, 0))
	accounts, err := authtypes.PackAccounts(genAccounts)
	s.Require().NoError(err)
	authData.Accounts = accounts
	authDataBz, err := s.cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	now := time.Now().UTC()
	s.startingTriggerID = uint64(7)
	s.startingQueueIndex = uint64(1)
	s.triggers = []triggertypes.Trigger{
		s.CreateTrigger(1, s.accountAddresses[0].String(), &triggertypes.BlockHeightEvent{BlockHeight: 100}, &triggertypes.MsgDestroyTriggerRequest{Id: 3, Authority: s.accountAddresses[0].String()}),
		s.CreateTrigger(2, s.accountAddresses[1].String(), &triggertypes.BlockHeightEvent{BlockHeight: 100}, &triggertypes.MsgDestroyTriggerRequest{Id: 4, Authority: s.accountAddresses[1].String()}),
		s.CreateTrigger(3, s.accountAddresses[0].String(), &triggertypes.BlockHeightEvent{BlockHeight: 1000}, &triggertypes.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
		s.CreateTrigger(4, s.accountAddresses[1].String(), &triggertypes.BlockHeightEvent{BlockHeight: 1000}, &triggertypes.MsgDestroyTriggerRequest{Id: 2, Authority: s.accountAddresses[1].String()}),
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
	s.Require().NoError(err)
	genesisState[triggertypes.ModuleName] = triggerDataBz

	s.cfg.GenesisState = genesisState

	s.cfg.ChainID = antewrapper.SimAppChainID

	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New")

	_, err = s.network.WaitForHeight(6)
	s.Require().NoError(err, "WaitForHeight")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.Require().NoError(s.network.WaitForNextBlock())
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) GenerateAccountsWithKeyrings(number int) {
	path := hd.CreateHDPath(118, 0, 0).String()
	s.keyringDir = s.T().TempDir()
	kr, err := keyring.New(s.T().Name(), "test", s.keyringDir, nil, s.cfg.Codec)
	s.Require().NoError(err)
	s.keyring = kr
	for i := 0; i < number; i++ {
		keyId := fmt.Sprintf("test_key%v", i)
		info, _, err := kr.NewMnemonic(keyId, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
		s.Require().NoError(err)
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
		expectedIds  []uint64
	}{
		{
			name: "query all triggers",
			args: []string{
				"all",
			},
			byId:         false,
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []uint64{1, 2},
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
			expectedIds:  []uint64{1},
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
			expectedIds:  []uint64{1, 2},
		},
		{
			name: "query trigger by id",
			args: []string{
				"1",
			},
			byId:         true,
			expectErrMsg: "",
			expectedCode: 0,
			expectedIds:  []uint64{1},
		},
		{
			name: "query trigger by invalid id",
			args: []string{
				"1000",
			},
			byId:         true,
			expectErrMsg: "failed to query trigger 1000: rpc error: code = Unknown desc = trigger not found: unknown request",
			expectedCode: types.ErrTriggerNotFound.ABCICode(),
			expectedIds:  []uint64{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			clientCtx := s.network.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, triggercli.GetTriggersCmd(), append(tc.args, []string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)}...))
			if len(tc.expectErrMsg) > 0 {
				s.EqualError(err, tc.expectErrMsg)
			} else if tc.byId {
				var response types.QueryTriggerByIDResponse
				s.NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.NoError(err)
				s.Equal(tc.expectedIds[0], response.Trigger.Id)
			} else {
				var response types.QueryTriggersResponse
				s.NoError(err)
				err = s.cfg.Codec.UnmarshalJSON(out.Bytes(), &response)
				s.NoError(err)
				var triggerIDs []uint64
				for _, rp := range response.Triggers {
					triggerIDs = append(triggerIDs, rp.Id)
				}
				s.ElementsMatch(tc.expectedIds, triggerIDs)
			}
		})
	}
}
