package keeper_test

import (
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

func (s *TestSuite) TestCheckAndUpdateRateLimits() {
	tests := []struct {
		name       string
		contract   string
		msgType    string
		packet     exported.PacketI
		err        string
		mockKeeper *MockPermissionedKeeper
	}{
		{
			name:       "success - rate limit checked and updated on send",
			contract:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			msgType:    ibcratelimit.MsgSendPacket,
			packet:     NewMockPacket(NewMockSerializedPacketData(), true),
			mockKeeper: NewMockPermissionedKeeper(true),
		},
		{
			name:       "success - rate limit checked and updated on recv",
			contract:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			msgType:    ibcratelimit.MsgRecvPacket,
			packet:     NewMockPacket(NewMockSerializedPacketData(), true),
			mockKeeper: NewMockPermissionedKeeper(true),
		},
		{
			name:     "failure - an invalid contract throws error",
			contract: "",
			msgType:  ibcratelimit.MsgSendPacket,
			packet:   nil,
			err:      "empty address string is not allowed: contract error",
		},
		{
			name:     "failure - throws error on bad packet",
			contract: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			msgType:  ibcratelimit.MsgSendPacket,
			packet:   NewMockPacket(NewMockSerializedPacketData(), false),
			err:      "bad message: contract error",
		},
		{
			name:     "failure - throws error on invalid message type",
			contract: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			msgType:  "bad message type",
			packet:   NewMockPacket(NewMockSerializedPacketData(), true),
			err:      "bad message: contract error",
		},
		{
			name:     "failure - throws error on bad packet data",
			contract: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			msgType:  ibcratelimit.MsgSendPacket,
			packet:   NewMockPacket([]byte("badpacketdata"), true),
			err:      "invalid character 'b' looking for beginning of value: contract error",
		},
		{
			name:       "failure - throws error on nil packet",
			contract:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			msgType:    ibcratelimit.MsgRecvPacket,
			packet:     nil,
			mockKeeper: NewMockPermissionedKeeper(true),
			err:        "bad message: contract error",
		},
		{
			name:       "failure - throws error on bad contract operation",
			contract:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			msgType:    ibcratelimit.MsgSendPacket,
			packet:     NewMockPacket(NewMockSerializedPacketData(), true),
			mockKeeper: NewMockPermissionedKeeper(false),
			err:        "rate limit exceeded: rate limit exceeded",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			permissionedKeeper := s.app.RateLimitingKeeper.PermissionedKeeper
			s.app.RateLimitingKeeper.SetParams(s.ctx, ibcratelimit.NewParams(tc.contract))
			if tc.mockKeeper != nil {
				s.app.RateLimitingKeeper.PermissionedKeeper = tc.mockKeeper
			}
			err := s.app.RateLimitingKeeper.CheckAndUpdateRateLimits(s.ctx, tc.msgType, tc.packet)
			if len(tc.err) > 0 {
				s.Assert().EqualError(err, tc.err, "should return the correct error")
			} else {
				s.Assert().NoError(err)
			}

			s.app.RateLimitingKeeper.PermissionedKeeper = permissionedKeeper
		})
	}
}

func (s *TestSuite) TestUndoSendRateLimit() {
	tests := []struct {
		name       string
		contract   string
		packet     exported.PacketI
		err        string
		mockKeeper *MockPermissionedKeeper
	}{
		{
			name:       "success - undo rate limit",
			contract:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			packet:     NewMockPacket(NewMockSerializedPacketData(), true),
			mockKeeper: NewMockPermissionedKeeper(true),
		},
		{
			name:     "failure - an invalid contract throws error",
			contract: "",
			packet:   NewMockPacket(NewMockSerializedPacketData(), false),
			err:      "empty address string is not allowed",
		},
		{
			name:     "failure - throws error on bad packet",
			contract: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			packet:   NewMockPacket(NewMockSerializedPacketData(), false),
			err:      "bad message",
		},
		{
			name:     "failure - throws error on nil packet",
			contract: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			packet:   NewMockPacket(NewMockSerializedPacketData(), false),
			err:      "bad message",
		},
		{
			name:       "failure - throws error on bad contract operation",
			contract:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			packet:     NewMockPacket(NewMockSerializedPacketData(), true),
			mockKeeper: NewMockPermissionedKeeper(false),
			err:        "rate limit exceeded: contract error",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			permissionedKeeper := s.app.RateLimitingKeeper.PermissionedKeeper
			if tc.mockKeeper != nil {
				s.app.RateLimitingKeeper.PermissionedKeeper = tc.mockKeeper
			}

			err := s.app.RateLimitingKeeper.UndoSendRateLimit(s.ctx, tc.contract, tc.packet)
			if len(tc.err) > 0 {
				s.Assert().EqualError(err, tc.err, "should return the correct error")
			} else {
				s.Assert().NoError(err)
			}

			s.app.RateLimitingKeeper.PermissionedKeeper = permissionedKeeper
		})
	}
}

func (s *TestSuite) TestRevertSentPacket() {
	tests := []struct {
		name string
	}{}

	for _, tc := range tests {
		s.Run(tc.name, func() {

		})
	}
}
