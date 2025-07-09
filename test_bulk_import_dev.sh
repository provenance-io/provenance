#!/bin/bash

# Development test script for bulk import functionality
# This script assumes the chain is already running and just sets up components

set -e  # Exit on any error

# Configuration
PIO_HOME=
CHAIN_ID="testing"
KEY_NAME="validator"
NODE="http://localhost:26657"
GAS="auto"
GAS_ADJUSTMENT="1.2"
GAS_PRICES="2000nhash"
FEE="1000nhash"
CONTRACT_SPEC_ID="contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"
SCOPE_SPEC_ID="scopespec1qj5hx4l3vgryhp5g3ks68wh53jkq3net7n"
SCOPE_ID="scope1qzqqqnucvdf5gu49t7agzh3pw4lsjaju7y"
LEDGER_CLASS_ID="figure_servicing_1.0"
DENOM="nhash"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if provenanced is available
check_provenanced() {
    if ! command -v provenanced &> /dev/null; then
        log_error "provenanced command not found. Please build it first: go build ./cmd/provenanced"
        exit 1
    fi
    log_success "provenanced found"
}

# Check if chain is running
check_chain() {
    log_info "Checking if chain is running..."
    
    if ! curl -s $NODE/status &> /dev/null; then
        log_error "Chain is not running. Please start it first."
        exit 1
    fi
    
    log_success "Chain is running"
}

# Create test key if it doesn't exist
create_key() {
    log_info "Creating test key..."
    
    # Check if key exists
    if ! provenanced keys show $KEY_NAME -ta --keyring-backend test &> /dev/null; then
        echo "y" | provenanced keys add $KEY_NAME --keyring-backend test
        log_success "Test key created"
    else
        log_warning "Test key already exists, skipping..."
    fi
    
    # Get the address
    ADDRESS=$(provenanced keys show $KEY_NAME -ta --keyring-backend test)
    log_info "Using address: $ADDRESS"
}

# Create contract specification
create_contract_spec() {
    log_info "Creating contract specification..."
    
    provenanced tx metadata write-contract-specification \
        $CONTRACT_SPEC_ID \
        $ADDRESS \
        "owner" \
        "FigureTestContract" \
        "Figure Test Asset" \
        "Contract specification for Figure test assets" \
        "https://figure.com" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes | provenanced q wait-tx --chain-id testing --node http://localhost:26657
    
    log_success "Scope specification created: $SCOPE_SPEC_ID"
}

# Create scope specification
create_scope_spec() {
    log_info "Creating scope specification..."    
    
    provenanced tx metadata write-scope-specification \
        $SCOPE_SPEC_ID \
        $ADDRESS \
        "owner" \
        $CONTRACT_SPEC_ID \
        "Figure Servicing Asset" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    log_success "Scope specification created: $SCOPE_SPEC_ID"
}

# Create scope
create_scope() {
    log_info "Creating scope..."
    
    provenanced tx metadata write-scope \
        $SCOPE_ID \
        $SCOPE_SPEC_ID \
        $ADDRESS \
        $ADDRESS \
        $ADDRESS \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    log_success "Scope created: $SCOPE_ID"
}

# Create ledger class
create_ledger_class() {
    log_info "Creating ledger class..."
    
    provenanced tx ledger create-class \
        $LEDGER_CLASS_ID \
        $SCOPE_SPEC_ID \
        $DENOM \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    log_success "Ledger class created: $LEDGER_CLASS_ID"
}

# Create status types
create_status_types() {
    log_info "Creating status types..."
    
    # Status Type 1: IN_REPAYMENT
    provenanced tx ledger add-status-type \
        $LEDGER_CLASS_ID \
        1 \
        IN_REPAYMENT \
        "In Repayment" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    # Status Type 24: DEFAULTED
    provenanced tx ledger add-status-type \
        $LEDGER_CLASS_ID \
        24 \
        DEFAULTED \
        "Defaulted" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    log_success "Status types created"
}

# Create entry types
create_entry_types() {
    log_info "Creating entry types..."
    
    # Entry Type 1: DISBURSEMENT
    provenanced tx ledger add-entry-type \
        $LEDGER_CLASS_ID \
        1 \
        DISBURSEMENT \
        "Disbursement" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    # Entry Type 10: PAYMENT
    provenanced tx ledger add-entry-type \
        $LEDGER_CLASS_ID \
        10 \
        PAYMENT \
        "Payment" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    # Entry Type 12: FEE
    provenanced tx ledger add-entry-type \
        $LEDGER_CLASS_ID \
        12 \
        FEE \
        "Fee" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    # Entry Type 31: INTEREST
    provenanced tx ledger add-entry-type \
        $LEDGER_CLASS_ID \
        31 \
        INTEREST \
        "Interest" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    log_success "Entry types created"
}

# Create bucket types
create_bucket_types() {
    log_info "Creating bucket types..."
    
    # Bucket Type 4: INTEREST
    provenanced tx ledger add-bucket-type \
        $LEDGER_CLASS_ID \
        4 \
        INTEREST \
        "Interest" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    # Bucket Type 7: PRINCIPAL
    provenanced tx ledger add-bucket-type \
        $LEDGER_CLASS_ID \
        7 \
        PRINCIPAL \
        "Principal" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    # Bucket Type 9: FEES
    provenanced tx ledger add-bucket-type \
        $LEDGER_CLASS_ID \
        9 \
        FEES \
        "Fees" \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes  | provenanced q wait-tx --chain-id testing
    
    log_success "Bucket types created"
}

# Perform bulk import
perform_bulk_import() {
    log_info "Performing bulk import..."
    
    provenanced tx ledger chunked-bulk-import \
        ledgers.json \
        --import-id 1 \
        --from $KEY_NAME \
        --keyring-backend test \
        --chain-id $CHAIN_ID \
        --node $NODE \
        --gas $GAS \
        --gas-adjustment $GAS_ADJUSTMENT \
        --gas-prices $GAS_PRICES \
        --testnet \
        --yes
    
    log_success "Bulk import initialized successfully!"
}

# Verify the import
verify_import() {
    log_info "Verifying import..."
    
   
    # Query for entries
    provenanced query ledger entries $SCOPE_SPEC_ID $SCOPE_ID \
        --node $NODE \
        --chain-id $CHAIN_ID \
        --testnet \
        --output json | jq '.entries | length' > /tmp/entry_count
    
    ENTRY_COUNT=$(cat /tmp/entry_count)

    if [ "$ENTRY_COUNT" -eq 0 ]; then
        log_error "No entries found"
        exit 1
    else
        log_success "Found $ENTRY_COUNT entries"
    fi
    
    log_success "Verification complete"
}

# Main execution
main() {
    log_info "Starting bulk import development test..."
    
    # Check prerequisites
    check_provenanced
    check_chain
    create_key
    
    # # Create all necessary components
    # create_contract_spec
    # create_scope_spec
    # create_scope
    # create_ledger_class
    # create_status_types
    # create_entry_types
    # create_bucket_types
    
    # # Wait a bit for all transactions to be processed
    # log_info "Waiting for transactions to be processed..."
    # sleep 5
    
    # Perform the bulk import
    perform_bulk_import
    
    # Wait for import to be processed
    # log_info "Waiting for bulk import to be processed..."
    # sleep 5
    
    # Verify the results
    # verify_import
    
    log_success "Bulk import development test completed successfully!"
}

# Run the main function
main "$@" 