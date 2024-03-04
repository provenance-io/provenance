# Building Provenance Blockchain

<!-- TOC -->
  - [Overview](#overview)
  - [Prerequisites](#prerequisites)
    - [Go](#go)
  - [Building or Installing `provenanced`](#building-or-installing-provenanced)
  - [Build Options](#build-options)
  - [Building `dbmigrate`](#building-dbmigrate)



## Overview

Provenance uses `make` to define build operations.
Built executables are placed in the `build/` directory.

## Prerequisites

### Go

Building `provenanced` requires [Go 1.21+](https://golang.org/dl/) (or higher).

## Building or Installing `provenanced`

To build the `provenanced` executable and place it in the `build/` directory:
```console
$ make build
```

To build the `provenanced` executable and place it in your system's default Go `bin/` directory.
```console
$ make install
```

To use a specific version of `provenanced`, check out that version's tag, then build or install it.
For example:
```console
$ git checkout "v1.7.6" -b "tag-v1.7.6"
$ make install
```

## Build Options

A few aspects of `make build` and `make install` can be controlled through environment variables.

* `WITH_LEDGER`: Enables/Disables building with Ledger hardware wallet support.
  The default is `true`.
  If this is not `true` the built `provenanced`, executable will not work with Ledger hardware wallets.
* `GO`: The GoLang executable.
  The default is `go`.
* `BINDIR`: The path to the Go binary directory.
  The default is `${GOPATH}/bin`.
* `BUILDDIR`: The path to the directory where the built executable should be placed.
  The default is `./build`.
* `VERSION`: The string to use as the output of `provenanced version`.
  The default is `{branch name}-{short commit hash}`.
* `BUILD_TAGS`: Any extra `-tags` to supply to the `go build` or `go install` invocations.
  These are appended to a list constructed by the Makefile.
* `LDFLAGS`: Any extra `-ldflags` to supply to the `go build` or `go install` invocations.
  These are appended to a list constructed by the Makefile.
* `CGO_LDFLAGS`: Anything extra to include in the CGO_LDFLAGS env var when invoking `go build` or `go install`.
  These are appended to a list constructed by the Makefile.
* `CGO_CFLAGS`: Anything extra to include in the CGO_CFLAGS env var when invoking `go build` or `go install`.
  These are appended to a list constructed by the Makefile.
* `BUILD_FLAGS`: Any extra flags to include when invoking `go build` or `go install.`.
  These are appended to a list constructed by the Makefile.

## Building `dbmigrate`

The `dbmigrate` utility can be used to migrate a node's data directory to a use a different db backend.

To build the `dbmigrate` executable and place it in the `build/` directory:
```console
$ make build-dbmigrate
```

To build the `dbmigrate` executable and place it in your system's default Go `bin/` directory.
```console
$ make install-dbmigrate
```

Building `dbmigrate` uses the same [Build Options](#build-options) as `provenanced`.

The dbmigrate program will:
1. Create a new `data/` directory, and copy the contents of the existing `data/` directory into it, converting the database files appropriately.
2. Back up the existing `data/` directory to `${home}/data-dbmigrate-backup-{timestamp}-{dbtypes}/`.
3. Move the newly created `data/` directory into place.
4. Update the config's `db_backend` value to the new db backend type.

The `dbmigrate` utility uses the same configs, environment variables, and flags as `provenanced`.
For example, if you have the environment variable PIO_HOME defined, then `dbmigrate` will use that as the `--home` directory (unless a `--home` is provided in the command line arguments).
