#!/usr/bin/make -f
export GO111MODULE=on

PACKAGES               := $(shell go list ./... 2>/dev/null || true)
PACKAGES_NOSIMULATION  := $(filter-out %/simulation%,$(PACKAGES))
PACKAGES_SIMULATION    := $(filter     %/simulation%,$(PACKAGES))

BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build

WITH_LEDGER ?= true
WITH_CLEVELDB ?= true
WITH_ROCKSDB ?= false
WITH_BADGERDB ?= true

# We used to use 'yes' on these flags, so at least for now, change 'yes' into 'true'
ifeq ($(WITH_LEDGER),yes)
  WITH_LEDGER=true
endif
ifeq ($(WITH_CLEVELDB),yes)
  WITH_CLEVELDB=true
endif
ifeq ($(WITH_ROCKSDB),yes)
  WITH_ROCKSDB=true
endif
ifeq ($(WITH_BADGERDB),yes)
  WITH_BADGERDB=true
endif


BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BRANCH_PRETTY := $(subst /,-,$(BRANCH))
TM_VERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::') # grab everything after the space in "github.com/tendermint/tendermint v0.34.7"
COMMIT := $(shell git log -1 --format='%h')
# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH_PRETTY)-$(COMMIT)
  endif
endif

GOLANGCI_LINT=$(shell which golangci-lint)
ifeq ("$(wildcard $(GOLANGCI_LINT))","")
    GOLANGCI_LINT = $(BINDIR)/golangci-lint
endif

GO ?= go

HTTPS_GIT := https://github.com/provenance-io/provenance.git
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

GO_MAJOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
MINIMUM_SUPPORTED_GO_MAJOR_VERSION = 1
MINIMUM_SUPPORTED_GO_MINOR_VERSION = 17
GO_VERSION_VALIDATION_ERR_MSG = Golang version $(GO_MAJOR_VERSION).$(GO_MINOR_VERSION) is not supported, please update to at least $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION).$(MINIMUM_SUPPORTED_GO_MINOR_VERSION)

# The below include contains the tools target.
include contrib/devtools/Makefile

#Identify the system and if gcc is available.
ifeq ($(OS),Windows_NT)
  UNAME_S = 'windows_nt'
  UNAME_M = 'unknown'
else
  UNAME_S = $(shell uname -s | tr '[A-Z]' '[a-z]')
  UNAME_M = $(shell uname -m | tr '[A-Z]' '[a-z]')
endif

ifeq ($(UNAME_S),windows_nt)
  ifneq ($(shell where gcc.exe 2> NUL),)
    have_gcc = true
  endif
else
  ifneq ($(shell command -v gcc 2> /dev/null),)
    have_gcc = true
  endif
endif

##############################
# Build Flags/Tags
##############################

build_tags = netgo
ifeq ($(UNAME_S),darwin)
	ifeq ($(UNAME_M),arm64)
		# Needed on M1 macs due to kafka issue: https://github.com/confluentinc/confluent-kafka-go/issues/591#issuecomment-811705552
		build_tags += dynamic
	endif
endif

ifeq ($(WITH_CLEVELDB),true)
  ifneq ($(have_gcc),true)
    $(error gcc not installed for cleveldb support, please install or set WITH_CLEVELDB=false)
  else
    build_tags += gcc
    build_tags += cleveldb
  endif
endif
ifeq ($(WITH_ROCKSDB),true)
  build_tags += rocksdb
endif
ifeq ($(WITH_BADGERDB),true)
  build_tags += badgerdb
endif

ifeq ($(WITH_LEDGER),true)
  ifeq ($(UNAME_S),openbsd)
    $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    WITH_LEDGER = false
  else ifneq ($(have_gcc),true)
    $(error gcc not installed for ledger support, please install or set WITH_LEDGER=false)
  else
    build_tags += ledger
  endif
endif

### CGO Settings
ifeq ($(UNAME_S),darwin)
  # osx linker settings
  cgo_ldflags += -Wl,-rpath,@loader_path/.
