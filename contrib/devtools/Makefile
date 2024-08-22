########################################
### Developer Tools
###
### runsim  --------> Used by some of the simulation targets.
### golangci-lint --> Used to lint our go code.


TOOLS_DESTDIR  ?= $(GOPATH)/bin
RUNSIM         = $(TOOLS_DESTDIR)/runsim

tools: tools-stamp
tools-stamp: runsim
	# Create dummy file to satisfy dependency and avoid
	# rebuilding when this Makefile target is hit twice
	# in a row.
	touch $@

proto-tools: proto-tools-stamp
proto-tools-stamp:
	bash contrib/devtools/proto-tools-installer.sh
	# Create dummy file to satisfy dependency and avoid
	# rebuilding when this Makefile target is hit twice
	# in a row.
	touch $@


# Install the runsim binary with a temporary workaround of entering an outside
# directory as the "go get" command ignores the -mod option and will polute the
# go.{mod, sum} files.
#
# ref: https://github.com/golang/go/issues/30515
runsim: $(RUNSIM)
$(RUNSIM):
	@echo "Installing runsim..."
	cd /tmp && go install github.com/cosmos/tools/cmd/runsim@v1.0.0

tools-clean:
	rm -f $(GOLANGCI_LINT) $(RUNSIM)
	rm -f proto-tools-stamp tools-stamp

# If you change this version, be sure to also change it in .github/workflows/lint.yml.
GOLANGCI_LINT_VERSION = v1.60.2
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT):
	@$(MAKE) golangci-lint-update
golangci-lint-update:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) ..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DESTDIR) $(GOLANGCI_LINT_VERSION)

.PHONY: tools-clean runsim golangci-lint golangci-lint-update