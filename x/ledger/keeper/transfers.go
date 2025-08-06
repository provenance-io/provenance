package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// TransferFundsWithSettlement processes a fund transfer request with settlement instructions
func (k Keeper) ProcessTransferFundsWithSettlement(goCtx context.Context, authorityAddr sdk.AccAddress, transfer *types.FundTransferWithSettlement) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate that the ledger entry correlation id exists for the ledger
	_, err := k.RequireGetLedgerEntry(ctx, transfer.Key, transfer.LedgerEntryCorrelationId)
	if err != nil {
		return err
	}

	// Store the transfer in the FundTransfersWithSettlement collection
	keyStr := transfer.Key.String()

	existingSettlements, err := k.GetSettlements(ctx, &keyStr, transfer.LedgerEntryCorrelationId)
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
			return types.NewErrCodeInvalidField("recipient_address", "invalid recipient address")
		}

		// This has to be done prior to bank sends to avoid sending coins to a module account
		if k.BankKeeper.BlockedAddr(recipientAddr) {
			return types.NewErrCodeInvalidField("recipient_address", "bank blocked address")
		}

		if err := k.BankKeeper.SendCoins(ctx, authorityAddr, recipientAddr, sdk.NewCoins(inst.Amount)); err != nil {
			return err
		}

		// Set the xfer status to completed
		inst.Status = types.FundingTransferStatus_FUNDING_TRANSFER_STATUS_COMPLETED

		// Add the new transfer to the existing transfer list.
		existingSettlements.SettlementInstructions = append(existingSettlements.SettlementInstructions, inst)
	}

	sk := collections.Join(keyStr, transfer.LedgerEntryCorrelationId)
	if err := k.FundTransfersWithSettlement.Set(ctx, sk, *existingSettlements); err != nil {
		return types.NewErrCodeInternal("failed to store transfer")
	}

	// Emit an event for the transfer
	ctx.EventManager().EmitTypedEvent(types.NewEventFundTransferWithSettlement(transfer.Key, transfer.LedgerEntryCorrelationId))

	return nil
}

func (k Keeper) GetAllSettlements(ctx context.Context, keyStr *string) ([]*types.StoredSettlementInstructions, error) {
	prefix := collections.NewPrefixedPairRange[string, string](*keyStr)

	existingTransfers := make([]*types.StoredSettlementInstructions, 0)
	err := k.FundTransfersWithSettlement.Walk(ctx, prefix, func(key collections.Pair[string, string], value types.StoredSettlementInstructions) (stop bool, err error) {
		existingTransfers = append(existingTransfers, &value)
		return false, nil
	})
	if err != nil {
		return nil, err
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
