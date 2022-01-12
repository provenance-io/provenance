package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, request *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	c := sdk.UnwrapSDKContext(ctx)
	var params types.Params
	k.paramSpace.GetParamSet(c, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) QueryAllMsgFees(c context.Context, req *types.QueryAllMsgFeesRequest) (*types.QueryAllMsgFeesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	var msgFees []*types.MsgFee
	store := ctx.KVStore(k.storeKey)
	msgFeeStore := prefix.NewStore(store, types.MsgFeeKeyPrefix)
	pageRes, err := query.Paginate(msgFeeStore, req.Pagination, func(key []byte, value []byte) error {
		var msgFee types.MsgFee

		if err := k.cdc.Unmarshal(value, &msgFee); err != nil {
			return err
		}

		msgFees = append(msgFees, &msgFee)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMsgFeesResponse{MsgFees: msgFees, Pagination: pageRes}, nil
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
	// based on Carlton H's comment this is only for testing, has no real value in practical usage.
	baseDenom := k.defaultFeeDenom
	if request.DefaultBaseDenom != "" {
		baseDenom = request.DefaultBaseDenom
	}

	minGasPrice := k.GetFloorGasPrice(ctx)
	gasAdjustment := request.GasAdjustment
	if gasAdjustment <= 0 {
		gasAdjustment = 1.0
	}
	gasUsed := sdk.NewInt(int64(float64(gasInfo.GasUsed) * float64(gasAdjustment)))
	totalFees := gasMeter.FeeConsumed().Add(sdk.NewCoin(baseDenom, minGasPrice.Amount.Mul(gasUsed)))

	return &types.CalculateTxFeesResponse{
		AdditionalFees: gasMeter.FeeConsumed(),
		TotalFees:      totalFees,
		EstimatedGas:   gasUsed.Uint64(),
	}, nil
}
