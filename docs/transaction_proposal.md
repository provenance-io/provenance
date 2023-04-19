# Transaction Proposal Submission
In this tutorial we will learn how to submit a proposal and then vote on it. Proposals can take a lengthy amount of time, and in order to bypass this we update the genesis file to have a shorter voting period.

## Author
- Matthew Witkowski

## Prequisites
The reader should first have understanding of what Provenance is and how to use its CLI. Additionally, the user should have the following setup.

- A local Provence directory that can be built and run
- An env variable named ACCOUNT that contains an address with funds
- A smart contract named `contract.wasm`

## Setup
First navigate to your Provenance directory and begin by making a new build and configuration.

`make clean`
`make build`
`make run-config`

Next, we need to modify the `voting_period` within `build/run/provenanced/config/genesis.json`. Open up the genesis file, set the following field, and then save your changes.

`"voting_period": "60s"`

This will allow users to vote on a proposal for a minute. This should be more than enough time for testing. The blockchain can now be started and use the voting period modifications.

`make run`

### Note
A new proposal can be made if the voting time expires.

## Creating the Proposal
In order to propose we need a json file containing our messages to run if the proposal passes. We can make use of the `--generate-only` flag on a transaction to easily give us the JSON. In this example, the `wasm store` transaction is being proposed. The code for the contract will only be stored if the proposal passes. In order to obtain the JSON for this transaction the following command can be ran.

`provenanced -t tx wasm store contract.wasm --from $ACCOUNT --gas auto --gas-adjustment 1.5 --gas-prices 1905nhash --instantiate-everybody "true" --generate-only`

Take the output of this and insert it into the `messages` array in the following JSON.

```
{
    "messages": [],
    "metadata": "",
    "deposit": "10000000nhash"
}
```

It should now look like the following...

```
{
    "messages": [
        {
            "@type": "/cosmwasm.wasm.v1.MsgStoreCode",
            "sender": "tp10d07y265gmmuvt4z0w9aw880jnsr700jq8ave9",
            "wasm_byte_code": "xyz=",
            "instantiate_permission": {
                "permission": "Everybody",
                "address": "",
                "addresses": []
            }
        }
    ],
    "metadata": "",
    "deposit": "10000000nhash"
}
```

The only remaining issue is the `sender`. In order for the transaction to pass it must have the same address as the governance module. We can obtain the module's address with the following command.

`provenanced -t q auth module-account gov`

Update the sender and save the file as `proposal.json`

## Proposing
The json file has been written to disk, but the proposal has not yet been sent out to Provenance. To propose it the following command can be ran:

`provenanced -t tx gov submit-proposal proposal.json --gas-prices 1905nhash --gas auto --gas-adjustment 1.5 --from $ACCOUNT`

The state and id of the proposal can be found with the following command:

`provenanced -t q gov proposals`

The proposal id in this example is 4 so the following transaction will vote on proposal id 4.
`provenanced -t tx gov vote 4 yes --from $ACCOUNT --gas auto --gas-adjustment 1.5 --gas-prices 1905nhash`

Lastly, proposal 4 can be monitored by using the following command. The proposal will eventually succeed or fail depending on the transaction message that was passed in.
`provenanced -t q gov proposal 4`