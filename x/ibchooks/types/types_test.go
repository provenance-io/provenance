package types

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type IbcHooksTypesTestSuite struct {
	suite.Suite
}

func (s *IbcHooksTypesTestSuite) SetupTest() {}

func TestSITestSuite(t *testing.T) {
	suite.Run(t, new(IbcHooksTypesTestSuite))
}

func (s *IbcHooksTypesTestSuite) TestIbcLifecycleCompleteAckJsonSerialization() {
	ack := RequestAckI{SourceChannel: "channel", PacketSequence: 100}
	ackBz, err := json.Marshal(ack)
	s.Require().NoError(err)
	ibcLifecycleCompleteAck := NewIbcLifecycleCompleteAck("channel-1", 100, ackBz, true)
	actualAck, err := json.Marshal(ibcLifecycleCompleteAck)
	s.Require().NoError(err, "Marshal() error")
	s.Require().Equal(`{"ibc_lifecycle_complete":{"ibc_ack":{"channel":"channel-1","sequence":100,"ack":{"packet_sequence":100,"source_channel":"channel"},"success":true}}}`, string(actualAck), "Serialized json doesn't match")
}

func (s *IbcHooksTypesTestSuite) TestIbcLifecycleCompleteTimeoutJsonSerialization() {
	ibcLifecycleCompleteAck := NewIbcLifecycleCompleteTimeout("channel-1", 100)
	actualAck, err := json.Marshal(ibcLifecycleCompleteAck)
	s.Require().NoError(err, "Marshal() error")
	s.Require().Equal(`{"ibc_lifecycle_complete":{"ibc_timeout":{"channel":"channel-1","sequence":100}}}`, string(actualAck), "Serialized json doesn't match")
}

func (s *IbcHooksTypesTestSuite) TestNewMarkerPayloadSerialization() {
	testCases := []struct {
		name    string
		addrs   []sdk.AccAddress
		expJson string
	}{
		{
			name:    "empty address array",
			addrs:   []sdk.AccAddress{},
			expJson: `{"transfer-auths":[]}`,
		},
		{
			name:    "single address array",
			addrs:   []sdk.AccAddress{sdk.AccAddress("address1")},
			expJson: `{"transfer-auths":["cosmos1v9jxgun9wdenzc33zgq"]}`,
		},
		{
			name:    "multiple address array",
			addrs:   []sdk.AccAddress{sdk.AccAddress("address1"), sdk.AccAddress("address2")},
			expJson: `{"transfer-auths":["cosmos1v9jxgun9wdenzc33zgq","cosmos1v9jxgun9wdenyy7j85h"]}`,
		},
	}
	for _, tc := range testCases {
		markerPayload := NewMarkerPayload(tc.addrs)
		s.T().Run(tc.name, func(t *testing.T) {
			actualJson, err := json.Marshal(markerPayload)
			s.Assert().NoError(err, "Marshal() error")
			s.Assert().Equal(tc.expJson, string(actualJson), "Serialized json doesn't match")
		})
	}
}
