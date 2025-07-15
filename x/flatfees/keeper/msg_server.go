package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/flatfees/types"
)

// MsgKeeper is an interface with all the keeper methods needed for the msg server endpoints.
type MsgKeeper interface {
	ValidateAuthority(authority string) error
	SetParams(ctx sdk.Context, params types.Params) error
	SetMsgFee(ctx sdk.Context, msgFee types.MsgFee) error
	RemoveMsgFee(ctx sdk.Context, msgType string) error
	SetConversionFactor(ctx sdk.Context, conversionFactor types.ConversionFactor) error
}

type msgServer struct {
	MsgKeeper
}

// NewMsgServer returns an implementation of the x/flatfees MsgServer interface for the provided Keeper.
func NewMsgServer(keeper MsgKeeper) types.MsgServer {
	return &msgServer{MsgKeeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateParams is a governance endpoint for updating the x/flatfees params.
func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParamsRequest) (*types.MsgUpdateParamsResponse, error) {
	if err := m.ValidateAuthority(req.Authority); err != nil {
		return nil, err
	}

	err := m.SetParams(sdk.UnwrapSDKContext(goCtx), req.Params)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// UpdateConversionFactor is a governance endpoint for updating just the conversion factor in the x/flatfees params.
func (m msgServer) UpdateConversionFactor(goCtx context.Context, req *types.MsgUpdateConversionFactorRequest) (*types.MsgUpdateConversionFactorResponse, error) {
	if err := m.ValidateAuthority(req.Authority); err != nil {
		return nil, err
	}

	err := m.SetConversionFactor(sdk.UnwrapSDKContext(goCtx), req.ConversionFactor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &types.MsgUpdateConversionFactorResponse{}, nil
}

// UpdateMsgFees is a governance endpoint for updating fees for specific msgs.
func (m msgServer) UpdateMsgFees(goCtx context.Context, req *types.MsgUpdateMsgFeesRequest) (*types.MsgUpdateMsgFeesResponse, error) {
	if err := m.ValidateAuthority(req.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	for _, msgFee := range req.ToSet {
		if err := m.SetMsgFee(ctx, *msgFee); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "could not set msg fee: %v", err)
		}
	}

	for _, msgType := range req.ToUnset {
		err := m.RemoveMsgFee(ctx, msgType)
		if err != nil && !errors.Is(err, types.ErrMsgFeeDoesNotExist) {
			return nil, status.Errorf(codes.InvalidArgument, "could not remove msg fee: %v", err)
		}
	}

	return &types.MsgUpdateMsgFeesResponse{}, nil
}
