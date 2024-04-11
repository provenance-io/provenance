package testutil

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/queries"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg       network.Config
	network   *network.Network
	clientCtx client.Context

	commonFlags []string
	val0        *network.Validator
	valAddr     sdk.AccAddress

	addrCodec address.Codec
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, s.bondCoins(10).String()),
	}
	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.clientCtx = s.network.Validators[0].ClientCtx
	s.valAddr = s.network.Validators[0].Address
	s.val0 = s.network.Validators[0]

	sdkcfg := sdk.GetConfig()
	s.addrCodec = addresscodec.NewBech32Codec(sdkcfg.GetBech32AccountAddrPrefix())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) stopIfFailed() {
	if s.T().Failed() {
		s.T().FailNow()
	}
}

// bondCoin creates an sdk.Coin with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(s.cfg.BondDenom, amt)
}

// bondCoins creates an sdk.Coins with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoins(amt int64) sdk.Coins {
	return sdk.NewCoins(s.bondCoin(amt))
}

func (s *IntegrationTestSuite) addAccountToKeyring(index, count int) string {
	memberNumber := uuid.New().String()

	info, _, err := s.clientCtx.Keyring.NewMnemonic(
		fmt.Sprintf("member%s", memberNumber),
		keyring.English, sdk.FullFundraiserPath,
		keyring.DefaultBIP39Passphrase, hd.Secp256k1,
	)
	s.Require().NoError(err, "[%d/%d] NewMnemonic", index, count)

	pk, err := info.GetPubKey()
	s.Require().NoError(err, "[%d/%d] GetPubKey", index, count)

	addr := pk.Address()
	rv, err := s.addrCodec.BytesToString(addr)
	s.Require().NoError(err, "[%d/%d] BytesToString(%v)", index, count, addr)

	return rv
}

// createAndFundAccount creates an account, adding the key to the keyring, funded with the provided amount of bond-denom coins.
func (s *IntegrationTestSuite) createAndFundAccount(bondCoinAmt int64) string {
	addr := s.addAccountToKeyring(1, 1)
	out, err := clitestutil.MsgSendExec(
		s.clientCtx,
		s.valAddr,
		asStringer(addr),
		s.bondCoins(bondCoinAmt),
		s.addrCodec,
		s.commonFlags...,
	)
	s.Require().NoError(err, "MsgSendExec")
	outBz := out.Bytes()
	s.T().Logf("MsgSendExec response:\n%s", string(outBz))
	s.Require().NoError(s.network.WaitForNextBlock(), "WaitForNextBlock after MsgSendExec")
	resp := queries.GetTxFromResponse(s.T(), s.val0, outBz)
	s.Require().Equal(0, int(resp.Code), "MsgSendExec tx response code. Tx response:\n%#v", resp)

	return addr
}

// createAndFundAccounts creates count account, adding the keys to the keyring, each funded with the provided amount of bond-denom coins.
func (s *IntegrationTestSuite) createAndFundAccounts(count int, bondCoinAmt int64) []string {
	addrs := make([]string, count)
	for i := range addrs {
		addrs[i] = s.addAccountToKeyring(i+1, count)
	}

	amount := s.bondCoins(bondCoinAmt).String()

	cmd := bankcli.NewMultiSendTxCmd(s.addrCodec)
	var args []string
	args = append(args, s.valAddr.String())
	args = append(args, addrs...)
	args = append(args, amount)
	args = s.appendCommonFlagsTo(args...)

	out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, args)
	s.Require().NoError(err, "ExecTestCLICmd bank multisend")
	outBz := out.Bytes()
	s.T().Logf("Multisend response:\n%s", string(outBz))
	s.Require().NoError(s.network.WaitForNextBlock(), "WaitForNextBlock after multisend")
	resp := queries.GetTxFromResponse(s.T(), s.val0, outBz)
	s.Require().Equal(0, int(resp.Code), "Multisend tx response code. Tx response:\n%#v", resp)

	return addrs
}

// appendCommonFlagsTo adds this suite's common flags to the end of the provided arguments.
func (s *IntegrationTestSuite) appendCommonFlagsTo(args ...string) []string {
	return append(args, s.commonFlags...)
}

// assertErrorContents calls AssertErrorContents using this suite's t.
func (s *IntegrationTestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

var _ fmt.Stringer = asStringer("")

// asStringer is a string that has a String() function on it so that we can provide a string to MsgSendExec.
type asStringer string

// String implements the Stringer interface.
func (s asStringer) String() string {
	return string(s)
}
