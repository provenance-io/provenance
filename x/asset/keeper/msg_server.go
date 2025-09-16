package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	nft "cosmossdk.io/x/nft"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/asset/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the asset MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// BurnAsset burns an NFT and removes its registry for the asset.
func (m msgServer) BurnAsset(goCtx context.Context, msg *types.MsgBurnAsset) (*types.MsgBurnAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify the asset exists and get the current owner
	ownerResp, err := m.nftKeeper.Owner(goCtx, &nft.QueryOwnerRequest{ClassId: msg.Asset.ClassId, Id: msg.Asset.Id})
	if err != nil {
		return nil, fmt.Errorf("asset does not exist: %w", err)
	}

	// Verify the signer is the current owner of the asset
	if msg.Signer != ownerResp.Owner {
		return nil, fmt.Errorf("signer %s is not the owner of asset %s/%s, current owner: %s",
			msg.Signer, msg.Asset.ClassId, msg.Asset.Id, ownerResp.Owner)
	}

	// Burn the NFT using the nft module
	err = m.nftKeeper.Burn(ctx, msg.Asset.ClassId, msg.Asset.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to burn NFT: %w", err)
	}

	// Note: Registry entries are preserved after asset burn for historical/audit purposes

	// Emit event for asset burn
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventAssetBurned(msg.Asset.ClassId, msg.Asset.Id, msg.Signer)); err != nil {
		return nil, fmt.Errorf("failed to emit asset burned event: %w", err)
	}

	return &types.MsgBurnAssetResponse{}, nil
}

// CreateAsset creates an NFT and a default registry for the asset and validates the data against the class schema.
func (m msgServer) CreateAsset(goCtx context.Context, msg *types.MsgCreateAsset) (*types.MsgCreateAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify the asset class exists
	classResp, err := m.nftKeeper.Class(ctx, &nft.QueryClassRequest{ClassId: msg.Asset.ClassId})
	if err != nil {
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
		if classResp.Class.Data != nil {
			jsonSchema, err := types.AnyToJSONSchema(m.cdc, classResp.Class.Data)
			if err != nil {
				return nil, fmt.Errorf("failed to convert class data to JSON schema: %w", err)
			}

			// Validate the data against the JSON schema
			err = types.ValidateJSONSchema(jsonSchema, []byte(msg.Asset.Data))
			if err != nil {
				return nil, fmt.Errorf("invalid data: %w", err)
			}
		}

		// Convert string to Any type
		anyValue, err := types.StringToAny(msg.Asset.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to create Any from data: %w", err)
		}
		token.Data = anyValue
	}

	// Get the owner address
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	// Mint the NFT with the owner address
	err = m.nftKeeper.Mint(ctx, token, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to mint NFT: %w", err)
	}

	// Create a default registry for the asset
	registryKey := &registrytypes.RegistryKey{
		AssetClassId: msg.Asset.ClassId,
		NftId:        msg.Asset.Id,
	}

	err = m.registryKeeper.CreateDefaultRegistry(ctx, owner.String(), registryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create default registry: %w", err)
	}

	// Emit event for asset creation
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventAssetCreated(msg.Asset.ClassId, msg.Asset.Id, owner.String())); err != nil {
		return nil, fmt.Errorf("failed to emit asset created event: %w", err)
	}

	return &types.MsgCreateAssetResponse{}, nil
}

// CreateAssetClass creates an NFT class and validates the json schema data field.
func (m msgServer) CreateAssetClass(goCtx context.Context, msg *types.MsgCreateAssetClass) (*types.MsgCreateAssetClassResponse, error) {
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
		anyMsg, err := types.StringToAny(msg.AssetClass.Data)
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

	// Emit event for asset class creation
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventAssetClassCreated(class.Id, class.Name, class.Symbol)); err != nil {
		return nil, fmt.Errorf("failed to emit asset class created event: %w", err)
	}

	return &types.MsgCreateAssetClassResponse{}, nil
}

// CreatePool creates a marker for the pool and transfers the assets to the pool marker.
func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	denom := fmt.Sprintf("pool.%s", msg.Pool.Denom)

	// Ensure the pool marker doesn't already exist
	if _, err := m.markerKeeper.GetMarkerByDenom(ctx, denom); err == nil {
		return nil, fmt.Errorf("pool marker with denom %s already exists", denom)
	}

	// Create the marker
	marker, err := m.createMarker(goCtx, sdk.NewCoin(denom, msg.Pool.Amount), msg.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool marker: %w", err)
	}

	// Get the nfts and count them
	var assetCount uint32
	for _, asset := range msg.Assets {
		// Get the owner of the nft and verify it matches the from address
		ownerResp, err := m.nftKeeper.Owner(goCtx, &nft.QueryOwnerRequest{ClassId: asset.ClassId, Id: asset.Id})
		if err != nil {
			return nil, fmt.Errorf("failed to get owner of asset: %w", err)
		}
		if ownerResp.Owner != msg.Signer {
			return nil, fmt.Errorf("asset class %s, id %s owner %s does not match from address %s", asset.ClassId, asset.Id, ownerResp.Owner, msg.Signer)
		}

		// Transfer the nft to the pool marker address
		err = m.nftKeeper.Transfer(goCtx, asset.ClassId, asset.Id, marker.GetAddress())
		if err != nil {
			return nil, fmt.Errorf("failed to transfer nft: %w", err)
		}
		assetCount++
	}

	// Emit event for pool creation
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventPoolCreated(msg.Pool.String(), assetCount, msg.Signer)); err != nil {
		return nil, fmt.Errorf("failed to emit pool created event: %w", err)
	}

	return &types.MsgCreatePoolResponse{}, nil
}

