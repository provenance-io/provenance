I wonder if we can run the rosetta docker images in cosmos?


```bash
provenanced rosetta  --blockchain="provenance" --network="testing" --tendermint="localhost:26657" --grpc="localhost:9090" --addr=":8080" --home ./build/run/provenanced
```

```bash
./bin/rosetta-cli check:construction --configuration-file ./code/provenance/client/rosetta/default.json
```
