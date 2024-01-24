package keeper

import (
	"context"
	"errors"
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

// OrderFeeCalc calculates the fees that will be associated with the provided order.
func (k QueryServer) OrderFeeCalc(goCtx context.Context, req *exchange.QueryOrderFeeCalcRequest) (*exchange.QueryOrderFeeCalcResponse, error) {
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
		if err := validateMarketExists(store, order.MarketId); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
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
		if err := validateMarketExists(store, order.MarketId); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
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
		panic(fmt.Errorf("missing OrderFeeCalc case"))
	}

	return resp, nil
}

// GetOrder looks up an order by id.
func (k QueryServer) GetOrder(goCtx context.Context, req *exchange.QueryGetOrderRequest) (*exchange.QueryGetOrderResponse, error) {
	if req == nil || req.OrderId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	order, err := k.Keeper.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if order == nil {
		return nil, status.Errorf(codes.InvalidArgument, "order %d not found", req.OrderId)
	}

	return &exchange.QueryGetOrderResponse{Order: order}, nil
}

// GetOrderByExternalID looks up an order by market id and external id.
func (k QueryServer) GetOrderByExternalID(goCtx context.Context, req *exchange.QueryGetOrderByExternalIDRequest) (*exchange.QueryGetOrderByExternalIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.MarketId == 0 || len(req.ExternalId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	order, err := k.Keeper.GetOrderByExternalID(ctx, req.MarketId, req.ExternalId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if order == nil {
		return nil, status.Errorf(codes.InvalidArgument, "order not found in market %d with external id %q",
			req.MarketId, req.ExternalId)
	}

	return &exchange.QueryGetOrderByExternalIDResponse{Order: order}, nil
}

// GetMarketOrders looks up the orders in a market.
func (k QueryServer) GetMarketOrders(goCtx context.Context, req *exchange.QueryGetMarketOrdersRequest) (*exchange.QueryGetMarketOrdersResponse, error) {
	if req == nil || req.MarketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetIndexKeyPrefixMarketToOrder(req.MarketId)

	resp := &exchange.QueryGetMarketOrdersResponse{}
	var err error
	resp.Pagination, resp.Orders, err = k.getPageOfOrdersFromIndex(ctx, pre, req.Pagination, req.OrderType, req.AfterOrderId)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for market %d: %v", req.MarketId, err)
	}

	return resp, nil
}

// GetOwnerOrders looks up the orders from the provided owner address.
func (k QueryServer) GetOwnerOrders(goCtx context.Context, req *exchange.QueryGetOwnerOrdersRequest) (*exchange.QueryGetOwnerOrdersResponse, error) {
	if req == nil || len(req.Owner) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	owner, aErr := sdk.AccAddressFromBech32(req.Owner)
	if aErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid owner %q: %v", req.Owner, aErr)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetIndexKeyPrefixAddressToOrder(owner)

	resp := &exchange.QueryGetOwnerOrdersResponse{}
	var err error
	resp.Pagination, resp.Orders, err = k.getPageOfOrdersFromIndex(ctx, pre, req.Pagination, req.OrderType, req.AfterOrderId)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for owner %s: %v", req.Owner, err)
	}

	return resp, nil
}

// GetAssetOrders looks up the orders for a specific asset denom.
func (k QueryServer) GetAssetOrders(goCtx context.Context, req *exchange.QueryGetAssetOrdersRequest) (*exchange.QueryGetAssetOrdersResponse, error) {
	if req == nil || len(req.Asset) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	pre := GetIndexKeyPrefixAssetToOrder(req.Asset)

	resp := &exchange.QueryGetAssetOrdersResponse{}
	var err error
	resp.Pagination, resp.Orders, err = k.getPageOfOrdersFromIndex(ctx, pre, req.Pagination, req.OrderType, req.AfterOrderId)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating orders for asset %s: %v", req.Asset, err)
	}

	return resp, nil
}

