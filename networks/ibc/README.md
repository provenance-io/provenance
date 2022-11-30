## Inter-Blockchain Communication (IBC) Test Cluster
# Overview

This folder contains the files to create 3 containers for testing IBC. The first container
is a single node running a blockchain with chain-id testing. The second container is also a
single node running a blockchain, but this chain has chain-id testing2. Lastly, the third container
holds the relayer used to forward packets between chains.

# Relayer Account

Each chain has an account for the relayer that is funded. The relayer uses this account
to publish packets from the sending chain to the receiving chain.

```
- type: local
  address: tp18uev5722xrwpfd2hnqducmt3qdjsyktmtw558y
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aj/FqZi4LKRTT+q0tByQsqk27uIMO9nf1UnkTyUaQRV5"}'
  mnemonic: ""
- balances:
  amount: "99999998039505445"
  denom: nhash
```

# Networking

It's always important to first verify that your relayer is working before doing anything
IBC related. You can verify that your relayer is working as inteded by running the following
command on the relayer container.

```
rly --home /relayer paths list
0: local_local2         -> chns(✔) clnts(✔) conn(✔) (testing<>testing2)
```

Notice that the output above has 3 checkmarks. This is saying that the channels, clients,
and connections are all properly working.

# Transferring Currency

The simplest test that can be used to verify that IBC is working is using the Fungible Token
Transfer (ICS-20) subprotocol. This subprotocol is exposed through provenance with the `ibc-transfer`
transaction. It transfers coins from an account on one chain to an account on another chain.
In this example we will transfer from an account on ibc0-0 to an account on ibc1-0.

## Note
The following tutorial assumes the user has built provenanced, and their current working directory
is within the root of their provenance source directory.

1. First, obtain the receiving address from ibc1-0. In this example our receiving address is
`tp1vtvgsl9je747twlxkh4ycl2g3td6g5gcpc6t0y`.

```
./build/provenanced -t --home ./build/ibc1-0/ keys list

- address: tp1vtvgsl9je747twlxkh4ycl2g3td6g5gcpc6t0y
  name: ibc1-0
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AjOKuV/CkfDBmZgSHZFrCN2PXz68ZyUssvBiWfe4A/ut"}'
  type: local
```

2. Next, obtain your sending address from ibc0-0. In this example our sending address is
`tp1u3ry0ry80hvj9vcfa8h5e30wkx9ec4l5jsqujd`.

```
./build/provenanced -t --home ./build/ibc0-0/ keys list

- address: tp1u3ry0ry80hvj9vcfa8h5e30wkx9ec4l5jsqujd
  name: ibc0-0
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A5RlZL3Gn3uEyzrVIkk9nyDV5LxJeUqPF/kOQmGX6nnQ"}'
  type: local
```

3. Now, we can transfer currency from ibc0-0 to ibc1-0. The following command sends 500nhash from our
sending account `tp1u3ry0ry80hvj9vcfa8h5e30wkx9ec4l5jsqujd` to our receiving account
`tp1vtvgsl9je747twlxkh4ycl2g3td6g5gcpc6t0y`.

```
./build/provenanced -t --home ./build/ibc0-0/ tx ibc-transfer transfer transfer channel-0 tp1vtvgsl9je747twlxkh4ycl2g3td6g5gcpc6t0y 500nhash --from tp1u3ry0ry80hvj9vcfa8h5e30wkx9ec4l5jsqujd --gas auto --gas-prices 1905nhash --gas-adjustment 1.5 --chain-id testing --node http://localhost:26657 -y
```

4. Lastly, let's check the balances for the receiving account on container ibc1-0. You should see the 500nhash sent
from ibc0-0. It will have its own denom specified with ibc/denom...

```
./build/provenanced -t --home ./build/ibc1-0/ q bank balances tp1vtvgsl9je747twlxkh4ycl2g3td6g5gcpc6t0y --node http://localhost:26660

balances:
- amount: "500"
  denom: ibc/319937B2FDA7A07031DBE22EA76C34CAC9DCFBD9AA1A922FA2B87421107B545D
- amount: "99899999900000000000"
  denom: nhash
```

# Smart Contracts

Smart Contracts can be stored and instantiated on this chain like any other. The only benefit
that this chain offers for testing smart contracts is support for the IBC entry points. Users
do not have to worry about setting up multiple chains and a relayer just to test out the IBC
portions of their smart contracts.

Please see the provwasm tutorial for more information on uploading and using contracts.
https://github.com/provenance-io/provwasm/blob/main/docs/tutorial/01-overview.md


NOTE: This folder contains files to run an image using a locally built binary 
that leverages Go's ability to target platforms during builds.  These docker
files are _not_ the ones used to build the release image.  For those images
see the `docker/blockchain` folder in the project root.

