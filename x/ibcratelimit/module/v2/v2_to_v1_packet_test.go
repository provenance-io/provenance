package v2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
)

func sampleData() transfertypes.FungibleTokenPacketData {
	return transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   "cosmos1sender",
		Receiver: "cosmos1receiver",
		Memo:     "test memo",
	}
}

func makePayload(t *testing.T, data transfertypes.FungibleTokenPacketData, encoding string) channeltypesv2.Payload {
	t.Helper()
	bz, err := transfertypes.MarshalPacketData(data, transfertypes.V1, encoding)
	require.NoError(t, err, "MarshalPacketData")
	return channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        encoding,
		Value:           bz,
	}
}

func TestV2ToV1Packet_JSONEncoding(t *testing.T) {
	data := sampleData()
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	packet, err := v2ToV1Packet(payload, "client-src", "client-dst", 42)
	require.NoError(t, err)

	assert.Equal(t, uint64(42), packet.Sequence, "Sequence")
	assert.Equal(t, "transfer", packet.SourcePort, "SourcePort")
	assert.Equal(t, "client-src", packet.SourceChannel, "SourceChannel")
	assert.Equal(t, "transfer", packet.DestinationPort, "DestinationPort")
	assert.Equal(t, "client-dst", packet.DestinationChannel, "DestinationChannel")

	var got transfertypes.FungibleTokenPacketData
	require.NoError(t, json.Unmarshal(packet.Data, &got), "Unmarshal packet data")
	assert.Equal(t, data, got, "packet data content")
}

func TestV2ToV1Packet_ProtobufEncoding(t *testing.T) {
	data := sampleData()
	payload := makePayload(t, data, transfertypes.EncodingProtobuf)

	packet, err := v2ToV1Packet(payload, "client-src", "client-dst", 7)
	require.NoError(t, err)

	assert.Equal(t, uint64(7), packet.Sequence)
	assert.Equal(t, "client-src", packet.SourceChannel)
	assert.Equal(t, "client-dst", packet.DestinationChannel)

	var got transfertypes.FungibleTokenPacketData
	require.NoError(t, json.Unmarshal(packet.Data, &got))
	assert.Equal(t, data, got)
}

func TestV2ToV1Packet_ABIEncoding(t *testing.T) {
	data := sampleData()
	payload := makePayload(t, data, transfertypes.EncodingABI)

	packet, err := v2ToV1Packet(payload, "src", "dst", 1)
	require.NoError(t, err)

	var got transfertypes.FungibleTokenPacketData
	require.NoError(t, json.Unmarshal(packet.Data, &got))
	assert.Equal(t, data, got)
}

func TestV2ToV1Packet_NilValue(t *testing.T) {
	payload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           nil,
	}
	_, err := v2ToV1Packet(payload, "src", "dst", 1)
	assert.Error(t, err, "should error on nil value")
}

func TestV2ToV1Packet_EmptyValue(t *testing.T) {
	payload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte{},
	}
	_, err := v2ToV1Packet(payload, "src", "dst", 1)
	assert.Error(t, err, "should error on empty value")
}

func TestV2ToV1Packet_InvalidValue(t *testing.T) {
	payload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte("not valid json"),
	}
	_, err := v2ToV1Packet(payload, "src", "dst", 1)
	assert.Error(t, err, "should error on invalid value")
}
