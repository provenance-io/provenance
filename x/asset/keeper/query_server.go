package keeper

import (
	"context"
	"fmt"

	nfttypes "cosmossdk.io/x/nft"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/asset/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the asset QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Asset gets a specified asset by its class ID and asset ID.
func (q queryServer) Asset(ctx context.Context, req *types.QueryAssetRequest) (*types.QueryAssetResponse, error) {
	nftResp, err := q.nftKeeper.NFT(ctx, &nfttypes.QueryNFTRequest{ClassId: req.ClassId, Id: req.Id})
	if err != nil {
		return nil, types.NewErrCodeInternal(fmt.Sprintf("failed to get asset: %s", err))
	}
	n := nftResp.Nft
	asset := &types.Asset{ClassId: n.ClassId, Id: n.Id, Uri: n.Uri, UriHash: n.UriHash}
	if n.Data != nil {
		asset.Data, err = types.AnyToString(q.cdc, n.Data)
		if err != nil {
			asset.Data = fmt.Sprintf("error reading nft data: %q", err.Error())
		}
	}
	return &types.QueryAssetResponse{Asset: asset}, nil
}

// Assets gets all assets for a given address and class.
func (q queryServer) Assets(ctx context.Context, req *types.QueryAssetsRequest) (*types.QueryAssetsResponse, error) {
	if req.Owner != "" {
		if _, err := sdk.AccAddressFromBech32(req.Owner); err != nil {
			return nil, types.NewErrCodeInvalidField("owner", "%s", err)
		}
	}
	// Delegate to NFT query server
	nftReq := &nfttypes.QueryNFTsRequest{Owner: req.Owner, ClassId: req.ClassId, Pagination: req.Pagination}
	nftResp, err := q.nftKeeper.NFTs(ctx, nftReq)
	if err != nil {
		return nil, types.NewErrCodeInternal(fmt.Sprintf("failed to get assets: %s", err))
	}
	assets := make([]*types.Asset, 0, len(nftResp.Nfts))
	for _, n := range nftResp.Nfts {
		asset := &types.Asset{ClassId: n.ClassId, Id: n.Id, Uri: n.Uri, UriHash: n.UriHash}
		if n.Data != nil {
			asset.Data, err = types.AnyToString(q.cdc, n.Data)
			if err != nil {
				asset.Data = fmt.Sprintf("error reading nft data: %q", err.Error())
			}
		}
		assets = append(assets, asset)
	}
	return &types.QueryAssetsResponse{Assets: assets, Pagination: nftResp.Pagination}, nil
}

// AssetClass gets a specific asset class by its ID.
func (q queryServer) AssetClass(ctx context.Context, req *types.QueryAssetClassRequest) (*types.QueryAssetClassResponse, error) {
	nftResp, err := q.nftKeeper.Class(ctx, &nfttypes.QueryClassRequest{ClassId: req.Id})
	if err != nil {
		return nil, types.NewErrCodeInternal(fmt.Sprintf("failed to get asset class: %s", err))
	}
	c := nftResp.Class
	ac := &types.AssetClass{Id: c.Id, Name: c.Name, Symbol: c.Symbol, Description: c.Description, Uri: c.Uri, UriHash: c.UriHash}
	if c.Data != nil {
		ac.Data, err = types.AnyToString(q.cdc, c.Data)
		if err != nil {
			ac.Data = fmt.Sprintf("error reading class data: %q", err.Error())
		}
	}
	return &types.QueryAssetClassResponse{AssetClass: ac}, nil
}

// AssetClasses gets all asset classes.
func (q queryServer) AssetClasses(ctx context.Context, req *types.QueryAssetClassesRequest) (*types.QueryAssetClassesResponse, error) {
	nftResp, err := q.nftKeeper.Classes(ctx, &nfttypes.QueryClassesRequest{Pagination: req.Pagination})
	if err != nil {
		return nil, types.NewErrCodeInternal(fmt.Sprintf("failed to get asset classes: %s", err))
	}
	classes := make([]*types.AssetClass, 0, len(nftResp.Classes))
	for _, c := range nftResp.Classes {
		ac := &types.AssetClass{Id: c.Id, Name: c.Name, Symbol: c.Symbol, Description: c.Description, Uri: c.Uri, UriHash: c.UriHash}
		if c.Data != nil {
			ac.Data, err = types.AnyToString(q.cdc, c.Data)
			if err != nil {
				ac.Data = fmt.Sprintf("error reading class data: %q", err.Error())
			}
		}
		classes = append(classes, ac)
	}
	return &types.QueryAssetClassesResponse{AssetClasses: classes, Pagination: nftResp.Pagination}, nil
}
