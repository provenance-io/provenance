#!/usr/bin/make -f
export GO111MODULE=on

# Find the go executable if it wasn't pre-set (e.g. via env var).
ifeq (,$(GO))
  ifeq ($(OS),Windows_NT)
    GO := $(shell where go.exe 2> NUL)
  else
    GO := $(shell command -v go 2> /dev/null)
  endif
endif
# Make sure we have a working go executable since most stuff in here needs it.
ifeq ("$(shell $(GO) version > /dev/null || echo nogo)","nogo")
  $(error Could not find go. Is it in PATH? $(GO))
endif
ifeq (,$(GOPATH))
  GOPATH := $(shell $(GO) env GOPATH)
endif
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build

WITH_LEDGER ?= true
# We used to use 'yes' on these flags, so at least for now, change 'yes' into 'true'
ifeq ($(WITH_LEDGER),yes)
  WITH_LEDGER=true
endif

COMMIT := $(shell git log -1 --format='%h' 2> /dev/null)
ifeq (,$(COMMIT))
  COMMIT := unknown
endif
ifeq (,$(VERSION))
  # If VERSION wasn't provided (e.g. via env var), look for a tag on HEAD.
  VERSION := $(shell git describe --exact-match 2> /dev/null)
  # If there isn't a tag, use the branch name and commit hash.
  ifeq (,$(VERSION))
    VERSION := $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD 2> /dev/null || echo nobranch))-$(COMMIT)
  endif
endif

ifeq (,$(GOLANGCI_LINT))
  GOLANGCI_LINT := $(shell which golangci-lint 2> /dev/null)
  # If golangci-lint isn't found, use a common location and let whatever needs it fail on its own.
  ifeq (,$(GOLANGCI_LINT))
    GOLANGCI_LINT := $(BINDIR)/golangci-lint
  endif
endif

HTTPS_GIT := https://github.com/provenance-io/provenance.git
DOCKER := $(shell which docker)

# The below include contains the tools target.
include contrib/devtools/Makefile

#Identify the system and if gcc is available.
ifeq ($(OS),Windows_NT)
  UNAME_S = windows_nt
  UNAME_M = unknown
else
  UNAME_S := $(shell uname -s | tr '[A-Z]' '[a-z]')
  UNAME_M := $(shell uname -m | tr '[A-Z]' '[a-z]')
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

export CMTVERSION := $(shell $(GO) list -m github.com/cometbft/cometbft 2> /dev/null | sed 's:.* ::')

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

base_ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=Provenance \
	-X github.com/cosmos/cosmos-sdk/version.AppName=provenanced \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
	-X github.com/cometbft/cometbft/version.TMCoreSemVer=$(CMTVERSION)
base_ldflags += $(LDFLAGS)
ver_var = github.com/cosmos/cosmos-sdk/version.Version

ldflags = -w -s $(base_ldflags) -X $(ver_var)=$(VERSION)
ldflags := $(strip $(ldflags))

ldflags_debug = $(base_ldflags) -X $(ver_var)=$(VERSION)-debug
ldflags_debug := $(strip $(ldflags_debug))

base_build_flags := -mod=readonly -tags "$(build_tags)" $(BUILD_FLAGS)

build_flags = $(base_build_flags) -trimpath -ldflags '$(ldflags)'
build_flags := $(strip $(build_flags))
BUILD_FLAGS := $(build_flags)

GCFLAGS ?= all=-N -l
build_debug_flags = $(base_build_flags) -gcflags '$(GCFLAGS)' -ldflags '$(ldflags_debug)'
build_debug_flags := $(strip $(build_debug_flags))
BUILD_DEBUG_FLAGS := $(build_debug_flags)

all: build format lint test

.PHONY: all


##############################
# Build
##############################

# Install puts the binaries in the local environment path.
install: validate-go-version go.sum
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" $(GO) install $(BUILD_FLAGS) ./cmd/provenanced

build: validate-go-version go.sum
	mkdir -p $(BUILDDIR)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" $(GO) build -o $(BUILDDIR)/ $(BUILD_FLAGS) ./cmd/provenanced

