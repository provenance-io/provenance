package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"
	nft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/asset/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	registry "github.com/provenance-io/provenance/x/registry"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the asset MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

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

	// Verify the ledger class exists and it's asset class is the same as this class
	ledgerClass, err := m.ledgerKeeper.GetLedgerClass(ctx, msg.LedgerClass)
	if err != nil {
		return nil, fmt.Errorf("ledger class %s does not exist: %w", msg.LedgerClass, err)
	}
	if ledgerClass.AssetClassId != class.Id {
		return nil, fmt.Errorf("ledger class %s asset class id %s does not match asset class id %s", msg.LedgerClass, ledgerClass.AssetClassId, class.Id)
	}
	
	// Save the NFT class
	err = m.nftKeeper.SaveClass(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("failed to save NFT class: %w", err)
	}

	// Emit event for asset class creation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAssetClassCreated,
			sdk.NewAttribute(types.AttributeKeyAssetClassId, class.Id),
			sdk.NewAttribute(types.AttributeKeyAssetName, class.Name),
			sdk.NewAttribute(types.AttributeKeyAssetSymbol, class.Symbol),
			sdk.NewAttribute(types.AttributeKeyLedgerClass, msg.LedgerClass),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.FromAddress),
		),
	)

	m.Logger(ctx).Info("Created new asset class as NFT class",
		"class_id", class.Id,
		"name", class.Name)

	return &types.MsgCreateAssetClassResponse{}, nil
}

func (m msgServer) CreateAsset(goCtx context.Context, msg *types.MsgCreateAsset) (*types.MsgCreateAssetResponse, error) {
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
	owner, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	// Mint the NFT with the module account as owner
	err = m.nftKeeper.Mint(ctx, token, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to mint NFT: %w", err)
	}

	// Create a ledger for this asset
	ledgerKey := &ledger.LedgerKey{
		AssetClassId: msg.Asset.ClassId,
		NftId:        msg.Asset.Id,
	}

	ledgerObj := ledger.Ledger{
		Key:           ledgerKey,
		LedgerClassId: msg.Asset.ClassId,
		StatusTypeId:  1, // Using 1 as the default status type
	}

	// Create the ledger
	err = m.ledgerKeeper.CreateLedger(ctx, owner, ledgerObj)
	if err != nil {
		return nil, fmt.Errorf("failed to create ledger: %w", err)
	}

	// Create a default registry for the asset
	registryKey := &registry.RegistryKey{
		AssetClassId: msg.Asset.ClassId,
		NftId:        msg.Asset.Id,
	}

	err = m.registryKeeper.CreateDefaultRegistry(ctx, owner, registryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create default registry: %w", err)
	}

	// Emit event for asset creation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAssetCreated,
			sdk.NewAttribute(types.AttributeKeyAssetClassId, msg.Asset.ClassId),
			sdk.NewAttribute(types.AttributeKeyAssetId, msg.Asset.Id),
			sdk.NewAttribute(types.AttributeKeyOwner, owner.String()),
		),
	)

	m.Logger(ctx).Info("Created new asset as NFT",
		"class_id", token.ClassId,
		"token_id", token.Id,
		"owner", owner.String())

	return &types.MsgCreateAssetResponse{}, nil
}

// CreatePool creates a new pool marker
func (m msgServer) CreatePool(goCtx context.Context, msg *types.MsgCreatePool) (*types.MsgCreatePoolResponse, error) {

	// Create the marker
	marker, err := m.createMarker(goCtx, sdk.NewCoin(fmt.Sprintf("pool.%s", msg.Pool.Denom), msg.Pool.Amount), msg.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool marker: %w", err)
	}

	// Get the nfts
	for _, nft := range msg.Nfts {
		// Get the owner of the nft and verify it matches the from address
		owner := m.nftKeeper.GetOwner(goCtx, nft.ClassId, nft.Id)
		if owner.String() != msg.FromAddress {
			return nil, fmt.Errorf("nft class %s, id %s owner %s does not match from address %s", nft.ClassId, nft.Id, owner.String(), msg.FromAddress)
		}

		// Transfer the nft to the pool marker address
		err = m.nftKeeper.Transfer(goCtx, nft.ClassId, nft.Id, marker.GetAddress())
		if err != nil {
			return nil, fmt.Errorf("failed to transfer nft: %w", err)
		}
	}

	// Emit event for pool creation
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolDenom, msg.Pool.Denom),
			sdk.NewAttribute(types.AttributeKeyPoolAmount, msg.Pool.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNftCount, fmt.Sprintf("%d", len(msg.Nfts))),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.FromAddress),
		),
	)

	return &types.MsgCreatePoolResponse{}, nil
}

