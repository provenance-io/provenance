package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

var _ FundTransferKeeper = (*BaseFundTransferKeeper)(nil)

type FundTransferKeeper interface {
	FundAsset(ctx context.Context, authorityAddr sdk.AccAddress, assetAddr sdk.AccAddress, amount sdk.Coin) error
	FundAssetByRegistry(ctx context.Context, authorityAddr sdk.AccAddress, assetAddr sdk.AccAddress, amount sdk.Coin) error
	TransferFunds(ctx context.Context, authorityAddr sdk.AccAddress, transfer *ledger.FundTransfer) error
	TransferFundsWithSettlement(ctx context.Context, authorityAddr sdk.AccAddress, transfer *ledger.FundTransferWithSettlement) error
	ValidateTransfer(ctx context.Context, authorityAddr sdk.AccAddress, transfer *ledger.FundTransfer) error
	GetTransferHistory(ctx context.Context, nftAddress string) ([]*ledger.FundTransfer, error)
}

type BaseFundTransferKeeper struct {
	BankKeeper
	BaseViewKeeper
}

// Fund an asset by transferring funds from the authority account to the asset account
func (k BaseFundTransferKeeper) FundAsset(ctx context.Context, authorityAddr sdk.AccAddress, assetAddr sdk.AccAddress, amount sdk.Coin) error {
	// TODO: Implement fund asset logic
	return nil
}

// Fund an asset based on the registered borrower in the registry
func (k BaseFundTransferKeeper) FundAssetByRegistry(ctx context.Context, authorityAddr sdk.AccAddress, assetAddr sdk.AccAddress, amount sdk.Coin) error {
	// TODO: Implement fund asset based on borrower logic
	return nil
}

// TransferFunds processes a fund transfer request
func (k BaseFundTransferKeeper) TransferFunds(ctx context.Context, authorityAddr sdk.AccAddress, transfer *ledger.FundTransfer) error {
	// TODO: Implement fund transfer logic
	return nil
}

// TransferFundsWithSettlement processes a fund transfer request with settlement instructions
func (k BaseFundTransferKeeper) TransferFundsWithSettlement(ctx context.Context, authorityAddr sdk.AccAddress, transfer *ledger.FundTransferWithSettlement) error {
	// Validate the transfer
	if transfer == nil {
		return errors.New("transfer cannot be nil")
	}
	if transfer.Key == nil {
		return errors.New("transfer key cannot be nil")
	}
	if len(transfer.SettlementInstructions) == 0 {
		return errors.New("settlement instructions cannot be empty")
	}

	// Validate the authority has permission to perform this transfer
	if err := k.ValidateTransfer(ctx, authorityAddr, &ledger.FundTransfer{
		Key:                      transfer.Key,
		LedgerEntryCorrelationId: transfer.LedgerEntryCorrelationId,
	}); err != nil {
		return fmt.Errorf("unauthorized transfer: %w", err)
	}

	// Get the ledger to ensure it exists
	ledger, err := k.GetLedger(sdk.UnwrapSDKContext(ctx), transfer.Key)
	if err != nil {
		return fmt.Errorf("failed to get ledger: %w", err)
	}
	if ledger == nil {
		return errors.New("ledger not found")
	}

	// Validate each settlement instruction
	totalAmount := sdk.NewCoin(transfer.SettlementInstructions[0].Amount.Denom, math.ZeroInt())
	for i, instruction := range transfer.SettlementInstructions {
		if instruction.Amount.Denom != totalAmount.Denom {
			return fmt.Errorf("mismatched denominations in settlement instructions: %s vs %s",
				instruction.Amount.Denom, totalAmount.Denom)
		}
		if !instruction.Amount.IsPositive() {
			return fmt.Errorf("settlement instruction %d has non-positive amount", i)
		}
		if instruction.RecipientAddress == "" {
			return fmt.Errorf("settlement instruction %d has empty recipient address", i)
		}
		totalAmount = totalAmount.Add(instruction.Amount)
	}

	// Store the transfer in the FundTransfersWithSettlement collection
	keyStr, err := LedgerKeyToString(transfer.Key)
	if err != nil {
		return fmt.Errorf("failed to convert ledger key to string: %w", err)
	}

	key := collections.Join(*keyStr, transfer.LedgerEntryCorrelationId)
	if err := k.FundTransfersWithSettlement.Set(sdk.UnwrapSDKContext(ctx), key, *transfer); err != nil {
		return fmt.Errorf("failed to store transfer: %w", err)
	}

	// Emit an event for the transfer
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"fund_transfer_with_settlement",
			sdk.NewAttribute("ledger_key", *keyStr),
			sdk.NewAttribute("correlation_id", transfer.LedgerEntryCorrelationId),
			sdk.NewAttribute("total_amount", totalAmount.String()),
			sdk.NewAttribute("num_instructions", fmt.Sprintf("%d", len(transfer.SettlementInstructions))),
		),
	)

	return nil
}

// ValidateTransfer validates if a fund transfer is allowed
func (k BaseFundTransferKeeper) ValidateTransfer(ctx context.Context, authorityAddr sdk.AccAddress, transfer *ledger.FundTransfer) error {
	// TODO: Implement transfer validation logic
	return nil
}

// GetTransferHistory returns the transfer history for an account
func (k BaseFundTransferKeeper) GetTransferHistory(ctx context.Context, nftAddress string) ([]*ledger.FundTransfer, error) {
	// TODO: Implement transfer history retrieval
	return nil, nil
}
