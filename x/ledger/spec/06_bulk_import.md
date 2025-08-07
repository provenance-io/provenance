# Bulk Import

## Overview

The ledger module supports bulk importing ledger data through dedicated endpoints that accept `GenesisState` proto messages. This provides several benefits:

1. **Clean Architecture**: Uses proper proto messages instead of JSON files
2. **Type Safety**: Compile-time validation of data structures
3. **Flexibility**: Can be called at any time, not just during upgrades
4. **Maintainability**: Data is properly structured and validated
5. **Reusability**: Can be used for testing, development, and production
6. **Large Dataset Support**: Includes chunked import for handling large datasets
7. **Flat Fee Integration**: Uses flat fees for predictable transaction costs
8. **Robust Resume Functionality**: Advanced resume capabilities with transaction tracking

## Available Commands

The ledger module provides two main bulk import commands:

### 1. Standard Bulk Import (`bulk-import`)

For smaller datasets that fit within transaction limits.

**Note**: Standard bulk import only supports camelCase for the root key (`ledgerToEntries`). Use chunked bulk import for snake_case support.

```bash
# Basic usage
provenanced tx ledger bulk-import <json_file> --from <key> --chain-id <chain_id>

# Example with testnet
provenanced tx ledger bulk-import test.json \
    --from validator \
    --keyring-backend test \
    --chain-id testing \
    --gas-prices 1nhash \
    --testnet \
    --yes

# Example with mainnet
provenanced tx ledger bulk-import production.json \
    --from mykey \
    --chain-id pio-mainnet-1 \
    --gas-prices 1nhash \
    --yes
```

### 2. Chunked Bulk Import (`chunked-bulk-import`)

For large datasets that need to be split into manageable chunks.

**Note**: Chunked bulk import supports both camelCase (`ledgerToEntries`) and snake_case (`ledger_to_entries`) for the root key.

```bash
# Import with default chunk size (10MB memory limit, optimized for gas efficiency)
provenanced tx ledger chunked-bulk-import large_dataset.json --from mykey

# Import with custom chunk size (5MB memory limit)
provenanced tx ledger chunked-bulk-import large_dataset.json 5000000 --from mykey

# Example with import ID for resume functionality
provenanced tx ledger chunked-bulk-import large_dataset.json \
    --import-id my_import_123 \
    --from mykey \
    --chain-id pio-mainnet-1 \
    --yes

# Example with full options
provenanced tx ledger chunked-bulk-import large_dataset.json 7500000 \
    --import-id production_import_456 \
    --from mykey \
    --chain-id pio-mainnet-1 \
    --yes
```

### 3. Import Status Query (`bulk-import-status`)

Check the status of a chunked bulk import operation.

```bash
# Check local import status
provenanced query ledger bulk-import-status <import_id>

# Example
provenanced query ledger bulk-import-status import_1751488486228131162
```

## Large Data Import Strategies

When importing large datasets into the ledger module, you may encounter block size and gas limitations. The following strategies help handle large data imports effectively.

### Current Limitations

#### Block and Transaction Limits

1. **Transaction Gas Limit**: 4,000,000 gas per transaction
2. **Block Gas Limit**: Configurable via consensus parameters
3. **Block Size**: Limited by `max_bytes` in consensus parameters
4. **Transaction Size**: Limited by block size and gas constraints
5. **Flat Fees**: Fixed transaction costs regardless of gas usage

#### Typical Constraints

- **Small datasets**: < 100 ledgers, < 1,000 entries
- **Medium datasets**: 100-1,000 ledgers, 1,000-10,000 entries  
- **Large datasets**: > 1,000 ledgers, > 10,000 entries

### Strategy 1: Chunked Bulk Import (Recommended)

#### Overview

Split large datasets into manageable chunks that fit within block size and gas limitations using the `chunked-bulk-import` command.

#### Implementation

The chunked bulk import automatically:
- Processes files using streaming JSON parsing for memory efficiency
- Splits data into configurable chunk sizes based on memory limits
- Uses flat fees for predictable transaction costs
- Tracks import progress locally with advanced resume functionality
- Provides detailed status reporting with transaction tracking

#### Default Configuration

