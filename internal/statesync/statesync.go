package statesync

import (
	cmtrpccore "github.com/cometbft/cometbft/rpc/core"
	server "github.com/cometbft/cometbft/rpc/jsonrpc/server"
	cmtrpctypes "github.com/cometbft/cometbft/rpc/jsonrpc/types"

	"github.com/cosmos/cosmos-sdk/version"
)

type ProvenanceEnvironment struct {
	cmtrpccore.Environment
}

func RegisterSyncStatus(env cmtrpccore.Environment) {
	routes := env.GetRoutes()
	routes["sync_info"] = server.NewRPCFunc(env.Header, "height")
}

func (env *ProvenanceEnvironment) GetSyncInfoAtBlock(ctx *cmtrpctypes.Context, height *int64) (*GetSyncInfo, error) {
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
