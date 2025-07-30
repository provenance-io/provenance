package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// TransferFundsWithSettlement processes a fund transfer request with settlement instructions
func (k Keeper) ProcessTransferFundsWithSettlement(goCtx context.Context, authorityAddr sdk.AccAddress, transfer *types.FundTransferWithSettlement) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// print the transfer key as json
	transferKeyJSON, err := json.MarshalIndent(transfer, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transfer key to JSON: %w", err)
	}
	fmt.Println(string(transferKeyJSON))

	// Get the ledger to ensure it exists
	ledger, err := k.GetLedger(ctx, transfer.Key)
	if err != nil {
		return fmt.Errorf("failed to get ledger: %w", err)
	}
	if ledger == nil {
		return errors.New("ledger not found")
	}

	// Validate that the ledger entry correlation id exists for the ledger
	ledgerEntry, err := k.GetLedgerEntry(ctx, transfer.Key, transfer.LedgerEntryCorrelationId)
	if err != nil {
		return fmt.Errorf("failed to get ledger entry: %w", err)
	}
	if ledgerEntry == nil {
		return errors.New("ledger entry not found")
	}

	// Store the transfer in the FundTransfersWithSettlement collection
	keyStr, err := LedgerKeyToString(transfer.Key)
	if err != nil {
		return fmt.Errorf("failed to convert ledger key to string: %w", err)
	}

	existingSettlements, err := k.GetSettlements(ctx, keyStr, transfer.LedgerEntryCorrelationId)
	if err != nil {
		return err
	}

	if existingSettlements == nil {
		existingSettlements = &types.StoredSettlementInstructions{}
	}

	// Transfer funds per the settlement instructions
	for _, inst := range transfer.SettlementInstructions {
		recipientAddr, err := sdk.AccAddressFromBech32(inst.RecipientAddress)
		if err != nil {
			return fmt.Errorf("failed to convert recipient address to bech32: %w", err)
		}

		if err := k.BankKeeper.SendCoins(ctx, authorityAddr, recipientAddr, sdk.NewCoins(inst.Amount)); err != nil {
			return fmt.Errorf("failed to send coins: %w", err)
		}

		// Set the xfer status to completed
		inst.Status = types.FundingTransferStatus_FUNDING_TRANSFER_STATUS_COMPLETED

		// Add the new transfer to the existing transfer list.
		existingSettlements.SettlementInstructions = append(existingSettlements.SettlementInstructions, inst)
	}

	sk := collections.Join(*keyStr, transfer.LedgerEntryCorrelationId)
	if err := k.FundTransfersWithSettlement.Set(ctx, sk, *existingSettlements); err != nil {
		return fmt.Errorf("failed to store transfer: %w", err)
	}

	// Emit an event for the transfer
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fund_transfer_with_settlement",
			sdk.NewAttribute("ledger_key", *keyStr),
			sdk.NewAttribute("correlation_id", transfer.LedgerEntryCorrelationId),
			// sdk.NewAttribute("total_amount", transfer.Amount.String()),
			sdk.NewAttribute("num_instructions", fmt.Sprintf("%d", len(transfer.SettlementInstructions))),
		),
	)

	return nil
}

func (k Keeper) GetAllSettlements(ctx context.Context, keyStr *string) ([]*types.StoredSettlementInstructions, error) {
	prefix := collections.NewPrefixedPairRange[string, string](*keyStr)
	iter, err := k.FundTransfersWithSettlement.Iterate(ctx, prefix)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	existingTransfers := make([]*types.StoredSettlementInstructions, 0)
	for ; iter.Valid(); iter.Next() {
		transfer, err := iter.Value()
		if err != nil {
			return nil, err
		}

		existingTransfers = append(existingTransfers, &transfer)
	}

	return existingTransfers, nil
}

func (k Keeper) GetSettlements(ctx context.Context, keyStr *string, correlationId string) (*types.StoredSettlementInstructions, error) {
	searchKey := collections.Join(*keyStr, correlationId)
	settlements, err := k.FundTransfersWithSettlement.Get(ctx, searchKey)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &settlements, nil
}