```go
type ChunkConfig struct {
    MaxChunkSizeBytes int // Default: 10MB (memory safety limit during parsing)
    MaxGasPerTx       int // Default: 4M gas per transaction
    MaxTxSizeBytes    int // Default: 1MB max transaction size
}

// Effective gas limit with safety margin
GetEffectiveGasLimit() int // Returns 3.8M gas (4M - 200k safety margin)
```

#### Benefits

- **Reliable**: Each chunk fits within transaction limits
- **Resumable**: Failed chunks can be retried individually with transaction tracking
- **Progress tracking**: Monitor import progress with detailed status files
- **Flexible**: Adjustable chunk sizes based on data characteristics
- **Memory efficient**: Uses streaming JSON parsing for large files
- **Predictable costs**: Flat fees provide consistent transaction costs
- **Advanced resume**: Robust resume functionality with transaction hash tracking

#### Example Workflow

1. **Prepare data**: Ensure ledger classes, status types, and entry types exist
2. **Estimate size**: Use helper functions to estimate chunk requirements
3. **Configure chunking**: Set appropriate chunk sizes (default 10MB recommended)
4. **Execute import**: Run chunked import with progress monitoring
5. **Verify results**: Check import status and validate data

### Strategy 2: Incremental Import

#### Overview

Import data incrementally over multiple transactions, focusing on specific subsets.

#### Implementation

```bash
# Import ledgers first
provenanced tx ledger bulk-import ledgers_only.json --from mykey

# Import entries for specific ledgers
provenanced tx ledger bulk-import entries_batch_1.json --from mykey
provenanced tx ledger bulk-import entries_batch_2.json --from mykey
```

#### Benefits

- **Controlled**: Import specific data types or ranges
- **Flexible**: Can prioritize critical data first
- **Debuggable**: Easier to identify issues in specific batches

#### Use Cases

- Import ledgers first, then entries
- Import by date ranges
- Import by ledger class or asset type
- Import critical data first, then supplementary data

### Strategy 3: Parallel Import

#### Overview

Import multiple chunks in parallel across different transactions.

#### Implementation

```bash
# Terminal 1: Import first chunk
provenanced tx ledger bulk-import chunk_1.json --from mykey

# Terminal 2: Import second chunk (different key)
provenanced tx ledger bulk-import chunk_2.json --from mykey2

# Terminal 3: Import third chunk (different key)
provenanced tx ledger bulk-import chunk_3.json --from mykey3
```

#### Benefits

- **Faster**: Multiple chunks processed simultaneously
- **Efficient**: Better resource utilization
- **Scalable**: Can use multiple validators or accounts

#### Considerations

- **Nonce management**: Each account needs proper nonce sequencing
- **Gas competition**: Multiple transactions may compete for gas
- **Order dependency**: Ensure no conflicts between parallel imports

### Strategy 4: Genesis Import

#### Overview

Import large datasets during chain initialization or upgrades.

#### Implementation

```json
{
  "app_state": {
    "ledger": {
      "ledger_to_entries": [
        // Large dataset here
      ]
    }
  }
}
```

#### Benefits

- **No limits**: Bypasses transaction size constraints
- **Atomic**: All data imported in single operation
- **Efficient**: No transaction overhead

#### Considerations

- **Chain restart**: Requires chain restart or upgrade
- **Timing**: Only available during genesis or upgrades
- **Validation**: Must validate entire dataset before import

### Strategy 5: Hybrid Approach

#### Overview

Combine multiple strategies based on data characteristics and requirements.

#### Implementation

```bash
# 1. Import critical ledgers via genesis
# 2. Import remaining ledgers via chunked import
provenanced tx ledger chunked-bulk-import remaining_ledgers.json \
    --import-id hybrid_import_123 \
    --from mykey

# 3. Import entries in parallel batches
provenanced tx ledger bulk-import entries_batch_1.json --from mykey1 &
provenanced tx ledger bulk-import entries_batch_2.json --from mykey2 &
provenanced tx ledger bulk-import entries_batch_3.json --from mykey3 &
```

## Flat Fees Integration

The bulk import system now uses flat fees for predictable transaction costs, while maintaining gas limit validation for execution constraints.

### How Flat Fees Work

1. **Flat Fees**: Determine transaction cost (how much you pay)
   - Queried from `x/flatfees` module
   - Fixed cost per message type regardless of gas usage
   - Stored in `FlatFeeInfo` struct for resume functionality