// CreateTokenization creates a new tokenization marker
func (m msgServer) CreateTokenization(goCtx context.Context, msg *types.MsgCreateTokenization) (*types.MsgCreateTokenizationResponse, error) {

	// Create the marker
	_, err := m.createMarker(goCtx, msg.Denom, msg.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenization marker: %w", err)
	}

	// Verify the NFT exists and is owned by the from address
	owner := m.nftKeeper.GetOwner(goCtx, msg.Nft.ClassId, msg.Nft.Id)
	if owner.String() != msg.FromAddress {
		return nil, fmt.Errorf("nft class %s, id %s owner %s does not match from address %s", msg.Nft.ClassId, msg.Nft.Id, owner.String(), msg.FromAddress)
	}

	// Emit event for tokenization creation
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTokenizationCreated,
			sdk.NewAttribute(types.AttributeKeyTokenizationDenom, msg.Denom.Denom),
			sdk.NewAttribute(types.AttributeKeyPoolAmount, msg.Denom.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyNftClassId, msg.Nft.ClassId),
			sdk.NewAttribute(types.AttributeKeyNftId, msg.Nft.Id),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.FromAddress),
		),
	)

	return &types.MsgCreateTokenizationResponse{}, nil
}

// CreateSecuritization creates a new securitization marker and tranches
func (m msgServer) CreateSecuritization(goCtx context.Context, msg *types.MsgCreateSecuritization) (*types.MsgCreateSecuritizationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Create the securitization marker
	_, err := m.createMarker(goCtx, sdk.NewCoin(fmt.Sprintf("sec.%s", msg.Id), sdkmath.NewInt(0)), msg.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create securitization marker: %w", err)
	}

	// Create the tranches
	for _, tranche := range msg.Tranches {
		_, err := m.createMarker(goCtx, sdk.NewCoin(fmt.Sprintf("sec.%s.tranche.%s", msg.Id, tranche.Denom), tranche.Amount), msg.FromAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create tranche marker: %w", err)
		}
	}

	// Reassign the pools permissions to the asset module account (prevent the pools from being transferred)
	for _, pool := range msg.Pools {
		pool, err := m.markerKeeper.GetMarkerByDenom(ctx, fmt.Sprintf("pool.%s", pool))
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
		accessList := pool.GetAccessList()
		for _, access := range accessList {
			accessAcc, err := sdk.AccAddressFromBech32(access.Address)
			if err != nil {
				return nil, fmt.Errorf("invalid from pool marker access address: %w", err)
			}
			err = pool.RevokeAccess(accessAcc)
			if err != nil {
				return nil, fmt.Errorf("failed to revoke access: %w", err)
			}
		}

		// Grant the module account access to the pool marker
		err = pool.GrantAccess(moduleAccessGrant)
		if err != nil {
			return nil, fmt.Errorf("failed to update pool marker access: %w", err)
		}

		// Save the updated marker
		m.markerKeeper.SetMarker(ctx, pool)
	}

	// Emit event for securitization creation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSecuritizationCreated,
			sdk.NewAttribute(types.AttributeKeySecuritizationId, msg.Id),
			sdk.NewAttribute(types.AttributeKeyTrancheCount, fmt.Sprintf("%d", len(msg.Tranches))),
			sdk.NewAttribute(types.AttributeKeyPoolCount, fmt.Sprintf("%d", len(msg.Pools))),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.FromAddress),
		),
	)

	return &types.MsgCreateSecuritizationResponse{}, nil
}

// CreatePool creates a new pool marker
func (m msgServer) createMarker(goCtx context.Context, denom sdk.Coin, fromAddr string) (*markertypes.MarkerAccount, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	marker, err := types.NewDefaultMarker(denom, fromAddr)
	if err != nil {
		return &markertypes.MarkerAccount{}, fmt.Errorf("failed to create marker: %w", err)
	}

	// Add the marker account by setting it
	err = m.Keeper.markerKeeper.AddFinalizeAndActivateMarker(ctx, marker)
	if err != nil {
		return &markertypes.MarkerAccount{}, fmt.Errorf("failed to add marker account: %w", err)
	}

	// Log the creation of the new pool marker
	ctx.Logger().Info("Created new pool marker", "pool_id", denom.Denom)

	return marker, nil
}
