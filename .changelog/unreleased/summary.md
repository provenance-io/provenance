Provenance Blockchain version `v1.20.1` contains several improvements and bug fixes. Validators will switch to it with the `viridian` upgrade, which is tentatively scheduled for 2024-11-14 5:15 PM Eastern time.

This version fixes a [security vulnerability](https://github.com/cometbft/cometbft/security/advisories/GHSA-p7mv-53f2-4cwj) in the cometbft library.

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.
