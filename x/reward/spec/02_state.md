<!--
order: 2
-->

# State

## Reward Program

A reward program is the main data structure uses by the Application performance and engagement module(aka rewards module)


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

  uint64                   id                      = 1;
  string                   title                   = 2;
  string                   description             = 3;
  string                   distribute_from_address = 4; // community pool for now (who provides the money)
  cosmos.base.v1beta1.Coin total_reward_pool       = 5 [(gogoproto.nullable) = false];
  cosmos.base.v1beta1.Coin remaining_pool_balance  = 6 [(gogoproto.nullable) = false];
  cosmos.base.v1beta1.Coin claimed_amount          = 7 [(gogoproto.nullable) = false];
  cosmos.base.v1beta1.Coin max_reward_by_address   = 8
      [(gogoproto.nullable) = false]; // maximum reward per claim per address
  cosmos.base.v1beta1.Coin minimum_rollover_amount = 9
      [(gogoproto.nullable) = false]; // minimum amount of coins for program to rollover

  uint64 claim_period_seconds =
      10; // claim_period_seconds defines the type of claim_period attributed to this program.(e.g day,week,month)
  google.protobuf.Timestamp program_start_time = 11 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "program_start_time,omitempty",
    (gogoproto.moretags) = "yaml:\"program_start_time,omitempty\""
  ]; // When the reward program starts
  google.protobuf.Timestamp expected_program_end_time = 12 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "expected_program_end_time,omitempty",
    (gogoproto.moretags) = "yaml:\"expected_program_end_time,omitempty\""
  ]; // Time that the reward program MUST end

  google.protobuf.Timestamp claim_period_end_time = 13 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "claim_period_end_time,omitempty",
    (gogoproto.moretags) = "yaml:\"claim_period_end_time,omitempty\""
  ]; // This can be calculated by us and its when the current claim period ends

  google.protobuf.Timestamp actual_program_end_time = 14 [
    (gogoproto.stdtime)  = true,
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "finished_time,omitempty",
    (gogoproto.moretags) = "yaml:\"finished_time,omitempty\""
  ]; // when the program actually ended, will be empty at start

  uint64 claim_periods = 15; // number of claim periods this program will run for

  uint64 current_claim_period =
      16; // the current claim_period for the reward program is on(claim periods start at 1 <-- fisrt period)

  State state = 17; // the current state of the reward program

  uint64 reward_claim_expiration_offset = 18; // Used to calculate the expiration time of a reward claim in seconds. The
                                              // expiration timer begins when the program ends.

  repeated QualifyingAction qualifying_actions = 19 [
    (gogoproto.nullable) = false,
    (gogoproto.moretags) = "yaml:\"qualifying_actions"
  ]; // The actions that count towards the reward
}
```

## ClaimPeriod Reward Distribution

## RewardAccount State

## Qualifying Actions