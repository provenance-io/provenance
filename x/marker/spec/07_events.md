# Events

The marker module emits the following events:
## Marker Added

Fires when a marker is added using the Add Marker Msg.

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_added           | denom                 | {denom string} |
| marker_added           | amount                | {supply amount}           |
| marker_added           | administrator         | {admin account address}   |
| marker_added           | status                | {current marker status}   |
| marker_added           | type                  | {type of marker}          |
| marker_added           | marker_AccessGrant    | {access grant string}     |
| marker_added           | marker_access_revoked | {address removed}         |

## Marker Updated

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_updated         | denom                 | {denom string} |
| marker_updated         | amount                | {supply amount}           |
| marker_updated         | administrator         | {admin account address}   |
| marker_updated         | status                | {current marker status}   |
| marker_updated         | type                  | {type of marker}          |
| marker_updated         | marker_AccessGrant    | {access grant string}     |
| marker_updated         | marker_access_revoked | {address removed}         |

## Grant Access

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_AccessGranted   | denom                 | {denom string} |
| marker_AccessGranted   | amount                | {supply amount}           |
| marker_AccessGranted   | administrator         | {admin account address}   |
| marker_AccessGranted   | status                | {current marker status}   |
| marker_AccessGranted   | type                  | {type of marker}          |
| marker_AccessGranted   | marker_AccessGrant    | {access grant string}     |
| marker_AccessGranted   | marker_access_revoked | {address removed}         |


## Revoke Access

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_access_revoked  | denom                 | {denom string} |
| marker_access_revoked  | amount                | {supply amount}           |
| marker_access_revoked  | administrator         | {admin account address}   |
| marker_access_revoked  | status                | {current marker status}   |
| marker_access_revoked  | type                  | {type of marker}          |
| marker_access_revoked  | marker_AccessGrant    | {access grant string}     |
| marker_access_revoked  | marker_access_revoked | {address removed}         |


## Finalize

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_finalized       | denom                 | {denom string} |
| marker_finalized       | amount                | {supply amount}           |
| marker_finalized       | administrator         | {admin account address}   |
| marker_finalized       | status                | {current marker status}   |
| marker_finalized       | type                  | {type of marker}          |
| marker_finalized       | marker_AccessGrant    | {access grant string}     |
| marker_finalized       | marker_access_revoked | {address removed}         |


## Activate

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_activated       | denom                 | {denom string} |
| marker_activated       | amount                | {supply amount}           |
| marker_activated       | administrator         | {admin account address}   |
| marker_activated       | status                | {current marker status}   |
| marker_activated       | type                  | {type of marker}          |
| marker_activated       | marker_AccessGrant    | {access grant string}     |
| marker_activated       | marker_access_revoked | {address removed}         |

## Cancel

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_cancelled       | denom                 | {denom string} |
| marker_cancelled       | amount                | {supply amount}           |
| marker_cancelled       | administrator         | {admin account address}   |
| marker_cancelled       | status                | {current marker status}   |
| marker_cancelled       | type                  | {type of marker}          |
| marker_cancelled       | marker_AccessGrant    | {access grant string}     |
| marker_cancelled       | marker_access_revoked | {address removed}         |

## Destroy


| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_destroyed       | denom                 | {denom string} |
| marker_destroyed       | amount                | {supply amount}           |
| marker_destroyed       | administrator         | {admin account address}   |
| marker_destroyed       | status                | {current marker status}   |
| marker_destroyed       | type                  | {type of marker}          |
| marker_destroyed       | marker_AccessGrant    | {access grant string}     |
| marker_destroyed       | marker_access_revoked | {address removed}         |

## Mint

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_minted_coins    | denom                 | {denom string} |
| marker_minted_coins    | amount                | {supply amount}           |
| marker_minted_coins    | administrator         | {admin account address}   |
| marker_minted_coins    | status                | {current marker status}   |
| marker_minted_coins    | type                  | {type of marker}          |
| marker_minted_coins    | marker_AccessGrant    | {access grant string}     |
| marker_minted_coins    | marker_access_revoked | {address removed}         |


## Burn

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_burned_coins    | denom                 | {denom string} |
| marker_burned_coins    | amount                | {supply amount}           |
| marker_burned_coins    | administrator         | {admin account address}   |
| marker_burned_coins    | status                | {current marker status}   |
| marker_burned_coins    | type                  | {type of marker}          |
| marker_burned_coins    | marker_AccessGrant    | {access grant string}     |
| marker_burned_coins    | marker_access_revoked | {address removed}         |


## Withdraw

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_withdraw_coins  | denom                 | {denom string} |
| marker_withdraw_coins  | amount                | {supply amount}           |
| marker_withdraw_coins  | administrator         | {admin account address}   |
| marker_withdraw_coins  | status                | {current marker status}   |
| marker_withdraw_coins  | type                  | {type of marker}          |
| marker_withdraw_coins  | marker_AccessGrant    | {access grant string}     |
| marker_withdraw_coins  | marker_access_revoked | {address removed}         |


## Transfer

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_tranfer_coin    | denom                 | {denom string} |
| marker_tranfer_coin    | amount                | {supply amount}           |
| marker_tranfer_coin    | administrator         | {admin account address}   |
| marker_tranfer_coin    | status                | {current marker status}   |
| marker_tranfer_coin    | type                  | {type of marker}          |
| marker_tranfer_coin    | marker_AccessGrant    | {access grant string}     |
| marker_tranfer_coin    | marker_access_revoked | {address removed}         |


## Deposit Asset


| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_asset_deposited | denom                 | {denom string} |
| marker_asset_deposited | amount                | {supply amount}           |
| marker_asset_deposited | administrator         | {admin account address}   |
| marker_asset_deposited | status                | {current marker status}   |
| marker_asset_deposited | type                  | {type of marker}          |
| marker_asset_deposited | marker_AccessGrant    | {access grant string}     |
| marker_asset_deposited | marker_access_revoked | {address removed}         |

## Withdraw Asset

| Type                   | Attribute Key         | Attribute Value           |
| ---------------------- | --------------------- | ------------------------- |
| marker_asset_withdrawn | denom                 | {denom string} |
| marker_asset_withdrawn | amount                | {supply amount}           |
| marker_asset_withdrawn | administrator         | {admin account address}   |
| marker_asset_withdrawn | status                | {current marker status}   |
| marker_asset_withdrawn | type                  | {type of marker}          |
| marker_asset_withdrawn | marker_AccessGrant    | {access grant string}     |
| marker_asset_withdrawn | marker_access_revoked | {address removed}         |