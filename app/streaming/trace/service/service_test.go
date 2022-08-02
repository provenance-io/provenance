package service

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	types1 "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TraceServiceTestSuite struct {
	suite.Suite

	ctx sdk.Context

	testStreamingService *TraceStreamingService

	testBeginBlockReq abci.RequestBeginBlock
	testBeginBlockRes abci.ResponseBeginBlock
	testEndBlockReq   abci.RequestEndBlock
	testEndBlockRes   abci.ResponseEndBlock
}

func (s *TraceServiceTestSuite) SetupTest() {
	s.ctx = sdk.NewContext(nil, types1.Header{}, false, log.TestingLogger())
	marshaller := codec.NewProtoCodec(codecTypes.NewInterfaceRegistry())
	s.testStreamingService = NewTraceStreamingService(true, marshaller)

	// test abci message types
	s.testBeginBlockReq = abci.RequestBeginBlock{
		Header: types1.Header{
			Height: 1,
		},
		ByzantineValidators: []abci.Evidence{},
		Hash:                []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
		LastCommitInfo: abci.LastCommitInfo{
			Round: 1,
			Votes: []abci.VoteInfo{},
		},
	}
	s.testBeginBlockRes = abci.ResponseBeginBlock{
		Events: []abci.Event{
			{
				Type: "testEventType1",
			},
			{
				Type: "testEventType2",
			},
		},
	}
	s.testEndBlockReq = abci.RequestEndBlock{
		Height: 1,
	}
	s.testEndBlockRes = abci.ResponseEndBlock{
		Events:                []abci.Event{},
		ConsensusParamUpdates: &abci.ConsensusParams{},
		ValidatorUpdates:      []abci.ValidatorUpdate{},
	}
}

func TestTraceServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TraceServiceTestSuite))
}

func (s *TraceServiceTestSuite) TestListenBeginBlocker() {
	sbb := s.testStreamingService.StreamBeginBlocker
	assert.NotPanicsf(s.T(), func() { sbb(s.ctx, s.testBeginBlockReq, s.testBeginBlockRes) }, "StreamBeginBlocker did not panic")
}

func (s *TraceServiceTestSuite) TestListenEndBlocker() {
	seb := s.testStreamingService.StreamEndBlocker
	assert.NotPanicsf(s.T(), func() { seb(s.ctx, s.testEndBlockReq, s.testEndBlockRes) }, "StreamEndBlocker did not panic")
}
