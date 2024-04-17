package keeper

import (
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	abci "github.com/cometbft/cometbft/abci/types"

	cerrs "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"

	"github.com/provenance-io/provenance/x/oracle/types"
)

// QueryOracle sends an ICQ to the other chain's module
func (k Keeper) QueryOracle(ctx sdk.Context, query wasmtypes.RawContractMessage, channel string) (uint64, error) {
	chanCap, found := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(k.GetPort(ctx), channel))
	if !found {
		return 0, cerrs.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	q := types.QueryOracleRequest{
		Query: query,
	}

	reqs := []abci.RequestQuery{
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

	return seq, nil
}
