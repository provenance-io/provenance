package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

var _ porttypes.ICS4Wrapper = &ICS4WrapperMock{}

type ICS4WrapperMock struct{}

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

func (m *ICS4WrapperMock) WriteAcknowledgement(
	_ sdk.Context,
	_ *capabilitytypes.Capability,
	_ exported.PacketI,
	_ exported.Acknowledgement,
) error {
	return nil
}

func (m *ICS4WrapperMock) GetAppVersion(
	_ sdk.Context,
	_,
	_ string,
) (string, bool) {
	return "", false
}
