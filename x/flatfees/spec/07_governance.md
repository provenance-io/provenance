<!--
order: 7
-->

# Governance Proposal Control

The msgfee module supports addition, update, and deletion of Msg Type which are assessed fees via governance proposal.

<!-- TOC -->
  - [Add MsgFee Proposal](#add-msgfee-proposal)
  - [Update MsgFee Proposal](#update-msgfee-proposal)
  - [Remove MsgFee Proposal](#remove-msgfee-proposal)



## Add MsgFee Proposal

AddMsgFeeProposal defines a governance proposal to create a new msgfee entry for a specific `MsgType`.

Add proposal [AddMsgFeeProposal](../../../proto/provenance/msgfees/v1/proposals.proto#L13-L34):

```protobuf
// AddMsgFeeProposal defines a governance proposal to add additional msg based fee
message AddMsgFeeProposal {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = true;

  // propsal title
  string title = 1;
  // propsal description
  string description = 2;

  // type url of msg to add fee
  string msg_type_url = 3;

  // additional fee for msg type
  cosmos.base.v1beta1.Coin additional_fee = 4
  [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];

  // optional recipient to recieve basis points
  string recipient = 5;
  // basis points to use when recipient is present (1 - 10,000)
  string recipient_basis_points = 6;
}
```

Sample command to add an additional fee locally:

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

Update proposal [UpdateMsgFeeProposal](../../../proto/provenance/msgfees/v1/proposals.proto#L36-L55):

```protobuf
// UpdateMsgFeeProposal defines a governance proposal to update a current msg based fee
message UpdateMsgFeeProposal {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = true;

  // propsal title
  string title = 1;
  // propsal description
  string description = 2;
  // type url of msg to update fee
  string msg_type_url = 3;

  // additional fee for msg type
  cosmos.base.v1beta1.Coin additional_fee = 4
  [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
  // optional recipient to recieve basis points
  string recipient = 5;
  // basis points to use when recipient is present (1 - 10,000)
  string recipient_basis_points = 6;
}
```

## Remove MsgFee Proposal

Remove proposal [RemoveMsgFeeProposal](../../../proto/provenance/msgfees/v1/proposals.proto#L57-L68):

```protobuf
// RemoveMsgFeeProposal defines a governance proposal to delete a current msg based fee
message RemoveMsgFeeProposal {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = true;

  // propsal title
  string title = 1;
  // propsal description
  string description = 2;
  // type url of msg fee to remove
  string msg_type_url = 3;
}
```
