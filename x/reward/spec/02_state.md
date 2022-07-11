<!--
order: 2
-->

# State

## Reward Program

A reward program is the main data structure used by the Active Participation and Engagement (APE) module.

```go
// RewardProgram
message RewardProgram {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = false;
  enum State {
    PENDING  = 0;
    STARTED  = 1;
    FINISHED = 2;
    EXPIRED  = 3;
  }

  uint64                   id                      = 1; // An integer to uniquely identify the reward program.
  string                   title                   = 2; // Name to help identify the Reward Program.
  string                   description             = 3; // Short summary describing the Reward Program.
  string                   distribute_from_address = 4; // Community pool for now (who provides the money)
  cosmos.base.v1beta1.Coin total_reward_pool       = 5
      [(gogoproto.nullable) = false]; // The total amount of funding given to the RewardProgram.
  cosmos.base.v1beta1.Coin remaining_pool_balance = 6
      [(gogoproto.nullable) = false]; // The remaining funds available to distribute.
  cosmos.base.v1beta1.Coin claimed_amount = 7
      [(gogoproto.nullable) = false]; // The total amount of funds claimed by participants.
  cosmos.base.v1beta1.Coin max_reward_by_address = 8
      [(gogoproto.nullable) = false]; // Maximum reward per claim per address
  cosmos.base.v1beta1.Coin minimum_rollover_amount = 9
      [(gogoproto.nullable) = false]; // Minimum amount of coins for a program to rollover

  uint64                    claim_period_seconds = 10; // Number of seconds that a claim period lasts.
  google.protobuf.Timestamp program_start_time   = 11 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "program_start_time,omitempty",
    (gogoproto.moretags) = "yaml:\"program_start_time,omitempty\""
  ]; // Time that a RewardProgram should start and switch to STARTED state.
  google.protobuf.Timestamp expected_program_end_time = 12 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "expected_program_end_time,omitempty",
    (gogoproto.moretags) = "yaml:\"expected_program_end_time,omitempty\""
  ]; // Time that a RewardProgram MUST end.

  google.protobuf.Timestamp claim_period_end_time = 13 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "claim_period_end_time,omitempty",
    (gogoproto.moretags) = "yaml:\"claim_period_end_time,omitempty\""
  ]; // Used internally to calculate and track the current claim period's ending time.

  google.protobuf.Timestamp actual_program_end_time = 14 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "finished_time,omitempty",
    (gogoproto.moretags) = "yaml:\"finished_time,omitempty\""
  ]; // Time the RewardProgram switched to FINISHED state. Initially set as empty.

  uint64 claim_periods = 15; // Number of claim periods this program will run for

  uint64 current_claim_period = 16; // Current claim period of the RewardProgram. Uses 1-based indexing.

  State state = 17; // Current state of the RewardProgram.

  uint64 reward_claim_expiration_offset = 18; // Grace period after a RewardProgram FINISHED. It is the number of
                                              // seconds until a RewardProgram enters the EXPIRED state.

  repeated QualifyingAction qualifying_actions = 19 [
    (gogoproto.nullable) = false,
    (gogoproto.moretags) = "yaml:\"qualifying_actions"
  ]; // Actions that count towards the reward
}
```

## ClaimPeriod Reward Distribution

## RewardAccount State

## Qualifying Actions