<!--
order: 7
-->

# Governance Proposal Control

The msgfee module supports addition, update, and deletion of Msg Type which are assessed fees via governance proposal.



## Add MsgFee Proposal

AddMsgFeeProposal defines a governance proposal to create a new msgfee entry for a specific `MsgType`.

Add proposal [AddMsgFeeProposal](../../../proto/provenance/msgfees/v1/proposals.proto#L19-L34):
```protobuf
// AddMsgFeeProposal defines a governance proposal to add additional msg based fee
message AddMsgFeeProposal {
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
sample command to add an additional fee locally 

```bash
  ${PROVENANCE_DEV_DIR}/build/provenanced -t tx msgfees proposal add "adding" "adding bank send addition fee" 10000000000nhash \
    --msg-type=/cosmos.bank.v1beta1.MsgSend --additional-fee 99usd.local\
		--from node0 \
    --home ${PROVENANCE_DEV_DIR}/build/node0 \
    --chain-id chain-local \
	--keyring-backend test \
    --gas auto \
    --fees 250990180nhash \
    --broadcast-mode block \
    --yes \
    --testnet
```
## Update MsgFee Proposal

Update proposal [UpdateMsgFeeProposal](../../../proto/provenance/msgfees/v1/proposals.proto#L36-L51):
```protobuf
// UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee
message UpdateMsgFeeProposal {
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

## Remove MsgFee Proposal

Remove proposal [RemoveMsgFeeProposal](../../../proto/provenance/msgfees/v1/proposals.proto#L53-L62):
```protobuf
// RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee
message RemoveMsgFeeProposal {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = false;

  string title       = 1;
  string description = 2;

  string msg_type_url = 3;
}
```
