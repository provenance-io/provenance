package statesync

import (
	// tmrpccore "github.com/cometbft/cometbft/rpc/core" // TODO[1760]: sync-info
	// tmrpc "github.com/cometbft/cometbft/rpc/jsonrpc/server" // TODO[1760]: sync-info
	tmrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"

	"github.com/cosmos/cosmos-sdk/version"
)

func RegisterSyncStatus() {
	// TODO[1760]: sync-info: Figure out how to still set a custom route.
	// tmrpccore.Routes["sync_info"] = tmrpc.NewRPCFunc(GetSyncInfoAtBlock, "height")
}

func GetSyncInfoAtBlock(ctx *tmrpctypes.Context, height *int64) (*GetSyncInfo, error) {
	// TODO[1760]: sync-info: Figure out the new way to get the current block.
	// block, err := tmrpccore.Block(ctx, height)
	// if err != nil {
	// 	return nil, err
	// }
	versionInfo := version.NewInfo()
	si := &GetSyncInfo{
		BlockHeight: 123,        // block.Block.Header.Height, // TODO[1760]: sync-info
		BlockHash:   "finishme", // block.Block.Header.Hash().String(), // TODO[1760]: sync-info
		Version:     versionInfo.Version,
	}
	return si, nil
}

type GetSyncInfo struct {
	BlockHeight int64  `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	Version     string `json:"version"`
}
