package ibc

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	provenanceapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/contracts"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

type TestChain struct {
	*ibctesting.TestChain
}

func SetupTestingApp(t *testing.T) (ibctesting.TestingApp, map[string]json.RawMessage) {
	provenanceApp := provenanceapp.Setup(t)
	return provenanceApp, provenanceApp.DefaultGenesis()
}

func (chain *TestChain) StoreContractCounterDirect(suite *suite.Suite) uint64 {
	codeID, err := contracts.StoreContractCode(chain.GetProvenanceApp(), chain.GetContext(), contracts.CounterWasm())
	suite.Require().NoError(err, "counter contract direct code load failed", err)
	println("loaded counter contract with code id: ", codeID)
	return codeID
}

func (chain *TestChain) StoreContractEchoDirect(suite *suite.Suite) uint64 {
	codeID, err := contracts.StoreContractCode(chain.GetProvenanceApp(), chain.GetContext(), contracts.EchoWasm())
	suite.Require().NoError(err, "echo contract direct code load failed", err)
	println("loaded echo contract with code id: ", codeID)
	return codeID
}

func (chain *TestChain) StoreContractRateLimiterDirect(suite *suite.Suite) uint64 {
	codeID, err := contracts.StoreContractCode(chain.GetProvenanceApp(), chain.GetContext(), contracts.RateLimiterWasm())
	suite.Require().NoError(err, "rate limiter contract direct code load failed")
	println("loaded rate limiter contract with code id: ", codeID)
	return codeID
}

func (chain *TestChain) InstantiateContract(suite *suite.Suite, msg string, codeID uint64) sdk.AccAddress {
	addr, err := contracts.InstantiateContract(chain.GetProvenanceApp(), chain.GetContext(), msg, codeID)
	suite.Require().NoError(err, "contract instantiation failed", err)
	println("instantiated contract '", codeID, "' with address: ", addr)
	return addr
}

func (chain *TestChain) PinContract(suite *suite.Suite, codeID uint64) {
	err := contracts.PinContract(chain.GetProvenanceApp(), chain.GetContext(), codeID)
	suite.Require().NoError(err, "contract pin failed")
}

func (chain *TestChain) QueryContract(suite *suite.Suite, contract sdk.AccAddress, key []byte) string {
	state, err := contracts.QueryContract(chain.GetProvenanceApp(), chain.GetContext(), contract, key)
	suite.Require().NoError(err, "contract query failed", err)
	println("got query result of ", state)
	return state
}

func (chain *TestChain) RegisterRateLimiterContract(suite *suite.Suite, addr []byte) {
	addrStr, err := sdk.Bech32ifyAddressBytes("cosmos", addr)
	suite.Require().NoError(err)
	provenanceApp := chain.GetProvenanceApp()
	provenanceApp.RateLimitingKeeper.SetParams(chain.GetContext(), ibcratelimit.Params{
		ContractAddress: addrStr,
	})
}

// SendMsgsNoCheck is an alternative to ibctesting.TestChain.SendMsgs so that it doesn't check for errors. That should be handled by the caller
func (chain *TestChain) SendMsgsNoCheck(msgs ...sdk.Msg) (*sdk.Result, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	_, r, err := SignAndDeliver(
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return r, nil
}

// SignAndDeliver signs and delivers a transaction without asserting the results. This overrides the function
// from ibctesting
func SignAndDeliver(
	txCfg client.TxConfig, app *baseapp.BaseApp, _ cmtproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, _ := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 2500)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)

	// Simulate a sending a transaction and committing a block
	gInfo, res, err := app.SimDeliver(txCfg.TxEncoder(), tx)

	return gInfo, res, err
}

// GetProvenanceApp returns the current chain's app as an ProvenanceApp
func (chain *TestChain) GetProvenanceApp() *provenanceapp.App {
	v, _ := chain.App.(*provenanceapp.App)
	return v
}
