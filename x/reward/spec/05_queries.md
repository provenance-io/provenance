<!--
order: 5
-->

# Rewards Queries
In this section we describe the queries available for looking up rewards information.

<!-- TOC 2 -->
  - [Query Reward Program By ID](#query-reward-program-by-id)
  - [Query Reward Programs](#query-reward-programs)
  - [Query Claim Period Reward Distribution By ID](#query-claim-period-reward-distribution-by-id)
  - [Query Claim Period Reward Distributions](#query-claim-period-reward-distributions)
  - [Query Rewards By Address](#query-rewards-by-address)


---
## Query Reward Program By ID
The `QueryRewardProgramByID` query is used to obtain the content of a specific Reward Program.

### Request
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L47-L51

The `id` is the unique identifier for the Reward Program.

### Response
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L53-L57


---
## Query Reward Programs
The `QueryRewardPrograms` query is used to obtain the content of all Reward Programs matching the supplied `query_type`.

### Request
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L59-L80

The `query_type` is used to filter on the Reward Program state. The following are a list of `query_types`.
* ALL - All Reward Programs will be returned.
* PENDING - All Reward Programs that are in the `PENDING` state will be returned.
* ACTIVE - All Reward Programs that are in the `STARTED` state will be returned.
* OUTSTANDING - All Reward Programs that are either in the `PENDING` or `STARTED` state will be returned.
* FINISHED - All Reward Programs that are in the `FINISHED` or `EXPIRED` state will be returned.

### Response
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L82-L88


---
## Query Claim Period Reward Distribution By ID
The `QueryClaimPeriodRewardDistributionByID` query is used to obtain the content of a specific `Claim Period Reward Distribution`.

### Request
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L104-L110

The `reward_id` is a unique identifier for the Reward Program and the `claim_id` is a unique identifier for the Reward Program's Claim Period.

### Response
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L112-L116


---
## Query Claim Period Reward Distributions
The `QueryClaimPeriodRewardDistributions` query is used to obtain the content of all `Claim Period Reward Distributions` matching the supplied `query_type`.

### Request
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L90-L94

The `pagination` field is used to help limit the number of results.

### Response
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L96-L102


---
## Query Rewards By Address
The `QueryRewardsByAddress` query is used to obtain the status of the address' Reward Claims.

### Request
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L118-L126

The `address` field is the bech32 address of the user to list Reward Claims for. The `claim_status` is used to filter on the Reward Claim. The following are a list of `claim_status`.
* ALL - All Reward Claims are returned.
* UNCLAIMABLE - All Reward Claims that are not yet eligible to be claimed.
* CLAIMABLE - All Reward Claims that are still eligible to be claimed.
* CLAIMED - All Reward Claims that have been claimed.
* EXPIRED - All Reward Claims that have expired.

### Response
+++ https://github.com/provenance-io/provenance/blob/243a89c76378bb5af8a8017e099ee04ac22e99ce/proto/provenance/reward/v1/query.proto#L128-L136
