package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	ctx := sdk.UnwrapSDKContext(goCtx)
	s.Keeper.SetOracleContract(ctx, sdk.MustAccAddressFromBech32(msg.Address))
	return &types.MsgUpdateOracleResponse{}, nil
}

func (k msgServer) SendQueryOracle(goCtx context.Context, msg *types.MsgSendQueryOracleRequest) (*types.MsgSendQueryOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(k.GetPort(ctx), msg.Channel))
	if !found {
		return nil, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	q := types.QueryOracleRequest{
		Query: msg.GetQuery(),
	}

	reqs := []abcitypes.RequestQuery{
		{
			Path: "/provenance.oracle.v1.Query/Oracle",
			Data: k.cdc.MustMarshal(&q),
		},
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	seq, err := k.SendQuery(ctx, types.PortID, msg.Channel, chanCap, reqs, clienttypes.ZeroHeight(), uint64(timeoutTimestamp))
	if err != nil {
		return nil, err
	}

	k.SetQueryRequest(ctx, seq, q)

	return &types.MsgSendQueryOracleResponse{
		Sequence: seq,
	}, nil
}
