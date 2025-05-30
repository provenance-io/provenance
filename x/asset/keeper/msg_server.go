package keeper

import (
	"context"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	nft "cosmossdk.io/x/nft"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/x/asset/types"
	ledger "github.com/provenance-io/provenance/x/ledger"
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

	// Get the asset module account address as the owner
	owner, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid owner address: %w", err)
	}

	// Create a ledger class for this asset class if it doesn't exist
	ledgerClassId := fmt.Sprintf("ledgert_%s", msg.AssetClass.Id)
	ledgerClass := ledger.LedgerClass{
		LedgerClassId:     ledgerClassId,
		AssetClassId:      msg.AssetClass.Id,
		Denom:             "nhash", // Using nhash as the default denom
		MaintainerAddress: owner.String(),
	}

	// Create the ledger class
	err = m.ledgerKeeper.CreateLedgerClass(ctx, owner, ledgerClass)
	if err != nil {
		// If the error is not that the class already exists, return the error
		return nil, fmt.Errorf("failed to create ledger class: %w", err)
	}

	// Add provided entry types, or default if none provided
	if len(msg.EntryTypes) > 0 {
		for _, entryType := range msg.EntryTypes {
			err = m.ledgerKeeper.AddClassEntryType(ctx, owner, ledgerClassId, *entryType)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return nil, fmt.Errorf("failed to add ledger class entry type: %w", err)
			}
		}
	} else {
		entryType := ledger.LedgerClassEntryType{
			Id:          1,
			Code:        "DEFAULT",
			Description: "Default entry type for asset ledger",
		}
		err = m.ledgerKeeper.AddClassEntryType(ctx, owner, ledgerClassId, entryType)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("failed to add ledger class entry type: %w", err)
		}
	}

	// Add provided status types, or default if none provided
	if len(msg.StatusTypes) > 0 {
		for _, statusType := range msg.StatusTypes {
			err = m.ledgerKeeper.AddClassStatusType(ctx, owner, ledgerClassId, *statusType)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				return nil, fmt.Errorf("failed to add ledger class status type: %w", err)
			}
		}
	} else {
		statusType := ledger.LedgerClassStatusType{
			Id:          1,
			Code:        "ACTIVE",
			Description: "Active status for asset ledger",
		}
		err = m.ledgerKeeper.AddClassStatusType(ctx, owner, ledgerClassId, statusType)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("failed to add ledger class status type: %w", err)
		}
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

	ledgerClassId := fmt.Sprintf("ledgert_%s", msg.Asset.ClassId)
	ledgerObj := ledger.Ledger{
		Key:           ledgerKey,
		LedgerClassId: ledgerClassId,
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

	m.Logger(ctx).Info("Created new asset as NFT",
		"class_id", token.ClassId,
		"token_id", token.Id,
		"owner", owner.String())

	return &types.MsgAddAssetResponse{}, nil
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
	
	return &types.MsgCreatePoolResponse{}, nil
}

// CreateParticipation creates a new participation marker
func (m msgServer) CreateParticipation(goCtx context.Context, msg *types.MsgCreateParticipation) (*types.MsgCreateParticipationResponse, error) {

	// Create the marker
	_, err := m.createMarker(goCtx, msg.Denom, msg.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create participation marker: %w", err)
	}

	return &types.MsgCreateParticipationResponse{}, nil
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

	return &types.MsgCreateSecuritizationResponse{}, nil
}

// CreatePool creates a new pool marker
func (m msgServer) createMarker(goCtx context.Context, denom sdk.Coin, fromAddr string) (*markertypes.MarkerAccount, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get the from address
	fromAcc, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return &markertypes.MarkerAccount{}, fmt.Errorf("invalid from address: %w", err)
	}

	// Create a new marker account
	markerAddr := markertypes.MustGetMarkerAddress(denom.Denom)
	marker := markertypes.NewMarkerAccount(
		authtypes.NewBaseAccountWithAddress(markerAddr),
		denom,
		fromAcc,
		[]markertypes.AccessGrant{
			{
				Address: fromAcc.String(),
				Permissions: markertypes.AccessList{
					markertypes.Access_Admin,
					markertypes.Access_Mint,
					markertypes.Access_Burn,
					markertypes.Access_Withdraw,
					markertypes.Access_Transfer,
				},
			},
		},
		markertypes.StatusProposed,
		markertypes.MarkerType_RestrictedCoin,
		true,       // Supply fixed
		false,      // Allow governance control
		false,      // Don't allow forced transfer
		[]string{}, // No required attributes
	)

	// Add the marker account by setting it
	err = m.Keeper.markerKeeper.AddFinalizeAndActivateMarker(ctx, marker)
	if err != nil {
		return &markertypes.MarkerAccount{}, fmt.Errorf("failed to add marker account: %w", err)
	}

	// Log the creation of the new pool marker
	ctx.Logger().Info("Created new pool marker", "pool_id", denom.Denom)

	return marker, nil
}

