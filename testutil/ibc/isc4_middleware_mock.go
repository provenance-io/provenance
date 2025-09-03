// Package ibc provides mocks and helpers for IBC testing.
package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
)

var _ porttypes.ICS4Wrapper = &ICS4WrapperMock{}

// ICS4WrapperMock is a mock implementation of the ICS4Wrapper interface for testing.
type ICS4WrapperMock struct{}

// SendPacket mocks sending an IBC packet.
func (m *ICS4WrapperMock) SendPacket(
	_ sdk.Context,
	_ *capabilitytypes.Capability,
	_ string,
	_ string,
	_ ibcclienttypes.Height,
	_ uint64,
	_ []byte,
) (sequence uint64, err error) {
	return 1, nil
}

// WriteAcknowledgement mocks writing an acknowledgement for a packet.
func (m *ICS4WrapperMock) WriteAcknowledgement(
	_ sdk.Context,
	_ *capabilitytypes.Capability,
	_ exported.PacketI,
	_ exported.Acknowledgement,
) error {
	return nil
}

// GetAppVersion mocks retrieving the application version for a given port and channel.
func (m *ICS4WrapperMock) GetAppVersion(
	_ sdk.Context,
	_,
	_ string,
) (string, bool) {
	return "", false
}
