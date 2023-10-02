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
	var err error
	resp.Pagination, resp.Orders, err = k.GetPageOfOrdersFromIndex(store, req.Pagination, req.OrderType, req.AfterOrderId)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for market %d: %v", req.MarketId, err)
	}

	return resp, nil
}

// QueryGetOwnerOrders looks up the orders from the provided owner address.
func (k QueryServer) QueryGetOwnerOrders(goCtx context.Context, req *exchange.QueryGetOwnerOrdersRequest) (*exchange.QueryGetOwnerOrdersResponse, error) {
	if req == nil || len(req.Owner) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	owner, aErr := sdk.AccAddressFromBech32(req.Owner)
	if aErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner: %v", aErr)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetIndexKeyPrefixAddressToOrder(owner)
	store := prefix.NewStore(k.getStore(ctx), pre)

	resp := &exchange.QueryGetOwnerOrdersResponse{}
	var err error
	resp.Pagination, resp.Orders, err = k.GetPageOfOrdersFromIndex(store, req.Pagination, req.OrderType, req.AfterOrderId)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for owner %s: %v", req.Owner, err)
	}

	return resp, nil
}

// QueryGetAssetOrders looks up the orders for a specific asset denom.
func (k QueryServer) QueryGetAssetOrders(goCtx context.Context, req *exchange.QueryGetAssetOrdersRequest) (*exchange.QueryGetAssetOrdersResponse, error) {
	if req == nil || len(req.Asset) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetIndexKeyPrefixAssetToOrder(req.Asset)
	store := prefix.NewStore(k.getStore(ctx), pre)

	resp := &exchange.QueryGetAssetOrdersResponse{}
	var err error
	resp.Pagination, resp.Orders, err = k.GetPageOfOrdersFromIndex(store, req.Pagination, req.OrderType, req.AfterOrderId)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for asset %s: %v", req.Asset, err)
	}

	return resp, nil
}

// QueryGetAllOrders gets all orders in the exchange module.
func (k QueryServer) QueryGetAllOrders(goCtx context.Context, req *exchange.QueryGetAllOrdersRequest) (*exchange.QueryGetAllOrdersResponse, error) {
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetKeyPrefixOrder()
	store := prefix.NewStore(k.getStore(ctx), pre)
	resp := &exchange.QueryGetAllOrdersResponse{}
	var pageErr error

	resp.Pagination, pageErr = query.FilteredPaginate(store, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// If we can't get the order id from the key, just pretend like it doesn't exist.
		orderID, ok := ParseKeyOrder(key)
		if !ok {
			return false, nil
		}
		if accumulate {
			// Only add it to the result if we can read it. This might result in fewer results than the limit,
			// but at least one bad entry won't block others by causing the whole thing to return an error.
			order, oerr := k.parseOrderStoreValue(orderID, value)
			if oerr != nil {
				resp.Orders = append(resp.Orders, order)
			}
		}
		return true, nil
	})

	if pageErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating all orders: %v", pageErr)
	}

	return resp, nil
}

// QueryGetMarket returns all the information and details about a market.
func (k QueryServer) QueryGetMarket(goCtx context.Context, req *exchange.QueryGetMarketRequest) (*exchange.QueryGetMarketResponse, error) {
	if req == nil || req.MarketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	market := k.GetMarket(ctx, req.MarketId)
	if market == nil {
		return nil, status.Errorf(codes.InvalidArgument, "market %d not found", req.MarketId)
	}

	return &exchange.QueryGetMarketResponse{Market: market}, nil
}

// QueryGetAllMarkets returns brief information about each market.
func (k QueryServer) QueryGetAllMarkets(goCtx context.Context, req *exchange.QueryGetAllMarketsRequest) (*exchange.QueryGetAllMarketsResponse, error) {
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetKeyPrefixKnownMarketID()
	store := prefix.NewStore(k.getStore(ctx), pre)

	resp := &exchange.QueryGetAllMarketsResponse{}
	var pageErr error
	resp.Pagination, pageErr = query.FilteredPaginate(store, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// If we can't get the market id from the key, just pretend like it doesn't exist.
		marketID, ok := ParseKeySuffixKnownMarketID(key)
		if !ok {
			return false, nil
		}
		if accumulate {
			// Only add it to the result if we can read it. This might result in fewer results than the limit,
			// but at least one bad entry won't block others by causing the whole thing to return an error.
			brief := k.GetMarketBrief(ctx, marketID)
			if brief != nil {
				resp.Markets = append(resp.Markets, brief)
			}
		}
		return true, nil
	})

	if pageErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating all known markets: %v", pageErr)
	}

	return resp, nil
}

// QueryParams returns the exchange module parameters.
func (k QueryServer) QueryParams(goCtx context.Context, _ *exchange.QueryParamsRequest) (*exchange.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &exchange.QueryParamsResponse{Params: k.GetParamsOrDefaults(ctx)}
	return resp, nil
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
