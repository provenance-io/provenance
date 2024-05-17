package ibc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/cometbft/cometbft/abci/types"
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
func (chain *TestChain) SendMsgsNoCheck(suite *suite.Suite, msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain.TestChain)

	resp, err := SignAndDeliver(
		// chain.TB,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		true,
		chain.CurrentHeader.GetTime(),
		chain.NextVals.Hash(),
		chain.SenderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	chain.commitBlock(suite, resp)

	chain.Coordinator.IncrementTime()

	suite.Require().Len(resp.TxResults, 1)
	txResult := resp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	return txResult, nil
}

// Copied from ibctesting because it's private
func (chain *TestChain) commitBlock(suite *suite.Suite, res *abci.ResponseFinalizeBlock) {
	_, err := chain.App.Commit()
	suite.Require().NoError(err)

	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentTMClientHeader()

	// val set changes returned from previous block get applied to the next validators
	// of this block. See tendermint spec for details.
	chain.Vals = chain.NextVals
	chain.NextVals = ibctesting.ApplyValSetChanges(chain, chain.Vals, res.ValidatorUpdates)

	// increment the current header
	chain.CurrentHeader = cmtproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.NextVals.Hash(),
		ProposerAddress:    chain.CurrentHeader.ProposerAddress,
	}
}

// SignAndDeliver signs and delivers a transaction without asserting the results. This overrides the function
// from ibctesting
func SignAndDeliver(
	txCfg client.TxConfig, app *baseapp.BaseApp, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, _ bool, blockTime time.Time, nextValHash []byte, priv ...cryptotypes.PrivKey,
) (*abci.ResponseFinalizeBlock, error) {
	// tb.Helper()
	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		simtestutil.DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)

	if err != nil {
		return nil, err
	}

	txBytes, err := txCfg.TxEncoder()(tx)
	if err != nil {
		return nil, err
	}

	return app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             app.LastBlockHeight() + 1,
		Time:               blockTime,
		NextValidatorsHash: nextValHash,
		Txs:                [][]byte{txBytes},
	})
}

// GetProvenanceApp returns the current chain's app as an ProvenanceApp
func (chain *TestChain) GetProvenanceApp() *provenanceapp.App {
	v, _ := chain.App.(*provenanceapp.App)
	return v
}
