package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// statusFileName returns the filename for storing bulk import status
func statusFileName(importID string) string {
	return ".bulk_import_status." + importID + ".json"
}

// readLocalBulkImportStatus reads the local bulk import status from a file
func readLocalBulkImportStatus(importID string) (*LocalBulkImportStatus, error) {
	file := statusFileName(importID)
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var status LocalBulkImportStatus
	if err := json.Unmarshal(b, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// writeLocalBulkImportStatus writes the local bulk import status to a file
func writeLocalBulkImportStatus(status *LocalBulkImportStatus) error {
	if status == nil {
		return fmt.Errorf("cannot write nil status")
	}

	file := statusFileName(status.ImportID)
	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Use atomic write to ensure the file is not corrupted if the process is interrupted
	// Write to a temporary file first, then rename to the final filename
	tempFile := file + ".tmp"
	err = os.WriteFile(tempFile, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write temporary status file: %w", err)
	}

	// Rename the temporary file to the final filename (atomic operation)
	err = os.Rename(tempFile, file)
	if err != nil {
		// Clean up the temporary file if rename fails
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to rename temporary status file: %w", err)
	}

	return nil
}
