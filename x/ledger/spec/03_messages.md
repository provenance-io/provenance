# Ledger Messages

The Ledger module provides several message types for creating and managing ledger classes, ledgers, entries, and transfers. These messages allow authorized users to perform various operations on ledger data.

---
<!-- TOC 2 2 -->
  - [CreateLedger](#createledger)
  - [UpdateStatus](#updatestatus)
  - [UpdateInterestRate](#updateinterestrate)
  - [UpdatePayment](#updatepayment)
  - [UpdateMaturityDate](#updatematuritydate)
  - [Append](#append)
  - [UpdateBalances](#updatebalances)
  - [TransferFundsWithSettlement](#transferfundswithsettlement)
  - [Destroy](#destroy)
  - [CreateLedgerClass](#createledgerclass)
  - [AddLedgerClassStatusType](#addledgerclassstatustype)
  - [AddLedgerClassEntryType](#addledgerclassentrytype)
  - [AddLedgerClassBucketType](#addledgerclassbuckettype)
  - [BulkCreate](#bulkcreate)


## CreateLedger

To create a new ledger, use a `MsgCreateLedgerRequest`.

This request is expected to fail if:
- The ledger already exists.
- The NFT does not exist.
- The ledger class does not exist.
- The ledger class status type does not exist.
- The `signer` does not have the authority to create a ledger for the provided NFT.
- The msg is invalid.

### MsgCreateLedgerRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L62-L70

#### Ledger

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L71-L111

See also: DayCountConvention, InterestAccrualMethod, PaymentFrequency // TODO: Convert to links.

#### LedgerKey

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L58-L69

### MsgCreateLedgerResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L72-L73


## UpdateStatus

To update the status of a ledger, use a `MsgUpdateStatusRequest`.

This request is expected to fail if: // TODO: UpdateStatus

### MsgUpdateStatusRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L75-L87

See also: [LedgerKey](#ledgerkey)

### MsgUpdateStatusResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L89-L90


## UpdateInterestRate

To update the interest rate, interest day count convention, and interest accrual method of a ledger, use a `MsgUpdateInterestRateRequest`.

This request is expected to fail if: // TODO: UpdateInterestRate

### MsgUpdateInterestRateRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L92-L110

See also: [LedgerKey](#ledgerkey), DayCountConvention, InterestAccrualMethod. // TODO: Convert to links.

### MsgUpdateInterestRateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L112-L113


## UpdatePayment

To update the next payment amount, next payment date, and payment frequency of a ledger, use a `MsgUpdatePaymentRequest`.

This request is expected to fail if: // TODO: UpdatePayment

### MsgUpdatePaymentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L115-L139

See also: [LedgerKey](#ledgerkey), PaymentFrequency // TODO: Convert to links.

### MsgUpdatePaymentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L141-L142


## UpdateMaturityDate

To update a ledger's maturity date, use a `MsgUpdateMaturityDateRequest`.

This request is expected to fail if: // TODO: UpdateMaturityDate

### MsgUpdateMaturityDateRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L144-L156

See also: [LedgerKey](#ledgerkey).

### MsgUpdateMaturityDateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L158-L159


## Append

Entries are added to a ledger using a `MsgAppendRequest`.

This request is expected to fail if: // TODO: Append
1. **MsgAppend**
   - Asset identifiers must be valid
   - Ledger must exist
   - Entries must be valid
   - Signer must have permission
   - Correlation IDs must be unique
   - Sequences must be valid
   - Bucket types must be valid
   - Amounts must be valid

### MsgAppendRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L161-L173

See also: [LedgerKey](#ledgerkey).

#### LedgerEntry

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L125-L164

#### LedgerBucketAmount

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L166-L180

#### BucketBalance

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L182-L195

### MsgAppendResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L175-L176


## UpdateBalances

To update the applied amounts or balances amounts of a ledger entry, use a `MsgUpdateBalancesRequest`.

This request is expected to fail if: // TODO: UpdateBalances

### MsgUpdateBalancesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L178-L196

See also: [LedgerKey](#ledgerkey), [LedgerBucketAmount](#ledgerbucketamount), [BucketBalance](#BucketBalance).

### MsgUpdateBalancesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L198-L199


## TransferFundsWithSettlement

To transfer funds for a ledger based on settlement instructions, use a `MsgTransferFundsWithSettlementRequest`.

This request is expected to fail if: // TODO: TransferFundsWithSettlement
1. **MsgTransferFundsWithSettlement**
   - Asset identifiers must be valid
   - Ledger must exist
   - Correlation ID must be valid
   - Settlement instructions must be valid
   - Signer must have permission

### MsgTransferFundsWithSettlementRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L201-L210

#### FundTransferWithSettlement

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger_settlement.proto#L27-L34

#### SettlementInstruction

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger_settlement.proto#L36-L49

#### FundingTransferStatus

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger_settlement.proto#L13-L25

### MsgTransferFundsWithSettlementResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L212-L213


## Destroy

To delete a ledger and it's entries, use a `MsgDestroyRequest`.

This request is expected to fail if: // TODO: Destroy
2. **MsgDestroy**
   - Asset identifiers must be valid
   - Ledger must exist
   - Signer must have permission
   - All associated data must be properly cleaned up

### MsgDestroyRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L215-L224

See also: [LedgerKey](#ledgerkey).

### MsgDestroyResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L226-L227


## CreateLedgerClass

Ledger classes are created using a `MsgCreateLedgerClassRequest`.

This request is expected to fail if: // TODO: CreateLedgerClass
1. **MsgCreateLedgerClass**
   - Ledger class configuration must be valid
   - Asset class ID must be valid
   - Denomination must be valid
   - Maintainer address must be valid
   - Signer must have permission

### MsgCreateLedgerClassRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L229-L238

#### LedgerClass

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L11-L28

### MsgCreateLedgerClassResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L240-L241


## AddLedgerClassStatusType

To create a ledger class status type, use a `MsgAddLedgerClassStatusTypeRequest`.

This request is expected to fail if: // TODO: AddLedgerClassStatusType
3. **MsgAddLedgerClassStatusType**
   - Ledger class must exist
   - Status type must be valid
   - Signer must have permission

### MsgAddLedgerClassStatusTypeRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L243-L255

#### LedgerClassStatusType

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L43-L56

### MsgAddLedgerClassStatusTypeResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L257-L258


## AddLedgerClassEntryType

A ledger class entry type is created using a `MsgAddLedgerClassEntryTypeRequest`.

This request is expected to fail if: // TODO: AddLedgerClassEntryType
2. **MsgAddLedgerClassEntryType**
   - Ledger class must exist
   - Entry type must be valid
   - Signer must have permission

### MsgAddLedgerClassEntryTypeRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L260-L272

#### LedgerClassEntryType

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L30-L41

### MsgAddLedgerClassEntryTypeResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L274-L275


## AddLedgerClassBucketType

To create a ledger class bucket type, use `MsgAddLedgerClassBucketTypeRequest`.

This request is expected to fail if: // TODO: AddLedgerClassBucketType
4. **MsgAddLedgerClassBucketType**
   - Ledger class must exist
   - Bucket type must be valid
   - Signer must have permission

### MsgAddLedgerClassBucketTypeRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L277-L289

#### LedgerClassBucketType

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L113-L123

### MsgAddLedgerClassBucketTypeResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L291-L292


## BulkCreate

Ledger and ledger entries can be created in bulk using a `MsgBulkCreateRequest`.

This request is expected to fail if: // TODO: BulkCreate
1. **MsgBulkCreate**
   - All ledger configurations must be valid
   - All entries must be valid
   - Signer must have permission
   - Transaction size must be within limits

### MsgBulkCreateRequest

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L294-L303

#### LedgerAndEntries

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/ledger.proto#L197-L205

### MsgBulkCreateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.25.0/proto/provenance/ledger/v1/tx.proto#L305-L306
