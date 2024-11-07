Provenance Blockchain version `v1.20.0` contains several improvements and bug fixes.

This version is superseded by `v1.20.1` due to a vulnerability found in the cometbft library.

Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.
