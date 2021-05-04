#!/usr/bin/make -f
export GO111MODULE=on

PACKAGES               := $(shell go list ./... 2>/dev/null || true)
PACKAGES_NOSIMULATION  := $(filter-out %/simulation%,$(PACKAGES))
PACKAGES_SIMULATION    := $(filter     %/simulation%,$(PACKAGES))

LEVELDB_PATH = $(shell brew --prefix leveldb 2>/dev/null || echo "$(HOME)/Cellar/leveldb/1.22/include")
CGO_CFLAGS   = -I$(LEVELDB_PATH)/include
CGO_LDFLAGS  = "-L$(LEVELDB_PATH)/lib -Wl,-rpath,\$$ORIGIN"

BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build

LEDGER_ENABLED ?= true
WITH_CLEVELDB ?= yes

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
TM_VERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::') # grab everything after the space in "github.com/tendermint/tendermint v0.34.7"
COMMIT := $(shell git log -1 --format='%h')
# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif


GO := go

HTTPS_GIT := https://github.com/provenance-io/provenance.git
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf


# The below include contains the tools target.
include contrib/devtools/Makefile


##############################
# Build Flags/Tags
##############################
build_tags = netgo
ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
  build_tags += cleveldb
endif

ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

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

ifeq ($(WITH_CLEVELDB),yes)
	ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))
BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)' -trimpath

all: build format lint test

.PHONY: all


##############################
# Build
##############################

# Install puts the binaries in the local environment path.
install: go.sum
	CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_CFLAGS=$(CGO_CFLAGS) $(GO) install -mod=readonly $(BUILD_FLAGS) ./cmd/provenanced

build: go.sum
	mkdir -p $(BUILDDIR)
	CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_CFLAGS=$(CGO_CFLAGS) $(GO) build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./cmd/provenanced

build-linux: go.sum
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

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
		$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced collect-gentxs; \
	fi ;

run: check-built run-config;
	$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced start

.PHONY: install build build-linux run

##############################
# Release artifacts and plan #
##############################

RELEASE_BIN=$(BUILDDIR)/bin
RELEASE_PROTO_NAME=protos-$(VERSION).zip
RELEASE_PROTO=$(BUILDDIR)/$(RELEASE_PROTO_NAME)
RELEASE_PLAN=$(BUILDDIR)/plan-$(VERSION).json
RELEASE_CHECKSUM_NAME=sha256sum.txt
RELEASE_CHECKSUM=$(BUILDDIR)/$(RELEASE_CHECKSUM_NAME)
RELEASE_ZIP_NAME=provenance-linux-amd64-$(VERSION).zip
RELEASE_ZIP=$(BUILDDIR)/$(RELEASE_ZIP_NAME)

.PHONY: build-release-clean
build-release-clean:
	rm -rf $(RELEASE_BIN) $(RELEASE_PLAN) $(RELEASE_CHECKSUM) $(RELEASE_ZIP)

.PHONY: build-release-checksum
build-release-checksum: build-release-zip
	cd $(BUILDDIR) && \
	  shasum -a 256 $(RELEASE_ZIP_NAME) > $(RELEASE_CHECKSUM) && \
	cd ..

.PHONY: build-release-plan
build-release-plan: build-release-zip build-release-checksum
	cd $(BUILDDIR) && \
	  sum="$(firstword $(shell cat $(RELEASE_CHECKSUM)))" && \
	  echo "sum=$$sum" && \
	  echo "{\"binaries\":{\"linux/amd64\":\"https://github.com/provenance-io/provenance/releases/download/$(VERSION)/$(RELEASE_ZIP_NAME)?checksum=sha256:$$sum\"}}" > $(RELEASE_PLAN) && \
	cd ..

.PHONY: build-release-bin
build-release-bin: build
	go mod vendor && \
	mkdir -p $(RELEASE_BIN) && \
	cp $(BUILDDIR)/provenanced $(RELEASE_BIN) && \
	cp vendor/github.com/CosmWasm/wasmvm/api/libwasmvm.so $(RELEASE_BIN) && \
	chmod +x $(RELEASE_BIN)/provenanced

.PHONY: build-release-zip
build-release-zip: build-release-bin
	cd $(BUILDDIR) && \
	  zip -r $(RELEASE_ZIP_NAME) bin/ && \
	cd ..

.PHONY: build-release-proto
build-release-proto:
	scripts/protoball.sh $(RELEASE_PROTO)

.PHONY: build-release
build-release: build-release-zip build-release-plan build-release-proto

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
	$(BINDIR)/golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" | xargs gofmt -d -s
	$(GO) mod verify