// CreateTokenization creates a marker for a tokenization and transfers the asset to the tokenization marker.
func (m msgServer) CreateTokenization(goCtx context.Context, msg *types.MsgCreateTokenization) (*types.MsgCreateTokenizationResponse, error) {
	// Create the marker
	marker, err := m.createMarker(goCtx, msg.Token, msg.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenization marker: %w", err)
	}

	// Verify the Asset exists and is owned by the from address
	ownerResp, err := m.nftKeeper.Owner(goCtx, &nft.QueryOwnerRequest{ClassId: msg.Asset.ClassId, Id: msg.Asset.Id})
	if err != nil {
		return nil, fmt.Errorf("failed to get owner of asset: %w", err)
	}
	if ownerResp.Owner != msg.Signer {
		return nil, fmt.Errorf("asset class %s, id %s owner %s does not match from address %s", msg.Asset.ClassId, msg.Asset.Id, ownerResp.Owner, msg.Signer)
	}

	// Transfer the Asset to the tokenization marker address
	err = m.nftKeeper.Transfer(goCtx, msg.Asset.ClassId, msg.Asset.Id, marker.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to transfer asset: %w", err)
	}

	// Emit event for tokenization creation
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventTokenizationCreated(msg.Token.String(), msg.Asset.ClassId, msg.Asset.Id, msg.Signer)); err != nil {
		return nil, fmt.Errorf("failed to emit tokenization created event: %w", err)
	}

	return &types.MsgCreateTokenizationResponse{}, nil
}

// CreateSecuritization creates markers for the securitization and tranches and transfers the assets to the securitization marker.
func (m msgServer) CreateSecuritization(goCtx context.Context, msg *types.MsgCreateSecuritization) (*types.MsgCreateSecuritizationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create the securitization marker
	_, err := m.createMarker(goCtx, sdk.NewCoin(fmt.Sprintf("sec.%s", msg.Id), sdkmath.NewInt(0)), msg.Signer)
	if err != nil {
		return nil, fmt.Errorf("failed to create securitization marker: %w", err)
	}

	// Create the tranches and count them
	var trancheCount uint32
	for _, tranche := range msg.Tranches {
		_, err := m.createMarker(goCtx, sdk.NewCoin(fmt.Sprintf("sec.%s.tranche.%s", msg.Id, tranche.Denom), tranche.Amount), msg.Signer)
		if err != nil {
			return nil, fmt.Errorf("failed to create tranche marker: %w", err)
		}
		trancheCount++
	}

	// Reassign the pools permissions to the asset module account (prevent the pools from being transferred)
	var poolCount uint32
	for _, pool := range msg.Pools {
		poolMarker, err := m.markerKeeper.GetMarkerByDenom(ctx, fmt.Sprintf("pool.%s", pool))
		if err != nil {
			return nil, fmt.Errorf("failed to get pool marker: %w", err)
		}

		// Create a new access grant with the desired permissions
		moduleAccessGrant := markertypes.NewAccessGrant(
			m.GetModuleAddress(),
			[]markertypes.Access{
				markertypes.Access_Admin,
				markertypes.Access_Mint,
				markertypes.Access_Burn,
				markertypes.Access_Withdraw,
				markertypes.Access_Transfer,
			},
		)

		// Revoke all access from the pool marker
		accessList := poolMarker.GetAccessList()
		for _, access := range accessList {
			accessAcc, err := sdk.AccAddressFromBech32(access.Address)
			if err != nil {
				return nil, fmt.Errorf("invalid from pool marker access address: %w", err)
			}
			err = poolMarker.RevokeAccess(accessAcc)
			if err != nil {
				return nil, fmt.Errorf("failed to revoke access: %w", err)
			}
		}

		// Grant the module account access to the pool marker
		err = poolMarker.GrantAccess(moduleAccessGrant)
		if err != nil {
			return nil, fmt.Errorf("failed to update pool marker access: %w", err)
		}

		// Save the updated marker
		m.markerKeeper.SetMarker(ctx, poolMarker)
		poolCount++
	}

	// Emit event for securitization creation
	if err := ctx.EventManager().EmitTypedEvent(types.NewEventSecuritizationCreated(msg.Id, trancheCount, poolCount, msg.Signer)); err != nil {
		return nil, fmt.Errorf("failed to emit securitization created event: %w", err)
	}

	return &types.MsgCreateSecuritizationResponse{}, nil
}

// createMarker creates a new marker. It creates a marker for the token and address.
func (m msgServer) createMarker(goCtx context.Context, token sdk.Coin, addr string) (*markertypes.MarkerAccount, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	marker, err := types.NewDefaultMarker(token, addr)
	if err != nil {
		return &markertypes.MarkerAccount{}, fmt.Errorf("failed to create marker: %w", err)
	}

	// Add the marker account by setting it
	err = m.Keeper.markerKeeper.AddFinalizeAndActivateMarker(ctx, marker)
	if err != nil {
		return &markertypes.MarkerAccount{}, fmt.Errorf("failed to add marker account: %w", err)
	}

	return marker, nil
}