else ifeq ($(UNAME_S),linux)
  # linux liner settings
  cgo_ldflags += -Wl,-rpath,\$$ORIGIN
endif

# cleveldb linker settings
ifeq ($(WITH_CLEVELDB),true)
  ifeq ($(UNAME_S),darwin)
    LEVELDB_PATH ?= $(shell brew --prefix leveldb 2> /dev/null)
    # Only do stuff if that LEVELDB_PATH exists. Otherwise, leave it up to already installed libraries.
    ifneq ($(wildcard $(LEVELDB_PATH)/.),)
      cgo_cflags  += -I$(LEVELDB_PATH)/include
	  cgo_ldflags += -L$(LEVELDB_PATH)/lib
	endif
  else ifeq ($(UNAME_S),linux)
    # Intentionally left blank to leave it up to already installed libraries.
  endif
endif

cgo_ldflags += $(CGO_LDFLAGS)
cgo_ldflags := $(strip $(cgo_ldflags))
CGO_LDFLAGS := $(cgo_ldflags)

cgo_cflags += $(CGO_CFLAGS)
cgo_cflags := $(strip $(cgo_cflags))
CGO_CFLAGS := $(cgo_cflags)

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))
whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -w -s \
	-X github.com/cosmos/cosmos-sdk/version.Name=Provenance \
	-X github.com/cosmos/cosmos-sdk/version.AppName=provenanced \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
	-X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TM_VERSION)

ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_flags = -mod=readonly -tags "$(build_tags)" -ldflags '$(ldflags)' -trimpath
build_flags += $(BUILD_FLAGS)
build_flags := $(strip $(build_flags))
BUILD_FLAGS := $(build_flags)

all: build format lint test

.PHONY: all


##############################
# Build
##############################

# Install puts the binaries in the local environment path.
install: go.sum
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" $(GO) install $(BUILD_FLAGS) ./cmd/provenanced

build: validate-go-version go.sum
	mkdir -p $(BUILDDIR)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" $(GO) build -o $(BUILDDIR)/ $(BUILD_FLAGS) ./cmd/provenanced

build-linux: go.sum
	WITH_LEDGER=false GOOS=linux GOARCH=amd64 $(MAKE) build

# Run an instance of the daemon against a local config (create the config if it does not exit.)
run-config: check-built
	@if [ ! -d "$(BUILDDIR)/run/provenanced/config" ]; then \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced init --chain-id=testing testing ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced keys add validator --keyring-backend test ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-root-name validator pio --keyring-backend test ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-root-name validator pb --restrict=false --keyring-backend test ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-root-name validator io --restrict --keyring-backend test ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-root-name validator provenance --keyring-backend test ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-account validator 100000000000000000000nhash --keyring-backend test ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced gentx validator 1000000000000000nhash --keyring-backend test --chain-id=testing; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-marker 100000000000000000000nhash --manager validator --access mint,burn,admin,withdraw,deposit --activate --keyring-backend test; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-msg-fee /provenance.name.v1.MsgBindNameRequest 10000000000nhash ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-msg-fee /provenance.marker.v1.MsgAddMarkerRequest 100000000000nhash ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-msg-fee /provenance.attribute.v1.MsgAddAttributeRequest 10000000000nhash ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-msg-fee /provenance.metadata.v1.MsgWriteScopeRequest 10000000000nhash ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced add-genesis-msg-fee /provenance.metadata.v1.MsgP8eMemorializeContractRequest 10000000000nhash ; \
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced collect-gentxs; \
	fi ;

run: check-built run-config;
	$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced start

.PHONY: install build build-linux run

##############################
# Build DB Migration Tools   #
##############################

install-dbmigrate: go.sum
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" $(GO) install $(BUILD_FLAGS) ./cmd/dbmigrate

build-dbmigrate: validate-go-version go.sum
	mkdir -p $(BUILDDIR)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" $(GO) build -o $(BUILDDIR)/ $(BUILD_FLAGS) ./cmd/dbmigrate

##############################
# Release artifacts and plan #
##############################

LIBWASMVM := libwasmvm

