package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

var _ FundTransferKeeper = (*BaseFundTransferKeeper)(nil)

type FundTransferKeeper interface {
	FundAsset(ctx context.Context, authorityAddr sdk.AccAddress, assetAddr sdk.AccAddress, amount sdk.Coin) error
	FundAssetByRegistry(ctx context.Context, authorityAddr sdk.AccAddress, assetAddr sdk.AccAddress, amount sdk.Coin) error
	TransferFundsWithSettlement(ctx context.Context, authorityAddr sdk.AccAddress, transfer *types.FundTransferWithSettlement) error
	GetTransferHistory(ctx context.Context, nftAddress string) ([]*types.FundTransferWithSettlement, error)
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

// TransferFundsWithSettlement processes a fund transfer request with settlement instructions
func (k BaseFundTransferKeeper) TransferFundsWithSettlement(goCtx context.Context, authorityAddr sdk.AccAddress, transfer *types.FundTransferWithSettlement) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := types.ValidateFundTransferWithSettlementBasic(transfer); err != nil {
		return err
	}

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

	// transfers := make([]*types.FundTransferWithSettlement, 0)
	// for _, inst := range transfer.SettlementInstructions {

	// 	transfers = append(transfers, &types.FundTransferWithSettlement{
	// 		Key:                      transfer.Key,
	// 		LedgerEntryCorrelationId: transfer.LedgerEntryCorrelationId,
	// 		Amount:                   instruction.Amount,
	// 		Memo:                     instruction.Memo,
	// 		SettlementInstructions:   transfer.SettlementInstructions,
	// 	})
	// }

	// key := collections.Join(*keyStr, transfer.LedgerEntryCorrelationId)
	// if err := k.FundTransfersWithSettlement.Set(sdk.UnwrapSDKContext(ctx), key, transfer.SettlementInstructions); err != nil {
	// 	return fmt.Errorf("failed to store transfer: %w", err)
	// }

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

// GetTransferHistory returns the transfer history for an account
func (k BaseFundTransferKeeper) GetTransferHistory(ctx context.Context, nftAddress string) ([]*types.FundTransferWithSettlement, error) {
	// TODO: Implement transfer history retrieval
	return nil, nil
}
