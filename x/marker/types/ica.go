package types

import context "context"

type IcaServer interface {
	ReflectMarker(goCtx context.Context, msg *MsgIcaReflectMarkerRequest) (*MsgIcaReflectMarkerResponse, error)
}
