package keeper

import (
	"context"
	"math/big"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/x/flatfees/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServer returns an implementation of the x/flatfees QueryServer interface for the provided keeper.
func NewQueryServer(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// Params queries the parameters for x/flatfees.
func (k queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	c := sdk.UnwrapSDKContext(ctx)
	return &types.QueryParamsResponse{Params: k.GetParams(c)}, nil
}

// AllMsgFees returns info on all msg types that have a customized msg fee.
func (k queryServer) AllMsgFees(ctx context.Context, req *types.QueryAllMsgFeesRequest) (*types.QueryAllMsgFeesResponse, error) {
	convert := true
	var pageReq *query.PageRequest
	if req != nil {
		convert = !req.DoNotConvert
		pageReq = req.Pagination
	}

	var cf *types.ConversionFactor
	if convert {
		params := k.GetParams(sdk.UnwrapSDKContext(ctx))
		cf = &params.ConversionFactor
	}

	rv := &types.QueryAllMsgFeesResponse{}
	var err error
	rv.MsgFees, rv.Pagination, err = query.CollectionPaginate(ctx, k.msgFees, pageReq, func(_ string, msgFee types.MsgFee) (*types.MsgFee, error) {
		if convert {
			return cf.ConvertMsgFee(&msgFee), nil
		}
		return &msgFee, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return rv, nil
}

// MsgFee will return information about what it will cost to execute a given msg type.
// If the provided msg type does not have a specific fee defined, the default is returned.
func (k queryServer) MsgFee(ctx context.Context, req *types.QueryMsgFeeRequest) (*types.QueryMsgFeeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msgFee, err := k.GetMsgFee(sdkCtx, req.MsgTypeUrl)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if msgFee != nil && req.DoNotConvert {
		return &types.QueryMsgFeeResponse{MsgFee: msgFee}, nil
	}

	// If we don't have an entry for it, make sure it's a valid msg type url.
	if msgFee == nil {
		if _, err = k.cdc.InterfaceRegistry().Resolve(req.MsgTypeUrl); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unknown msg type url %q", req.MsgTypeUrl)
		}
	}

	params := k.GetParams(sdkCtx)
	if msgFee == nil {
		// If we don't have an entry, we know it's an actual msg type, so return the default cost.
		msgFee = &types.MsgFee{
			MsgTypeUrl: req.MsgTypeUrl,
			Cost:       params.DefaultCostCoins(),
		}
	}

	if !req.DoNotConvert {
		msgFee = params.ConversionFactor.ConvertMsgFee(msgFee)
	}

	return &types.QueryMsgFeeResponse{MsgFee: msgFee}, nil
}

// CalculateTxFees simulates executing a transaction for estimating gas usage and fees.
func (k queryServer) CalculateTxFees(_ context.Context, req *types.QueryCalculateTxFeesRequest) (*types.QueryCalculateTxFeesResponse, error) {
	if req.GasAdjustment > 10.0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("gas adjustment cannot be larger than 10.0")
	}
	if req.GasAdjustment < 0.0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("gas adjustment cannot be less than 0.0")
	}
	// If it's effectively 0, change it to the default.
	if req.GasAdjustment < 0.0001 {
		req.GasAdjustment = 1.0
	}

	gasInfo, _, txCtx, err := k.simulate(req.TxBytes)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
	}

	gasMeter, err := antewrapper.GetFlatFeeGasMeter(txCtx)
	if err != nil {
		return nil, err
	}

	// Golang's floating point math is deterministic only in that "The same floating-point operations,
	// run on the same hardware, always produce the same result." I couldn't find anything asserting
	// that the same code will produce the same result on two different sets of hardware, though.
	// Thankfully, the math/big library is designed to be deterministic. So we use that here (in order
	// to be deterministic) even though the values would fit in normal types just fine .
	// SetUint64 will set the precision to 64 bits, which is enough for us here.
	// GasUsed should have a range of 5 to 7 digits, so the result should be 5 to 8 digits.
	// Some rare cases might have significantly more gas, but if this gets larger than a uint64, the
	// Uint64() method just returns max uint64, which is acceptable behavior.
	gas := new(big.Float).SetUint64(gasInfo.GasUsed)
	adj := new(big.Float).SetFloat64(float64(req.GasAdjustment))
	gas.Mul(gas, adj)
	est, _ := gas.Uint64() // The ignored value is "Accuracy" (i.e. Above, Below, Exact) which we don't care about.
	return &types.QueryCalculateTxFeesResponse{TotalFees: gasMeter.GetRequiredFee(), EstimatedGas: est}, nil
}