2. **Gas Limits**: Determine execution limits (how much computation is allowed)
   - Still enforced at 4M gas per transaction
   - Used for validation to ensure chunks don't exceed limits
   - Gas estimation still performed for validation purposes

### Benefits of Flat Fees

- **Simplified Fee Calculation**: No need for complex gas estimation for fees
- **Predictable Costs**: Flat fees provide consistent transaction costs
- **Higher Success Rate**: Transactions have predictable fees and better chance of success
- **Proper Gas Information**: Fixed transaction response parsing to show actual gas usage
- **Backward Compatibility**: Resume functionality works with both old and new fee systems

### Transaction Flow

1. **Startup**: Query flat fees for `MsgBulkImportRequest`
2. **Chunk Processing**: 
   - Validate chunk size and estimated gas usage
   - Build transaction with flat fees via command flags
   - Broadcast using `GenerateOrBroadcastTxCLI` for proper response parsing
   - Get detailed gas information in response
3. **Status Tracking**: Store both gas costs and flat fee info for resume

## JSON Format Support

The bulk import supports **both** snake_case and camelCase field names, as well as **both** integer and string enum values. This provides maximum flexibility for JSON input:

**Note**: For the root key `ledger_to_entries`/`ledgerToEntries`, both formats are supported in chunked imports, but standard bulk imports currently only support camelCase (`ledgerToEntries`) due to protobuf JSON tag limitations.

### Field Naming Support

| Proto Field (snake_case) | JSON Field (snake_case) | JSON Field (camelCase) | Description |
|--------------------------|-------------------------|----------------------|-------------|
| `ledger_to_entries` | `ledger_to_entries` | `ledgerToEntries` | Array of ledger entries |
| `ledger_key` | `ledger_key` | `ledgerKey` | Ledger key information |
| `nft_id` | `nft_id` | `nftId` | NFT identifier |
| `asset_class_id` | `asset_class_id` | `assetClassId` | Asset class identifier |
| `ledger_class_id` | `ledger_class_id` | `ledgerClassId` | Ledger class identifier |
| `status_type_id` | `status_type_id` | `statusTypeId` | Status type identifier |
| `next_pmt_date` | `next_pmt_date` | `nextPmtDate` | Next payment date |
| `next_pmt_amt` | `next_pmt_amt` | `nextPmtAmt` | Next payment amount |
| `interest_rate` | `interest_rate` | `interestRate` | Interest rate |
| `maturity_date` | `maturity_date` | `maturityDate` | Maturity date |
| `correlation_id` | `correlation_id` | `correlationId` | Correlation identifier |
| `entry_type_id` | `entry_type_id` | `entryTypeId` | Entry type identifier |
| `total_amt` | `total_amt` | `totalAmt` | Total amount |
| `applied_amounts` | `applied_amounts` | `appliedAmounts` | Applied amounts |
| `balance_amounts` | `balance_amounts` | `balanceAmounts` | Balance amounts |

### Enum Value Support

The following enums support **both** integer and string values:

#### DayCountConvention
| Integer | String | Description |
|---------|--------|-------------|
| `0` | `"LEDGER_DAY_COUNT_UNSPECIFIED"` | Unspecified |
| `1` | `"LEDGER_DAY_COUNT_ACTUAL_365"` | Actual/365 |
| `2` | `"LEDGER_DAY_COUNT_ACTUAL_360"` | Actual/360 |
| `3` | `"LEDGER_DAY_COUNT_THIRTY_360"` | 30/360 |
| `4` | `"LEDGER_DAY_COUNT_ACTUAL_ACTUAL"` | Actual/Actual |
| `5` | `"LEDGER_DAY_COUNT_DAYS_365"` | 365/365 |
| `6` | `"LEDGER_DAY_COUNT_DAYS_360"` | 360/360 |

