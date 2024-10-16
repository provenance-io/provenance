Building or installing `provenanced` from source now requires you to use [Go 1.23](https://golang.org/dl/).
Linting now requires `golangci-lint` v1.60.2. You can update yours using `make golangci-lint-update` or install it using `make golangci-lint`.

Version `v1.20.0-rc3` should be used in place of `v1.20.0-rc2`. Version `v1.20.0-rc2` doesn't allow restarting a node once it has been stopped (after applying the `viridian-rc1` upgrade). Switching to `v1.20.0-rc3` will fix the error `failed to load latest version: version of store params mismatch root store's version`. It is also safe to use `v1.20.0-rc3` to apply the upgrade (even though the upgrade says to use `v1.20.0-rc2`).
