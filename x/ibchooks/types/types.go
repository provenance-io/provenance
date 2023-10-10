package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
)

// Async: The following types represent the response sent by a contract on OnRecvPacket when it wants the ack to be async

// OnRecvPacketAsyncAckResponse the response a contract sends to instruct the module to make the ack async
type OnRecvPacketAsyncAckResponse struct {
	IsAsyncAck bool `json:"is_async_ack"`
}

// Async The following types are used to ask a contract that has sent a packet to generate an ack for it

// RequestAckI internals of IBCAsync
type RequestAckI struct {
	PacketSequence uint64 `json:"packet_sequence"`
	SourceChannel  string `json:"source_channel"`
}

// RequestAck internals of IBCAsync
type RequestAck struct {
	RequestAckI `json:"request_ack"`
}

// IBCAsync is the sudo message to be sent to the contract for it to generate  an ack for a sent packet
type IBCAsync struct {
	RequestAck `json:"ibc_async"`
}

// General

// ContractAck is the response to be stored when a wasm hook is executed
type ContractAck struct {
	ContractResult []byte `json:"contract_result"`
	IbcAck         []byte `json:"ibc_ack"`
}

// IBCAckResponse is the response that a contract returns from the sudo() call on OnRecvPacket or RequestAck
type IBCAckResponse struct {
	Packet      channeltypes.Packet `json:"packet"`
	ContractAck ContractAck         `json:"contract_ack"`
}

// IBCAckError is the error that a contract returns from the sudo() call on RequestAck
type IBCAckError struct {
	Packet           channeltypes.Packet `json:"packet"`
	ErrorDescription string              `json:"error_description"`
	ErrorResponse    string              `json:"error_response"`
}

// IBCAck is the parent IBC ack response structure
type IBCAck struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
	// Note: These two fields have to be pointers so that they can be null
	// If they are not pointers, they will be empty structs when null,
	// which will cause issues with json.Unmarshal.
	AckResponse *IBCAckResponse `json:"response,omitempty"`
	AckError    *IBCAckError    `json:"error,omitempty"`
}

// UnmarshalIBCAck unmashals Ack to either response or error type
func UnmarshalIBCAck(bz []byte) (*IBCAck, error) {
	var ack IBCAck
	if err := json.Unmarshal(bz, &ack); err != nil {
		return nil, err
	}

	switch ack.Type {
	case "ack_response":
		ack.AckResponse = &IBCAckResponse{}
		if err := json.Unmarshal(ack.Content, ack.AckResponse); err != nil {
			return nil, err
		}
	case "ack_error":
		ack.AckError = &IBCAckError{}
		if err := json.Unmarshal(ack.Content, ack.AckError); err != nil {
			return nil, err
		}
	}

	return &ack, nil
}

// IbcAck ibc ack struct with json fields defined
type IbcAck struct {
	Channel  string    `json:"channel"`
	Sequence uint64    `json:"sequence"`
	Ack      JSONBytes `json:"ack"`
	Success  bool      `json:"success"`
}

// IbcLifecycleCompleteAck ibc lifcycle complete ack with json fields defined
type IbcLifecycleCompleteAck struct {
	IbcAck IbcAck `json:"ibc_ack"`
}

// IbcTimeout ibc timeout struct with json fields defined
type IbcTimeout struct {
	Channel  string `json:"channel"`
	Sequence uint64 `json:"sequence"`
}

// IbcLifecycleCompleteTimeout ibc lifecycle complete struct with json fields defined
type IbcLifecycleCompleteTimeout struct {
	IbcTimeout IbcTimeout `json:"ibc_timeout"`
}

// IbcLifecycleComplete ibc lifecycle complete struct with json fields defined
type IbcLifecycleComplete struct {
	IbcLifecycleComplete interface{} `json:"ibc_lifecycle_complete"`
}

// MarkerMemo parent marker struct for memo json
type MarkerMemo struct {
	Marker MarkerPayload `json:"marker"`
}

// MarkerPayload child structure for marker memo
type MarkerPayload struct {
	TransferAuths      []string `json:"transfer-auths"`
	AllowForceTransfer bool     `json:"allow-force-transfer"`
}

// NewMarkerPayload returns a marker payload with transfer authorities and allow force transfer flag
func NewMarkerPayload(transferAuthAddrs []sdk.AccAddress, allowForceTransfer bool) MarkerPayload {
	addresses := make([]string, len(transferAuthAddrs))
	for i := 0; i < len(transferAuthAddrs); i++ {
		addresses[i] = transferAuthAddrs[i].String()
	}
	return MarkerPayload{
		TransferAuths:      addresses,
		AllowForceTransfer: allowForceTransfer,
	}
}

// NewIbcLifecycleCompleteAck returns a new ibc lifecycle complete acknowledgment object for json serialization
func NewIbcLifecycleCompleteAck(sourceChannel string, sequence uint64, ackAsJSON []byte, success bool) IbcLifecycleComplete {
	ibcLifecycleCompleteAck := IbcLifecycleCompleteAck{
		IbcAck: IbcAck{
			Channel:  sourceChannel,
			Sequence: sequence,
			Ack:      ackAsJSON,
			Success:  success,
		},
	}
	return IbcLifecycleComplete{IbcLifecycleComplete: ibcLifecycleCompleteAck}
}

// NewIbcLifecycleCompleteTimeout return a new ibc lifecycle complete timeout object for json serialization
func NewIbcLifecycleCompleteTimeout(sourceChannel string, sequence uint64) IbcLifecycleComplete {
	ibcLifecycleCompleteAck := IbcLifecycleCompleteTimeout{
		IbcTimeout: IbcTimeout{
			Channel:  sourceChannel,
			Sequence: sequence,
		},
	}
	return IbcLifecycleComplete{IbcLifecycleComplete: ibcLifecycleCompleteAck}
}

// JSONBytes is a byte array of a json string
type JSONBytes []byte

// MarshalJSON returns empty json object bytes when bytes are empty
func (jb JSONBytes) MarshalJSON() ([]byte, error) {
	if len(jb) == 0 {
		return []byte("{}"), nil
	}
	return jb, nil
}

// PreSendPacketDataProcessingFn is function signature used for custom data processing before ibc's PacketSend executed in middleware
type PreSendPacketDataProcessingFn func(ctx sdk.Context, data []byte, processData map[string]interface{}) ([]byte, error)
