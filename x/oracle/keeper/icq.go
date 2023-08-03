package keeper

import (
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	"github.com/provenance-io/provenance/x/oracle/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

// QueryOracle sends an ICQ to the other chain's module
func (k Keeper) QueryOracle(ctx sdk.Context, query wasmtypes.RawContractMessage, channel string) (uint64, error) {
	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(k.GetPort(ctx), channel))
	if !found {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	q := types.QueryOracleRequest{
		Query: query,
	}

	reqs := []abcitypes.RequestQuery{
		{
			Path: "/provenance.oracle.v1.Query/Oracle",
			Data: k.cdc.MustMarshal(&q),
		},
	}

	timeoutTimestamp := ctx.BlockTime().Add(time.Minute).UnixNano()
	seq, err := k.SendQuery(ctx, types.PortID, channel, chanCap, reqs, clienttypes.ZeroHeight(), uint64(timeoutTimestamp))
	if err != nil {
		return 0, err
	}

	k.SetQueryRequest(ctx, seq, q)
	return seq, nil
}
