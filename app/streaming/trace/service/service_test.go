package service

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	types1 "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	interfaceRegistry    = codecTypes.NewInterfaceRegistry()
	testMarshaller       = codec.NewProtoCodec(interfaceRegistry)
	testStreamingService *TraceStreamingService
	testingCtx           sdk.Context

	// test abci message types
	mockHash          = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	testBeginBlockReq = abci.RequestBeginBlock{
		Header: types1.Header{
			Height: 1,
		},
		ByzantineValidators: []abci.Evidence{},
		Hash:                mockHash,
		LastCommitInfo: abci.LastCommitInfo{
			Round: 1,
			Votes: []abci.VoteInfo{},
		},
	}
	testBeginBlockRes = abci.ResponseBeginBlock{
		Events: []abci.Event{
			{
				Type: "testEventType1",
			},
			{
				Type: "testEventType2",
			},
		},
	}
	testEndBlockReq = abci.RequestEndBlock{
		Height: 1,
	}
	testEndBlockRes = abci.ResponseEndBlock{
		Events:                []abci.Event{},
		ConsensusParamUpdates: &abci.ConsensusParams{},
		ValidatorUpdates:      []abci.ValidatorUpdate{},
	}
)

// change this to write to in-memory io.Writer (e.g. bytes.Buffer)
func TestStreamingService(t *testing.T) {
	testingCtx = sdk.NewContext(nil, types1.Header{}, false, log.TestingLogger())
	testStreamingService = NewTraceStreamingService(true, testMarshaller)
	require.NotNil(t, testStreamingService)
	require.IsType(t, &TraceStreamingService{}, testStreamingService)
	testListenBeginBlocker(t)
	testListenEndBlocker(t)
}

func testListenBeginBlocker(t *testing.T) {
	sbb := testStreamingService.StreamBeginBlocker
	assert.NotPanicsf(t, func() { sbb(testingCtx, testBeginBlockReq, testBeginBlockRes) }, "StreamBeginBlocker did not panic")
}

func testListenEndBlocker(t *testing.T) {
	seb := testStreamingService.StreamEndBlocker
	assert.NotPanicsf(t, func() { seb(testingCtx, testEndBlockReq, testEndBlockRes) }, "StreamEndBlocker did not panic")
}