// GetAllOrders gets all orders in the exchange module.
func (k QueryServer) GetAllOrders(goCtx context.Context, req *exchange.QueryGetAllOrdersRequest) (*exchange.QueryGetAllOrdersResponse, error) {
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
			if oerr == nil {
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

// GetCommitment gets the funds in an account that are committed to the market.
func (k QueryServer) GetCommitment(goCtx context.Context, req *exchange.QueryGetCommitmentRequest) (*exchange.QueryGetCommitmentResponse, error) {
	if req == nil || len(req.Account) == 0 || req.MarketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account %q: %v", req.Account, err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &exchange.QueryGetCommitmentResponse{
		Amount: k.GetCommitmentAmount(ctx, req.MarketId, addr),
	}
	return resp, nil
}

// GetAccountCommitments gets all the funds in an account that are committed to any market.
func (k QueryServer) GetAccountCommitments(goCtx context.Context, req *exchange.QueryGetAccountCommitmentsRequest) (*exchange.QueryGetAccountCommitmentsResponse, error) {
	if req == nil || len(req.Account) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	addr, err := sdk.AccAddressFromBech32(req.Account)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid account %q: %v", req.Account, err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	store := k.getStore(ctx)
	resp := &exchange.QueryGetAccountCommitmentsResponse{}
	k.IterateKnownMarketIDs(ctx, func(marketID uint32) bool {
		amount := getCommitmentAmount(store, marketID, addr)
		if !amount.IsZero() {
			resp.Commitments = append(resp.Commitments, &exchange.MarketAmount{MarketId: marketID, Amount: amount})
		}
		return false
	})

	return resp, nil
}

// GetMarketCommitments gets all the funds committed to a market from any account.
func (k QueryServer) GetMarketCommitments(goCtx context.Context, req *exchange.QueryGetMarketCommitmentsRequest) (*exchange.QueryGetMarketCommitmentsResponse, error) {
	if req == nil || req.MarketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	keyPrefix := GetKeyPrefixCommitmentsToMarket(req.MarketId)
	store := prefix.NewStore(k.getStore(ctx), keyPrefix)

	resp := &exchange.QueryGetMarketCommitmentsResponse{}
	var pageErr error
	resp.Pagination, pageErr = query.Paginate(store, req.Pagination, func(keySuffix []byte, value []byte) error {
		com, _ := parseCommitmentKeyValue(keyPrefix, keySuffix, value)
		if com != nil && !com.Amount.IsZero() {
			resp.Commitments = append(resp.Commitments, &exchange.AccountAmount{Account: com.Account, Amount: com.Amount})
		}
		return nil
	})

	if pageErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating commitments for market %d: %v", req.MarketId, pageErr)
	}

	return resp, nil
}