RELEASE_BIN=$(BUILDDIR)/bin
RELEASE_PROTO_NAME=protos-$(VERSION).zip
RELEASE_PROTO=$(BUILDDIR)/$(RELEASE_PROTO_NAME)
RELEASE_PLAN=$(BUILDDIR)/plan-$(VERSION).json
RELEASE_CHECKSUM_NAME=sha256sum.txt
RELEASE_CHECKSUM=$(BUILDDIR)/$(RELEASE_CHECKSUM_NAME)

UNAME_M = $(shell uname -m)
ifeq ($(UNAME_S),darwin)
    LIBWASMVM := $(LIBWASMVM).dylib
else ifeq ($(UNAME_S),linux)
	ifeq ($(UNAME_M),x86_64)
		LIBWASMVM := $(LIBWASMVM).$(UNAME_M).so
	else
		LIBWASMVM := $(LIBWASMVM).aarch64.so
	endif
endif

ifeq ($(UNAME_M),x86_64)
	ARCH=amd64
endif

RELEASE_WASM=$(RELEASE_BIN)/$(LIBWASMVM)
RELEASE_PIO=$(RELEASE_BIN)/provenanced
RELEASE_ZIP_BASE=provenance-$(UNAME_S)-$(ARCH)
RELEASE_ZIP_NAME=$(RELEASE_ZIP_BASE)-$(VERSION).zip
RELEASE_ZIP=$(BUILDDIR)/$(RELEASE_ZIP_NAME)

.PHONY: build-release-clean
build-release-clean:
	rm -rf $(RELEASE_BIN) $(RELEASE_PLAN) $(RELEASE_CHECKSUM) $(RELEASE_ZIP)

.PHONY: build-release-checksum
build-release-checksum: $(RELEASE_CHECKSUM)

$(RELEASE_CHECKSUM):
	cd $(BUILDDIR) && \
	  shasum -a 256 *.zip  > $(RELEASE_CHECKSUM) && \
	cd ..

.PHONY: build-release-plan
build-release-plan: $(RELEASE_PLAN)

$(RELEASE_PLAN): $(RELEASE_CHECKSUM)
	scripts/release-plan $(RELEASE_CHECKSUM) $(VERSION) > $(RELEASE_PLAN)

.PHONY: build-release-libwasm
build-release-libwasm: $(RELEASE_WASM)

$(RELEASE_WASM): $(RELEASE_BIN)
	go mod vendor && \
	cp vendor/github.com/CosmWasm/wasmvm/api/$(LIBWASMVM) $(RELEASE_BIN)

.PHONY: build-release-bin
build-release-bin: $(RELEASE_PIO)

$(RELEASE_PIO): $(BUILDDIR)/provenanced $(RELEASE_BIN)
	cp $(BUILDDIR)/provenanced $(RELEASE_BIN) && \
	chmod +x $(RELEASE_PIO)

$(BUILDDIR)/provenanced:
	$(MAKE) build

.PHONY: build-release-zip
build-release-zip: $(RELEASE_ZIP)

$(RELEASE_ZIP): $(RELEASE_PIO) $(RELEASE_WASM)
	cd $(BUILDDIR) && \
	  zip -u $(RELEASE_ZIP_NAME) bin/$(LIBWASMVM) bin/provenanced && \
	cd ..

# gon packages the zip wrong. need bin/provenanced and bin/libwasmvm
.PHONY: build-release-rezip
build-release-rezip: ZIP_FROM = $(BUILDDIR)/$(RELEASE_ZIP_BASE).zip
build-release-rezip: ZIP_TO   = $(BUILDDIR)/$(RELEASE_ZIP_NAME)
build-release-rezip:
	scripts/fix-gon-zip $(ZIP_FROM) && \
		mv -v $(ZIP_FROM) $(ZIP_TO)

.PHONY: build-release-proto
build-release-proto:
	scripts/protoball.sh $(RELEASE_PROTO)

$(RELEASE_BIN):
	mkdir -p $(RELEASE_BIN)

##############################
# Tools / Dependencies
##############################

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download
.PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify
	@go mod tidy

