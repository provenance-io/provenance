# Provenance Attribute Module

Manage attributes (typed key-value pairs) of provenance accounts. The supported attribute types are `json`, `string`, `uri`, `int`, `float`, `proto`, and `bytes`.

## CLI Example

__These examples assume you have set up and are runnning a local node using the [README](https://github.com/provenance-io/provenance/blob/main/README.md#local-setup) in the base directory.__

In addition, you should have bound names to account addresses. See the `Provenance Name Module` [README](https://github.com/provenance-io/provenance/blob/main/x/name/README.md).

### Transaction Commands

Add a the uuid attribute `provenance.pio`  to the `node0` account.

    build/provenanced tx attribute add \
		provenance.pio \
		$(build/provenanced keys show -t -a node0 --home build/node0 --keyring-backend test) \
		uuid \
		9cb1c906-6d59-44c3-91f3-ddb94d549632 \
		--testnet \
		--from node0 \
		--home build/node0 \
		--keyring-backend test \
		--chain-id chain-local \
		--fees 5000nhash \
		--broadcast-mode block --yes

Delete the `provenance.pio` attribute

     build/provenanced tx attribute delete \
        provenance.pio \
       	$(build/provenanced keys show -t -a node0 --home build/node0 --keyring-backend test) \
        --testnet \
		--from node0 \
		--home build/node0 \
		--keyring-backend test \
		--chain-id chain-local \
		--fees 5000nhash \
		--broadcast-mode block --yes
	
### Query Commands

Get all account attributes for `node0` account

    build/provenanced query attribute list \
		$(build/provenanced keys show -t -a node0 --home build/node0 --keyring-backend test) \
		--testnet

Get account attributes by name

	build/provenanced query attribute get \
		$(build/provenanced keys show -t -a node0 --home build/node0 --keyring-backend test) \
		provenance.pio \
		--testnet

Scan account attributes by name suffix
	
	build/provenanced query attribute scan \
		$(build/provenanced keys show -t -a node0 --home build/node0 --keyring-backend test) \
		pio \
		--testnet

Query the current name parameters
	
	build/provenanced query attribute params
	
