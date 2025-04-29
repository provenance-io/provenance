package keeper

import (
	"context"
	"fmt"

	nft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		strMsg := &wrapperspb.StringValue{Value: msg.AssetClass.Data}
		anyMsg, err := cdctypes.NewAnyWithValue(strMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Any from data: %w", err)
		}
		class.Data = anyMsg

		// Validate the data is valid JSON schema
		_, err = types.AnyToJSONSchema(m.cdc, anyMsg)
		if err != nil {
			return nil, fmt.Errorf("invalid data: %w", err)
		}
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

		// Validate the data against the Class schema if it exists
		// otherwise it's an invalid Class id
		class, ok := m.nftKeeper.GetClass(ctx, msg.Asset.ClassId)
		if !ok {
			return nil, fmt.Errorf("asset class %s does not exist", msg.Asset.ClassId)
		}

		jsonSchema, err := types.AnyToJSONSchema(m.cdc, class.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to convert class data to JSON schema: %w", err)
		}

		// Validate the data against the JSON schema
		err = types.ValidateJSONSchema(jsonSchema, []byte(msg.Asset.Data))
		if err != nil {
			return nil, fmt.Errorf("invalid data: %w", err)
		}

		// Convert string to Any type
		strMsg := &wrapperspb.StringValue{Value: msg.Asset.Data}
		any, err := cdctypes.NewAnyWithValue(strMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to create Any from data: %w", err)
		}
		token.Data = any
	}

	// Get the asset module account address as the owner
	owner := sdk.AccAddress(msg.FromAddress)

	// Mint the NFT with the module account as owner
	err := m.nftKeeper.Mint(ctx, token, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to mint NFT: %w", err)
	}

	m.Logger(ctx).Info("Created new asset as NFT",
		"class_id", token.ClassId,
		"token_id", token.Id,
		"owner", owner.String())

	return &types.MsgAddAssetResponse{}, nil
}
