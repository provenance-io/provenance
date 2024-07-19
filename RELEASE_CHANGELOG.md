## [v1.19.0](https://github.com/provenance-io/provenance/releases/tag/v1.19.0) - 2024-07-19

Provenance Blockchain `v1.19.0` is primarily focused on updating our use of [Cosmos-SDK](https://github.com/cosmos/cosmos-sdk) to `v0.50` (from `v0.46`) which includes several performance improvements.

Details about changes made by Cosmos-SDK can be found in their [releases](https://github.com/cosmos/cosmos-sdk/releases), e.g. [v0.50.1](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.50.1) or [v0.47.0](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.47.0).

### Features

* Bump cosmos-SDK to `v0.50.2` (from `v0.46.13-pio-2`) [#1772](https://github.com/provenance-io/provenance/issues/1772).
* Add store for crisis module for sdk v0.50 [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Add PreBlocker support for sdk v0.50 [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Add the Sanction module back in [#1922](https://github.com/provenance-io/provenance/pull/1922).
* Add the Quarantine module back in [#1926](https://github.com/provenance-io/provenance/pull/1926).
* Bump wasmd to `v0.50.0` [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Update stargate queries for Attribute, Exchange, Marker, IBCRateLimit, Metadata, Msgfees, and Oracle modules [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Update stargate queries for Quarantine and Sanction modules [#2016](https://github.com/provenance-io/provenance/pull/2016).
* Add the circuit breaker module [#2031](https://github.com/provenance-io/provenance/pull/2031).
* Add upgrade handler to set scope net asset values and update block height for pio-testnet-1 [#2046](https://github.com/provenance-io/provenance/pull/2046), [#2050](https://github.com/provenance-io/provenance/pull/2050).
* Create a script for updating links in the spec docs to proto messages [#2068](https://github.com/provenance-io/provenance/pull/2068).
* Add grpc querier for cosmwasm 2.1.0 smart contracts [#2083](https://github.com/provenance-io/provenance/pull/2083).

### Improvements

* Remove unsupported database types [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Update ibc and migrate params [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Replace ModuleBasics with BasicModuleManager [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Remove handlers from provenance modules [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Updated app.go to use RegisterStreamingServices on BaseApp [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Bump the SDK to `v0.50.5-pio-1` (from an earlier ephemeral version) [#1897](https://github.com/provenance-io/provenance/pull/1897).
* Removed `rewards` module [#1905](https://github.com/provenance-io/provenance/pull/1905).
* Remove unused navs [#1920](https://github.com/provenance-io/provenance/issues/1920).
* Remove emitting of EventTypeMessage [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Update genutil for sdk 50 [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Migrate module params from param space to module store.[#1760](https://github.com/provenance-io/provenance/issues/1935)
  *  Attribute module param migration [#1927](https://github.com/provenance-io/provenance/pull/1927).
  *  Marker module param migration [#1934](https://github.com/provenance-io/provenance/pull/1934).
  *  Metadata module param migration [#1932](https://github.com/provenance-io/provenance/pull/1932).
  *  Msgfees module param migration [#1936](https://github.com/provenance-io/provenance/pull/1936).
  *  Name module param migration [#1937](https://github.com/provenance-io/provenance/pull/1937).
  *  IbcHooks module param migration [#1939](https://github.com/provenance-io/provenance/pull/1939).
  *  Bank module param migration [#1967](https://github.com/provenance-io/provenance/pull/1967).
* Remove `msgfees` legacy gov proposals [#1953](https://github.com/provenance-io/provenance/pull/1953).
* Remove `marker` legacy gov proposals [#1961](https://github.com/provenance-io/provenance/pull/1961).
* Restore the hold module [#1930](https://github.com/provenance-io/provenance/pull/1930).
* Restore gov-prop cli commands and fix next key decoding [#1930](https://github.com/provenance-io/provenance/pull/1930).
* Switch to InputOutputCoinsProv for exchange transfers [#1930](https://github.com/provenance-io/provenance/pull/1930).
* Use fields of the SimulationState for the encoders needed for simulations [#1931](https://github.com/provenance-io/provenance/pull/1931).
* Removes sync-info code for sdk v0.50 [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Remove `name` legacy gov proposals [#1963](https://github.com/provenance-io/provenance/pull/1963).
* Fix most of the failing unit tests [#1943](https://github.com/provenance-io/provenance/pull/1943).
* Clean up ReadFromClient [#1760](https://github.com/provenance-io/provenance/issues/1760).
* Enhance the config get and changed commands to make it easier to find fields [#1968](https://github.com/provenance-io/provenance/pull/1968).
* Change the default keyring backend to "os", but leave it as "test" for testnets [#1968](https://github.com/provenance-io/provenance/pull/1968).
* Change the default broadcast mode to "sync" [#1968](https://github.com/provenance-io/provenance/pull/1968).
* The pre-upgrade command now updates the client config's broadcast mode to "sync" if it's set to "block" [#1968](https://github.com/provenance-io/provenance/pull/1968).
* Remove all `GetSigners()` methods [#1957](https://github.com/provenance-io/provenance/pull/1957).
* Ensure all `Msg`s have correctly identified `signer` fields [#1957](https://github.com/provenance-io/provenance/pull/1957).
* Clean up all the module codecs [#1957](https://github.com/provenance-io/provenance/pull/1957).
* Switch to auto-generated `String` and `Equal` methods for most proto messages [#1957](https://github.com/provenance-io/provenance/pull/1957).
* Clean up the marker module's expected BankKeeper interface [#1954](https://github.com/provenance-io/provenance/pull/1954).
* Add the auto-cli commands and a few others newly added by the SDK [#1971](https://github.com/provenance-io/provenance/pull/1971).
* Fix unit tests for ibcratelimit [#1977](https://github.com/provenance-io/provenance/pull/1977).
* Fix unit tests for ibchooks [#1980](https://github.com/provenance-io/provenance/pull/1980).
* Replace deprecated wasm features [#1988](https://github.com/provenance-io/provenance/pull/1988).
* Add `UpdateParams` and `Params` query rpc endpoints to modules.
  * `exchange` add `UpdateParams` endpoint and deprecate `GovUpdateParams` [#2017](https://github.com/provenance-io/provenance/pull/2017).
  * `ibchooks` add `UpdateParams` endpoint and `Params` query endpoint [#2006](https://github.com/provenance-io/provenance/pull/2006).
  * `ibcratelimit` add `UpdateParams` endpoint and deprecate `GovUpdateParams` [#1984](https://github.com/provenance-io/provenance/pull/1984).
  * `attribute` add `UpdateParams` endpoint and cli [#1987](https://github.com/provenance-io/provenance/pull/1987).
  * `marker` add `UpdateParams` endpoint and cli [#1991](https://github.com/provenance-io/provenance/pull/1991).
  * `name` add `UpdateParams` endpoint and cli [#2004](https://github.com/provenance-io/provenance/pull/2004).
* Update the exchange `commitment-settlement-fee-calc` cli query to utilize the keyring [#2001](https://github.com/provenance-io/provenance/pull/2001).
* Implement the ProposalMsgs module interface for the internal/provwasm, ibcratelimit, oracle, and sanction modules [#1993](https://github.com/provenance-io/provenance/pull/1993).
* Update ibcnet relayer to v2.5.2 and fix chain ids for localnet and ibcnet [#2021](https://github.com/provenance-io/provenance/pull/2021).
* Breakdown internal/helpers into multiple internal packages [#2019](https://github.com/provenance-io/provenance/pull/2019).
* Update `app.New` to get the home directory, invariant check period, and skip-upgrade heights from the appOptions instead of arguments [#2015](https://github.com/provenance-io/provenance/pull/2015).
* Simplify the module lists (e.g. `SetOrderEndBlockers`) by removing unneeded entries [#2015](https://github.com/provenance-io/provenance/pull/2015).
* Update the `upgrade-test.sh` script to work with v0.50 commands [#2026](https://github.com/provenance-io/provenance/pull/2026).
* Set the new gov params fields during the umber upgrades [#2027](https://github.com/provenance-io/provenance/pull/2027).
* Add a bunch of queries to the stargate whitelist [#2037](https://github.com/provenance-io/provenance/pull/2037).
  * /cosmos.auth.v1beta1.Query/Accounts
  * /cosmos.auth.v1beta1.Query/AccountAddressByID
  * /cosmos.auth.v1beta1.Query/ModuleAccounts
  * /cosmos.auth.v1beta1.Query/ModuleAccountByName
  * /cosmos.auth.v1beta1.Query/Bech32Prefix
  * /cosmos.auth.v1beta1.Query/AddressBytesToString
  * /cosmos.auth.v1beta1.Query/AddressStringToBytes
  * /cosmos.auth.v1beta1.Query/AccountInfo
  * /cosmos.authz.v1beta1.Query/Grants
  * /cosmos.authz.v1beta1.Query/GranterGrants
  * /cosmos.authz.v1beta1.Query/GranteeGrants
  * /cosmos.bank.v1beta1.Query/AllBalances
  * /cosmos.bank.v1beta1.Query/SpendableBalances
  * /cosmos.bank.v1beta1.Query/SpendableBalanceByDenom
  * /cosmos.bank.v1beta1.Query/TotalSupply
  * /cosmos.bank.v1beta1.Query/DenomMetadataByQueryString
  * /cosmos.bank.v1beta1.Query/DenomsMetadata
  * /cosmos.bank.v1beta1.Query/DenomOwners
  * /cosmos.bank.v1beta1.Query/DenomOwnersByQuery
  * /cosmos.bank.v1beta1.Query/SendEnabled
  * /cosmos.circuit.v1.Query/Account
  * /cosmos.circuit.v1.Query/Accounts
  * /cosmos.circuit.v1.Query/DisabledList
  * /cosmos.consensus.v1.Query/Params
  * /cosmos.distribution.v1beta1.Query/ValidatorDistributionInfo
  * /cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards
  * /cosmos.distribution.v1beta1.Query/ValidatorSlashes
  * /cosmos.distribution.v1beta1.Query/DelegationRewards
  * /cosmos.distribution.v1beta1.Query/DelegationTotalRewards
  * /cosmos.distribution.v1beta1.Query/DelegatorValidators
  * /cosmos.distribution.v1beta1.Query/CommunityPool
  * /cosmos.evidence.v1beta1.Query/Evidence
  * /cosmos.evidence.v1beta1.Query/AllEvidence
  * /cosmos.feegrant.v1beta1.Query/Allowance
  * /cosmos.feegrant.v1beta1.Query/Allowances
  * /cosmos.feegrant.v1beta1.Query/AllowancesByGranter
  * /cosmos.gov.v1beta1.Query/Proposal
  * /cosmos.gov.v1beta1.Query/Proposals
  * /cosmos.gov.v1beta1.Query/Vote
  * /cosmos.gov.v1beta1.Query/Votes
  * /cosmos.gov.v1beta1.Query/Params
  * /cosmos.gov.v1beta1.Query/Deposit
  * /cosmos.gov.v1beta1.Query/Deposits
  * /cosmos.gov.v1beta1.Query/TallyResult
  * /cosmos.gov.v1.Query/Constitution
  * /cosmos.group.v1.Query/GroupInfo
  * /cosmos.group.v1.Query/GroupPolicyInfo
  * /cosmos.group.v1.Query/GroupMembers
  * /cosmos.group.v1.Query/GroupsByAdmin
  * /cosmos.group.v1.Query/GroupPoliciesByGroup
  * /cosmos.group.v1.Query/GroupPoliciesByAdmin
  * /cosmos.group.v1.Query/Proposal
  * /cosmos.group.v1.Query/ProposalsByGroupPolicy
  * /cosmos.group.v1.Query/VoteByProposalVoter
  * /cosmos.group.v1.Query/VotesByProposal
  * /cosmos.group.v1.Query/VotesByVoter
  * /cosmos.group.v1.Query/GroupsByMember
  * /cosmos.group.v1.Query/TallyResult
  * /cosmos.group.v1.Query/Groups
  * /cosmos.mint.v1beta1.Query/Params
  * /cosmos.mint.v1beta1.Query/Inflation
  * /cosmos.mint.v1beta1.Query/AnnualProvisions
  * /cosmos.slashing.v1beta1.Query/SigningInfos
  * /cosmos.staking.v1beta1.Query/Validators
  * /cosmos.staking.v1beta1.Query/Validator
  * /cosmos.staking.v1beta1.Query/ValidatorDelegations
  * /cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations
  * /cosmos.staking.v1beta1.Query/DelegatorDelegations
  * /cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations
  * /cosmos.staking.v1beta1.Query/Redelegations
  * /cosmos.staking.v1beta1.Query/DelegatorValidators
  * /cosmos.staking.v1beta1.Query/DelegatorValidator
  * /cosmos.staking.v1beta1.Query/HistoricalInfo
  * /cosmos.staking.v1beta1.Query/Pool
* Update the Swagger API documentation [#2063](https://github.com/provenance-io/provenance/pull/2063).
* Update all the proto links in the spec docs to point to `v1.19.0` versions of the proto files [#2068](https://github.com/provenance-io/provenance/pull/2068).
* Add the (empty) `umber-rc2`, `umber-rc3` and `umber-rc4` upgrades [#2069](https://github.com/provenance-io/provenance/pull/2069), [#2091](https://github.com/provenance-io/provenance/pull/2091).
* Remove the warnings about some config settings [2095](https://github.com/provenance-io/provenance/pull/2095).
* Record several scope NAVs with the umber upgrade [#2085](https://github.com/provenance-io/provenance/pull/2085).
* Store the Figure Funding Trading Bridge Smart Contract as part of the umber upgrade [2102](https://github.com/provenance-io/provenance/pull/2102).

### Bug Fixes

* The `add-net-asset-values` command now correctly uses the from `flag`'s `AccAddress` [#1995](https://github.com/provenance-io/provenance/issues/1995).
* Fix the `umber` and `umber-rc1` upgrades [#2033](https://github.com/provenance-io/provenance/pull/2033).
* Fix the heighliner docker image build [#2052](https://github.com/provenance-io/provenance/pull/2052).
* Put the location of the wasm directory back to where it is in previous versions: data/wasm/wasm [#2071](https://github.com/provenance-io/provenance/pull/2071).

### Deprecated

* In the config commands, the "tendermint" and "tm" options are deprecated, replaced with "cometbft", "comet", and "cmt" [#1968](https://github.com/provenance-io/provenance/pull/1968).
* All of the old governance proposals are now either deprecated or unusable. They all have new `Msg`-style endpoints for use in a `gov.v1.MsgSubmitProposal`.

### Client Breaking

* The `provenanced query account` command has been removed. It is still available as `provenanced query auth account` [#1971](https://github.com/provenance-io/provenance/pull/1971).
* Move the genesis-related commands into a new `genesis` sub-command, and remove the `genesis-` parts of their names [#1971](https://github.com/provenance-io/provenance/pull/1971).
  * `provenanced add-genesis-account` is now `provenanced genesis add-account`
  * `provenanced add-genesis-custom-floor` is now `provenanced genesis add-custom-floor`
  * `provenanced add-genesis-custom-market` is now `provenanced genesis add-custom-market`
  * `provenanced add-genesis-default-market` is now `provenanced genesis add-default-market`
  * `provenanced add-genesis-marker` is now `provenanced genesis add-marker`
  * `provenanced add-genesis-msg-fee` is now `provenanced genesis add-msg-fee`
  * `provenanced add-genesis-root-name` is now `provenanced genesis add-root-name`
  * `provenanced collect-gentxs` is now `provenanced genesis collect-gentxs`
  * `provenanced gentx` is now `provenanced genesis gentx`
  * `provenanced validate-genesis` is now `provenanced genesis validate`
* Many of the SDK's query commands have had their usage altered [#1971](https://github.com/provenance-io/provenance/pull/1971).
* Rosetta has been removed from the `provenanced` executable [#1981](https://github.com/provenance-io/provenance/pull/1981).
  It is now a stand-alone service. See: <https://github.com/cosmos/rosetta> for more info.
* When submitting a governance proposal, at least 1000 hash (`1000000000000nhash`) of the deposit must be included. The rest can still be added later.
* When submitting a governance proposal, the new `title` and `summary` fields are required.
* The `broadcast-mode` value `block` has been removed and can now only be either `sync` or `async`.
  The `provenanced query wait-tx` command can be used to achieve similar functionality as `block`, E.g. `provenanced tx <whatever> --output text | provenanced query wait-tx`.

### Dependencies

- Bump `bufbuild/buf-breaking-action` from 1.1.3 to 1.1.4 ([#1894](https://github.com/provenance-io/provenance/pull/1894))
- Bump `bufbuild/buf-lint-action` from 1.1.0 to 1.1.1 ([#1895](https://github.com/provenance-io/provenance/pull/1895))
- Bump `bufbuild/buf-setup-action` from 1.30.0 to 1.34.0 ([#1904](https://github.com/provenance-io/provenance/pull/1904), [#1949](https://github.com/provenance-io/provenance/pull/1949), [#1979](https://github.com/provenance-io/provenance/pull/1979), [#1990](https://github.com/provenance-io/provenance/pull/1990), [#2011](https://github.com/provenance-io/provenance/pull/2011), [#2036](https://github.com/provenance-io/provenance/pull/2036), [#2049](https://github.com/provenance-io/provenance/pull/2049))
- Bump `cosmossdk.io/api` from 0.7.4 to 0.7.5 ([#2025](https://github.com/provenance-io/provenance/pull/2025))
- Bump `cosmossdk.io/client/v2` from 2.0.0-beta.1 to 2.0.0-beta.2 ([#2042](https://github.com/provenance-io/provenance/pull/2042))
- Bump `cosmossdk.io/store` from 1.0.2 to 1.1.0 [#2026](https://github.com/provenance-io/provenance/pull/2026)
- Bump `cosmossdk.io/x/circuit` from 0.1.0 to 0.1.1 ([#2035](https://github.com/provenance-io/provenance/pull/2035))
- Bump `cosmossdk.io/x/evidence` from 0.1.0 to 0.1.1 [#2026](https://github.com/provenance-io/provenance/pull/2026)
- Bump `cosmossdk.io/x/feegrant` from 0.1.0 to 0.1.1 [#2026](https://github.com/provenance-io/provenance/pull/2026)
- Bump `cosmossdk.io/x/tx` from 0.13.1 to 0.13.3 ([#1928](https://github.com/provenance-io/provenance/pull/1928), [#1944](https://github.com/provenance-io/provenance/pull/1944))
- Bump `cosmossdk.io/x/upgrade` from 0.1.0 to 0.1.3 ([#1913](https://github.com/provenance-io/provenance/pull/1913), [#2026](https://github.com/provenance-io/provenance/pull/2026))
- Bump `cosmwasm-std` from 1.4.1 to 1.4.4 ([#1950](https://github.com/provenance-io/provenance/pull/1950))
- Bump `docker/build-push-action` from 5 to 6 ([#2039](https://github.com/provenance-io/provenance/pull/2039))
- Bump `docker/setup-qemu-action` from 2 to 3 ([#1983](https://github.com/provenance-io/provenance/pull/1983))
- Bump `github.com/CosmWasm/wasmd` from `v0.50.0-pio-2` to `v0.52.0-pio-1` ([#2045](https://github.com/provenance-io/provenance/pull/2045), [#2077](https://github.com/provenance-io/provenance/pull/2077))
- Bump `github.com/CosmWasm/wasmvm` from `v1.5.0` to `v2.1.0`` ([#2045](https://github.com/provenance-io/provenance/pull/2045), [#2077](https://github.com/provenance-io/provenance/pull/2077))
- Bump `github.com/cometbft/cometbft` from 0.38.5 to v0.38.10 ([#1912](https://github.com/provenance-io/provenance/pull/1912), [#1959](https://github.com/provenance-io/provenance/pull/1959), [#2061](https://github.com/provenance-io/provenance/pull/2061), [#2087](https://github.com/provenance-io/provenance/pull/2087), [#2097](https://github.com/provenance-io/provenance/pull/2097))
- Bump `github.com/cosmos/cosmos-sdk` from 0.50.5-pio-3 to 0.50.7-pio-1 [#2026](https://github.com/provenance-io/provenance/pull/2026)
- Bump `github.com/cosmos/gogoproto` from 1.4.12 to 1.5.0 ([#2024](https://github.com/provenance-io/provenance/pull/2024))
- Bump `github.com/cosmos/iavl` from 1.1.2 to 1.2.0 ([#2076](https://github.com/provenance-io/provenance/pull/2076))
- Bump `github.com/cosmos/ibc-go/modules/capability` from 1.0.0 to 1.0.1 ([#2064](https://github.com/provenance-io/provenance/pull/2064))
- Bump `github.com/cosmos/ibc-go/v8` from 8.0.0 to `github.com/provenance-io/ibc-go/v8` v8.3.2-pio-1 ([#1910](https://github.com/provenance-io/provenance/pull/1910), [#1956](https://github.com/provenance-io/provenance/pull/1956), [#1998](https://github.com/provenance-io/provenance/pull/1998), [#2043](https://github.com/provenance-io/provenance/pull/2043))
- Bump `github.com/hashicorp/go-getter` from 1.7.3 to 1.7.5 ([#1958](https://github.com/provenance-io/provenance/pull/1958), [#2057](https://github.com/provenance-io/provenance/pull/2057))
- Bump `github.com/hashicorp/go-metrics` from 0.5.2 to 0.5.3 ([#1914](https://github.com/provenance-io/provenance/pull/1914))
- Bump `github.com/rs/cors` from 1.10.1 to 1.11.0 ([#2066](https://github.com/provenance-io/provenance/pull/2066))
- Bump `github.com/rs/zerolog` from 1.32.0 to 1.33.0 ([#1994](https://github.com/provenance-io/provenance/pull/1994))
- Bump `github.com/spf13/cobra` from 1.8.0 to 1.8.1 ([#2038](https://github.com/provenance-io/provenance/pull/2038))
- Bump `github.com/spf13/viper` from 1.18.2 to 1.19.0 ([#2020](https://github.com/provenance-io/provenance/pull/2020))
- Bump `golang.org/x/text` from 0.14.0 to 0.16.0 ([#1964](https://github.com/provenance-io/provenance/pull/1964), [#2023](https://github.com/provenance-io/provenance/pull/2023))
- Bump `golangci/golangci-lint-action` from 4 to 6 ([#1951](https://github.com/provenance-io/provenance/pull/1951), [#1965](https://github.com/provenance-io/provenance/pull/1965))
- Bump `google.golang.org/grpc` from 1.62.1 to 1.65.0 ([#1903](https://github.com/provenance-io/provenance/pull/1903), [#1918](https://github.com/provenance-io/provenance/pull/1918), [#1972](https://github.com/provenance-io/provenance/pull/1972), [#2065](https://github.com/provenance-io/provenance/pull/2065))
- Bump `google.golang.org/protobuf` from 1.33.0 to 1.34.2 ([#1960](https://github.com/provenance-io/provenance/pull/1960), [#1966](https://github.com/provenance-io/provenance/pull/1966), [#2028](https://github.com/provenance-io/provenance/pull/2028))
- Bump `peter-evans/create-pull-request` from 6.0.2 to 6.1.0 ([#1929](https://github.com/provenance-io/provenance/pull/1929), [#1940](https://github.com/provenance-io/provenance/pull/1940), [#1955](https://github.com/provenance-io/provenance/pull/1955), [#2040](https://github.com/provenance-io/provenance/pull/2040))

### Full Commit History

* https://github.com/provenance-io/provenance/compare/v1.18.0...v1.19.0
