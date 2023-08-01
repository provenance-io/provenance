package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	"github.com/provenance-io/provenance/x/oracle/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the account MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateOracle changes the oracle's address to the provided one
func (s msgServer) UpdateOracle(goCtx context.Context, msg *types.MsgUpdateOracleRequest) (*types.MsgUpdateOracleResponse, error) {
	//ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.MsgUpdateOracleResponse{}, nil
}

func (s msgServer) QueryOracle(goCtx context.Context, msg *types.MsgQueryOracleRequest) (*types.MsgQueryOracleResponse, error) {
	return &types.MsgQueryOracleResponse{}, nil
}

func (k msgServer) SendQueryAllBalances(goCtx context.Context, msg *types.MsgSendQueryAllBalances) (*types.MsgSendQueryAllBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(k.GetPort(ctx), msg.ChannelId))
	if !found {
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	q := banktypes.QueryAllBalancesRequest{
		Address:    msg.Address,
		Pagination: msg.Pagination,
	}
	reqs := []abcitypes.RequestQuery{
		{
			Path: "/cosmos.bank.v1beta1.Query/AllBalances",
			Data: k.cdc.MustMarshal(&q),
		},
	}

	// timeoutTimestamp set to max value with the unsigned bit shifted to sastisfy hermes timestamp conversion
	// it is the responsibility of the auth module developer to ensure an appropriate timeout timestamp
	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	seq, err := k.SendQuery(ctx, types.PortID, msg.ChannelId, chanCap, reqs, clienttypes.ZeroHeight(), uint64(timeoutTimestamp))
	if err != nil {
		return nil, err
	}

	k.SetQueryRequest(ctx, seq, q)

	return &types.MsgSendQueryAllBalancesResponse{
		Sequence: seq,
	}, nil
}
