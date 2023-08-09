package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/oracle/keeper"
	"github.com/provenance-io/provenance/x/oracle/types"
)

func (s *KeeperTestSuite) TestUpdateOracle() {
	authority := s.app.OracleKeeper.GetAuthority()
	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
	)

	tests := []struct {
		name  string
		req   *types.MsgUpdateOracleRequest
		res   *types.MsgUpdateOracleResponse
		event *sdk.Event
		err   string
	}{
		{
			name: "failure - authority does not match module authority",
			req: &types.MsgUpdateOracleRequest{
				Address:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Authority: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			res: nil,
			err: fmt.Sprintf("expected authority %s got cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma: unauthorized", authority),
		},
		{
			name: "success - oracle is updated",
			req: &types.MsgUpdateOracleRequest{
				Address:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Authority: authority,
			},
			res:   &types.MsgUpdateOracleResponse{},
			event: &event,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			res, err := s.msgServer.UpdateOracle(s.ctx, tc.req)

			if tc.event != nil {
				s.Assert().Equal(1, len(s.ctx.EventManager().Events()), "should emit the correct number of events")
				s.Assert().Equal(*tc.event, s.ctx.EventManager().Events()[0], "should emit the correct event")
			} else {
				s.Assert().Equal(0, len(s.ctx.EventManager().Events()), "should not emit events")
			}

			if len(tc.err) > 0 {
				s.Assert().Nil(res, "should have nil response")
				s.Assert().EqualError(err, tc.err, "should have correct error")
			} else {
				s.Assert().NoError(err, "should not have error")
				s.Assert().Equal(tc.res, res, "should have the correct response")
			}
		})
	}
}

func (s *KeeperTestSuite) TestSendQueryOracle() {
	event := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
	)

	s.app.OracleKeeper.SetMockICS4Wrapper(keeper.MockICS4Wrapper{})
	s.app.OracleKeeper.SetMockScopedKeeper(keeper.MockScopedKeeper{})

	tests := []struct {
		name        string
		req         *types.MsgSendQueryOracleRequest
		res         *types.MsgSendQueryOracleResponse
		event       *sdk.Event
		err         string
		mockChannel bool
	}{
		{
			name: "failure - a packet should not be sent on invalid channel",
			req: &types.MsgSendQueryOracleRequest{
				Query:     []byte("{}"),
				Channel:   "invalid-channel",
				Authority: "authority",
			},
			res: nil,
			err: "port ID (oracle) channel ID (invalid-channel): channel not found",
		},
		{
			name: "success - a packet should be sent",
			req: &types.MsgSendQueryOracleRequest{
				Query:     []byte("{}"),
				Channel:   "channel-1",
				Authority: "authority",
			},
			res: &types.MsgSendQueryOracleResponse{
				Sequence: 1,
			},
			event:       &event,
			mockChannel: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.mockChannel {
				s.app.OracleKeeper.SetMockChannelKeeper(&keeper.MockChannelKeeper{})
			}
			res, err := s.msgServer.SendQueryOracle(s.ctx, tc.req)

			if tc.event != nil {
				s.Assert().Equal(1, len(s.ctx.EventManager().Events()), "should emit the correct number of events")
				s.Assert().Equal(*tc.event, s.ctx.EventManager().Events()[0], "should emit the correct event")
			} else {
				s.Assert().Equal(0, len(s.ctx.EventManager().Events()), "should not emit events")
			}

			if len(tc.err) > 0 {
				s.Assert().Nil(res, "should have nil response")
				s.Assert().EqualError(err, tc.err, "should have correct error")
			} else {
				s.Assert().NoError(err, "should not have error")
				s.Assert().Equal(tc.res, res, "should have the correct response")
			}
		})
	}
}
