# Bulk Import

## Overview

The ledger module supports bulk importing ledger data through a dedicated endpoint that accepts a `GenesisState` proto message. This provides several benefits:

1. **Clean Architecture**: Uses proper proto messages instead of JSON files
2. **Type Safety**: Compile-time validation of data structures
3. **Flexibility**: Can be called at any time, not just during upgrades
4. **Maintainability**: Data is properly structured and validated
5. **Reusability**: Can be used for testing, development, and production

## JSON Format Support

The bulk import now supports **both** snake_case and camelCase field names, as well as **both** integer and string enum values. This provides maximum flexibility for JSON input:

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

## Command Usage

```bash
# Basic usage
provenanced tx ledger bulk-import <json_file> --from <key> --chain-id <chain_id>

# Example with testnet
provenanced tx ledger bulk-import test.json \
    --from validator \
    --keyring-backend test \
    --chain-id testing \
    --gas-prices 1905nhash \
    --testnet \
    --yes

# Example with mainnet
provenanced tx ledger bulk-import production.json \
    --from mykey \
    --chain-id pio-mainnet-1 \
    --gas-prices 1905nhash \
    --yes
```

## Validation

The bulk import validates:

1. **JSON Format**: Ensures the JSON structure matches the expected proto message format
2. **Field Names**: Accepts both snake_case and camelCase field names
3. **Enum Values**: Accepts both integer and string enum values
4. **Data Types**: Validates all data types match the proto definitions
5. **Required Fields**: Ensures all required fields are present
6. **Ledger Class**: Verifies the ledger class exists before importing data

## Error Handling

Common error messages and their solutions:

- **`failed to get ledger class: collections: not found`**: The ledger class doesn't exist. Create it first using the ledger class commands.
- **`failed to unmarshal JSON`**: Check that the JSON format is valid and field names are correct.
- **`account not found`**: The signing account doesn't exist on the target network. Create the account or use a different key.

## Best Practices

1. **Use camelCase for new JSON files**: It's more readable and follows modern JSON conventions
2. **Use string enums for readability**: String enum values are self-documenting
3. **Validate JSON before importing**: Use JSON validators to check format
4. **Test with small datasets first**: Import a few records before bulk importing large datasets
5. **Backup existing data**: Always backup existing ledger data before bulk imports
6. **Use dry-run for testing**: Test the import with `--dry-run` flag before actual import 