// GetAllCommitments gets all fund committed to any market from any account.
func (k QueryServer) GetAllCommitments(goCtx context.Context, req *exchange.QueryGetAllCommitmentsRequest) (*exchange.QueryGetAllCommitmentsResponse, error) {
	var pageReq *query.PageRequest
	if req != nil {
		pageReq = req.Pagination
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	keyPrefix := GetKeyPrefixCommitments()
	store := prefix.NewStore(k.getStore(ctx), keyPrefix)

	resp := &exchange.QueryGetAllCommitmentsResponse{}
	var pageErr error
	resp.Pagination, pageErr = query.Paginate(store, pageReq, func(keySuffix []byte, value []byte) error {
		com, _ := parseCommitmentKeyValue(keyPrefix, keySuffix, value)
		if com != nil && !com.Amount.IsZero() {
			resp.Commitments = append(resp.Commitments, com)
		}
		return nil
	})

	if pageErr != nil {
		return nil, status.Errorf(codes.InvalidArgument, "error iterating all commitments: %v", pageErr)
	}

	return resp, nil
}

// GetMarket returns all the information and details about a market.
func (k QueryServer) GetMarket(goCtx context.Context, req *exchange.QueryGetMarketRequest) (*exchange.QueryGetMarketResponse, error) {
	if req == nil || req.MarketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	market := k.Keeper.GetMarket(ctx, req.MarketId)
	if market == nil {
		return nil, status.Errorf(codes.InvalidArgument, "market %d not found", req.MarketId)
	}

	resp := &exchange.QueryGetMarketResponse{
		Address: exchange.GetMarketAddress(req.MarketId).String(),
		Market:  market,
	}

	return resp, nil
}

// GetAllMarkets returns brief information about each market.
func (k QueryServer) GetAllMarkets(goCtx context.Context, req *exchange.QueryGetAllMarketsRequest) (*exchange.QueryGetAllMarketsResponse, error) {
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

// Params returns the exchange module parameters.
func (k QueryServer) Params(goCtx context.Context, _ *exchange.QueryParamsRequest) (*exchange.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &exchange.QueryParamsResponse{Params: k.GetParamsOrDefaults(ctx)}
	return resp, nil
}

// CommitmentSettlementFeeCalc calculates the fees a market will pay for a commitment settlement using current NAVs.
func (k QueryServer) CommitmentSettlementFeeCalc(goCtx context.Context, req *exchange.QueryCommitmentSettlementFeeCalcRequest) (*exchange.QueryCommitmentSettlementFeeCalcResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp, err := k.CalculateCommitmentSettlementFee(ctx, req.Settlement)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return resp, nil
}

// ValidateCreateMarket checks the provided MsgGovCreateMarketResponse and returns any errors it might have.
func (k QueryServer) ValidateCreateMarket(goCtx context.Context, req *exchange.QueryValidateCreateMarketRequest) (*exchange.QueryValidateCreateMarketResponse, error) {
	if req == nil || req.CreateMarketRequest == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	msg := req.CreateMarketRequest
	resp := &exchange.QueryValidateCreateMarketResponse{}

	if err := msg.ValidateBasic(); err != nil {
		resp.Error = err.Error()
		return resp, nil
	}

	if err := k.ValidateAuthority(msg.Authority); err != nil {
		resp.Error = err.Error()
		return resp, nil
	}

	// The SDK *should* already be using a cache context for queries, but I'm doing it here too just to be on the safe side.
	ctx, _ := sdk.UnwrapSDKContext(goCtx).CacheContext()
	marketID, err := k.Keeper.CreateMarket(ctx, msg.Market)
	if err != nil {
		resp.Error = err.Error()
		return resp, nil
	}

	resp.GovPropWillPass = true

	var errs []error
	if err = exchange.ValidateReqAttrsAreNormalized("create ask", msg.Market.ReqAttrCreateAsk); err != nil {
		errs = append(errs, err)
	}
	if err = exchange.ValidateReqAttrsAreNormalized("create bid", msg.Market.ReqAttrCreateBid); err != nil {
		errs = append(errs, err)
	}

	if err = k.Keeper.ValidateMarket(ctx, marketID); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		resp.Error = errors.Join(errs...).Error()
	}

	return resp, nil
}

// ValidateMarket checks for any problems with a market's setup.
func (k QueryServer) ValidateMarket(goCtx context.Context, req *exchange.QueryValidateMarketRequest) (*exchange.QueryValidateMarketResponse, error) {
	if req == nil || req.MarketId == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &exchange.QueryValidateMarketResponse{}
	if err := k.Keeper.ValidateMarket(ctx, req.MarketId); err != nil {
		resp.Error = err.Error()
	}

	return resp, nil
}

// ValidateManageFees checks the provided MsgGovManageFeesRequest and returns any errors that it might have.
func (k QueryServer) ValidateManageFees(goCtx context.Context, req *exchange.QueryValidateManageFeesRequest) (*exchange.QueryValidateManageFeesResponse, error) {
	if req == nil || req.ManageFeesRequest == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	msg := req.ManageFeesRequest
	resp := &exchange.QueryValidateManageFeesResponse{}

	if err := msg.ValidateBasic(); err != nil {
		resp.Error = err.Error()
		return resp, nil
	}

	if err := k.ValidateAuthority(msg.Authority); err != nil {
		resp.Error = err.Error()
		return resp, nil
	}

	// The SDK *should* already be using a cache context for queries, but I'm doing it here too just to be on the safe side.
	ctx, _ := sdk.UnwrapSDKContext(goCtx).CacheContext()
	store := k.getStore(ctx)
	if err := validateMarketExists(store, msg.MarketId); err != nil {
		resp.Error = err.Error()
		return resp, nil
	}

	resp.GovPropWillPass = true

	var errs []error
	if len(msg.AddFeeCreateAskFlat) > 0 || len(msg.RemoveFeeCreateAskFlat) > 0 {
		createAskFlats := getCreateAskFlatFees(store, msg.MarketId)
		errs = append(errs, exchange.ValidateAddRemoveFeeOptionsWithExisting("create-ask",
			createAskFlats, msg.AddFeeCreateAskFlat, msg.RemoveFeeCreateAskFlat)...)
	}

	if len(msg.AddFeeCreateBidFlat) > 0 || len(msg.RemoveFeeCreateBidFlat) > 0 {
		createBidFlats := getCreateBidFlatFees(store, msg.MarketId)
		errs = append(errs, exchange.ValidateAddRemoveFeeOptionsWithExisting("create-bid",
			createBidFlats, msg.AddFeeCreateBidFlat, msg.RemoveFeeCreateBidFlat)...)
	}

	if len(msg.AddFeeCreateCommitmentFlat) > 0 || len(msg.RemoveFeeCreateCommitmentFlat) > 0 {
		createCommitmentFlatKeyMakers := getCreateCommitmentFlatFees(store, msg.MarketId)
		errs = append(errs, exchange.ValidateAddRemoveFeeOptionsWithExisting("create-commitment",
			createCommitmentFlatKeyMakers, msg.AddFeeCreateCommitmentFlat, msg.RemoveFeeCreateCommitmentFlat)...)
	}

	if len(msg.AddFeeSellerSettlementFlat) > 0 || len(msg.RemoveFeeSellerSettlementFlat) > 0 {
		sellerFlats := getSellerSettlementFlatFees(store, msg.MarketId)
		errs = append(errs, exchange.ValidateAddRemoveFeeOptionsWithExisting("seller settlement",
			sellerFlats, msg.AddFeeSellerSettlementFlat, msg.RemoveFeeSellerSettlementFlat)...)
	}

	if len(msg.AddFeeSellerSettlementRatios) > 0 || len(msg.RemoveFeeSellerSettlementRatios) > 0 {
		sellerRatios := getSellerSettlementRatios(store, msg.MarketId)
		errs = append(errs, exchange.ValidateAddRemoveFeeRatiosWithExisting("seller settlement",
			sellerRatios, msg.AddFeeSellerSettlementRatios, msg.RemoveFeeSellerSettlementRatios)...)
	}

	if len(msg.AddFeeBuyerSettlementFlat) > 0 || len(msg.RemoveFeeBuyerSettlementFlat) > 0 {
		buyerFlats := getBuyerSettlementFlatFees(store, msg.MarketId)
		errs = append(errs, exchange.ValidateAddRemoveFeeOptionsWithExisting("buyer settlement",
			buyerFlats, msg.AddFeeBuyerSettlementFlat, msg.RemoveFeeBuyerSettlementFlat)...)
	}

	if len(msg.AddFeeBuyerSettlementRatios) > 0 || len(msg.RemoveFeeBuyerSettlementRatios) > 0 {
		buyerRatios := getBuyerSettlementRatios(store, msg.MarketId)
		errs = append(errs, exchange.ValidateAddRemoveFeeRatiosWithExisting("buyer settlement",
			buyerRatios, msg.AddFeeBuyerSettlementRatios, msg.RemoveFeeBuyerSettlementRatios)...)
	}

	k.UpdateFees(ctx, msg)
	if err := k.Keeper.ValidateMarket(ctx, msg.MarketId); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		resp.Error = errors.Join(errs...).Error()
	}

	return resp, nil
}
