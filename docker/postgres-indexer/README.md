# Postgres Indexing Database Docker

This docker container is for using the postgres indexer.  

To configure postgres indexing change indexer configuration in the `config.toml`

```toml
[tx_index]

indexer = "psql"
psql-conn = "postgresql://postgres:password1@localhost:5432/tendermint?sslmode=disable"

```

```console
docker-compose -f docker/postgres-indexer/docker-compose.yaml
```