#### InterestAccrualMethod
| Integer | String | Description |
|---------|--------|-------------|
| `0` | `"LEDGER_ACCRUAL_UNSPECIFIED"` | Unspecified |
| `1` | `"LEDGER_ACCRUAL_SIMPLE_INTEREST"` | Simple Interest |
| `2` | `"LEDGER_ACCRUAL_COMPOUND_INTEREST"` | Compound Interest |
| `3` | `"LEDGER_ACCRUAL_DAILY_COMPOUNDING"` | Daily Compounding |
| `4` | `"LEDGER_ACCRUAL_MONTHLY_COMPOUNDING"` | Monthly Compounding |
| `5` | `"LEDGER_ACCRUAL_QUARTERLY_COMPOUNDING"` | Quarterly Compounding |
| `6` | `"LEDGER_ACCRUAL_ANNUAL_COMPOUNDING"` | Annual Compounding |
| `7` | `"LEDGER_ACCRUAL_CONTINUOUS_COMPOUNDING"` | Continuous Compounding |

#### PaymentFrequency
| Integer | String | Description |
|---------|--------|-------------|
| `0` | `"LEDGER_PAYMENT_FREQUENCY_UNSPECIFIED"` | Unspecified |
| `1` | `"LEDGER_PAYMENT_FREQUENCY_DAILY"` | Daily |
| `2` | `"LEDGER_PAYMENT_FREQUENCY_WEEKLY"` | Weekly |
| `3` | `"LEDGER_PAYMENT_FREQUENCY_MONTHLY"` | Monthly |
| `4` | `"LEDGER_PAYMENT_FREQUENCY_QUARTERLY"` | Quarterly |
| `5` | `"LEDGER_PAYMENT_FREQUENCY_ANNUALLY"` | Annually |

## Proto Structure

The bulk import uses the `GenesisState` proto message with the following structure that matches the `test.json` format:

```protobuf
message GenesisState {
  repeated LedgerToEntries ledger_to_entries = 1;
}

message LedgerToEntries {
  LedgerKey            ledger_key = 1;
  Ledger               ledger     = 2;
  repeated LedgerEntry entries    = 3;
}
```

## JSON Examples

### Example 1: Snake Case with Integer Enums (Traditional Format)
```json
{
    "ledger_to_entries": [
        {
            "ledger_key": {
                "nft_id": "scope1qzqqqnucvdf5gu49t7agzh3pw4lsjaju7y",
                "asset_class_id": "scopespec1qj5hx4l3vgryhp5g3ks68wh53jkq3net7n"
            },
            "ledger": {
                "key": {
                    "nft_id": "scope1qzqqqnucvdf5gu49t7agzh3pw4lsjaju7y",
                    "asset_class_id": "scopespec1qj5hx4l3vgryhp5g3ks68wh53jkq3net7n"
                },
                "ledger_class_id": "figure_servicing_1.0",
                "status_type_id": 1,
                "next_pmt_date": 20264,
                "next_pmt_amt": 790140000,
                "interest_rate": 99900,
                "maturity_date": 21970,
                "interest_day_count_convention": 4,
                "interest_accrual_method": 1,
                "payment_frequency": 3
            },
            "entries": [
                {
                    "correlation_id": "test-entry-1",
                    "reverses_correlation_id": "",
                    "is_void": false,
                    "sequence": 1,
                    "entry_type_id": 1,
                    "posted_date": 20264,
                    "effective_date": 20264,
                    "total_amt": "790140000",
                    "applied_amounts": [
                        {
                            "bucket_type_id": 1,
                            "applied_amt": "790140000"
                        }
                    ],
                    "balance_amounts": [
                        {
                            "bucket_type_id": 1,
                            "balance_amt": "790140000"
                        }
                    ]
                }
            ]
        }
    ]
}
```

### Example 2: Camel Case with String Enums (Modern Format)
```json
{
    "ledgerToEntries": [
        {
            "ledgerKey": {
                "nftId": "scope1qzqqqnucvdf5gu49t7agzh3pw4lsjaju7y",
                "assetClassId": "scopespec1qj5hx4l3vgryhp5g3ks68wh53jkq3net7n"
            },
            "ledger": {
                "key": {
                    "nftId": "scope1qzqqqnucvdf5gu49t7agzh3pw4lsjaju7y",
                    "assetClassId": "scopespec1qj5hx4l3vgryhp5g3ks68wh53jkq3net7n"
                },
                "ledgerClassId": "figure_servicing_1.0",
                "statusTypeId": 1,
                "nextPmtDate": 20264,
                "nextPmtAmt": 790140000,
                "interestRate": 99900,
                "maturityDate": 21970,
                "interestDayCountConvention": "LEDGER_DAY_COUNT_ACTUAL_ACTUAL",
                "interestAccrualMethod": "LEDGER_ACCRUAL_SIMPLE_INTEREST",
                "paymentFrequency": "LEDGER_PAYMENT_FREQUENCY_MONTHLY"
            },
            "entries": [
                {
                    "correlationId": "test-entry-1",
                    "reversesCorrelationId": "",
                    "isVoid": false,
                    "sequence": 1,
                    "entryTypeId": 1,
                    "postedDate": 20264,
                    "effectiveDate": 20264,
                    "totalAmt": "790140000",
                    "appliedAmounts": [
                        {
                            "bucketTypeId": 1,
                            "appliedAmt": "790140000"
                        }
                    ],
                    "balanceAmounts": [
                        {
                            "bucketTypeId": 1,
                            "balanceAmt": "790140000"
                        }
                    ]
                }
            ]
        }
    ]
}
```

