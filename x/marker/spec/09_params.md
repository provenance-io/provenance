# Parameters

The marker module contains several settings that control operation of the module managed by the
Param module and available for control via Governance proposal to change parameters.

## Params

| Key                    | Type     | Example                           |
|------------------------|----------|-----------------------------------|
| MaxTotalSupply         | `uint64` | `"259200000000000"`               |
| EnableGovernance       | `bool`   | `true`                            |
| UnrestrictedDenomRegex | `string` | `"[a-zA-Z][a-zA-Z0-9\-\.]{7,83}"` |


## Definitions

- **Max Total Supply** (uint64) - A value indicating the maximum supply level allowed for any added marker

- **Enable Governance** (boolean) - A flag indicating if `allow_governance_control` setting on added markers must
  be set to `true`.

- **Unrestricted Denom Regex** (string) - A regular expression that is used to check the denom value on markers added
  by calling AddMarker.  This is intended to further restrict what may be used for a denom when a generic marker is
  created.