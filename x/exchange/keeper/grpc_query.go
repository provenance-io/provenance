package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/exchange"
)

// QueryServer is an alias for a Keeper that implements the exchange.QueryServer interface.
type QueryServer struct {
	Keeper
}

func NewQueryServer(k Keeper) exchange.QueryServer {
	return QueryServer{Keeper: k}
}

var _ exchange.QueryServer = QueryServer{}

// QueryOrderFeeCalc calculates the fees that will be associated with the provided order.
func (k QueryServer) QueryOrderFeeCalc(goCtx context.Context, req *exchange.QueryOrderFeeCalcRequest) (*exchange.QueryOrderFeeCalcResponse, error) {
	if req == nil || (req.AskOrder == nil && req.BidOrder == nil) {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.AskOrder != nil && req.BidOrder != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := k.getStore(ctx)
	resp := &exchange.QueryOrderFeeCalcResponse{}

	switch {
	case req.AskOrder != nil:
		order := req.AskOrder
		ratioFee, err := calculateSellerSettlementRatioFee(store, order.MarketId, order.Price)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to calculate seller ratio fee option: %v", err)
		}
		if ratioFee != nil {
			resp.SettlementRatioFeeOptions = append(resp.SettlementRatioFeeOptions, *ratioFee)
		}
		resp.SettlementFlatFeeOptions = getSellerSettlementFlatFees(store, order.MarketId)
		resp.CreationFeeOptions = getCreateAskFlatFees(store, order.MarketId)
	case req.BidOrder != nil:
		order := req.BidOrder
		ratioFees, err := calcBuyerSettlementRatioFeeOptions(store, order.MarketId, order.Price)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to calculate buyer ratio fee options: %v", err)
		}
		if len(ratioFees) > 0 {
			resp.SettlementRatioFeeOptions = append(resp.SettlementRatioFeeOptions, ratioFees...)
		}
		resp.SettlementFlatFeeOptions = getBuyerSettlementFlatFees(store, order.MarketId)
		resp.CreationFeeOptions = getCreateBidFlatFees(store, order.MarketId)
	default:
		// This case should have been caught right off the bat in this query.
		panic(fmt.Errorf("missing QueryOrderFeeCalc case"))
	}

	return resp, nil
}

// QueryGetOrder looks up an order by id.
func (k QueryServer) QueryGetOrder(goCtx context.Context, req *exchange.QueryGetOrderRequest) (*exchange.QueryGetOrderResponse, error) {
	if req == nil || req.OrderId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	order, err := k.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if order == nil {
		return nil, status.Errorf(codes.InvalidArgument, "order %d not found", req.OrderId)
	}

	return &exchange.QueryGetOrderResponse{Order: order}, nil
}

// QueryGetMarketOrders looks up the orders in a market.
func (k QueryServer) QueryGetMarketOrders(goCtx context.Context, req *exchange.QueryGetMarketOrdersRequest) (*exchange.QueryGetMarketOrdersResponse, error) {
	if req == nil || req.MarketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetIndexKeyPrefixMarketToOrder(req.MarketId)
	store := prefix.NewStore(k.getStore(ctx), pre)
	resp := &exchange.QueryGetMarketOrdersResponse{}
	var pageErr error
	resp.Pagination, pageErr = query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// If we can't get the order id from the key, just pretend like it doesn't exist.
		_, orderID, perr := ParseIndexKeyMarketToOrder(key)
		if perr != nil {
			return false, nil
		}
		if accumulate {
			// Only add them to the result if we can read it.
			// This might result in fewer results than the limit, but at least one bad entry won't block others.
			order, oerr := k.parseOrderStoreValue(orderID, value)
			if oerr != nil {
				resp.Orders = append(resp.Orders, order)
			}
		}
		return true, nil
	})

	if pageErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders: %v", pageErr)
	}

	return resp, nil
}

// QueryGetOwnerOrders looks up the orders from the provided owner address.
func (k QueryServer) QueryGetOwnerOrders(goCtx context.Context, req *exchange.QueryGetOwnerOrdersRequest) (*exchange.QueryGetOwnerOrdersResponse, error) {
	// TODO[1658]: Implement QueryGetOwnerOrders query
	panic("not implemented")
}

// QueryGetAssetOrders looks up the orders for a specific asset denom.
func (k QueryServer) QueryGetAssetOrders(goCtx context.Context, req *exchange.QueryGetAssetOrdersRequest) (*exchange.QueryGetAssetOrdersResponse, error) {
	// TODO[1658]: Implement QueryGetAssetOrders query
	panic("not implemented")
}

// QueryGetAllOrders gets all orders in the exchange module.
func (k QueryServer) QueryGetAllOrders(goCtx context.Context, req *exchange.QueryGetAllOrdersRequest) (*exchange.QueryGetAllOrdersResponse, error) {
	// TODO[1658]: Implement QueryGetAllOrders query
	panic("not implemented")
}

// QueryMarketInfo returns the information/details about a market.
func (k QueryServer) QueryMarketInfo(goCtx context.Context, req *exchange.QueryMarketInfoRequest) (*exchange.QueryMarketInfoResponse, error) {
	// TODO[1658]: Implement QueryMarketInfo query
	panic("not implemented")
}

// QueryParams returns the exchange module parameters.
func (k QueryServer) QueryParams(goCtx context.Context, req *exchange.QueryParamsRequest) (*exchange.QueryParamsResponse, error) {
	// TODO[1658]: Implement QueryParams query
	panic("not implemented")
}

// QueryValidateCreateMarket checks the provided MsgGovCreateMarketResponse and returns any errors it might have.
func (k QueryServer) QueryValidateCreateMarket(goCtx context.Context, req *exchange.QueryValidateCreateMarketRequest) (*exchange.QueryValidateCreateMarketResponse, error) {
	// TODO[1658]: Implement QueryValidateCreateMarket query
	panic("not implemented")
}

// QueryValidateManageFees checks the provided MsgGovManageFeesRequest and returns any errors that it might have.
func (k QueryServer) QueryValidateManageFees(goCtx context.Context, req *exchange.QueryValidateManageFeesRequest) (*exchange.QueryValidateManageFeesResponse, error) {
	// TODO[1658]: Implement QueryValidateManageFees query
	panic("not implemented")
}
