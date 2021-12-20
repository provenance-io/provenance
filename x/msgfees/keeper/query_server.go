package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	c := sdk.UnwrapSDKContext(ctx)
	var params types.Params
	k.paramSpace.GetParamSet(c, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) QueryAllMsgBasedFees(c context.Context, req *types.QueryAllMsgBasedFeesRequest) (*types.QueryAllMsgBasedFeesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	var msgFees []*types.MsgBasedFee
	store := ctx.KVStore(k.storeKey)
	msgFeeStore := prefix.NewStore(store, types.MsgBasedFeeKeyPrefix)
	pageRes, err := query.Paginate(msgFeeStore, req.Pagination, func(key []byte, value []byte) error {
		var msgFee types.MsgBasedFee

		if err := k.cdc.Unmarshal(value, &msgFee); err != nil {
			return err
		}

		msgFees = append(msgFees, &msgFee)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMsgBasedFeesResponse{MsgBasedFees: msgFees, Pagination: pageRes}, nil
}

func (k Keeper) CalculateTxFees(goCtx context.Context, request *types.CalculateTxFeesRequest) (*types.CalculateTxFeesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	gasInfo, _, txCtx, err := k.simulateFunc(request.TxBytes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	gasMeter, ok := txCtx.GasMeter().(*antewrapper.FeeGasMeter)
	if !ok {
		return nil, fmt.Errorf("unable to extract fee gas meter from transaction context")
	}
	baseDenom := k.defaultFeeDenom
	if request.DefaultBaseDenom != "" {
		baseDenom = request.DefaultBaseDenom
	}

	minGasPrice := int64(k.GetFloorGasPrice(ctx))
	gasAdjustment := request.GasAdjustment
	if gasAdjustment <= 0 {
		gasAdjustment = 1.0
	}
	gasUsed := int64(float64(gasInfo.GasUsed) * float64(gasAdjustment))
	totalFees := gasMeter.FeeConsumed().Add(sdk.NewCoin(baseDenom, sdk.NewInt(gasUsed*minGasPrice)))

	return &types.CalculateTxFeesResponse{
		AdditionalFees: gasMeter.FeeConsumed(),
		TotalFees:      totalFees,
		EstimatedGas:   uint64(gasUsed),
	}, nil
}
