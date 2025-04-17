package keeper

import (
	"context"
	"fmt"

	nft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/x/asset/types"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the asset MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddAssetClass(goCtx context.Context, msg *types.MsgAddAssetClass) (*types.MsgAddAssetClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create NFT class from asset class
	class := nft.Class{
		Id:          msg.AssetClass.Id,
		Name:        msg.AssetClass.Name,
		Symbol:      msg.AssetClass.Symbol,
		Description: msg.AssetClass.Description,
		Uri:         msg.AssetClass.Uri,
		UriHash:     msg.AssetClass.UriHash,
	}

	// If there's data, add it to the class
	if msg.AssetClass.Data != "" {
		// Convert string to Any type
		strMsg := wrapperspb.String(msg.AssetClass.Data)
		any, err := cdctypes.NewAnyWithValue(strMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Any from data: %w", err)
		}
		class.Data = any
	}

	// Save the NFT class
	err := m.nftKeeper.SaveClass(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("failed to save NFT class: %w", err)
	}

	m.Logger(ctx).Info("Created new asset class as NFT class",
		"class_id", class.Id,
		"name", class.Name)

	return &types.MsgAddAssetClassResponse{}, nil
}

func (m msgServer) AddAsset(goCtx context.Context, msg *types.MsgAddAsset) (*types.MsgAddAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify the asset class exists
	if !m.nftKeeper.HasClass(ctx, msg.Asset.ClassId) {
		return nil, fmt.Errorf("asset class %s does not exist", msg.Asset.ClassId)
	}

	// Create NFT from asset
	token := nft.NFT{
		ClassId: msg.Asset.ClassId,
		Id:      msg.Asset.Id,
		Uri:     msg.Asset.Uri,
		UriHash: msg.Asset.UriHash,
	}

	// If there's data, add it to the token
	if msg.Asset.Data != "" {
		// Convert string to Any type
		strMsg := wrapperspb.String(msg.Asset.Data)
		any, err := cdctypes.NewAnyWithValue(strMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Any from data: %w", err)
		}
		token.Data = any
	}

	// Get the asset module account address as the owner
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	if moduleAddr == nil {
		return nil, fmt.Errorf("asset module account not found")
	}

	// Mint the NFT with the module account as owner
	err := m.nftKeeper.Mint(ctx, token, moduleAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to mint NFT: %w", err)
	}

	m.Logger(ctx).Info("Created new asset as NFT",
		"class_id", token.ClassId,
		"token_id", token.Id,
		"module_address", moduleAddr.String())

	return &types.MsgAddAssetResponse{}, nil
}
