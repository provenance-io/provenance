# Building Provenance Blockchain

<!-- TOC -->
  - [Overview](#overview)
  - [Prerequisites](#prerequisites)
    - [Go](#go)
    - [CLevelDB](#cleveldb)
    - [RocksDB](#rocksdb)
    - [librdkafka](#librdkafka)
  - [Building or Installing `provenanced`](#building-or-installing-provenanced)
  - [Build Options](#build-options)
  - [Building `dbmigrate`](#building-dbmigrate)



## Overview

Provenance uses `make` to define build operations.
Built executables are placed in the `build/` directory.

## Prerequisites

### Go

Building `provenanced` requires [Go 1.17+](https://golang.org/dl/) (or higher).

### CLevelDB

By default, `provenanced` is built with CLevelDB support.
Building without CLevelDB support is also possible. See `WITH_CLEVELDB` in [Build Options](#build-options) below.

To download, build, and install the C LevelDB library on your system:
```console
$ make cleveldb
```

<details>
<summary>Environment variables that can control the behavior of this command:</summary>

* `CLEVELDBDB_VERSION` will install a version other than the one defined in the `Makefile`.
  Do not include the `v` at the beginning of the version number.
  Example: `CLEVELDBDB_VERSION=1.22 make cleveldb`.
  The default is `1.23`
* `CLEVELDB_JOBS` will control the number of parallel jobs used to build the library.
  The default is the result of the `nproc` command.
  More parallel jobs can speed up the build.
  Fewer parallel jobs can alleviate memory problems/crashes that can be encountered during a build.
* `CLEVELDB_DO_BUILD` defines whether to build cleveldb.
  The default is `true`.
* `CLEVELDB_DO_INSTALL` defines whether to install cleveldb.
  The default is `true`.
* `CLEVELDB_SUDO` defines whether to use `sudo` for the installation of the built library.
  The difference between `sudo make cleveldb` and `CLEVELDB_SUDO=true make cleveldb`
  is that the latter will use `sudo` only for the installation (the download and build still use your current user).
  Some systems (e.g. Ubuntu) might require this.
  The default is `true` if the `sudo` command is found, or `false` otherwise.
* `CLEVELDB_DO_CLEANUP` defines whether to delete the downloaded and unpacked repo when done.
  The default is `true`.
</details>

### RocksDB

By default, `provenanced` is built without RocksDB support.
Building with RocksDB support is also possible. See `WITH_ROCKSDB` in [Build Options](#build-options) below.

To download, build, and install the RocksDB library on your system:
```console
$ make rocksdb
```

<details>
<summary>Environment variables that can control the behavior of this command:</summary>

* `ROCKSDB_VERSION` will install a version other than the one defined in the `Makefile`.
  Do not include the `v` at the beginning of the version number.
  Example: `ROCKSDB_VERSION=6.17.3 make rocksdb`.
  The default is `6.29.4`
* `ROCKSDB_JOBS` will control the number of parallel jobs used to build the library.
  The default is the result of the `nproc` command.
  More parallel jobs can speed up the build.
  Fewer parallel jobs can alleviate memory problems/crashes that can be encountered during a build.
* `ROCKSDB_WITH_SHARED` defines whether to build and install the shared (dynamic) library.
  The default is `true`.
* `ROCKSDB_WITH_STATIC` defines whether to build and install the static library.
  The default is `false`.
* `ROCKSDB_DO_BUILD` defines whether to build rocksdb.
  The default is `true`.
* `ROCKSDB_DO_INSTALL` defines whether to install rocksdb.
  The default is `true`.
* `ROCKSDB_SUDO` defines whether to use `sudo` for the installation of the built library.
  The difference between `sudo make rocksdb` and `ROCKSDB_SUDO=true make rocksdb`
  is that the latter will use `sudo` only for the installation (the download and build still use your current user).
  Some systems (e.g. Ubuntu) might require this.
  The default is `true` if the `sudo` command is found, or `false` otherwise.
* `ROCKSDB_DO_CLEANUP` defines whether to delete the downloaded and unpacked repo when done.
  The default is `true`.
</details>

### librdkafka

On M1 Macs (arm64), `librdkafka` and its dependencies are required.

To download, build, and install librdkafka and its dependencies on your system:
```console
$ make librdkafka
```

Openssl's pkg-config files need to be included in the PKG_CONFIG_PATH.

```console
$ export PKG_CONFIG_PATH="$( brew --prefix openssl )"/lib/pkgconfig"${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
```

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

* `WITH_CLEVELDB`: Enables/Disables building with CLevelDB support.
  The default is `true`.
  If this is not `true` the built `provenanced`, executable will not be able to use CLevelDB as a database backend.
* `LEVELDB_PATH`: Defines the location of the leveldb library and includes.
  This is only used if compiling with CLevelDB support on a Mac.
  The default is the result of `brew --prefix leveldb`.
* `WITH_ROCKSDB`: Enables/Disables building with RocksDB support.
  The default is `false`.
  If this is not `true` the built `provenanced`, executable will not be able to use RocksDB as a database backend.
* `WITH_BADGERDB`: Enables/Disables building with BadgerDB support.
  The default is `true`.
  If this is not `true` the built `provenanced`, executable will not be able to use BadgerDB as a database backend.
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
