package types

import (
	"time"
)

// BulkImportStatus represents the status of a bulk import operation.
type BulkImportStatus struct {
	ImportID        string    `json:"import_id"`
	TotalChunks     uint32    `json:"total_chunks"`
	CompletedChunks uint32    `json:"completed_chunks"`
	TotalLedgers    uint32    `json:"total_ledgers"`
	TotalEntries    uint32    `json:"total_entries"`
	Status          string    `json:"status"` // "pending", "in_progress", "completed", "failed"
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Bulk import status constants
const (
	BulkImportStatusPending    = "pending"
	BulkImportStatusInProgress = "in_progress"
	BulkImportStatusCompleted  = "completed"
	BulkImportStatusFailed     = "failed"
)
