<!--
order: 4
-->

# Messages

In this section we describe the processing of the reward messages and the corresponding updates to the state.

<!-- TOC 2 -->
  - [Msg/CreateRewardProgramRequest](#msg-createrewardprogramrequest)
  - [Msg/EndRewardProgramRequest](#msg-endrewardprogramrequest)
  - [Msg/ClaimRewardRequest](#msg-claimrewardrequest)
  - [Msg/ClaimAllRewardsRequest](#msg-claimallrewardsrequest)


## Msg/CreateRewardProgramRequest

Creates a Reward Program that users can participate in.

### Request
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L30-L55

### Response
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L58-L60

The message will fail under the following conditions:
* The program start time is at the current block time or after
* The requester is unable to send the reward pool amount to module
* The title is empty or greater than 140 characters
* The description is empty or greater than 10000 characters
* The distribute from address is an invalid bech32 address
* The total reward pool amount is not positive
* The claim periods field is set to less than 1
* The denominations are not in nhash
* There are no qualifying actions
* The qualifying actions are not valid

## Msg/EndRewardProgramRequest

Ends a Reward Program that is in either the PENDING or STARTED state.

### Request
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L63-L69

### Response
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L72-L73

The message will fail under the following conditions:
* The Reward Program does not end
* The Reward Program is not in PENDING or STARTED state
* The Reward Program owner does not match the specified address

## Msg/ClaimRewardRequest

Allows a participant to claim all their rewards for all past claim periods on a reward program.

### Request
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L76-L79

### Response
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L82-L86

The message will fail under the following conditions:
* The Reward Program does not exist
* The Reward Program is expired
* The Reward Address does not exist

## Msg/ClaimAllRewardsRequest

Allows a participant to claim all their rewards for all past claim periods on all reward programs.

### Request
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L89-L91

### Response
+++ https://github.com/provenance-io/provenance/blob/f77baab1ffe688b05c9e9e587632e28aad723898/proto/provenance/reward/v1/tx.proto#L94-L99

The message will fail under the following conditions:
* The Reward Address does not exist