# Bulk Import

## Overview

The ledger module supports bulk importing ledger data through a dedicated endpoint that accepts a `GenesisState` proto message. This provides several benefits:

1. **Clean Architecture**: Uses proper proto messages instead of JSON files
2. **Type Safety**: Compile-time validation of data structures
3. **Flexibility**: Can be called at any time, not just during upgrades
4. **Maintainability**: Data is properly structured and validated
5. **Reusability**: Can be used for testing, development, and production

## Proto Structure

The bulk import uses the `GenesisState` proto message with the following structure that matches the `test.json` format:

```protobuf
message GenesisState {
  repeated LedgerToEntries ledger_to_entries = 1;
}

message LedgerToEntries {
  LedgerKey ledger_key = 1;
  Ledger ledger = 2;
  repeated LedgerEntry entries = 3;
}
```

## JSON Field Naming Requirements

**Critical**: The JSON file must use **snake_case** field names that exactly match the proto definitions. The proto-generated Go code expects snake_case field names, not camelCase.

### Common Field Name Mappings

| Proto Field (snake_case) | JSON Field (snake_case) | Incorrect (camelCase) |
|--------------------------|-------------------------|----------------------|
| `ledger_to_entries` | `ledger_to_entries` | `ledgerToEntries` |
| `ledger_key` | `ledger_key` | `ledgerKey` |
| `nft_id` | `nft_id` | `nftId` |
| `asset_class_id` | `asset_class_id` | `assetClassId` |
| `ledger_class_id` | `ledger_class_id` | `ledgerClassId` |
| `status_type_id` | `status_type_id` | `statusTypeId` |
| `next_pmt_date` | `next_pmt_date` | `nextPmtDate` |
| `next_pmt_amt` | `next_pmt_amt` | `nextPmtAmt` |
| `interest_rate` | `interest_rate` | `interestRate` |
| `maturity_date` | `maturity_date` | `maturityDate` |
| `interest_day_count_convention` | `interest_day_count_convention` | `interestDayCountConvention` |
| `interest_accrual_method` | `interest_accrual_method` | `interestAccrualMethod` |
| `payment_frequency` | `payment_frequency` | `paymentFrequency` |
| `correlation_id` | `correlation_id` | `correlationId` |
| `reverses_correlation_id` | `reverses_correlation_id` | `reversesCorrelationId` |
| `is_void` | `is_void` | `isVoid` |
| `entry_type_id` | `entry_type_id` | `entryTypeId` |
| `posted_date` | `posted_date` | `postedDate` |
| `effective_date` | `effective_date` | `effectiveDate` |
| `total_amt` | `total_amt` | `totalAmt` |
| `applied_amounts` | `applied_amounts` | `appliedAmounts` |
| `bucket_type_id` | `bucket_type_id` | `bucketTypeId` |
| `applied_amt` | `applied_amt` | `appliedAmt` |
| `balance_amounts` | `balance_amounts` | `balanceAmounts` |
| `balance_amt` | `balance_amt` | `balanceAmt` |

### Enum Values

Enum values must be specified as **integers**, not strings:

| Enum Type | Integer Value | String Value (Incorrect) |
|-----------|---------------|-------------------------|
| `LEDGER_DAY_COUNT_ACTUAL_ACTUAL` | `4` | `"LEDGER_DAY_COUNT_ACTUAL_ACTUAL"` |
| `LEDGER_ACCRUAL_SIMPLE_INTEREST` | `1` | `"LEDGER_ACCRUAL_SIMPLE_INTEREST"` |
| `LEDGER_PAYMENT_FREQUENCY_MONTHLY` | `3` | `"LEDGER_PAYMENT_FREQUENCY_MONTHLY"` |

### Common Error Messages

If you encounter these error messages, check your JSON field naming:

- `collections: not found: key '''' of type github.com/cosmos/gogoproto/provenance.ledger.v1.LedgerClass`
  - **Cause**: Empty ledger class ID due to field name mismatch
  - **Solution**: Use `ledger_class_id` instead of `ledgerClassId`

