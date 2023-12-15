package keeper

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
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
	// TODO[1760]: event-history: Put this back once our version of the SDK is back in with the updated baseapp.Simulate func.
	/*
		ctx := sdk.UnwrapSDKContext(goCtx)

		gasInfo, _, txCtx, err := k.simulateFunc(request.TxBytes)
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}

		gasMeter, err := antewrapper.GetFeeGasMeter(txCtx)
		if err != nil {
			return nil, err
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
		gasUsed := int64(float64(gasInfo.GasUsed) * float64(gasAdjustment))
		totalFees := gasMeter.FeeConsumed().Add(sdk.NewCoin(baseDenom, minGasPrice.Amount.MulRaw(gasUsed)))

		return &types.CalculateTxFeesResponse{
			AdditionalFees: gasMeter.FeeConsumed(),
			TotalFees:      totalFees,
			EstimatedGas:   uint64(gasUsed),
		}, nil
	*/
	return nil, errors.New("not yet updated")
}
