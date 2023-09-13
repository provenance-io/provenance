package types

import (
	"encoding/json"
	"testing"

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
	sudoMsg, err := json.Marshal(ibcLifecycleCompleteAck)
	s.Require().NoError(err)
	s.Require().Equal(`{"ibc_lifecycle_complete":{"ibc_ack":{"channel":"channel-1","sequence":100,"ack":{"packet_sequence":100,"source_channel":"channel"},"success":true}}}`, string(sudoMsg))
}

func (s *IbcHooksTypesTestSuite) TestIbcLifecycleCompleteTimeoutJsonSerialization() {
	ibcLifecycleCompleteAck := NewIbcLifecycleCompleteTimeout("channel-1", 100)
	sudoMsg, err := json.Marshal(ibcLifecycleCompleteAck)
	s.Require().NoError(err)
	s.Require().Equal(`{"ibc_lifecycle_complete":{"ibc_timeout":{"channel":"channel-1","sequence":100}}}`, string(sudoMsg))
}
