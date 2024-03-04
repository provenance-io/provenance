# Building Provenance Blockchain

<!-- TOC -->
  - [Overview](#overview)
  - [Prerequisites](#prerequisites)
    - [Go](#go)
  - [Building or Installing `provenanced`](#building-or-installing-provenanced)
  - [Build Options](#build-options)



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
