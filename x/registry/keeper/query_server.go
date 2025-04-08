package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/registry"
)

// QueryServer implements the gRPC querier service for the registry module
type QueryServer struct {
	keeper RegistryKeeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(keeper RegistryKeeper) *QueryServer {
	return &QueryServer{keeper: keeper}
}

// GetRegistryEntry returns a registry entry by address
func (qs QueryServer) GetRegistryEntry(ctx context.Context, req *registry.QueryGetRegistryEntryRequest) (*registry.QueryGetRegistryEntryResponse, error) {
	// TODO: Implement
	return &registry.QueryGetRegistryEntryResponse{}, nil
}

// ListRegistryEntries returns all registry entries
func (qs QueryServer) ListRegistryEntries(ctx context.Context, req *registry.QueryListRegistryEntriesRequest) (*registry.QueryListRegistryEntriesResponse, error) {
	// TODO: Implement
	return &registry.QueryListRegistryEntriesResponse{}, nil
}
