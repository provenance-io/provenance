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

	additionalFees := sdk.Coins{}
	gasInfo, _, err, txCtx := k.simulateFunc(request.TxBytes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	_, err = k.txDecoder(request.TxBytes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	gasMeter, ok := txCtx.GasMeter().(*antewrapper.FeeGasMeter)
	if !ok {
		return nil, fmt.Errorf("unable to extract fee gas meter from transaction context")
	}
	totalFees := gasMeter.FeeConsumed().Add(sdk.NewCoin(txCtx.MinGasPrices().GetDenomByIndex(0), sdk.NewInt(int64(gasInfo.GasUsed)*int64(k.GetMinGasPrice(ctx)))))
	totalFees = totalFees.Add(additionalFees...)

	return &types.CalculateTxFeesResponse{
		AdditionalFees: gasMeter.FeeConsumed(),
		TotalFees:      totalFees,
		EstimatedGas:   gasInfo.GasUsed,
	}, nil
}
