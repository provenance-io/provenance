package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/log"

	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// getFlatFeeForBulkImport queries the flat fee for MsgBulkImportRequest
func getFlatFeeForBulkImport(clientCtx client.Context) (*FlatFeeInfo, error) {
	queryClient := flatfeestypes.NewQueryClient(clientCtx)

	// Query the flat fee for the bulk import message type
	response, err := queryClient.MsgFee(
		context.Background(),
		&flatfeestypes.QueryMsgFeeRequest{
			MsgTypeUrl: "/provenance.ledger.v1.MsgBulkImportRequest",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query flat fee for bulk import: %w", err)
	}

	return &FlatFeeInfo{
		FeeAmount: response.MsgFee.Cost,
		MsgType:   response.MsgFee.MsgTypeUrl,
	}, nil
}

// getFlatFeeForChunk returns the flat fee for a chunk (same for all chunks)
func getFlatFeeForChunk(flatFeeInfo *FlatFeeInfo) sdk.Coins {
	if flatFeeInfo == nil {
		return nil
	}
	return flatFeeInfo.FeeAmount
}

// validateChunkWithFlatFees validates a chunk using flat fees and gas limits
func validateChunkWithFlatFees(chunk *types.GenesisState, gasUsed int, config ChunkConfig, logger log.Logger) error {
	// Validate chunk size against transaction limits
	chunkSize := getChunkSizeBytes(chunk)
	if chunkSize > config.MaxTxSizeBytes {
		return fmt.Errorf("chunk exceeds maximum transaction size: %d bytes > %d bytes", chunkSize, config.MaxTxSizeBytes)
	}

	// Validate chunk gas against gas limits (still needed for execution limits)
	if gasUsed > config.MaxGasPerTx-100000 { // 100k gas safety margin
		return fmt.Errorf("chunk exceeds maximum gas limit: %d gas > %d gas", gasUsed, config.MaxGasPerTx-100000)
	}

	logger.Info("Chunk validation passed",
		"size_bytes", chunkSize,
		"gas_used", gasUsed,
		"max_size_bytes", config.MaxTxSizeBytes,
		"max_gas", config.MaxGasPerTx-100000)

	return nil
}

// logFlatFeeInfo logs information about the flat fee being used
func logFlatFeeInfo(flatFeeInfo *FlatFeeInfo, logger log.Logger) {
	if flatFeeInfo == nil {
		logger.Warn("No flat fee information available")
		return
	}

	logger.Info("Using flat fee for bulk import",
		"msg_type", flatFeeInfo.MsgType,
		"fee_amount", flatFeeInfo.FeeAmount.String())
}
