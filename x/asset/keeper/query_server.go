package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	nfttypes "cosmossdk.io/x/nft"
	"github.com/provenance-io/provenance/x/asset/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the asset QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Asset queries for a specified asset by its class ID and asset ID.
func (q queryServer) Asset(ctx context.Context, req *types.QueryAssetRequest) (*types.QueryAssetResponse, error) {
	nftResp, err := q.nftKeeper.NFT(ctx, &nfttypes.QueryNFTRequest{ClassId: req.ClassId, Id: req.Id})
	if err != nil {
		return nil, err
	}
	n := nftResp.Nft
	asset := &types.Asset{ClassId: n.ClassId, Id: n.Id, Uri: n.Uri, UriHash: n.UriHash}
	if n.Data != nil {
		if strValue, err := types.AnyToString(q.cdc, n.Data); err == nil {
			asset.Data = strValue
		}
	}
	return &types.QueryAssetResponse{Asset: asset}, nil
}

// Assets queries all assets for a given address.
func (q queryServer) Assets(ctx context.Context, req *types.QueryAssetsRequest) (*types.QueryAssetsResponse, error) {
	if req.Owner != "" {
		if _, err := sdk.AccAddressFromBech32(req.Owner); err != nil {
			return nil, fmt.Errorf("invalid owner address: %w", err)
		}
	}
	// Delegate to NFT query server
	nftReq := &nfttypes.QueryNFTsRequest{Owner: req.Owner, ClassId: req.Id, Pagination: req.Pagination}
	nftResp, err := q.nftKeeper.NFTs(ctx, nftReq)
	if err != nil {
		return nil, err
	}
	assets := make([]*types.Asset, 0, len(nftResp.Nfts))
	for _, n := range nftResp.Nfts {
		asset := &types.Asset{ClassId: n.ClassId, Id: n.Id, Uri: n.Uri, UriHash: n.UriHash}
		if n.Data != nil {
			if strValue, err := types.AnyToString(q.cdc, n.Data); err == nil {
				asset.Data = strValue
			}
		}
		assets = append(assets, asset)
	}
	return &types.QueryAssetsResponse{Assets: assets, Pagination: nftResp.Pagination}, nil
}

// AssetClass queries a specific asset class by its ID.
func (q queryServer) AssetClass(ctx context.Context, req *types.QueryAssetClassRequest) (*types.QueryAssetClassResponse, error) {
	nftResp, err := q.nftKeeper.Class(ctx, &nfttypes.QueryClassRequest{ClassId: req.Id})
	if err != nil {
		return nil, err
	}
	c := nftResp.Class
	ac := &types.AssetClass{Id: c.Id, Name: c.Name, Symbol: c.Symbol, Description: c.Description, Uri: c.Uri, UriHash: c.UriHash}
	if c.Data != nil {
		if strValue, err := types.AnyToString(q.cdc, c.Data); err == nil {
			ac.Data = strValue
		}
	}
	return &types.QueryAssetClassResponse{Class: ac}, nil
}

// AssetClasses queries all asset classes.
func (q queryServer) AssetClasses(ctx context.Context, req *types.QueryAssetClassesRequest) (*types.QueryAssetClassesResponse, error) {
	nftResp, err := q.nftKeeper.Classes(ctx, &nfttypes.QueryClassesRequest{Pagination: req.Pagination})
	if err != nil {
		return nil, err
	}
	classes := make([]*types.AssetClass, 0, len(nftResp.Classes))
	for _, c := range nftResp.Classes {
		ac := &types.AssetClass{Id: c.Id, Name: c.Name, Symbol: c.Symbol, Description: c.Description, Uri: c.Uri, UriHash: c.UriHash}
		if c.Data != nil {
			if strValue, err := types.AnyToString(q.cdc, c.Data); err == nil {
				ac.Data = strValue
			}
		}
		classes = append(classes, ac)
	}
	return &types.QueryAssetClassesResponse{Classes: classes, Pagination: nftResp.Pagination}, nil
}
