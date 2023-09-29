package keeper

import (
	"context"

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
	// TODO[1658]: Implement QueryOrderFeeCalc query
	panic("not implemented")
}

// QuerySettlementFeeCalc calculates the fees that will be associated with the provided settlement.
func (k QueryServer) QuerySettlementFeeCalc(goCtx context.Context, req *exchange.QuerySettlementFeeCalcRequest) (*exchange.QuerySettlementFeeCalcResponse, error) {
	// TODO[1658]: Implement QuerySettlementFeeCalc query
	panic("not implemented")
}

// QueryGetOrder looks up an order by id.
func (k QueryServer) QueryGetOrder(goCtx context.Context, req *exchange.QueryGetOrderRequest) (*exchange.QueryGetOrderResponse, error) {
	// TODO[1658]: Implement QueryGetOrder query
	panic("not implemented")
}

// QueryGetMarketOrders looks up the orders in a market.
func (k QueryServer) QueryGetMarketOrders(goCtx context.Context, req *exchange.QueryGetMarketOrdersRequest) (*exchange.QueryGetMarketOrdersResponse, error) {
	// TODO[1658]: Implement QueryGetMarketOrders query
	panic("not implemented")
}

// QueryGetAddressOrders looks up the orders from the provided address.
func (k QueryServer) QueryGetAddressOrders(goCtx context.Context, req *exchange.QueryGetAddressOrdersRequest) (*exchange.QueryGetAddressOrdersResponse, error) {
	// TODO[1658]: Implement QueryGetAddressOrders query
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