## Monitoring and Verification

### Import Status

```bash
# Check chunked import status
provenanced query ledger bulk-import-status <import_id>

# Verify imported data
provenanced query ledger ledgers
provenanced query ledger entries
```

### Status File Format

The chunked import creates a local status file (`.bulk_import_status.<import_id>.json`) with the following structure:

```json
{
  "import_id": "import_1751488486228131162",
  "total_chunks": 5,
  "completed_chunks": 3,
  "total_ledgers": 500,
  "total_entries": 1500,
  "status": "in_progress",
  "error_message": "",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:15:00Z",
  "last_successful_correlation_id": "corr_12345",
  "file_hash": "sha256:abc123...",
  "last_attempted_chunk": {
    "first_correlation_id": "corr_12340",
    "last_correlation_id": "corr_12345",
    "confirmed": true,
    "transaction_hash": "0x123456..."
  },
  "gas_costs": {
    "ledger_with_key_gas": 150000,
    "entry_gas": 7500
  },
  "flat_fee_info": {
    "fee_amount": "1000000nhash",
    "msg_type": "/provenance.ledger.v1.MsgBulkImportRequest"
  }
}
```

## Resume Functionality

The chunked bulk import supports robust resume functionality that allows interrupted imports to be restarted from where they left off.

### Overview

Resume functionality provides:
- **Automatic detection** of interrupted imports
- **File integrity validation** to prevent data corruption
- **Gas cost persistence** for efficient chunking
- **Flat fee persistence** for consistent transaction costs
- **Transaction hash tracking** for reliable resume points
- **Correlation ID tracking** for precise data positioning

### How Resume Works

1. **Status File Detection**: When `--import-id` is provided, the system checks for an existing status file
2. **File Hash Validation**: Validates that the source file hasn't changed using SHA256 hash
3. **Resume Point Calculation**: Determines the correct starting point based on last successful chunk
4. **Gas Cost Reuse**: Uses stored gas costs to maintain consistent chunk structure
5. **Flat Fee Reuse**: Uses stored flat fee information for consistent transaction costs
6. **Transaction Verification**: Checks if the last attempted transaction was actually processed

### Import ID Behavior

The `--import-id` flag controls resume behavior:

```bash
# Fresh import (auto-generated ID)
provenanced tx ledger chunked-bulk-import data.json --yes

# Resume existing import
provenanced tx ledger chunked-bulk-import data.json --import-id my_import_123 --yes

# Fresh import with specific ID
provenanced tx ledger chunked-bulk-import data.json --import-id new_import_456 --yes
```

### Gas Cost Storage

The system stores component-based gas costs for efficient resume:

```json
{
  "gas_costs": {
    "ledger_with_key_gas": 150000,
    "entry_gas": 7500
  }
}
```

- **`ledger_with_key_gas`**: Base cost for creating a ledger with its key
- **`entry_gas`**: Cost per individual ledger entry

These costs are calculated from representative simulations and reused on resume to avoid re-simulation.

### Flat Fee Storage

The system stores flat fee information for consistent transaction costs:

```json
{
  "flat_fee_info": {
    "fee_amount": "1000000nhash",
    "msg_type": "/provenance.ledger.v1.MsgBulkImportRequest"
  }
}
```

- **`fee_amount`**: The flat fee amount for each chunk
- **`msg_type`**: The message type URL for the bulk import

This information is queried from the `x/flatfees` module and reused on resume for consistent costs.

### Transaction Hash Tracking

The system ensures transaction information is preserved:

