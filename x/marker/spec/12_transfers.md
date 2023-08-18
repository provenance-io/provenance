# Transfers

There are some complex interactions involved with transfers of restricted coins.

<!-- TOC -->
  - [General](#general)
  - [Definitions](#definitions)
    - [Transfer Permission](#transfer-permission)
    - [Forced Transfers](#forced-transfers)
    - [Required Attributes](#required-attributes)
    - [Individuality](#individuality)
    - [Deposits](#deposits)
    - [Bypass Accounts](#bypass-accounts)
  - [Send Restrictions](#send-restrictions)
    - [Flowcharts](#flowcharts)
    - [Quarantine Complexities](#quarantine-complexities)

## General

Accounting of restricted coins is handled by the bank module. Restricted funds can be moved using the bank module's `MsgSend` or `MsgMutliSend`. They can also be moved using the marker module's `MsgTransferRequest`.

During such transfers several things are checked using a `SendRestrictionFn` injected into the bank module. This restriction is applied in almost all instances when funds are being moved between accounts. The exceptions are delegations, undelegations, minting, burning, and marker withdrawals. A `MsgTransferRequest` also bypasses the `SendRestrictionFn` in order to include the `admin` account in the logic.

<!-- TODO: Add notes about IBC movement too -->

## Definitions

### Transfer Permission

One permission that can be granted to an address is `transfer`.  The `transfer` permission is granted to accounts that represent a "Transfer Agent" or "Transfer Authority" for restricted marker tokens. An address with `transfer` permission can utilize `MsgTransferRequest` to move restricted funds from one account to another. If the marker allows forced transfer, the source account can be any account, otherwise, it must be the admin's own account.

`MsgSend` and `MsgMultiSend` can also be used by an address with `transfer` permission to move funds out of their own account.

### Forced Transfers

A restricted coin marker can be configured to allow forced transfers. If allowed, an account with `transfer` permission can use a `MsgTransferRequest` to transfer the restricted coins out of almost any account to another. Forced transfer cannot be used to move restricted coins out of module accounts or smart contract accounts, though.

### Required Attributes

Required attributes allow a marker Transfer Authority to define a set of account attestations created with the name/attribute modules to certify an account as an approved holder of the token.  Accounts that possess all of the required attributes are considered authorized by the Transfer Authority to receive the token from normal bank send operations without a specific Transfer Authority approval. Required attributes are only supported on restricted markers.

For example, say account A has some restricted coins of a marker that has required attributes. Also say account B has all of those required attributes, and account C does not. Account A could use a `MsgSend` to send those restricted coins to account B. However, account B could not send them to account C (unless B also has `transfer` permission).

If a restricted coin marker does not have any required attributes defined, the only way the funds can be moved is by someone with `transfer` permisison.

### Individuality

If multiple restricted coin denoms are being moved at once, each denom is considered separately.
For example, if the sender has `transfer` permission on one of them, it does not also apply to the other(s).

### Deposits

A deposit is when any funds are being sent to a marker's account. The funds being sent do not have to be in the denom of the destination marker.

Whenever funds are being deposited into a marker, the sender must have `deposit` permission on the target marker. If the funds to deposit are restricted coins, the sender also needs `transfer` permission on the funds being moved; required attributes are not taken into account.

### Bypass Accounts

There are several hard-coded module account addresses that are given special consideration in the marker module's `SendRestrictionFn`:

* `authtypes.FeeCollectorName` - Allows paying fees with restricted coins.
* `reward` - Allows reward programs to use restricted coins.
* `quarantine` - Allows quarantine and acceptance of quarantined coins.
* `gov` - Allows deposits to have quarantined coins.
* `distribution` - Allows collection of delegation rewards in restricted coins.
* `stakingtypes.BondedPoolName` - Allows delegation of restricted coins.
* `stakingtypes.NotBondedPoolName` - Allows delegation of restricted coins.

All of these are treated equally in the application of a marker's send restrictions.

For restricted markers without required attributes:
* If the `toAddr` is a bypass account, the `fromAddr` must have transfer authority.
* If the `fromAddr` is a bypass account, it's assumed that the funds got where they currently are because someone with transfer authority got them there, so this transfer is allowed.

For restricted markers with required attributes:
* If the `toAddr` is a bypass account, the transfer is allowed regardless of whether the `fromAddr` has transfer authority. It's assumed that the next destination's attributes will be properly checked before allowing the funds to leave the bypass account.
* If the `fromAddr` is a bypass account, the `toAddr` must have the required attributes.

Bypass accounts are not considered during a `MsgTransferRequest`.

## Send Restrictions

The marker module injects a `SendRestrictionFn` into the bank module. This function is responsible for deciding whether any given transfer is allowed from the marker module's point of view. However, it is bypassed during a `MsgTransfer`.

### Flowcharts

#### The SendRestrictionFn

This `SendRestrictionFn` uses the following flow.

```mermaid
%%{ init: { 'flowchart': { 'curve': 'monotoneY'} } }%%
flowchart TD
    start[["SendRestrictionFn(Sender, Receiver, Amount)"]]
    qhasbp{{"Does context have bypass?"}}
    nextd["Get next Denom from Amount."]
    vsd[["validateSendDenom(Sender, Receiver, Denom)"]]
    isdok{{"Is Denom transfer allowed?"}}
    mored{{"Does Amount have another Denom?"}}
    ok(["Send allowed."])
    style ok fill:#bbffaa,stroke:#1b8500,stroke-width:3px
    denied(["Send denied."])
    style denied fill:#ffaaaa,stroke:#b30000,stroke-width:3px
    start --> qhasbp
    qhasbp ------>|yes| ok
    qhasbp -.->|no| denomloop
    subgraph denomloop ["Denom Loop"]
    isdok -->|yes| mored
    vsd --> isdok
    mored -->|yes| nextd
    nextd --> vsd
    end
    mored -.->|no| ok
    isdok -.->|no| denied

    style denomloop fill:#bbffff
    linkStyle 8 stroke:#b30000,color:#b30000
    linkStyle 1,7 stroke:#1b8500,color:#1b8500
```

#### validateSendDenom

Each `Denom` is checked using `validateSendDenom`, which has this flow:

```mermaid
%%{ init: { 'flowchart': { 'curve': 'monotoneY'} } }%%
flowchart TD
    start[["validateSendDenom(Sender, Receiver, Denom)"]]
    qisrdep{{"Is Receiver a restricted coin marker account?"}}
    qhasdep{{"Does Sender have Deposit\non Receiver marker?"}}
    qisrc{{"Is Denom a restricted coin?"}}
    qisdeny{{"Is Sender on marker's deny list?"}}
    qhastrans{{"Does Sender have\ntransfer for Denom?"}}
    qisdep{{"Is Receiver a marker account?"}}
    qmhasattr{{"Does Denom have\nrequired attributes?"}}
    qissbp{{"Is Sender a\nbypass account?"}}
    qisrbp{{"Is Receiver a\nbypass account?"}}
    qrhasattr{{"Does Receiver have\nthe required attributes?"}}
    ok(["Denom transfer allowed."])
    style ok fill:#bbffaa,stroke:#1b8500,stroke-width:3px
    denied(["Send denied."])
    style denied fill:#ffaaaa,stroke:#b30000,stroke-width:3px
    start --> qisrdep
    qisrdep -->|yes| qhasdep
    qisrdep -.->|no| qisrc
    qhasdep -.->|no| denied
    qhasdep -->|yes| qisrc
    qisrc -->|yes| qisdeny
    qisrc -.->|no| ok
    qisdeny -->|yes| denied
    qisdeny -.->|no| qhastrans
    qhastrans -.->|no| qisdep
    qhastrans -->|yes| ok
    qisdep -->|yes| denied
    qisdep -.->|no| qmhasattr
    qmhasattr -.->|no| qissbp
    qmhasattr -->|yes| qisrbp
    qissbp -..->|no| denied
    qissbp --->|yes| ok
    qisrbp -.->|no| qrhasattr
    qisrbp -->|yes| ok
    qrhasattr -.->|no| denied
    qrhasattr -->|yes| ok

    linkStyle 3,7,11,15,19 stroke:#b30000,color:#b30000
    linkStyle 6,10,16,18,20 stroke:#1b8500,color:#1b8500
```

#### MsgTransferRequest

A `MsgTransferRequest` bypasses the `SendRestrictionFn` and applies its own logic. A `MsgTransferRequest` only allows for a single coin amount, i.e. there's only one `Denom` to consider.

```mermaid
%%{ init: { 'flowchart': { 'curve': 'monotoneY'} } }%%
flowchart TD
    start[["TransferCoin(Sender, Receiver, Admin)"]]
    qisrc{{"Is Denom a restricted coin?"}}
    qhast{{"Does Admin have transfer for Denom?"}}
    qadminfrom{{"Does Sender == Admin?"}}
    qallowft{{"Is forced transfer allowed for Denom?"}}
    qauthz{{"Has Sender granted Admin\npermission with authz?"}}
    qmodacc{{"Is Sender a\nmodule account?"}}
    qblocked{{"Is Receiver an\naddress blocked by\nthe bank module?"}}
    ok(["Transfer allowed."])
    style ok fill:#bbffaa,stroke:#1b8500,stroke-width:3px
    denied(["Transfer denied."])
    style denied fill:#ffaaaa,stroke:#b30000,stroke-width:3px
    start --> qisrc
    qisrc -->|yes| qhast
    qisrc -.->|no| denied
    qhast -->|yes| qadminfrom
    qhast -.->|no| denied
    qadminfrom -->|yes| qblocked
    qadminfrom -.->|no| qallowft
    qallowft -->|yes| qmodacc
    qallowft -.->|no| qauthz
    qmodacc -.->|no| qblocked
    qmodacc -->|yes| denied
    qauthz -->|yes| qblocked
    qauthz -.->|no| denied
    qblocked -.->|no| ok
    qblocked -->|yes| denied

    linkStyle 2,4,10,12,14 stroke:#b30000,color:#b30000
    linkStyle 13 stroke:#1b8500,color:#1b8500
```

### Quarantine Complexities

There are some noteable complexities involving restricted coins and quarantined accounts.

#### Sending Restricted Coins to a Quarantined Account

The marker module's `SendRestrictionFn` is applied before the quarantine module's. So, when funds are being sent to a quarantined account, the marker module runs its check using the original `Sender` and `Receiver` (i.e. the `Receiver` is not `QFH`).

If the `Receiver` is a quarantined account, we can assume that it is neither a marker, nor a bypass account. Then, assuming the `Sender` is not on the deny list, the `validateSendDenom` flow can be simplified to this for restricted coins.

```mermaid
%%{ init: { 'flowchart': { 'curve': 'monotoneY'} } }%%
flowchart LR
    vsd[["validateSendDenom(Sender, Receiver, Denom)"]]
    transq{{"Does Sender have\ntransfer for Denom?"}}
    mreqattr{{"Does Denom have\nrequired attributes?"}}
    treqattr{{"Does Receiver have\nthose attributes?"}}
    ok(["Denom transfer allowed."])
    style ok fill:#bbffaa,stroke:#1b8500,stroke-width:3px
    denied(["Send denied."])
    style denied fill:#ffaaaa,stroke:#b30000,stroke-width:3px
    transq -->|yes| ok
    transq -.->|no| mreqattr
    mreqattr -->|yes| treqattr
    mreqattr -.->|no| denied
    treqattr -->|yes| ok
    treqattr -.->|no| denied

    linkStyle 3,5 stroke:#b30000,color:#b30000
    linkStyle 0,4 stroke:#1b8500,color:#1b8500
```

If the `Send` is allowed, and the `Receiver` is a quarantined account, the quarantine module's `SendRestrictionFn` will then change the `Send`'s destination to `QFH` (the Quarantined-funds-holder account) and make a record of the transfer. The `Send` then transfers funds from the `Sender` to `QFH`.

The marker's `SendRestrictionFn` should never have `QFH` as a `Receiver`. The only way this would happen is if `MsgSend` is used to send funds directly to `QFH`.

If `MsgTransferRequest` is used to transfer a restricted coin to a quarantined account, the standard `MsgTransferRequest` logic is applied (bypassing the marker module's `SendRestrictionFn`). The quarantine module's `SendRestrictionFn` is not bypassed, though, so the funds still go to the `QFH`.

#### Accepting Quarantined Restricted Coins

Once funds have been sent to `QFH`, the `Receiver` will probably want to accept them, and have them sent to their account. They issue an `Accept` to the quarantine module which utilizes the bank module's `Send` functionality to try to transfer funds from `QFH` to the `Receiver`.

`QFH` is a bypass account. Since `Receiver` is a quarantined account, we can assume that it is neither a marker nor bypass account. So, the `validateSendDenom` flow can be simplified to this for restricted coins.

```mermaid
%%{ init: { 'flowchart': { 'curve': 'monotoneY'} } }%%
flowchart LR
    vsd[["validateSendDenom(Sender, Receiver, Denom)"]]
    mreqattr{{"Does Denom have\nrequired attributes?"}}
    treqattr{{"Does Receiver have\nthose attributes?"}}
    ok(["Denom transfer allowed."])
    style ok fill:#bbffaa,stroke:#1b8500,stroke-width:3px
    denied(["Send denied."])
    style denied fill:#ffaaaa,stroke:#b30000,stroke-width:3px
    mreqattr -->|yes| treqattr
    mreqattr -.->|no| ok
    treqattr -->|yes| ok
    treqattr -.->|no| denied

    linkStyle 3 stroke:#b30000,color:#b30000
    linkStyle 1,2 stroke:#1b8500,color:#1b8500

```

If the `Send` is allowed, the requested funds are transferred from `QFH` to `Receiver`.

If the `Send` is denied, the funds remain with `QFH`.

An important subtle part of this process is the rechecking of `Receiver` attributes. It's possible for the initial send to be okay (causing funds to be quarantined), then later, during this `Accept`, the send is not okay, and the quarantined funds are effectively locked with`QFH` until the `Receiver` gets the required attributes.

If the marker does not have required attributes though, it's assumed that they were originally sent by someone with transfer authority, so they are allowed to continue from here too.

#### Successful Quarantine and Accept Sequence

When restricted coin funds are sent to a quarantined account (1), the marker's `SendRestrictionFn` is called using the original `Sender` and `Receiver` (2). Then, the quarantine's `SendRestrictionFn` is called (4) which will return `QFH` for the new destination (5). Funds are then transferred from `Sender` to `QFH` (6).

When the `Receiver` attempts to `Accept` those quarantined funds (7), the marker's `SendRestrictionFn` is called again, this time using `QFH` (as the sender) and `Receiver` (9). The quarantine's `SendRestrictionFn` is bypassed (11), so the destination is not changed (12). Funds are then transferred from `QFH` to `Receiver` (13).

```mermaid
sequenceDiagram
    autonumber
    actor Sender
    actor Receiver
    participant Bank Module
    participant Quarantine Module
    participant Marker Restriction
    participant Quarantine Restriction
    participant QFH
    Sender ->>+ Bank Module: Send(sender, receiver)
    Bank Module ->>+ Marker Restriction: Is this send  from Sender to Receiver allowed?
    Marker Restriction -->>- Bank Module: Yes
    Bank Module ->>+ Quarantine Restriction: Is Receiver quarantined?
    Quarantine Restriction -->>- Bank Module: Yes. Change destination to QFH.
    Sender ->> QFH: Funds transferred from Sender to QFH.
    deactivate Bank Module

    Note over Sender,QFH: Some Time Later

    Receiver ->>+ Quarantine Module: Accept(receiver, sender)
    Quarantine Module ->> Bank Module: Send(QFH, receiver)
    activate Bank Module
    Bank Module ->>+ Marker Restriction: Is this send  from QFH to Receiver allowed?
    Marker Restriction -->>- Bank Module: Yes
    Bank Module ->>+ Quarantine Restriction: Is Receiver quarantined?
    Quarantine Restriction -->>- Bank Module: Restriction bypassed. No change.
    QFH ->> Receiver: Funds transferred from QFH to Receiver.
    deactivate Bank Module
    deactivate Quarantine Module
```
