# Pre-authorized transfer on Provenance

In Provenance a user can pre-authorize another user to transfer a restricted coin, with a specified limit, from their account to any other account.  This authorization can be revoked or modified at a later date.  Also, this ability is only allowed on restricted coins, so the base coin, hash, does not have the ability to do this.

### Overview of how to do this
In order to grant authorization for a user to transfer restricted coin, a user can take advantage of the `grant-authz` command on the marker module.  `provenanced tx marker grant-authz --help` can be used in order to view all of the possible flags and inputs.  However, in order for it to run, it is necessary to give this command the address of the user that will be pre-authorized to conduct transfers, the type of action that is pre-authorized which in this case is `transfer`, a transfer-limit, and an address to sign the transaction with using the `--from` flag.

The flag `spend-limit` is used to set the total amount of coin that can be transfered.  Each transfer will deduct from this total until it is exhausted at which point no more transfers can be made. However, the spend limit can be reset by the user doing another pre-authorize transaction.  Note that this new spend limit does not take into account what has already been spent.  So, for example if you grant permissions to transfer 100 coins and 50 are transfered and then you set a new spend limit of 75, that allows the user to now transfer 75 coins.  The previously spent coins are not taken into account with the new spend limit. 

If a user wants to revoke authorization to transfer then they can use the `revoke-authz` command on the marker module.  All the flags and inputs can be found with: `provenanced tx marker revoke-authz --help`.  This command needs the address of the user whose authorization is being revoked, the type of action that is being revoked, and the signature of the address revoking permissions to itself with the `--from` flag.

### Practical example on a local node

All of this can be tested and demonstrated on a local node with the following steps:

#### Running a local node and adding keys with hash to sign txs

Run the following commands in your terminal to start a clean local build:

```bash 
make clean install build run
```

In a separate terminal after the node is running do the following to add users to test pre-authorization:
```bash
provenanced -t keys add user1 --home=./build/run/provenanced
provenanced -t keys add user2 --home=./build/run/provenanced
provenanced -t keys add user3 --home=./build/run/provenanced
```

Verify that all the keys were created with the following command:
```bash
provenanced -t keys list --home=./build/run/provenanced
```

Run the following command to set env variables for the user addresses and the validator address to make running commands a lot easier:
```bash
addr1=$(provenanced -t keys show user1 -a --keyring-backend=test --home=./build/run/provenanced)
addr2=$(provenanced -t keys show user2 -a --keyring-backend=test --home=./build/run/provenanced)
addr3=$(provenanced -t keys show user3 -a --keyring-backend=test --home=./build/run/provenanced)
addrV=$(provenanced -t keys show validator -a --keyring-backend=test --home=./build/run/provenanced)
```

Then transfer some hash to user1 and user2 from the validator node so that they can pay gas fees for txs:

```bash
provenanced -t tx bank send  "$addrV" "$addr1" 4762500000000nhash --chain-id="testing" --node tcp://localhost:26657 --yes --keyring-backend=test  --home=./build/run/provenanced --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes
provenanced -t tx bank send  "$addrV" "$addr2" 4762500000000nhash --chain-id="testing" --node tcp://localhost:26657 --yes --keyring-backend=test  --home=./build/run/provenanced --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes
```

Then verify that the hash was transferred correctly with the following commands:

```bash
provenanced -t q bank balances "$addr1" --home=./build/run/provenanced
provenanced -t q bank balances "$addr2" --home=./build/run/provenanced
```

#### Creating a restricted coin
All of the information for this section and an explanation of how creating coins works can be found [here](https://docs.provenance.io/blockchain/basics/stablecoin)

In our local example we are going to create a restricted coin called `bitcoin` and give a large amount of it to user1.

To create `bitcoin` run the following command:
```bash
provenanced -t --chain-id="testing" tx marker new 1000000bitcoin  --type RESTRICTED --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
```

Give user1 all of the necessary privileges to work with this new currency:
```bash
provenanced -t --chain-id="testing" tx marker grant "$addr1" bitcoin admin --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
provenanced -t --chain-id="testing" tx marker grant "$addr1" bitcoin mint --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
provenanced -t --chain-id="testing" tx marker grant "$addr1" bitcoin burn --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
provenanced -t --chain-id="testing" tx marker grant "$addr1" bitcoin withdraw --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
provenanced -t --chain-id="testing" tx marker grant "$addr1" bitcoin transfer --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
```

We also need to give user2 the ability to transfer bitcoin so that he can be pre-authorized to transfer from user1 to user3.  Do that with the following command:
```bash
provenanced -t --chain-id="testing" tx marker grant "$addr2" bitcoin transfer --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
```

To activate and finalize this new coin run the following commands"
```bash
provenanced -t --chain-id="testing" tx marker finalize bitcoin --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
provenanced -t --chain-id="testing" tx marker activate bitcoin --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
```

To withdraw some of the currency to user1's account run the following command:

```bash
provenanced -t --chain-id="testing" tx marker withdraw bitcoin 500000bitcoin "$addr1" --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes --home=./build/run/provenanced
```

#### Granting pre-authorization
In order to pre-authorize user2 the ability to transfer up to 100 bitcoin from user1 run the following:

```bash
provenanced -t --chain-id="testing" tx marker grant-authz "$addr2" transfer --transfer-limit=100bitcoin --home=./build/run/provenanced --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes
```

#### Transferring coins
Run the following command for user2 to transfer 10bitcoin from user1 to user3 without need user1's signature:

```bash
provenanced -t tx marker transfer "$addr1" "$addr3" 10bitcoin --chain-id="testing" --home=./build/run/provenanced --from "$addr2" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes
```

Verify that the 10bitcoin successfully transferred to user3 with the following command:
```bash
provenanced -t q bank balances "$addr3" --home=./build/run/provenanced
```

If you want to verify that user2 can't transfer more than the 100bitcoin limit you can run the following command and verify that it fails with an error message stating that a transfer was attempted above the spend limit.
```bash
provenanced -t tx marker transfer "$addr1" "$addr3" 95bitcoin --chain-id="testing" --home=./build/run/provenanced --from "$addr2" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes
```

#### Revoking pre-authorization
In order to revoke permissions to user2 run the following command:
```bash
provenanced -t --chain-id="testing" tx marker revoke-authz "$addr2" transfer --home=./build/run/provenanced --from "$addr1" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes
```

Verify that user2 can no longer transfer bitcoin from user1 by running the following command and verifying that you receive an error message stating that user2 no longer has permissions to transfer bitcoin from user1.
```bash
provenanced -t tx marker transfer "$addr1" "$addr3" 10bitcoin --chain-id="testing" --home=./build/run/provenanced --from "$addr2" --gas-prices="1905nhash" --gas=auto --gas-adjustment=1.5 --yes
```
