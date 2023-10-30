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

type MockPacket struct {
	data        []byte
	validHeight bool
}

func NewMockPacket(data []byte, validHeight bool) *MockPacket {
	return &MockPacket{
		data:        data,
		validHeight: validHeight,
	}
}

func (m MockPacket) GetSequence() uint64 {
	return 1
}

func (m MockPacket) GetTimeoutHeight() exported.Height {
	if !m.validHeight {
		return nil
	}
	return clienttypes.Height{
		RevisionNumber: 5,
		RevisionHeight: 5,
	}
}

func (m MockPacket) GetTimeoutTimestamp() uint64 {
	return 1
}

func (m MockPacket) GetSourcePort() string {
	return "src-port"
}

func (m MockPacket) GetSourceChannel() string {
	return "src-channel"
}

func (m MockPacket) GetDestPort() string {
	return "dest-port"
}

func (m MockPacket) GetDestChannel() string {
	return "dest-channel"
}

func (m MockPacket) GetData() []byte {
	return m.data
}

func (m MockPacket) ValidateBasic() error {
	return nil
}

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

func NewMockSerializedPacketData() []byte {
	data := NewMockFungiblePacketData(false)
	bytes, _ := json.Marshal(data)
	return bytes
}

type MockPermissionedKeeper struct {
	valid bool
}

func NewMockPermissionedKeeper(valid bool) *MockPermissionedKeeper {
	return &MockPermissionedKeeper{
		valid: valid,
	}
}

func (m *MockPermissionedKeeper) Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error) {
	if !m.valid {
		return nil, ibcratelimit.ErrRateLimitExceeded
	}
	return []byte("success"), nil
}
