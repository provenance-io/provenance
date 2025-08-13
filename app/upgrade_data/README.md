# Upgrade Data Directory

This directory contains gzipped JSON data files that are embedded into the binary and loaded during blockchain upgrades.

## File Format

All data files must be gzipped JSON files with the following structure:

```json
{
  "ledgerToEntries": [
    {
      "ledgerKey": {
        "nftId": "string",
        "assetClassId": "string"
      },
      "ledger": {
        // Ledger object structure
      },
      "entries": [
        // Array of ledger entries
      ]
    }
  ]
}
```

## File Naming Convention

- Use descriptive names that indicate the upgrade and data type
- Include version information if applicable
- Use snake_case for file names
- **All files must have `.json.gz` extension**
- Examples:
  - `bouvardia_ledger_data.json.gz`
  - `bouvardia_entries_chunk1.json.gz`
  - `future_upgrade_placeholder.json.gz`

## Adding New Data Files

1. Create a new JSON file with the desired data
2. Compress it using gzip: `gzip -c your_file.json > your_file.json.gz`
3. Place the `.json.gz` file in this directory
4. Follow the naming convention above
5. Ensure the JSON structure matches the expected format
6. Test the data loading by running the upgrade handler
7. Update this README if adding new data types

## Validation

All data files are automatically validated during the upgrade process. The validation includes:

- JSON syntax validation
- Required field validation
- Data type validation
- Business logic validation

## File Size Guidelines

- **Small files (< 1MB uncompressed)**: Can be single files
- **Medium files (1-10MB uncompressed)**: Consider splitting into chunks
- **Large files (> 10MB uncompressed)**: Should be split into chunks
- **All files are automatically compressed** using gzip for efficient storage

## Compression

All files in this directory must be compressed using gzip:

```bash
# Compress a JSON file
gzip -c your_data.json > your_data.json.gz

# The loader automatically decompresses .json.gz files during upgrade
```

## Testing

To test data loading without running a full upgrade:

```go
func TestLoadLedgerData() {
    ledgers, err := loadLedgerDataFromFiles()
    if err != nil {
        t.Fatalf("Failed to load data: %v", err)
    }
    
    // Validate the loaded data
    for i, ledger := range ledgers {
        if err := ledger.Validate(); err != nil {
            t.Errorf("Invalid ledger at index %d: %v", i, err)
        }
    }
}
```

## Troubleshooting

### Common Issues

1. **File not found**: Ensure the file is in the correct directory and has a `.json.gz` extension
2. **JSON parsing errors**: Validate JSON syntax before compressing
3. **Compression errors**: Ensure files are properly gzipped
4. **Missing required fields**: Check that all required fields are present in the data
5. **Data validation errors**: Review the business logic validation rules



### Debugging

Enable debug logging to see detailed information about the data loading process:

```go
ctx.Logger().SetLevel("debug")
```

## Security Considerations

- All data files are embedded in the binary and cannot be modified at runtime
- Data is validated before being imported into the blockchain state
- No external file system access is required during upgrades 