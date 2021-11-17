package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
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
	totalFees := sdk.Coins{}
	gasInfo, _, err := k.simulateFunc(request.Tx)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	theTx, err := k.txDecoder(request.Tx)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}
	msgs := theTx.GetMsgs()
	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		msgFees, err := k.GetMsgBasedFee(ctx, typeURL)
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
		if msgFees == nil {
			continue
		}
		if msgFees.AdditionalFee.IsPositive() {
			additionalFees = additionalFees.Add(sdk.NewCoin(msgFees.AdditionalFee.Denom, msgFees.AdditionalFee.Amount))
		}
	}

	totalFees = totalFees.Add(sdk.NewCoin("nhash", sdk.NewInt(int64(gasInfo.GasUsed)*int64(k.GetMinGasPrice(ctx)))))
	totalFees = totalFees.Add(additionalFees...)

	return &types.CalculateTxFeesResponse{
		AdditionalFees: additionalFees,
		TotalFees:      totalFees,
		EstimatedGas:   gasInfo.GasUsed,
	}, nil
}
