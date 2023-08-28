package keeper_test

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/provenance-io/provenance/x/oracle/keeper"
	"github.com/provenance-io/provenance/x/oracle/types"
	icqtypes "github.com/strangelove-ventures/async-icq/v6/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (s *KeeperTestSuite) TestSendQuery() {
	tests := []struct {
		name        string
		err         string
		sequence    uint64
		req         []abci.RequestQuery
		enableMocks bool
	}{
		{
			name:     "failure - invalid channel",
			err:      "port ID (port) channel ID (channel): channel not found",
			sequence: 0,
			req:      nil,
		},
		{
			name:        "success - valid send query",
			sequence:    1,
			req:         []abci.RequestQuery{{Data: []byte("{}")}},
			enableMocks: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.enableMocks {
				s.app.OracleKeeper = s.app.OracleKeeper.WithMockChannelKeeper(&keeper.MockChannelKeeper{})
				s.app.OracleKeeper = s.app.OracleKeeper.WithMockICS4Wrapper(&keeper.MockICS4Wrapper{})
			}
			sequence, err := s.app.OracleKeeper.SendQuery(s.ctx, "port", "channel", nil, tc.req, clienttypes.ZeroHeight(), 0)
			if len(tc.err) > 0 {
				s.Assert().Equal(int(tc.sequence), int(sequence), "should have correct sequence number")
				s.Assert().EqualError(err, tc.err, "should have correct error")
			} else {
				s.Assert().Equal(int(tc.sequence), int(sequence), "should have correct sequence number")
				s.Assert().Nil(err, "should have no error")
			}
		})
	}
}

func (s *KeeperTestSuite) TestOnAcknowledgementPacket() {
	wasmError := sdkerrors.New("codespace", 2, "jackthecat ran away")
	_, code, _ := sdkerrors.ABCIInfo(wasmError, false)

	tests := []struct {
		name   string
		ack    channeltypes.Acknowledgement
		packet channeltypes.Packet
		event  proto.Message
		err    string
	}{
		{
			name:   "success - error event is emitted on ack error",
			ack:    channeltypes.NewErrorAcknowledgement(wasmError),
			packet: channeltypes.Packet{Sequence: 5, DestinationChannel: "oracle-channel"},
			event: &types.EventOracleQueryError{
				SequenceId: strconv.FormatUint(5, 10),
				Error:      fmt.Sprintf("ABCI code: %d: %s", code, "error handling packet: see events for details"),
				Channel:    "oracle-channel",
			},
		},
		{
			name:   "success - success event is emitted on ack",
			ack:    channeltypes.NewResultAcknowledgement(createICQResponse(s.app.AppCodec(), "{}")),
			packet: channeltypes.Packet{Sequence: 5, DestinationChannel: "oracle-channel"},
			event: &types.EventOracleQuerySuccess{
				SequenceId: strconv.FormatUint(5, 10),
				Result:     "{\"data\":\"CgY6BAoCe30=\"}",
				Channel:    "oracle-channel",
			},
		},
		{
			name:   "failure - invalid icq packet ack in result ack",
			ack:    channeltypes.NewResultAcknowledgement([]byte("baddata")),
			packet: channeltypes.Packet{Sequence: 5},
			event:  nil,
			err:    "failed to unmarshal interchain query packet ack: invalid character 'b' looking for beginning of value",
		},
		{
			name:   "failure - invalid cosmos response in icq packet ack",
			ack:    channeltypes.NewResultAcknowledgement(createInvalidICQPacketAck()),
			packet: channeltypes.Packet{Sequence: 5},
			event:  nil,
			err:    "could not deserialize data to cosmos response: unexpected EOF",
		},
		{
			name:   "failure - empty cosmos response in icq packet ack",
			ack:    channeltypes.NewResultAcknowledgement(createEmptyICQPacketAck()),
			packet: channeltypes.Packet{Sequence: 5},
			event:  nil,
			err:    "no responses in interchain query packet ack: invalid request",
		},
		{
			name:   "failure - invalid query response in cosmos response",
			ack:    channeltypes.NewResultAcknowledgement(createInvalidCosmosResponse()),
			packet: channeltypes.Packet{Sequence: 5},
			event:  nil,
			err:    "failed to unmarshal interchain query response to type *types.Acknowledgement_Result: unexpected EOF",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithEventManager(sdktypes.NewEventManager())
			err := s.app.OracleKeeper.OnAcknowledgementPacket(s.ctx, tc.packet, tc.ack)

			if len(tc.err) > 0 {
				s.Assert().EqualError(err, tc.err, "should return expected error")
			} else {
				s.Assert().NoError(err, "should not return error")
				event, _ := sdktypes.TypedEventToEvent(tc.event)
				events := s.ctx.EventManager().Events()
				s.Assert().Equal(event, events[0], "should emit correct event")
			}
		})
	}
}

func (s *KeeperTestSuite) TestOnTimeoutPacket() {
	packet := channeltypes.Packet{Sequence: 5, DestinationChannel: "oracle-channel"}
	err := s.app.OracleKeeper.OnTimeoutPacket(s.ctx, packet)
	event, _ := sdktypes.TypedEventToEvent(&types.EventOracleQueryTimeout{
		SequenceId: strconv.FormatUint(5, 10),
		Channel:    "oracle-channel",
	})
	s.Assert().NoError(err, "should not throw an error")
	emitted := s.ctx.EventManager().Events()
	s.Assert().Equal(event, emitted[0], "timeout event should be emitted")
}

func createICQResponse(cdc codec.Codec, response string) []byte {
	oracleResponse := types.QueryOracleResponse{
		Data: []byte("{}"),
	}
	value, _ := cdc.Marshal(&oracleResponse)
	bytes, _ := icqtypes.SerializeCosmosResponse([]abci.ResponseQuery{{
		Value: value,
	}})

	icqPacket := icqtypes.InterchainQueryPacketAck{
		Data: bytes,
	}
	icqBytes, _ := icqtypes.ModuleCdc.MarshalJSON(&icqPacket)
	return icqBytes
}

func createInvalidICQPacketAck() []byte {
	icqPacket := icqtypes.InterchainQueryPacketAck{
		Data: []byte("abc"),
	}
	icqBytes, _ := icqtypes.ModuleCdc.MarshalJSON(&icqPacket)
	return icqBytes
}

func createEmptyICQPacketAck() []byte {
	bytes, _ := icqtypes.SerializeCosmosResponse([]abci.ResponseQuery{})

	icqPacket := icqtypes.InterchainQueryPacketAck{
		Data: bytes,
	}

	icqBytes, _ := icqtypes.ModuleCdc.MarshalJSON(&icqPacket)
	return icqBytes
}

func createInvalidCosmosResponse() []byte {
	bytes, _ := icqtypes.SerializeCosmosResponse([]abci.ResponseQuery{{
		Value: []byte("baddata"),
	}})

	icqPacket := icqtypes.InterchainQueryPacketAck{
		Data: bytes,
	}
	icqBytes, _ := icqtypes.ModuleCdc.MarshalJSON(&icqPacket)
	return icqBytes
}
