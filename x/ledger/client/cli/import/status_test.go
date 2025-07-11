package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatusFileName(t *testing.T) {
	importID := "test_import_123"
	filename := statusFileName(importID)
	expected := ".bulk_import_status.test_import_123.json"

	require.Equal(t, expected, filename, "Status filename should match expected format")
}

func TestWriteAndReadLocalBulkImportStatus(t *testing.T) {
	// Create a test status
	status := &LocalBulkImportStatus{
		ImportID:        "test_import_123",
		TotalChunks:     5,
		CompletedChunks: 2,
		TotalLedgers:    100,
		TotalEntries:    500,
		Status:          "in_progress",
		ErrorMessage:    "test error",
		CreatedAt:       "2024-01-01T12:00:00Z",
		UpdatedAt:       "2024-01-01T12:15:00Z",
	}

	// Write status to file
	err := writeLocalBulkImportStatus(status)
	require.NoError(t, err, "Should write status without error")

	// Clean up after test
	defer func() {
		filename := statusFileName(status.ImportID)
		os.Remove(filename)
	}()

	// Read status from file
	readStatus, err := readLocalBulkImportStatus(status.ImportID)
	require.NoError(t, err, "Should read status without error")
	require.NotNil(t, readStatus, "Read status should not be nil")

	// Verify all fields match
	require.Equal(t, status.ImportID, readStatus.ImportID)
	require.Equal(t, status.TotalChunks, readStatus.TotalChunks)
	require.Equal(t, status.CompletedChunks, readStatus.CompletedChunks)
	require.Equal(t, status.TotalLedgers, readStatus.TotalLedgers)
	require.Equal(t, status.TotalEntries, readStatus.TotalEntries)
	require.Equal(t, status.Status, readStatus.Status)
	require.Equal(t, status.ErrorMessage, readStatus.ErrorMessage)
	require.Equal(t, status.CreatedAt, readStatus.CreatedAt)
	require.Equal(t, status.UpdatedAt, readStatus.UpdatedAt)
}

func TestReadLocalBulkImportStatusFileNotFound(t *testing.T) {
	// Try to read a non-existent status file
	_, err := readLocalBulkImportStatus("non_existent_import")
	require.Error(t, err, "Should return error for non-existent file")
}

func TestWriteLocalBulkImportStatusNilStatus(t *testing.T) {
	// Test writing nil status (should return error)
	err := writeLocalBulkImportStatus(nil)
	require.Error(t, err, "Should return error when writing nil status")
	require.Contains(t, err.Error(), "cannot write nil status", "Error message should mention nil status")
}

func TestWriteLocalBulkImportStatusEmptyImportID(t *testing.T) {
	// Test writing status with empty import ID
	status := &LocalBulkImportStatus{
		ImportID:        "",
		TotalChunks:     5,
		CompletedChunks: 2,
		TotalLedgers:    100,
		TotalEntries:    500,
		Status:          "in_progress",
		CreatedAt:       "2024-01-01T12:00:00Z",
		UpdatedAt:       "2024-01-01T12:15:00Z",
	}

	err := writeLocalBulkImportStatus(status)
	require.NoError(t, err, "Should write status with empty import ID")

	// Clean up
	defer func() {
		filename := statusFileName(status.ImportID)
		os.Remove(filename)
	}()

	// Verify file was created
	filename := statusFileName(status.ImportID)
	_, err = os.Stat(filename)
	require.NoError(t, err, "File should exist")
}

func TestStatusFilePermissions(t *testing.T) {
	status := &LocalBulkImportStatus{
		ImportID:        "test_permissions",
		TotalChunks:     1,
		CompletedChunks: 0,
		TotalLedgers:    10,
		TotalEntries:    20,
		Status:          "pending",
		CreatedAt:       "2024-01-01T12:00:00Z",
		UpdatedAt:       "2024-01-01T12:00:00Z",
	}

	err := writeLocalBulkImportStatus(status)
	require.NoError(t, err, "Should write status without error")

	// Clean up
	defer func() {
		filename := statusFileName(status.ImportID)
		os.Remove(filename)
	}()

	// Check file permissions (should be 0644)
	filename := statusFileName(status.ImportID)
	fileInfo, err := os.Stat(filename)
	require.NoError(t, err, "Should be able to stat file")

	// Check that file is readable and writable by owner
	mode := fileInfo.Mode()
	require.True(t, mode.IsRegular(), "Should be a regular file")
	require.Equal(t, os.FileMode(0644), mode.Perm(), "Should have 0644 permissions")
}