- `json: cannot unmarshal string into Go struct field Ledger.ledger_to_entries.ledger.interest_day_count_convention of type types.DayCountConvention`
  - **Cause**: Using string enum values instead of integers
  - **Solution**: Use integer values (e.g., `4` instead of `"LEDGER_DAY_COUNT_ACTUAL_ACTUAL"`)

## Key Associations

### Ledger Entries Storage Structure

Ledger entries are stored using a **composite key** system:

1. **Primary Key**: Ledger Key (bech32-encoded)
   - `AssetClassId`: The scope specification ID or NFT class ID
   - `NftId`: The specific NFT identifier (scope ID or NFT ID)
   - Encoded as: `ledger1<asset_class_id:nft_id>`

2. **Secondary Key**: Correlation ID
   - Each entry has a unique `correlation_id` for identification

3. **Storage Format**: `collections.Pair[ledger_key_string, correlation_id]`

### Required Associations

For ledger entries to be properly initialized, they must be associated with:

1. **Existing Ledger**: The ledger must exist before entries can be added
2. **Valid Ledger Class**: The ledger class must be registered
3. **Valid Entry Types**: Entry types must be defined for the ledger class
4. **Valid Bucket Types**: Bucket types must be defined for the ledger class
5. **Authority**: The maintainer address must have proper authority

**Important**: The bulk import function assumes that ledger classes, status types, entry types, and bucket types are already created before calling this function. These should be created separately using the appropriate keeper methods.

## Usage

### 1. Create the Genesis State

Create a `GenesisState` message with your ledger data:

```go
genesisState := types.GenesisState{
    LedgerToEntries: []types.LedgerToEntries{
        {
            LedgerKey: &types.LedgerKey{
                NftId:        "scope1qxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
                AssetClassId: "figure-heloc",
            },
            Ledger: &types.Ledger{
                Key: &types.LedgerKey{
                    NftId:        "scope1qxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
                    AssetClassId: "figure-heloc",
                },
                LedgerClassId: "ledger-class-1",
                StatusTypeId:  1,
                NextPmtDate:   20000,
                NextPmtAmt:    1000000,
                InterestRate:  5000000,
                MaturityDate:  30000,
                InterestDayCountConvention: types.LEDGER_DAY_COUNT_ACTUAL_365,
                InterestAccrualMethod:      types.LEDGER_ACCRUAL_SIMPLE_INTEREST,
                PaymentFrequency:           types.LEDGER_PAYMENT_FREQUENCY_MONTHLY,
            },
            Entries: []*types.LedgerEntry{
                {
                    CorrelationId: "entry-001",
                    Sequence:      1,
                    EntryTypeId:   1,
                    PostedDate:    19000,
                    EffectiveDate: 19000,
                    TotalAmt:      sdk.NewInt(1000000000),
                    AppliedAmounts: []*types.LedgerBucketAmount{
                        {
                            BucketTypeId: 1,
                            AppliedAmt:    sdk.NewInt(1000000000),
                        },
                    },
                    BalanceAmounts: []*types.BucketBalance{
                        {
                            BucketTypeId: 1,
                            BalanceAmt:    sdk.NewInt(1000000000),
                        },
                    },
                },
            },
        },
    },
}
```

### 2. Create the Message

Create a `MsgBulkImportRequest`:

```go
msg := &types.MsgBulkImportRequest{
    Authority:    authorityAddress,
    GenesisState: genesisState,
}
```

### 3. Send the Transaction

Send the transaction using your preferred method (CLI, REST API, gRPC, etc.):

```bash
# Using the provenance CLI
provenanced tx ledger bulk-import \
    --from my-key \
    --chain-id testing \
    --genesis-state-file genesis.json \
    --yes
```

## Example JSON File

For reference, here's an example JSON file that can be converted to the proto structure:

