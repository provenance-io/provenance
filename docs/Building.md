# Building Provenance Blockchain

## Overview

Provenance uses `make` to define build operations.
Built executables are placed in the `build/` directory.

## Prerequisites

### Go

Building `provenanced` requires [Go 1.17+](https://golang.org/dl/) (or higher).

### CLevelDB

By default, `provenanced` is built with CLevelDB support.
Building without CLevelDB support is also possible. See `WITH_CLEVELDB` in [Options](#options) below.

CLevelDB can usually be installed using your system's software package manager.

On a mac:
```console
$ brew install leveldb
```

With apt-get:
```console
$ apt-get install libleveldb-dev
```

### RocksDB

By default, `provenanced` is built with RocksDB support.
Building without RocksDB support is also possible. See `WITH_ROCKSDB` in [Options](#options) below.

To download, build, and install the RocksDB library on your system:
```console
$ make rocksdb
```

There are a few environment variables that can be used to control some behavior of this command.

* `ROCKSDB_VERSION` will install a version other than the one defined in the `Makefile`.
  Do not include the `v` at the beginning of the version number.
  Example: `ROCKSDB_VERSION=6.17.3 make rocksdb`.
* `ROCKSDB_JOBS` will control the number of parallel jobs used to build the library.
  The default is the result of the `nproc` command.
  More parallel jobs can speed up the build.
  Fewer parallel jobs can alleviate memory problems/crashes that can be encountered during a build.
* `ROCKSDB_SUDO` defines whether to use `sudo` for the installation of the built library.
  The difference between `sudo make rocksdb` and `ROCKSDB_SUDO=yes make rocksdb`
  is that the latter will use `sudo` only for the installation (the download and build still use your current user).
  Some systems (e.g. Ubuntu) might require this.
  The default is `ROCKSDB_SUDO=no`.

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

## Options

A few aspects of `make build` and `make install` can be controlled through environment variables.

* `WITH_CLEVELDB`: Enables/Disables building with CLevelDB support.
  The default is `yes`.
  If this is not `yes` the built `provenanced`, executable will not be able to use CLevelDB as a database backend.
* `LEVELDB_PATH`: Defines the location of the leveldb library and includes.
  This is only used if compiling with CLevelDB support on a Mac.
  The default is the result of `brew --prefix leveldb`.
* `WITH_ROCKSDB`: Enables/Disables building with RocksDB support.
  The default is `yes`.
  If this is not `yes` the built `provenanced`, executable will not be able to use RocksDB as a database backend.
* `WITH_BADGERDB`: Enables/Disables building with BadgerDB support.
  The default is `yes`.
  If this is not `yes` the built `provenanced`, executable will not be able to use BadgerDB as a database backend.
* `BINDIR`: The path to the Go binary directory.
  The default is `${GOPATH}/bin`.
* `BUILDDIR`: The path to the directory where the built executable should be placed.
  The default is `./build`.
* `VERSION`: The string to use as the output of `provenanced version`.
  The default is `{branch name}-{short commit hash}`.
* `BUILD_TAGS`: Any extra `-tags` to supply to the `go build` or `go install` invocations.
  These are added to a list constructed by the Makefile.
* `LDFLAGS`: Any extra `-ldflags` to supply to the `go build` or `go install` invocations.
  These are added to a list constructed by the Makefile.

## Building `dbmigrate`

The `dbmigrate` utility can be used to migrate a node's data directory from one backend database to another.

To build the `dbmigrate` executable and place it in the `build/` directory:
```console
$ make build-dbmigrate
```

Building `dbmigrate` uses the same [Options](#options) as `provenanced`.

It will:
1. Create a new `data/` directory, and copy the contents of the existing `data/` directory into it, converting the database files appropriately.
2. Back up the existing `data/` directory to `${home}/data-dbmigrate-backup-{timestamp}-{dbtypes}/`.
3. Move the newly created `data/` directory into place.
4. Update the config's `db_backend` value to the new db backend type.

The `dbmigrate` utility uses the same configs, environment variables, and flags as `provenanced`.