# look into .golangci.yml for enabling / disabling linters
lint:
	$(GOLANGCI_LINT) run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./client/*" -not -path "*.git*" -not -path "*.pb.go" | xargs gofmt -d -s
	$(GO) mod verify

clean:
	rm -rf $(BUILDDIR)/*

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" -not -path "*/statik*" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" -not -path "*/statik*" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" -not -path "*/statik*" | xargs goimports -w -local github.com/provenance-io/provenance

check-built:
	@if [ ! -f "$(BUILDDIR)/provenanced" ]; then \
		echo "\n fatal: Nothing to run.  Try 'make build' first.\n" ; \
		exit 1; \
	fi

linkify:
	python ./scripts/linkify.py CHANGELOG.md

update-tocs:
	scripts/update-toc.sh x docs

# Download, compile, and install rocksdb so that it can be used when doing a build.
rocksdb:
	scripts/rocksdb_build_and_install.sh

# Download, compile, and install cleveldb so that it can be used when doing a build.
cleveldb:
	scripts/cleveldb_build_and_install.sh

# Download and install librdkafka so that it can be used when doing a build.
librdkafka:
	@if [[ $(UNAME_S) == darwin && $(UNAME_M) == arm64 ]]; then \
		scripts/m1_librdkafka_install.sh;\
	fi

.PHONY: go-mod-cache go.sum lint clean format check-built linkify update-tocs rocksdb cleveldb librdkafka


