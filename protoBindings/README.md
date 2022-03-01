# Cosmos Blockchain Protobuf Build

The Cosmos SDK uses a combination of [Cosmos](https://github.com/cosmos/cosmos-sdk) 
and [IBC](https://github.com/cosmos/ibc-go) [protobuf](https://developers.google.com/protocol-buffers) definitions.
Protocol buffers (protobuf) are Google's language-neutral, platform-neutral,
extensible mechanism for serializing structured data.  The Cosmos
[gRPC](https://grpc.io) and protobuf provide the RPC mechanism that Cosmos SDK uses
to communicate with the blockchain.

## Development
If dependencies appear to be missing in your IDE, run
```bash
./gradlew clean build --refresh-dependencies
```

## Build Proto
Checkout `./bindings` for all supported language bindings.

### Java
```bash
./gradlew clean :java:jar
```

### Kotlin
```bash
./gradlew clean :kotlin:jar
```