```json
{
  "last_attempted_chunk": {
    "first_correlation_id": "corr_12340",
    "last_correlation_id": "corr_12345",
    "confirmed": true,
    "transaction_hash": "0x123456..."
  }
}
```

- **Immediate Storage**: Transaction hash saved immediately after broadcast
- **Pre-confirmation Safety**: Status written before confirmation wait
- **Robust Extraction**: Handles various output formats and parsing failures

### Resume Scenarios

#### Scenario 1: Normal Interruption
- **Cause**: Ctrl+C, network interruption, process kill
- **Status**: `"in_progress"` with complete `LastAttemptedChunk`
- **Resume**: Starts from next chunk using stored gas costs and flat fees
- **Result**: Seamless continuation

#### Scenario 2: Transaction Confirmation Wait
- **Cause**: Interruption during `waitForTransactionConfirmation`
- **Status**: `LastAttemptedChunk` contains transaction hash
- **Resume**: Verifies transaction status and continues appropriately
- **Result**: Reliable resume regardless of confirmation state

#### Scenario 3: Gas Cost Changes
- **Cause**: Network gas costs change between runs
- **Status**: Uses stored gas costs for consistent chunking
- **Resume**: Maintains same chunk structure
- **Result**: Predictable behavior

#### Scenario 4: Flat Fee Changes
- **Cause**: Flat fees change between runs
- **Status**: Uses stored flat fee info for consistent costs
- **Resume**: Maintains same transaction costs
- **Result**: Predictable transaction costs

#### Scenario 5: File Modification
- **Cause**: Source file modified between runs
- **Status**: File hash mismatch detected
- **Resume**: Prevents resume to avoid corruption
- **Result**: Clear error message

### Resume Algorithm

```go
// Resume detection and processing
func detectResume(importID string, sourceFile string) (*ResumeInfo, error) {
    // 1. Check for existing status file
    status, err := readLocalBulkImportStatus(importID)
    if err != nil {
        return nil, fmt.Errorf("no existing import found")
    }
    
    // 2. Validate file hash
    currentHash := calculateFileHash(sourceFile)
    if status.FileHash != currentHash {
        return nil, fmt.Errorf("source file modified")
    }
    
    // 3. Determine resume point
    resumePoint := calculateResumePoint(status)
    
    // 4. Load stored gas costs and flat fees
    gasCosts := status.GasCosts
    flatFeeInfo := status.FlatFeeInfo
    
    return &ResumeInfo{
        StartChunk: resumePoint.ChunkIndex,
        GasCosts:   gasCosts,
        FlatFeeInfo: flatFeeInfo,
        Status:     status,
    }, nil
}
```

### Best Practices for Resume

#### 1. Import ID Management
- Use descriptive, meaningful import IDs
- Maintain consistent naming conventions
- Document import IDs and their purposes
- Use version control for import configurations

#### 2. File Management
- Never modify source files during import
- Keep backups of source files
- Use version control for source files
- Validate file integrity before import

#### 3. Monitoring and Verification
- Regularly check import status
- Monitor logs for gas cost and flat fee calculations
- Verify transactions on-chain after completion
- Use status queries to track progress

#### 4. Error Recovery
- Understand different error conditions
- Use appropriate resume strategies
- Test resume functionality with sample data
- Keep detailed logs for troubleshooting

### Troubleshooting Resume Issues

#### Common Problems

1. **File Hash Mismatch**
   ```bash
   # Error: "file hash doesn't match"
   # Solution: Use new import ID or restore original file
   provenanced tx ledger chunked-bulk-import data.json --import-id new_import_123
   ```

2. **Missing Transaction Hash**
   ```bash
   # Check transaction broadcast logs
   # Verify network connectivity
   # Check gas estimation
   ```

3. **Gas Cost Extraction Failed**
   ```bash
   # Check network connectivity
   # Verify gas estimation parameters
   # Check simulation logs
   ```

4. **Flat Fee Query Failed**
   ```bash
   # Check flat fees module availability
   # Verify network connectivity
   # Check flat fee configuration
   ```

5. **Wrong Resume Point**
   ```bash
   # Verify correlation ID tracking
   # Check status file contents
   # Validate resume logic
   ```

#### Debug Commands