build-debug: validate-go-version go.sum
	mkdir -p $(BUILDDIR)
	CGO_LDFLAGS="$(CGO_LDFLAGS)" CGO_CFLAGS="$(CGO_CFLAGS)" $(GO) build -o $(BUILDDIR)/ $(BUILD_DEBUG_FLAGS) ./cmd/provenanced

build-linux: go.sum
	WITH_LEDGER=false GOOS=linux GOARCH=amd64 $(MAKE) build

DENOM ?= nhash
CHAIN_ID ?= testing
CHAIN_ID_DOCKER ?= chain-local

# Run an instance of the daemon against a local config (create the config if it does not exit.)
# if required to use something other than nhash, use: make run DENOM=vspn
run-config: check-built
	@if [ ! -d "$(BUILDDIR)/run/provenanced/config" ]; then \
		PROV_CMD='$(BUILDDIR)/provenanced' \
			PIO_HOME='$(BUILDDIR)/run/provenanced' \
			PIO_CHAIN_ID='$(CHAIN_ID)' \
			DENOM='$(DENOM)' \
			SHOW_START=false \
			./scripts/initialize.sh; \
	fi

run: check-built run-config ;
ifeq ($(DENOM),nhash)
	$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced start
else
	$(BUILDDIR)/provenanced -t --home $(BUILDDIR)/run/provenanced start --custom-denom $(DENOM)
endif

.PHONY: install build build-debug build-linux run

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

ifeq ($(UNAME_S),darwin)
    LIBWASMVM := $(LIBWASMVM).dylib
else ifeq ($(UNAME_S),linux)
	ifeq ($(UNAME_M),x86_64)
		LIBWASMVM := $(LIBWASMVM).$(UNAME_M).so
	else
		LIBWASMVM := $(LIBWASMVM).aarch64.so
	endif
endif

ARCH=$(UNAME_M)
ifeq ($(ARCH),x86_64)
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
	$(GO) mod vendor && \
	cp vendor/github.com/CosmWasm/wasmvm/v2/internal/api/$(LIBWASMVM) $(RELEASE_BIN)

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
	$(GO) mod download
.PHONY: go-mod-cache

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	$(GO) mod verify
	$(GO) mod tidy

