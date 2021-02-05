# Provenance Name Module

Bind human readable names to provenance addresses.

## CLI Example

__These examples assume you have set up and are runnning a local node using the [README](https://github.com/provenance-io/provenance/blob/main/README.md#local-setup) in the base directory.__


### Transaction Commands

Bind the unrestricted name "provenance" under "pio" using the root account

	build/provenanced tx name bind \
		provenance \
		$(build/provenanced keys show -t -a node0 --home build/node0 --keyring-backend test) \
		pio \
		--testnet \
		--from node0 \
		--restrict=false \
		--home build/node0 \
		--chain-id chain-local \
		--keyring-backend test \
		--fees 5000nhash \
		--broadcast-mode block --yes

Delete the unrestricted name "provenance" under "pio" using the root account

	build/provenanced tx name delete \
		provenance.pio \
		--testnet \
		--from node0 \
		--home build/node0 \
		--chain-id chain-local \
		--keyring-backend test \
		--fees 5000nhash \
		--broadcast-mode block --yes

### Query Commands

Query the current name parameters

	build/provenanced query name params
	
Reverse lookup of all names bound to a given address

    build/provenanced query name lookup $(build/provenanced keys show -t -a node0 --home build/node0 --keyring-backend test) -t
	
Resolve the address for a name

	build/provenanced query name resolve pio