validate-go-version: ## Validates the installed version of go against Provenance's minimum requirement.
	@if [ $(GO_MAJOR_VERSION) -gt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		exit 0 ;\
	elif [ $(GO_MAJOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MAJOR_VERSION) ]; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	elif [ $(GO_MINOR_VERSION) -lt $(MINIMUM_SUPPORTED_GO_MINOR_VERSION) ] ; then \
		echo '$(GO_VERSION_VALIDATION_ERR_MSG)';\
		exit 1; \
	fi

download-smart-contracts:
	./scripts/download_smart_contracts.sh

##############################
### Test
##############################

include sims.mk

test: test-unit
test-all: test-unit test-ledger-mock test-race test-cover

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-amino test-unit-proto test-ledger-mock test-race test-ledger test-race

ifeq ($(WITH_CLEVELDB),true)
	TAGS+= cleveldb
endif
ifeq ($(UNAME_S),darwin)
	ifeq ($(UNAME_M),arm64)
		# Needed on M1 macs due to kafka issue: https://github.com/confluentinc/confluent-kafka-go/issues/591#issuecomment-811705552
		TAGS += dynamic
	endif
endif

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise TAGS, ARGS and/or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: TAGS+=cgo ledger test_ledger_mock norace
test-unit-amino: TAGS+=ledger test_ledger_mock test_amino norace
test-ledger: TAGS+=cgo ledger norace
test-ledger-mock: TAGS+=ledger test_ledger_mock norace
test-race: ARGS+=-race
test-race: TAGS+=cgo ledger test_ledger_mock
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

# check-* compiles and collects tests without running them
# note: go test -c doesn't support multiple packages yet (https://github.com/golang/go/issues/15513)
CHECK_TEST_TARGETS := check-test-unit check-test-unit-amino
check-test-unit: TAGS+=cgo ledger test_ledger_mock norace
check-test-unit-amino: TAGS+=ledger test_ledger_mock test_amino norace
$(CHECK_TEST_TARGETS): ARGS+=-run=none
$(CHECK_TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	go test -mod=readonly -json $(ARGS) -tags='$(TAGS)'$(TEST_PACKAGES) | tparse
else
	go test -mod=readonly $(ARGS) -tags='$(TAGS)' $(TEST_PACKAGES)
endif

test-cover:
	@export VERSION=$(VERSION); bash -x contrib/test_cover.sh

benchmark:
	go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)

.PHONY: test test-all test-unit test-race test-cover benchmark run-tests  $(TEST_TARGETS)

##############################
# Test Network Targets
##############################
.PHONY: vendor
vendor:
	go mod vendor -v

# Full build inside a docker container for a clean release build
docker-build: vendor
	docker build --build-arg VERSION=$(VERSION) -t provenance-io/blockchain . -f docker/blockchain/Dockerfile

# Quick build using local environment and go platform target options.
docker-build-local: vendor
	docker build --target provenance-$(shell uname -m) --tag provenance-io/blockchain-local -f networks/local/blockchain-local/Dockerfile .

# Generate config files for a 4-node localnet
localnet-generate: localnet-stop docker-build-local
	@if ! [ -f build/node0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-local testnet --v 4 -o . --starting-ip-address 192.168.20.2 --keyring-backend=test --chain-id=chain-local ; fi

# Run a 4-node testnet locally
localnet-up:
	docker-compose -f networks/local/docker-compose.yml --project-directory ./ up -d

# Run a 4-node testnet locally (replace docker-build with docker-build local for better speed)
localnet-start: localnet-generate localnet-up

# Stop testnet
localnet-stop:
	docker-compose -f networks/local/docker-compose.yml --project-directory ./ down

# Quick build using devnet environment and go platform target options.
docker-build-dev: vendor
	docker build --tag provenance-io/blockchain-dev -f networks/dev/blockchain-dev/Dockerfile .

# Generate config files for a single node devnet
devnet-generate: devnet-stop docker-build-dev
	docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-dev keys list

# Run a single node devnet locally
devnet-up:
	docker-compose -f networks/dev/docker-compose.yml --project-directory ./ up -d

# Run a single node devnet locally (replace docker-build with docker-build local for better speed)
devnet-start: devnet-generate devnet-up

# Stop devnet
devnet-stop:
	docker-compose -f networks/dev/docker-compose.yml --project-directory ./ down

.PHONY: docker-build-local localnet-start localnet-stop docker-build-dev devnet-start devnet-stop


##############################
# Proto -> golang compilation
##############################
proto-all: proto-update-deps proto-format proto-lint proto-check-breaking proto-gen proto-swagger-gen

containerProtoVer=v0.2
containerProtoImage=tendermintdev/sdk-proto-gen:$(containerProtoVer)
containerProtoGen=cosmos-sdk-proto-gen-$(containerProtoVer)
containerProtoGenSwagger=cosmos-sdk-proto-gen-swagger-$(containerProtoVer)
containerProtoFmt=cosmos-sdk-proto-fmt-$(containerProtoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./scripts/protocgen.sh; fi

# This generates the SDK's custom wrapper for google.protobuf.Any. It should only be run manually when needed
proto-gen-any:
	@echo "Generating Protobuf Any"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) sh ./scripts/protocgen-any.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenSwagger}$$"; then docker start -a $(containerProtoGenSwagger); else docker run --name $(containerProtoGenSwagger) -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./scripts/protoc-swagger-gen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoFmt}$$"; then docker start -a $(containerProtoFmt); else docker run --name $(containerProtoFmt) -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -i {} \; ; fi

proto-lint:
	@echo "Linting Protobuf files"
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@echo "Check breaking Protobuf files"
	@$(DOCKER_BUF) breaking proto --against $(HTTPS_GIT)#branch=main,subdir=proto --error-format=json

proto-update-check:
	@echo "Checking for third_party Protobuf updates"
	sh ./scripts/proto-update-check.sh

proto-update-deps:
	@echo "Updating Protobuf files"
	sh ./scripts/proto-update-deps.sh

.PHONY: proto-all proto-gen proto-format proto-gen-any proto-lint proto-check-breaking proto-update-deps proto-update-check


##############################
### Docs
##############################
update-swagger-docs: statik proto-swagger-gen
	$(BINDIR)/statik -src=client/docs/swagger-ui -dest=client/docs -f -m

.PHONY: update-swagger-docs

test-rosetta:
	docker build -t rosetta-ci:latest -f client/rosetta/rosetta-ci/Dockerfile .
	docker-compose -f client/rosetta/docker-compose.yaml --project-directory ./ up --abort-on-container-exit --exit-code-from test_rosetta --build
.PHONY: test-rosetta