clean:
	rm -rf $(BUILDDIR)/*

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" | xargs goimports -w -local github.com/provenance-io/provenance

check-built:
	@if [ ! -f "$(BUILDDIR)/provenanced" ]; then \
		echo "\n fatal: Nothing to run.  Try 'make build' first.\n" ; \
		exit 1; \
	fi

statik:
	$(GO) get -u github.com/rakyll/statik
	$(GO) generate ./api/...

linkify:
	python ./scripts/linkify.py CHANGELOG.md

.PHONY: go-mod-cache go.sum lint clean format check-built statik linkify


##############################
### Test
##############################

include sims.mk

test: test-unit
test-all: test-unit test-ledger-mock test-race test-cover

TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-amino test-unit-proto test-ledger-mock test-race test-ledger test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: ARGS=-tags='cgo ledger test_ledger_mock norace'
test-unit-amino: ARGS=-tags='ledger test_ledger_mock test_amino norace'
test-ledger: ARGS=-tags='cgo ledger norace'
test-ledger-mock: ARGS=-tags='ledger test_ledger_mock norace'
test-race: ARGS=-race -tags='cgo ledger test_ledger_mock'
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

# check-* compiles and collects tests without running them
# note: go test -c doesn't support multiple packages yet (https://github.com/golang/go/issues/15513)
CHECK_TEST_TARGETS := check-test-unit check-test-unit-amino
check-test-unit: ARGS=-tags='cgo ledger test_ledger_mock norace'
check-test-unit-amino: ARGS=-tags='ledger test_ledger_mock test_amino norace'
$(CHECK_TEST_TARGETS): EXTRA_ARGS=-run=none
$(CHECK_TEST_TARGETS): run-tests

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	go test -mod=readonly -json $(ARGS) $(EXTRA_ARGS) $(TEST_PACKAGES) | tparse
else
	go test -mod=readonly $(ARGS)  $(EXTRA_ARGS) $(TEST_PACKAGES)
endif

test-cover:
	@export VERSION=$(VERSION); bash -x contrib/test_cover.sh

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)

.PHONY: test test-all test-unit test-race test-cover benchmark run-tests  $(TEST_TARGETS)

##############################
# Test Network Targets
##############################
.PHONY: vendor
vendor:
	go mod vendor -v

# Full build inside a docker container for a clean release build
docker-build: vendor
	docker build -t provenance-io/blockchain . -f docker/blockchain/Dockerfile
	docker build -t provenance-io/blockchain-gateway . -f docker/gateway/Dockerfile

# Quick build using local environment and go platform target options.
docker-build-local: vendor
	docker build --tag provenance-io/blockchain-local -f networks/local/blockchain-local/Dockerfile .

# Run a 4-node testnet locally (replace docker-build with docker-build local for better speed)
localnet-start: localnet-stop docker-build-local
	@if ! [ -f build/node0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-local testnet --v 4 -o . --starting-ip-address 192.168.20.2 --keyring-backend=test --chain-id=chain-local ; fi
	docker-compose -f networks/local/docker-compose.yml --project-directory ./ up -d

# Stop testnet
localnet-stop:
	docker-compose -f networks/local/docker-compose.yml --project-directory ./ down

.PHONY: docker-build-local localnet-start localnet-stop


##############################
# Proto -> golang compilation
##############################
proto-all: proto-tools proto-gen proto-lint proto-check-breaking proto-swagger-gen proto-format protoc-gen-gocosmos protoc-gen-grpc-gateway

proto-gen:
	@./scripts/protocgen.sh

proto-format:
	find ./ -not -path "*/third_party/*" -name *.proto -exec clang-format -i {} \;

# This generates the SDK's custom wrapper for google.protobuf.Any. It should only be run manually when needed
proto-gen-any:
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen-any.sh

proto-swagger-gen:
	@./scripts/protoc-swagger-gen.sh

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)#branch=main

TM_URL           = https://raw.githubusercontent.com/tendermint/tendermint/v0.34.x/proto/tendermint
GOGO_PROTO_URL   = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
COSMOS_PROTO_URL = https://raw.githubusercontent.com/regen-network/cosmos-proto/master
COSMOS_SDK_URL   = https://raw.githubusercontent.com/cosmos/cosmos-sdk/release/v0.40.x/proto/cosmos
CONFIO_URL       = https://raw.githubusercontent.com/confio/ics23/v0.6.3

TM_CRYPTO_TYPES     = third_party/proto/tendermint/crypto
TM_ABCI_TYPES       = third_party/proto/tendermint/abci
TM_TYPES            = third_party/proto/tendermint/types
TM_VERSION          = third_party/proto/tendermint/version
TM_LIBS             = third_party/proto/tendermint/libs/bits

GOGO_PROTO_TYPES     = third_party/proto/gogoproto
COSMOS_PROTO_TYPES   = third_party/proto/cosmos_proto
COSMOS_BASE_TYPES    = third_party/proto/cosmos/base
COSMOS_SIGNING_TYPES = third_party/proto/cosmos/tx/signing
COSMOS_CRYPTO_TYPES  = third_party/proto/cosmos/crypto
COSMOS_AUTH_TYPES    = third_party/proto/cosmos/auth
COSMOS_BANK_TYPES    = third_party/proto/cosmos/bank
CONFIO_TYPES         = third_party/proto/confio

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@curl -sSL $(COSMOS_PROTO_URL)/cosmos.proto > $(COSMOS_PROTO_TYPES)/cosmos.proto

	@mkdir -p $(COSMOS_BASE_TYPES)/v1beta1
	@curl -sSL $(COSMOS_SDK_URL)/base/v1beta1/coin.proto > $(COSMOS_BASE_TYPES)/v1beta1/coin.proto

	@mkdir -p $(COSMOS_BASE_TYPES)/query/v1beta1
	@curl -sSL $(COSMOS_SDK_URL)/base/query/v1beta1/pagination.proto > $(COSMOS_BASE_TYPES)/query/v1beta1/pagination.proto

	@mkdir -p $(COSMOS_SIGNING_TYPES)/v1beta1
	@curl -sSL $(COSMOS_SDK_URL)/tx/signing/v1beta1/signing.proto > $(COSMOS_SIGNING_TYPES)/v1beta1/signing.proto

	@mkdir -p $(COSMOS_CRYPTO_TYPES)/secp256k1
	@curl -sSL $(COSMOS_SDK_URL)/crypto/secp256k1/keys.proto > $(COSMOS_CRYPTO_TYPES)/secp256k1/keys.proto

	@mkdir -p $(COSMOS_CRYPTO_TYPES)/multisig/v1beta1
	@curl -sSL $(COSMOS_SDK_URL)/crypto//multisig/v1beta1/multisig.proto > $(COSMOS_CRYPTO_TYPES)/multisig/v1beta1/multisig.proto

	@mkdir -p $(COSMOS_AUTH_TYPES)/v1beta1
	@curl -sSL $(COSMOS_SDK_URL)/auth/v1beta1/auth.proto > $(COSMOS_AUTH_TYPES)/v1beta1/auth.proto

	@mkdir -p $(COSMOS_BANK_TYPES)/v1beta1
	@curl -sSL $(COSMOS_SDK_URL)/bank/v1beta1/bank.proto > $(COSMOS_BANK_TYPES)/v1beta1/bank.proto

	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types.proto > $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_VERSION)
	@curl -sSL $(TM_URL)/version/types.proto > $(TM_VERSION)/types.proto

	@mkdir -p $(TM_TYPES)
	@curl -sSL $(TM_URL)/types/types.proto > $(TM_TYPES)/types.proto
	@curl -sSL $(TM_URL)/types/evidence.proto > $(TM_TYPES)/evidence.proto
	@curl -sSL $(TM_URL)/types/params.proto > $(TM_TYPES)/params.proto
	@curl -sSL $(TM_URL)/types/validator.proto > $(TM_TYPES)/validator.proto

	@mkdir -p $(TM_CRYPTO_TYPES)
	@curl -sSL $(TM_URL)/crypto/proof.proto > $(TM_CRYPTO_TYPES)/proof.proto
	@curl -sSL $(TM_URL)/crypto/keys.proto > $(TM_CRYPTO_TYPES)/keys.proto

	@mkdir -p $(TM_LIBS)
	@curl -sSL $(TM_URL)/libs/bits/types.proto > $(TM_LIBS)/types.proto

	@mkdir -p $(CONFIO_TYPES)
	@curl -sSL $(CONFIO_URL)/proofs.proto > $(CONFIO_TYPES)/proofs.proto.orig
## insert go, java package option into proofs.proto file
## Issue link: https://github.com/confio/ics23/issues/32 (instead of a simple sed we need 4 lines cause bsd sed -i is incompatible)
	@head -n3 $(CONFIO_TYPES)/proofs.proto.orig > $(CONFIO_TYPES)/proofs.proto
	@echo 'option go_package = "github.com/confio/ics23/go";' >> $(CONFIO_TYPES)/proofs.proto
	@echo 'option java_package = "tech.confio.ics23";' >> $(CONFIO_TYPES)/proofs.proto
	@echo 'option java_multiple_files = true;' >> $(CONFIO_TYPES)/proofs.proto
	@tail -n+4 $(CONFIO_TYPES)/proofs.proto.orig >> $(CONFIO_TYPES)/proofs.proto
	@rm $(CONFIO_TYPES)/proofs.proto.orig

.PHONY: proto-all proto-gen proto-format proto-gen-any proto-swagger-gen proto-lint proto-check-breaking
.PHONY: proto-update-deps


##############################
### Docs
##############################
update-swagger-docs: statik
	$(BINDIR)/statik -src=client/docs/swagger-ui -dest=client/docs -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "Swagger docs are out of sync";\
        exit 1;\
    else \
    	echo "Swagger docs are in sync";\
    fi

.PHONY: update-swagger-docs

.PHONY: update-tocs
update-tocs:
	scripts/update-toc.sh x
