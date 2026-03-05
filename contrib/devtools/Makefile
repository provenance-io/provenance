########################################
### Developer Tools
###
### golangci-lint --> Used to lint our go code.


TOOLS_DESTDIR  ?= $(GOPATH)/bin

tools: tools-stamp
tools-stamp: golangci
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

tools-clean:
	rm -f $(GOLANGCI_LINT)
	rm -f proto-tools-stamp tools-stamp

# If you change this version, be sure to also change it in .github/workflows/lint.yml.
GOLANGCI_LINT_VERSION = v2.4.0
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT):
	@$(MAKE) golangci-lint-update
golangci-lint-update:
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION) ..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DESTDIR) $(GOLANGCI_LINT_VERSION)

.PHONY: tools-clean golangci-lint golangci-lint-update