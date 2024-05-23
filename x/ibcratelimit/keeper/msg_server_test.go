package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

func (s *TestSuite) TestUpdateParams() {
	authority := s.app.OracleKeeper.GetAuthority()

	tests := []struct {
		name  string
		req   *ibcratelimit.MsgUpdateParamsRequest
		res   *ibcratelimit.MsgUpdateParamsResponse
		event *sdk.Event
		err   string
	}{
		{
			name: "failure - authority does not match module authority",
			req: &ibcratelimit.MsgUpdateParamsRequest{
				Params:    ibcratelimit.NewParams("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				Authority: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			res: nil,
			err: fmt.Sprintf("expected \"%s\" got \"cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma\": expected gov account as only signer for proposal message", authority),
		},
		{
			name: "success - rate limiter is updated",
			req: &ibcratelimit.MsgUpdateParamsRequest{
				Params:    ibcratelimit.NewParams("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
				Authority: authority,
			},
			res:   &ibcratelimit.MsgUpdateParamsResponse{},
			event: typedEventToEvent(ibcratelimit.NewEventParamsUpdated()),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			res, err := s.msgServer.UpdateParams(s.ctx, tc.req)
			events := s.ctx.EventManager().Events()
			numEvents := len(events)

			if tc.event != nil {
				s.Assert().Equal(1, numEvents, "should emit the correct number of events")
				s.Assert().Equal(*tc.event, events[0], "should emit the correct event")
			} else {
				s.Assert().Empty(events, "should not emit events")
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

func typedEventToEvent(tev proto.Message) *sdk.Event {
	event, _ := sdk.TypedEventToEvent(tev)
	return &event
}
