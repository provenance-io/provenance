package keeper_test

import (
	"github.com/provenance-io/provenance/x/oracle/keeper"
)

func (s *KeeperTestSuite) TestQueryOracle() {
	tests := []struct {
		name        string
		query       []byte
		channel     string
		sequence    uint64
		err         string
		setMocks    bool
		setIcs4Mock bool
	}{
		{
			name:     "failure - missing channel capability",
			query:    []byte("{}"),
			channel:  "invalid",
			sequence: 0,
			err:      "module does not own channel capability: channel capability not found",
		},
		{
			name:     "failure - unable to send",
			query:    []byte("{}"),
			channel:  "channel-1",
			sequence: 0,
			err:      "channel-1: channel not found",
			setMocks: true,
		},
		{
			name:        "success - should send a packet",
			query:       []byte("{}"),
			channel:     "channel-1",
			sequence:    2,
			setIcs4Mock: true,
		},
		{
			name:     "success - should send a packet with the next sequence",
			query:    []byte("{}"),
			channel:  "channel-1",
			sequence: 3,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setMocks {
				s.app.OracleKeeper.SetMockScopedKeeper(keeper.MockScopedKeeper{})
				s.app.OracleKeeper.SetMockChannelKeeper(&keeper.MockChannelKeeper{})
			}
			if tc.setIcs4Mock {
				s.app.OracleKeeper.SetMockICS4Wrapper(keeper.MockICS4Wrapper{})
			}
			sequence, err := s.app.OracleKeeper.QueryOracle(s.ctx, tc.query, tc.channel)
			s.Assert().Equal(int(tc.sequence), int(sequence), "should have correct sequence")
			if len(tc.err) > 0 {
				s.Assert().EqualError(err, tc.err, "should have the correct error")

			} else {
				s.Assert().Nil(err, "should have nil error")
				s.Assert().Equal(int(tc.sequence), int(sequence), "should have correct sequence")
			}

		})
	}
}
