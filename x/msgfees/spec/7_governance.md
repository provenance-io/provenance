# Governance Proposal Control

The msgfee module supports addition, update, and deletion of Msg Type which are assessed fees via governance proposal.



## Add MsgFee Proposal

AddMsgBasedFeeProposal defines a governance proposal to create a new msgfee entry for a specific `MsgType`.

```protobuf
// AddMsgBasedFeeProposal defines a governance proposal to add additional msg based fee
message AddMsgBasedFeeProposal {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = false;

  string title       = 1;
  string description = 2;

  string msg_type_url = 3;

  cosmos.base.v1beta1.Coin additional_fee = 4 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (gogoproto.moretags)     = "yaml:\"additional_fee\""
  ];
}

```