# look into .golangci.yml for enabling / disabling linters
lint:
	$(GOLANGCI_LINT) run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./client/*" -not -path "*.git*" -not -path "*.pb.go" | xargs gofmt -d -s
	scripts/no-now-lint.sh
	$(GO) mod verify
	.changelog/lint-unreleased.sh

lint-fix:
	$(GOLANGCI_LINT) run --fix
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "./client/*" -not -path "*.git*" -not -path "*.pb.go" | xargs gofmt -w -s
	$(GO) mod tidy

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

linkify:
	python ./scripts/linkify.py CHANGELOG.md

update-tocs:
	scripts/update-toc.sh x docs CONTRIBUTING.md .changelog/README.md

get-valid-sections:
	.changelog/get-valid-sections.sh

.PHONY: go-mod-cache go.sum lint clean format check-built linkify update-tocs get-valid-sections

# Only support go version 1.23
SUPPORTED_GO_MAJOR_VERSION = 1
SUPPORTED_GO_MINOR_VERSION = 23
GO_MAJOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f1)
GO_MINOR_VERSION = $(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)

validate-go-version: ## Validates the installed version of go against Provenance's minimum requirement.
	@if [ "$(GO_MAJOR_VERSION)" -ne $(SUPPORTED_GO_MAJOR_VERSION) ] || [ "$(GO_MINOR_VERSION)" -ne $(SUPPORTED_GO_MINOR_VERSION) ]; then \
		echo 'Golang version $(GO_MAJOR_VERSION).$(GO_MINOR_VERSION) is not supported, you must use $(SUPPORTED_GO_MAJOR_VERSION).$(SUPPORTED_GO_MINOR_VERSION).'; \
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

PACKAGES               := $(shell $(GO) list ./... 2>/dev/null || true)
PACKAGES_NOSIMULATION  := $(filter-out %/simulation%,$(PACKAGES))
PACKAGES_SIMULATION    := $(filter     %/simulation%,$(PACKAGES))

TEST_PACKAGES ?= ./...
TEST_TARGETS := test-unit test-unit-proto test-ledger-mock test-race test-ledger build-tests

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise TAGS, ARGS and/or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: TAGS+=cgo ledger test_ledger_mock norace
build-tests: TAGS+=cgo ledger test_ledger_mock norace
build-tests: ARGS+=-run='ZYX_NOPE_NOPE_XYZ'
test-ledger: TAGS+=cgo ledger norace
test-ledger-mock: TAGS+=ledger test_ledger_mock norace
test-race: ARGS+=-race
test-race: TAGS+=cgo ledger test_ledger_mock
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

run-tests: go.sum
ifneq (,$(shell which tparse 2>/dev/null))
	$(GO) test -mod=readonly -json $(ARGS) -tags='$(TAGS)' $(TEST_PACKAGES) | tparse
else
	$(GO) test -mod=readonly $(ARGS) -tags='$(TAGS)' $(TEST_PACKAGES)
endif

test-cover:
	export VERSION=$(VERSION); bash -x contrib/test_cover.sh

benchmark:
	$(GO) test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)

.PHONY: test test-all test-cover benchmark run-tests build-tests $(TEST_TARGETS)

##############################
# Test Network Targets
##############################
.PHONY: vendor
vendor:
	$(GO) mod vendor -v

# Full build inside a docker container for a clean release build
docker-build: vendor
	docker build --build-arg VERSION=$(VERSION) -t provenance-io/blockchain . -f docker/blockchain/Dockerfile


# Quick build using local environment and go platform target options.
docker-build-local: vendor
	docker build --target provenance-$(shell uname -m) --tag provenance-io/blockchain-local -f networks/local/blockchain-local/Dockerfile .

# Generate config files for a 4-node localnet
localnet-generate: localnet-stop docker-build-local
ifeq ($(DENOM),nhash)
	@if ! [ -f build/node0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-local testnet --v 4 -o . --starting-ip-address 192.168.20.2 --keyring-backend=test --chain-id=$(CHAIN_ID_DOCKER) ; fi
else
	@if ! [ -f build/node0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-local testnet --v 4 -o . --starting-ip-address 192.168.20.2 --keyring-backend=test --chain-id=$(CHAIN_ID_DOCKER) --custom-denom=$(DENOM) --minimum-gas-prices=0$(DENOM) ; fi
endif

# Run a 4-node testnet locally
localnet-up:
	docker-compose -f networks/local/docker-compose.yml --project-directory ./ up -d

# Run a 4-node testnet locally (replace docker-build with docker-build local for better speed)
# to run custom denom network, `make clean localnet-start DENOM=vspn
localnet-start: localnet-generate localnet-up

# Stop testnet
localnet-stop:
	docker-compose -f networks/local/docker-compose.yml --project-directory ./ down

# Quick build using ibc environment and go platform target options.
RELAYER_MNEMONIC ?= "list judge day spike walk easily outer state fashion library review include leisure casino wagon zoo chuckle alien stock bargain stone wait squeeze fade"
RELAYER_KEY ?= tp18uev5722xrwpfd2hnqducmt3qdjsyktmtw558y
RELAYER_VERSION ?= v2.5.2
docker-build-ibc: vendor
	docker build --target provenance-$(shell uname -m) --tag provenance-io/blockchain-ibc -f networks/ibc/blockchain-ibc/Dockerfile .
	docker build --target relayer --tag provenance-io/blockchain-relayer -f networks/ibc/blockchain-relayer/Dockerfile --build-arg MNEMONIC=$(RELAYER_MNEMONIC) --build-arg VERSION=$(RELAYER_VERSION) .

# Generate config files for a 2-node ibcnet with relayer
ibcnet-generate: ibcnet-stop docker-build-ibc
	@if ! [ -f build/ibc0-0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-ibc testnet --v 1 -o . --starting-ip-address 192.168.20.2 --node-dir-prefix=ibc0- --keyring-backend=test --chain-id=testing ; fi
	mv build/gentxs/ibc0-0.json build/gentxs/tmp
	@if ! [ -f build/ibc1-0/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-ibc testnet --v 1 -o . --starting-ip-address 192.168.20.3 --node-dir-prefix=ibc1- --keyring-backend=test --chain-id=testing2 ; fi
	mv build/gentxs/tmp build/gentxs/ibc0-0.json
	scripts/ibcnet-add-relayer-key.sh $(RELAYER_KEY)

# Run a 2-node testnet locally with a relayer
ibcnet-up:
	docker-compose -f networks/ibc/docker-compose.yml --project-directory ./ up -d

# Run a 2-node testnet locally with a relayer
ibcnet-start: ibcnet-generate ibcnet-up

# Stop ibcnet
ibcnet-stop:
	docker-compose -f networks/ibc/docker-compose.yml --project-directory ./ down

# Quick build using devnet environment and go platform target options.
docker-build-dev: vendor
	docker build --target provenance-$(shell uname -m) --tag provenance-io/blockchain-dev -f networks/dev/blockchain-dev/Dockerfile .

# Generate config files for a single node devnet
devnet-generate: devnet-stop docker-build-dev
	@if ! [ -f build/nodedev/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/provenance:Z provenance-io/blockchain-dev keys list ; fi

# Run a single node devnet locally
devnet-up:
	docker-compose -f networks/dev/docker-compose.yml --project-directory ./ up -d

# Run a single node devnet locally (replace docker-build with docker-build local for better speed)
devnet-start: devnet-generate devnet-up

# Stop devnet
devnet-stop:
	docker-compose -f networks/dev/docker-compose.yml --project-directory ./ down

# Start postgres indexer instance
indexer-db-up:
	docker compose -f docker/postgres-indexer/docker-compose.yaml --project-directory ./ up -d

# Stop postgres indexer instance
indexer-db-down:
	docker compose -f docker/postgres-indexer/docker-compose.yaml --project-directory ./ down

.PHONY: docker-build-local localnet-start localnet-stop docker-build-dev devnet-start devnet-stop


##############################
# Proto -> golang compilation
##############################
proto-all: proto-update-deps proto-format proto-lint proto-check-breaking proto-check-breaking-third-party proto-gen proto-swagger-gen proto-doc-gen
proto-checks: proto-update-deps proto-lint proto-check-breaking proto-check-breaking-third-party
proto-regen: proto-format proto-gen proto-swagger-gen proto-doc-gen

protoVer=0.15.1
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

# The proto gen stuff will update go.mod and go.sum in ways we don't want (due to docker stuff).
# So we need to go mod tidy afterward, but it can't go in the scripts for the same reason that we need it.

proto-gen:
	@echo "Generating Protobuf files"
	cp go.mod .go.mod.bak
	$(protoImage) sh ./scripts/protocgen.sh
	mv .go.mod.bak go.mod
	$(GO) mod tidy

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	$(protoImage) sh ./scripts/protoc-swagger-gen.sh

proto-doc-gen:
	@echo "Generating Protobuf Markdown"
	./scripts/proto-doc-gen.sh

proto-format:
	@echo "Formatting Protobuf files"
	$(protoImage) find . \
			-not -path './third_party/*' \
			-not -path './vendor/*' \
			-not -path './protoBindings/*' \
			-name '*.proto' \
			-exec clang-format -i {} \; ;

proto-lint:
	@echo "Linting Protobuf files"
	$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@echo "Check breaking Protobuf files"
	$(protoImage) buf breaking proto --against '$(HTTPS_GIT)#branch=main,subdir=proto' --error-format=json

proto-check-breaking-third-party:
	@echo "Check breaking 3rd party Protobuf files"
	$(protoImage) buf breaking third_party/proto --against '$(HTTPS_GIT)#branch=main,subdir=third_party/proto' --error-format=json

proto-update-check:
	@echo "Checking for third_party Protobuf updates"
	sh ./scripts/proto-update-check.sh

proto-update-deps:
	@echo "Updating Protobuf files"
	sh ./scripts/proto-update-deps.sh

proto-legacy:
	docker run --rm \
		-v $(PWD):/workspace \
		--workdir /workspace \
		ghcr.io/cosmos/proto-builder:0.15.1 \
		buf generate --template legacy_protos/buf.gen.legacy.yaml legacy_protos

.PHONY: proto-all proto-checks proto-regen proto-gen proto-format proto-lint proto-check-breaking proto-check-breaking-third-party proto-update-deps proto-update-check proto-doc-gen proto-legacy


##############################
### Relayer
##############################
relayer-install:
	scripts/install-relayer.sh

relayer-start: relayer-install
	scripts/start-relayer.sh

.PHONY: relayer-install relayer-start
