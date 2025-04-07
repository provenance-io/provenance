# Events

The Ledger module emits events to track important state changes and activities. These events can be used by external systems to monitor ledger activities and maintain synchronization.

## Event Types

### Ledger Created
Emitted when a new ledger is created for an NFT.

Attributes:
- `nft_address`: The address of the NFT
- `denom`: The denomination used for the ledger

### Ledger Entry Added
Emitted when a new entry is added to a ledger.

Attributes:
- `nft_address`: The address of the NFT
- `entry_uuid`: The unique identifier of the entry
- `entry_type`: The type of the entry
- `posted_date`: The date the entry was posted
- `effective_date`: The date the entry takes effect
- `amount`: The total amount of the entry

### Balance Updated
Emitted when balances are updated due to a ledger entry.

Attributes:
- `nft_address`: The address of the NFT
- `principal_balance`: The new principal balance
- `interest_balance`: The new interest balance
- `other_balance`: The new other balance

### Ledger Configuration Updated
Emitted when ledger configuration is modified.

Attributes:
- `nft_address`: The address of the NFT
- `denom`: The new denomination
- `previous_denom`: The previous denomination (if changed)

## Event Usage

Events can be used to:
- Track financial activities in real-time
- Maintain external accounting systems
- Monitor NFT financial positions
- Trigger external processes based on ledger activities
- Audit trail creation and verification 