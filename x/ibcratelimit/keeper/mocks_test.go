package keeper_test

import (
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// MockPacket is a test struct that implements the PacketI interface.
type MockPacket struct {
	data        []byte
	validHeight bool
}

// NewMockPacket creates a new MockPacket.
func NewMockPacket(data []byte, validHeight bool) *MockPacket {
	return &MockPacket{
		data:        data,
		validHeight: validHeight,
	}
}

// GetSequence implements the PacketI interface and always returns 1.
func (m MockPacket) GetSequence() uint64 {
	return 1
}

// GetTimeoutHeight implements the PacketI interface and can return a valid or invalid height.
func (m MockPacket) GetTimeoutHeight() exported.Height {
	if !m.validHeight {
		return nil
	}
	return clienttypes.Height{
		RevisionNumber: 5,
		RevisionHeight: 5,
	}
}

// GetTimeoutTimestamp implements the PacketI interface and always returns 1.
func (m MockPacket) GetTimeoutTimestamp() uint64 {
	return 1
}

// GetSourcePort implements the PacketI interface and always returns "src-port".
func (m MockPacket) GetSourcePort() string {
	return "src-port"
}

// GetSourceChannel implements the PacketI interface and always returns "src-channel".
func (m MockPacket) GetSourceChannel() string {
	return "src-channel"
}

// GetDestPort implements the PacketI interface and always returns "dest-port".
func (m MockPacket) GetDestPort() string {
	return "dest-port"
}

// GetDestChannel implements the PacketI interface and always returns "dest-channel".
func (m MockPacket) GetDestChannel() string {
	return "dest-channel"
}

// GetData implements the PacketI interface and always returns provided data.
func (m MockPacket) GetData() []byte {
	return m.data
}

// ValidateBasic implements the PacketI interface and always returns nil.
func (m MockPacket) ValidateBasic() error {
	return nil
}

// NewMockFungiblePacketData creates a new NewFungibleTokenPacketData for testing.
func NewMockFungiblePacketData(invalidReceiver bool) transfertypes.FungibleTokenPacketData {
	data := transfertypes.NewFungibleTokenPacketData(
		"denom",
		"500",
		"sender",
		"receiver",
		"memo",
	)
	if invalidReceiver {
		data.Receiver = strings.Repeat("a", 4096)
	}
	return data
}

// NewMockSerializedPacketData creates a new serialized NewFungibleTokenPacketData for testing.
func NewMockSerializedPacketData() []byte {
	data := NewMockFungiblePacketData(false)
	bytes, _ := json.Marshal(data)
	return bytes
}

// MockPacket is a test struct that implements the PacketI interface.
type MockPermissionedKeeper struct {
	valid bool
}

// NewMockPermissionedKeeper is a test struct that implements the PermissionedKeeper interface.
func NewMockPermissionedKeeper(valid bool) *MockPermissionedKeeper {
	return &MockPermissionedKeeper{
		valid: valid,
	}
}

// GetTimeoutHeight implements the PermissionedKeeper interface and provides a basic error or success message.
func (m *MockPermissionedKeeper) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	if !m.valid {
		return nil, ibcratelimit.ErrRateLimitExceeded
	}
	return []byte("success"), nil
}
