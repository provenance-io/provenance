package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/types"
)

// ProcessTransferFundsWithSettlement processes a fund transfer request with settlement instructions.
// This function executes bank transfers, and stores settlement records.
// It ensures that funds are transferred according to settlement instructions while maintaining proper state tracking.
func (k Keeper) ProcessTransferFundsWithSettlement(goCtx context.Context, authorityAddr sdk.AccAddress, transfer *types.FundTransferWithSettlement) error {
	// Convert the context to SDK context for state operations.
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Store the transfer in the FundTransfersWithSettlement collection.
	// This collection tracks all settlement instructions for fund transfers.
	keyStr := transfer.Key.String()

	// Retrieve existing settlements for this ledger entry to append new instructions.
	existingSettlements, err := k.GetSettlements(ctx, &keyStr, transfer.LedgerEntryCorrelationId)
	if err != nil {
		return err
	}

	// Initialize empty settlement instructions if none exist yet.
	if existingSettlements == nil {
		existingSettlements = &types.StoredSettlementInstructions{}
	}

	// Process each settlement instruction in the transfer request.
	// This loop executes the actual fund transfers and updates settlement status.
	for _, inst := range transfer.SettlementInstructions {
		// Parse the recipient address from bech32 format.
		recipientAddr, err := sdk.AccAddressFromBech32(inst.RecipientAddress)
		if err != nil {
			return types.NewErrCodeInvalidField("recipient_address", "invalid recipient address")
		}

		// Check if the recipient address is blocked by the bank module.
		// This prevents sending coins to module accounts or other blocked addresses.
		if k.BankKeeper.BlockedAddr(recipientAddr) {
			return types.NewErrCodeInvalidField("recipient_address", "bank blocked address")
		}

		// Execute the actual bank transfer from authority to recipient.
		// This moves the specified amount of coins between accounts.
		if err := k.BankKeeper.SendCoins(ctx, authorityAddr, recipientAddr, sdk.NewCoins(inst.Amount)); err != nil {
			return err
		}

		// Mark the transfer as completed in the settlement instruction.
		// This tracks the status of each individual transfer within the settlement.
		inst.Status = types.FundingTransferStatus_FUNDING_TRANSFER_STATUS_COMPLETED

		// Add the new transfer to the existing transfer list.
		// This maintains a complete history of all settlements for this ledger entry.
		existingSettlements.SettlementInstructions = append(existingSettlements.SettlementInstructions, inst)
	}

	// Store the updated settlement instructions in the state store.
	// This preserves the complete settlement history for future reference.
	sk := collections.Join(keyStr, transfer.LedgerEntryCorrelationId)
	if err := k.FundTransfersWithSettlement.Set(ctx, sk, *existingSettlements); err != nil {
		return types.NewErrCodeInternal("failed to store transfer")
	}

	// Emit an event to notify other modules of the completed transfer.
	// This allows for proper event handling and external integrations.
	ctx.EventManager().EmitTypedEvent(types.NewEventFundTransferWithSettlement(transfer.Key, transfer.LedgerEntryCorrelationId))

	return nil
}

// GetAllSettlements retrieves all settlement instructions for a given ledger.
// This function walks through all settlement records associated with a specific ledger key.
// It provides a complete view of all fund transfers and settlements for the ledger.
func (k Keeper) GetAllSettlements(ctx context.Context, keyStr *string) ([]*types.StoredSettlementInstructions, error) {
	// Create a prefix range to find all settlements for this ledger.
	prefix := collections.NewPrefixedPairRange[string, string](*keyStr)

	// Initialize a slice to collect all settlement instructions.
	existingTransfers := make([]*types.StoredSettlementInstructions, 0)

	// Walk through all settlement records that match the ledger prefix.
	// This collects all settlement instructions for the specified ledger.
	err := k.FundTransfersWithSettlement.Walk(ctx, prefix, func(key collections.Pair[string, string], value types.StoredSettlementInstructions) (stop bool, err error) {
		existingTransfers = append(existingTransfers, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return existingTransfers, nil
}

// GetSettlements retrieves settlement instructions for a specific ledger entry.
// This function looks up settlement records by ledger key and correlation ID.
// It returns nil if no settlements exist for the specified entry.
func (k Keeper) GetSettlements(ctx context.Context, keyStr *string, correlationId string) (*types.StoredSettlementInstructions, error) {
	// Create the composite key for the specific settlement record.
	searchKey := collections.Join(*keyStr, correlationId)

	// Retrieve the settlement instructions from the state store.
	settlements, err := k.FundTransfersWithSettlement.Get(ctx, searchKey)
	if err != nil {
		// Return nil if no settlements are found for this entry.
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &settlements, nil
}
