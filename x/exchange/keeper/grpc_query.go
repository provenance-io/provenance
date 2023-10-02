package keeper

import (
	"context"
	"fmt"
	"strings"

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
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for market %d: %v", req.MarketId, pageErr)
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
	var pageErr error

	resp.Pagination, pageErr = query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		// If we can't get the order id from the key, just pretend like it doesn't exist.
		_, orderID, perr := ParseIndexKeyAddressToOrder(key)
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
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for owner %s: %v", req.Owner, pageErr)
	}

	return resp, nil
}

// QueryGetAssetOrders looks up the orders for a specific asset denom.
func (k QueryServer) QueryGetAssetOrders(goCtx context.Context, req *exchange.QueryGetAssetOrdersRequest) (*exchange.QueryGetAssetOrdersResponse, error) {
	if req == nil || len(req.Asset) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var orderTypeByte byte
	filterByType := false
	if len(req.OrderType) > 0 {
		orderType := strings.ToLower(req.OrderType)
		// only look at the first 3 chars to handle stuff like "asks" or "bidOrders" too.
		if len(orderType) > 3 {
			orderType = orderType[:3]
		}
		switch orderType {
		case exchange.OrderTypeAsk:
			orderTypeByte = OrderKeyTypeAsk
		case exchange.OrderTypeBid:
			orderTypeByte = OrderKeyTypeBid
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown order type %q", req.OrderType)
		}
		filterByType = true
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetIndexKeyPrefixAssetToOrder(req.Asset)
	store := prefix.NewStore(k.getStore(ctx), pre)
	resp := &exchange.QueryGetAssetOrdersResponse{}
	var pageErr error

	resp.Pagination, pageErr = query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		if filterByType && (len(value) == 0 || value[0] != orderTypeByte) {
			return false, nil
		}
		// If we can't get the order id from the key, just pretend like it doesn't exist.
		_, orderID, perr := ParseIndexKeyAssetToOrder(key)
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
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for asset %s: %v", req.Asset, pageErr)
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
		return nil, status.Errorf(codes.InvalidArgument, "error iterating all orders: %v", pageErr)
	}

	return resp, nil
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
