package cli

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

// ChunkConfig defines configuration for chunking large datasets
type ChunkConfig struct {
	MaxChunkSizeBytes int // Maximum chunk size in bytes (memory safety limit during parsing)
	MaxGasPerTx       int // Maximum gas consumption per transaction
	MaxTxSizeBytes    int // Maximum transaction size in bytes (blockchain limit)
}

// DefaultChunkConfig returns a reasonable default configuration
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		MaxChunkSizeBytes: 10000000, // 5MB per chunk (5x larger than max tx size for efficient gas optimization)
		MaxGasPerTx:       4000000,  // 4M gas per transaction (matching blockchain limit)
		MaxTxSizeBytes:    1000000,  // 1MB max transaction size (typical blockchain limit)
	}
}

// ChunkedGenesisState represents a chunked version of GenesisState
type ChunkedGenesisState struct {
	ImportID     string
	TotalChunks  int
	Chunks       []*types.GenesisState
	TotalLedgers int
	TotalEntries int
}

// ChunkStatus tracks the status of a single chunk
type ChunkStatus struct {
	FirstCorrelationID string `json:"first_correlation_id,omitempty"`
	LastCorrelationID  string `json:"last_correlation_id,omitempty"`
	Confirmed          bool   `json:"confirmed"`                  // Whether the transaction was confirmed on-chain
	TransactionHash    string `json:"transaction_hash,omitempty"` // Hash of the transaction for this chunk
}

// FlatFeeInfo represents the flat fee information for bulk import transactions
type FlatFeeInfo struct {
	FeeAmount sdk.Coins `json:"fee_amount"` // The flat fee amount for each chunk
	MsgType   string    `json:"msg_type"`   // The message type URL
}

// LocalBulkImportStatus tracks the status of a chunked bulk import on the client side.
type LocalBulkImportStatus struct {
	ImportID        string `json:"import_id"`
	TotalChunks     int    `json:"total_chunks"`
	CompletedChunks int    `json:"completed_chunks"`
	TotalLedgers    int    `json:"total_ledgers"`
	TotalEntries    int    `json:"total_entries"`
	Status          string `json:"status"` // "pending", "in_progress", "completed", "failed"
	ErrorMessage    string `json:"error_message,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	// Simple resume tracking
	LastSuccessfulCorrelationID string `json:"last_successful_correlation_id,omitempty"`
	FileHash                    string `json:"file_hash,omitempty"` // Hash of the source file for validation
	// Last attempted chunk status for resume safety
	LastAttemptedChunk *ChunkStatus `json:"last_attempted_chunk,omitempty"` // Status of the last chunk that was attempted
	// Stored flat fee info for deterministic chunking (can be reused on resume)
	FlatFeeInfo *FlatFeeInfo `json:"flat_fee_info,omitempty"`
	// Stored gas costs for gas limit validation (can be reused on resume)
	GasCosts *GasCosts `json:"gas_costs,omitempty"`
}

// ImportStats tracks import statistics
type ImportStats struct {
	TotalLedgers int
	TotalEntries int
	TotalChunks  int
}

// GasCosts represents the estimated gas costs for different components (for gas limit validation only)
type GasCosts struct {
	LedgerWithKeyGas int // Gas cost for a ledger key + ledger (base cost)
	EntryGas         int // Gas cost per entry
}

// generateImportID creates a unique identifier for an import operation
func generateImportID() string {
	// In a real implementation, you might want to use a more sophisticated ID generation
	// For now, we'll use a simple timestamp-based approach
	return fmt.Sprintf("import_%d", time.Now().UnixNano())
}