```bash
# Check status file contents
cat .bulk_import_status.<import_id>.json

# Verify file hash
sha256sum <source_file>

# Check transaction status
provenanced query tx <tx_hash>

# Validate gas costs
provenanced query ledger bulk-import-status <import_id> | jq '.gas_costs'

# Check flat fee info
provenanced query ledger bulk-import-status <import_id> | jq '.flat_fee_info'

# Check import progress
provenanced query ledger bulk-import-status <import_id> | jq '.completed_chunks'
```

#### Error Messages and Solutions

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `"file hash doesn't match"` | Source file modified | Use new import ID or restore file |
| `"no next correlation ID found"` | Import complete or file corrupted | Check if import is finished |
| `"failed to extract transaction hash"` | Transaction broadcast failed | Check network and gas settings |
| `"gas costs not found"` | First run needed | Let first run complete to calculate costs |
| `"flat fee query failed"` | Flat fees module unavailable | Check flat fees module status |
| `"last chunk was not processed"` | Transaction failed | Check transaction status and retry |

## Best Practices

### 1. Data Preparation

- **Validate data**: Ensure all required ledger classes exist
- **Estimate sizes**: Use helper functions to estimate chunk requirements
- **Test with samples**: Test import process with small datasets first

### 2. Chunking Strategy

- **Memory-based**: Split by memory limits (recommended, default 10MB)
- **Ledger-based**: Split by ledger count for ledger-heavy data
- **Entry-based**: Split by entry count for entry-heavy data
- **Size-based**: Split by estimated transaction size
- **Time-based**: Split by date ranges for time-series data

### 3. Error Handling

- **Retry logic**: Implement retry mechanisms for failed chunks
- **Rollback capability**: Ensure ability to rollback partial imports
- **Status tracking**: Monitor import progress and status

### 4. Performance Optimization

- **Gas estimation**: Accurately estimate gas consumption for validation
- **Flat fee usage**: Use flat fees for predictable transaction costs
- **Batch sizing**: Optimize chunk sizes for your data
- **Parallel processing**: Use multiple accounts for parallel imports

### 5. JSON Best Practices

- **Use camelCase for new JSON files**: It's more readable and follows modern JSON conventions
- **Use string enums for readability**: String enum values are self-documenting
- **Validate JSON before importing**: Use JSON validators to check format
- **Test with small datasets first**: Import a few records before bulk importing large datasets
- **Backup existing data**: Always backup existing ledger data before bulk imports
- **Use dry-run for testing**: Test the import with `--dry-run` flag before actual import

## Troubleshooting

### Common Issues

1. **Gas limit exceeded**: Reduce chunk size or optimize data
2. **Block size exceeded**: Split data into smaller chunks
3. **Missing dependencies**: Ensure ledger classes exist before import
4. **Duplicate data**: Check for existing ledgers/entries before import
5. **Flat fee query failed**: Check flat fees module availability

### Debugging Commands

```bash
# Check transaction gas usage
provenanced query tx <tx_hash> --output json | jq '.gas_used'

# Check block size
provenanced query block <height> --output json | jq '.block.data.txs | length'

# Check flat fee for bulk import
provenanced query flatfees msg-fee "/provenance.ledger.v1.MsgBulkImportRequest"

# Validate chunk before import
provenanced tx ledger validate-chunk chunk.json
```

### Error Handling

Common error messages and their solutions:

- **`failed to get ledger class: collections: not found`**: The ledger class doesn't exist. Create it first using the ledger class commands.
- **`failed to unmarshal JSON`**: Check that the JSON format is valid and field names are correct.
- **`account not found`**: The signing account doesn't exist on the target network. Create the account or use a different key.
- **`chunk size too large`**: Reduce the chunk size parameter for chunked imports.
- **`no data imported`**: If using snake_case (`ledger_to_entries`) with standard bulk import, switch to chunked bulk import or use camelCase (`ledgerToEntries`).
- **`flat fee query failed`**: The flat fees module may not be available. Check module status or use gas-prices instead.

## Conclusion

For large data imports, the **chunked bulk import** strategy is recommended as it provides the best balance of reliability, flexibility, and performance. The integration with flat fees provides predictable transaction costs, while the advanced resume functionality ensures robust handling of interruptions.

Always test your import strategy with sample data before running on production datasets, and ensure you have proper monitoring and rollback capabilities in place. 