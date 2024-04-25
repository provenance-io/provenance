package statesync

import (
	cmtrpccore "github.com/cometbft/cometbft/rpc/core"       // TODO[1760]: sync-info
	server "github.com/cometbft/cometbft/rpc/jsonrpc/server" // TODO[1760]: sync-info
	cmtrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"

	"github.com/cosmos/cosmos-sdk/version"
)

func RegisterSyncStatus(env cmtrpccore.Environment) {
	routes := env.GetRoutes()
	routes["sync_info"] = server.NewRPCFunc(GetSyncInfoAtBlock, "height")
}

func GetSyncInfoAtBlock(ctx *cmtrpctypes.Context, height *int64) (*GetSyncInfo, error) {
	// TODO[1760]: sync-info: Figure out the new way to get the current block.
	// How do we get env?
	var env cmtrpccore.Environment

	block, err := env.Block(ctx, height)
	if err != nil {
		return nil, err
	}
	versionInfo := version.NewInfo()
	si := &GetSyncInfo{
		BlockHeight: block.Block.Header.Height,
		BlockHash:   block.Block.Header.Hash().String(),
		Version:     versionInfo.Version,
	}
	return si, nil
}

type GetSyncInfo struct {
	BlockHeight int64  `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	Version     string `json:"version"`
}
