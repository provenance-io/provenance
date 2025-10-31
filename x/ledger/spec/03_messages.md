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
  - [UpdateLedgerClass](#updateledgerclass)
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

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L66-L74

#### Ledger

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L71-L111

#### LedgerKey

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L58-L69

#### DayCountConvention

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L207-L234

#### InterestAccrualMethod

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L236-L264

#### PaymentFrequency

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L266-L288

### MsgCreateLedgerResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L76-L77


## UpdateStatus

To update the status of a ledger, use a `MsgUpdateStatusRequest`.

This request is expected to fail if:
- The ledger does not exist.
- The status does not exist for the ledger class.
- The `signer` does not have the authority to update the ledger.
- The msg is invalid.

### MsgUpdateStatusRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L79-L91

See also: [LedgerKey](#ledgerkey).

### MsgUpdateStatusResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L93-L94


## UpdateInterestRate

To update the interest rate, interest day count convention, and interest accrual method of a ledger, use a `MsgUpdateInterestRateRequest`.

This request is expected to fail if:
- The ledger does not exist.
- The `signer` does not have the authority to update the ledger.
- The msg is invalid.

### MsgUpdateInterestRateRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L96-L114

See also: [LedgerKey](#ledgerkey), [DayCountConvention](#daycountconvention), [InterestAccrualMethod](#interestaccrualmethod).

### MsgUpdateInterestRateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L116-L117


## UpdatePayment

To update the next payment amount, next payment date, and payment frequency of a ledger, use a `MsgUpdatePaymentRequest`.

This request is expected to fail if:
- The ledger does not exist.
- The `signer` does not have the authority to update the ledger.
- The msg is invalid.

### MsgUpdatePaymentRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L119-L143

See also: [LedgerKey](#ledgerkey), [PaymentFrequency](#paymentfrequency).

### MsgUpdatePaymentResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L145-L146


## UpdateMaturityDate

To update a ledger's maturity date, use a `MsgUpdateMaturityDateRequest`.

This request is expected to fail if:
- The ledger does not exist.
- The `signer` does not have the authority to update the ledger.
- The msg is invalid.

### MsgUpdateMaturityDateRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L148-L160

See also: [LedgerKey](#ledgerkey).

### MsgUpdateMaturityDateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L162-L163


## Append

Entries are added to a ledger using a `MsgAppendRequest`.

This request is expected to fail if:
- The ledger does not exist.
- The `signer` does not have the authority to update the ledger.
- One or more provided correlation_id values equal one or more existing ones.
- The ledger class entry type doesn't exist for one or more entries.
- The msg is invalid.

### MsgAppendRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L165-L177

See also: [LedgerKey](#ledgerkey).

#### LedgerEntry

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L125-L164

#### LedgerBucketAmount

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L166-L180

#### BucketBalance

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L182-L195

### MsgAppendResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L179-L180


## UpdateBalances

To update the applied amounts or balances amounts of a ledger entry, use a `MsgUpdateBalancesRequest`.

This request is expected to fail if:
- The ledger does not exist.
- The ledger entry does not exist.
- The `signer` does not have the authority to update the ledger.
- The msg is invalid.

### MsgUpdateBalancesRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L182-L209

See also: [LedgerKey](#ledgerkey), [LedgerBucketAmount](#ledgerbucketamount), [BucketBalance](#BucketBalance).

### MsgUpdateBalancesResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L211-L212


## TransferFundsWithSettlement

To transfer funds for a ledger based on settlement instructions, use a `MsgTransferFundsWithSettlementRequest`.

This request is expected to fail if:
- One or more ledgers do not exist.
- The ledger entry does not exist.
- The `signer` does not have the authority to update one of the ledgers.
- The `signer` does not have the required funds.
- The msg is invalid.

### MsgTransferFundsWithSettlementRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L214-L223

#### FundTransferWithSettlement

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger_settlement.proto#L30-L37

#### SettlementInstruction

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger_settlement.proto#L39-L52

#### FundingTransferStatus

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger_settlement.proto#L13-L28

### MsgTransferFundsWithSettlementResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L225-L226


## Destroy

To delete a ledger and it's entries, use a `MsgDestroyRequest`.

This request is expected to fail if:
- The ledger does not exist.
- The `signer` does not have the authority to update the ledger.
- The msg is invalid.

### MsgDestroyRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L228-L237

See also: [LedgerKey](#ledgerkey).

### MsgDestroyResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L239-L240


## CreateLedgerClass

Ledger classes are created using a `MsgCreateLedgerClassRequest`.

This request is expected to fail if:
- The asset class does not exist.
- The ledger class already exists.
- The denom does not yet have a marker.
- The msg is invalid.

### MsgCreateLedgerClassRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L242-L251

#### LedgerClass

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L11-L28

### MsgCreateLedgerClassResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L253-L254


## UpdateLedgerClass

Ledger classes can be updated using a `MsgUpdateLedgerClassRequest`.

This request is expected to fail if:
- The ledger class does not exist.
- The signer is not the maintainer of the ledger class.
- The new asset class does not exist.
- The new denom does not exist.
- The msg is invalid.

### MsgUpdateLedgerClassRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L256-L278

### MsgUpdateLedgerClassResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L280-L281


## AddLedgerClassStatusType

To create a ledger class status type, use a `MsgAddLedgerClassStatusTypeRequest`.

This request is expected to fail if:
- The ledger class does not exist.
- The ledger class status type already exists.
- The `signer` is not the maintainer of the ledger class.
- The msg is invalid

### MsgAddLedgerClassStatusTypeRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L283-L295

#### LedgerClassStatusType

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L43-L56

### MsgAddLedgerClassStatusTypeResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L297-L298


## AddLedgerClassEntryType

A ledger class entry type is created using a `MsgAddLedgerClassEntryTypeRequest`.

This request is expected to fail if:
- The ledger class does not exist.
- The ledger class entry type already exists.
- The `signer` is not the maintainer of the ledger class.
- The msg is invalid

### MsgAddLedgerClassEntryTypeRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L300-L312

#### LedgerClassEntryType

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L30-L41

### MsgAddLedgerClassEntryTypeResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L314-L315


## AddLedgerClassBucketType

To create a ledger class bucket type, use `MsgAddLedgerClassBucketTypeRequest`.

This request is expected to fail if:
- The ledger class does not exist.
- The ledger class bucket type already exists.
- The `signer` is not the maintainer of the ledger class.
- The msg is invalid

### MsgAddLedgerClassBucketTypeRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L317-L329

#### LedgerClassBucketType

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L113-L123

### MsgAddLedgerClassBucketTypeResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L331-L332


## BulkCreate

Ledger and ledger entries can be created together and in bulk using a `MsgBulkCreateRequest`.
This is essentially a combination of the [CreateLedger](#createledger) and [Append](#append) endpoints.

This request is expected to fail if:
- One or more of the ledgers already exist.
- One or more of the NFTs do not exist.
- One or more of the ledger classes do not exist.
- One or more of the ledger class status types do not exist.
- The `signer` does not have the authority to create a ledger for one or more provided NFTs.
- The ledger does not exist for one or more ledger entries.
- The `signer` does not have the authority to update the ledger for one or more ledger entries.
- One or more provided correlation_id values equal one or more existing ones.
- The ledger class entry type doesn't exist for one or more ledger entries.
- The msg is invalid.

### MsgBulkCreateRequest

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L334-L343

#### LedgerAndEntries

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/ledger.proto#L197-L205

At least one of `Ledger` or `LedgerKey` must be provided.
If both are provided, the two `LedgerKey`s must be equal.

### MsgBulkCreateResponse

+++ https://github.com/provenance-io/provenance/blob/v1.26.0/proto/provenance/ledger/v1/tx.proto#L345-L346
