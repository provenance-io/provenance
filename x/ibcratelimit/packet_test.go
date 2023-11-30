package ibcratelimit_test

import (
	"encoding/json"
	"testing"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/stretchr/testify/assert"
)

func TestValidateReceiverAddress(t *testing.T) {
	tests := []struct {
		name     string
		packetFn func() []byte
		err      string
	}{
		{
			name: "success - packet is valid",
			packetFn: func() []byte {
				data := NewMockFungiblePacketData(false)
				bytes, _ := json.Marshal(data)
				return bytes
			},
		},
		{
			name: "failure - long receiver name",
			packetFn: func() []byte {
				data := NewMockFungiblePacketData(true)
				bytes, _ := json.Marshal(data)
				return bytes
			},
			err: "IBC Receiver address too long. Max supported length is 4096: invalid address",
		},
		{
			name: "failure - invalid packet type",
			packetFn: func() []byte {
				return []byte("garbage")
			},
			err: "invalid character 'g' looking for beginning of value",
		},
		{
			name: "failure - nil data",
			packetFn: func() []byte {
				return nil
			},
			err: "unexpected end of JSON input",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ibcratelimit.ValidateReceiverAddress(NewMockPacket(tc.packetFn(), false))
			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err, "should return correct error when invalid")
			} else {
				assert.NoError(t, err, "should not return an error when valid")
			}
		})
	}
}

func TestUnwrapPacket(t *testing.T) {
	tests := []struct {
		name        string
		packetFn    func() []byte
		validHeight bool
		err         string
		expected    ibcratelimit.UnwrappedPacket
	}{
		{
			name: "success - packet data and height are valid",
			packetFn: func() []byte {
				data := NewMockFungiblePacketData(false)
				bytes, _ := json.Marshal(data)
				return bytes
			},
			expected: ibcratelimit.UnwrappedPacket{
				Sequence:           1,
				SourcePort:         "src-port",
				SourceChannel:      "src-channel",
				DestinationPort:    "dest-port",
				DestinationChannel: "dest-channel",
				TimeoutHeight: clienttypes.Height{
					RevisionNumber: 5,
					RevisionHeight: 5,
				},
				TimeoutTimestamp: 1,
				Data:             NewMockFungiblePacketData(false),
			},
			validHeight: true,
		},
		{
			name: "failure - height is invalid",
			packetFn: func() []byte {
				data := NewMockFungiblePacketData(false)
				bytes, _ := json.Marshal(data)
				return bytes
			},
			err: "bad message",
		},
		{
			name: "failure - invalid packet data",
			packetFn: func() []byte {
				return []byte("garbage")
			},
			err: "invalid character 'g' looking for beginning of value",
		},
		{
			name: "failure - nil data",
			packetFn: func() []byte {
				return nil
			},
			err: "unexpected end of JSON input",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unwrapped, err := ibcratelimit.UnwrapPacket(NewMockPacket(tc.packetFn(), tc.validHeight))
			if len(tc.err) > 0 {
				assert.Equal(t, tc.expected, unwrapped, "should return an empty unwrapped packet on failure")
				assert.EqualError(t, err, tc.err, "should return correct error when invalid")
			} else {
				assert.Equal(t, tc.expected, unwrapped, "should return an unwrapped packet with correct data")
				assert.NoError(t, err, "should not return an error when valid")
			}
		})
	}
}