```json
{
  "ledger_to_entries": [
    {
      "ledger_key": {
        "nft_id": "scope1qxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "asset_class_id": "figure-heloc"
      },
      "ledger": {
        "key": {
          "nft_id": "scope1qxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
          "asset_class_id": "figure-heloc"
        },
        "ledger_class_id": "ledger-class-1",
        "status_type_id": 1,
        "next_pmt_date": 20000,
        "next_pmt_amt": 1000000,
        "interest_rate": 5000000,
        "maturity_date": 30000,
        "interest_day_count_convention": 4,
        "interest_accrual_method": 1,
        "payment_frequency": 3
      },
      "entries": [
        {
          "correlation_id": "entry-001",
          "sequence": 1,
          "entry_type_id": 1,
          "posted_date": 19000,
          "effective_date": 19000,
          "total_amt": "1000000000",
          "applied_amounts": [
            {
              "bucket_type_id": 1,
              "applied_amt": "1000000000"
            }
          ],
          "balance_amounts": [
            {
              "bucket_type_id": 1,
              "balance_amt": "1000000000"
            }
          ]
        }
      ]
    }
  ]
}
```

## Prerequisites

Before running the bulk import, ensure that:

1. **Ledger Classes** are created using `CreateLedgerClass`
2. **Status Types** are added using `AddClassStatusType`
3. **Entry Types** are added using `AddClassEntryType`
4. **Bucket Types** are added using `AddClassBucketType`

The bulk import function will fail if any of these prerequisites are not met.

## Error Handling

The bulk import function will return an error if:

- A ledger class referenced in the data does not exist
- A status type, entry type, or bucket type referenced in the data does not exist
- The authority address is invalid or lacks proper permissions
- Any ledger or entry data is invalid
- JSON field names do not match proto definitions (use snake_case)
- Enum values are specified as strings instead of integers

## Migration from Previous Structure

If you were using the previous genesis structure with separate arrays for ledger classes, status types, entry types, and bucket types, you should:

1. Create the ledger classes, status types, entry types, and bucket types separately using the keeper methods
2. Use only the `LedgerToEntries` array in your genesis state
3. Ensure all referenced IDs in the ledger data match the created types
4. **Convert all field names to snake_case**
5. **Convert all enum values to integers**

This simplified structure matches the `test.json` format and provides a cleaner, more focused approach to bulk importing ledger data.

## CLI Usage

The bulk import functionality is available through the Provenance CLI:

```bash
# Basic usage
provenanced tx ledger bulk-import <genesis_state_file> --from <key_name>

# Example with test.json
provenanced tx ledger bulk-import test.json --from mykey

# With additional flags
provenanced tx ledger bulk-import genesis.json --from mykey --gas auto --gas-adjustment 1.2
```

### Command Options

- `<genesis_state_file>`: Path to the JSON file containing the genesis state data
- `--from`: Name or address of the private key to sign the transaction
- `--gas`: Gas limit for the transaction (default: 200000)
- `--gas-adjustment`: Adjustment factor for gas estimation (default: 1.0)
- `--fees`: Fees to pay with the transaction
- `--chain-id`: The network chain ID
- `--node`: RPC node address (default: tcp://localhost:26657)

### Example Genesis State File

The genesis state file should contain data in the following format:

```json
{
  "ledger_to_entries": [
    {
      "ledger_key": {
        "asset_class_id": "test-asset-1",
        "nft_id": "test-nft-1"
      },
      "ledger": {
        "ledger_class_id": "ledger-class-1",
        "status_type_id": 1
      },
      "entries": [
        {
          "correlation_id": "entry1",
          "sequence": 1,
          "entry_type_id": 1,
          "posted_date": 5,
          "effective_date": 5,
          "total_amt": "80000",
          "applied_amounts": [
            {
              "bucket_type_id": 1,
              "applied_amt": "80000"
            }
          ],
          "balance_amounts": [
            {
              "bucket_type_id": 1,
              "balance_amt": "80000"
            }
          ]
        }
      ]
    }
  ]
}
```

## Go Usage 