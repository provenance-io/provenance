package cli

import (
	"encoding/json"
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
	file := statusFileName(status.ImportID)
	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, b, 0644)
}
