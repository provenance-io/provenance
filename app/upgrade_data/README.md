# Upgrade Data Directory

This directory contains gzipped JSON data files that are embedded into the binary and loaded during blockchain upgrades.

## File Format

All data files must be gzipped JSON files containing a `ledgerTypes.GenesisState` object with the following structure:

```json
{
  "ledgerClasses": [
    // Array of ledger class definitions
  ],
  "ledgerClassEntryTypes": [
    // Array of ledger class entry type definitions
  ],
  "ledgerClassStatusTypes": [
    // Array of ledger class status type definitions
  ],
  "ledgerClassBucketTypes": [
    // Array of ledger class bucket type definitions
  ],
  "ledgers": [
    // Array of ledger definitions
  ],
  "ledgerEntries": [
    // Array of ledger entries
  ],
  "settlementInstructions": [
    // Array of settlement instructions
  ]
}
```

## File Naming Convention

- The file must be named `bouvardia_ledger_genesis.json.gz`
- This is the single gzipped GenesisState file for the bouvardia upgrade
- The file must have `.json.gz` extension

## Adding New Data Files

1. Create a new JSON file with the desired GenesisState data
2. Compress it using gzip: `gzip -c bouvardia_ledger_genesis.json > bouvardia_ledger_genesis.json.gz`
3. Place the `.json.gz` file in this directory
4. Ensure the JSON structure matches the expected GenesisState format
5. Test the data loading by running the upgrade handler

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