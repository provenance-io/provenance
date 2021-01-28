package types

import (
	"encoding/base64"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*

A Contract is a special structure encapsulating the context of a coordinated execution made by the Contract
Execution Environment.  This context represents a piece of code and a set of methods, data that are replicated
between one or more parties then executed.   The results of this execution as well as all of the input and output
state is captured in a Contract object.


	Contract Type Overview

	 - Recital - An identity performing a role, associated with identification information

	 - RecordReference - A reference to a record on chain.  This reference can be as generic as a scope or
	    as specific as a particular execution context (record group) and hash output on a record.

	 - Condition - Preconditions satisfied before the contract was invoked
	 - Consideration - A set of outputs from the contract.  These will be added as Records in a Scope
	 	- ExecutionResult - Results of an execution, associated with Conditions and Considerations.
		- ProposedRecord - A reference to existing data on chain or a simple hash to off chain data referenced


*/

// ValidateBasic runs stateless validation checks on the message.
func (contract Contract) ValidateBasic() error {
	// Validate contract fields
	scopeID := strings.TrimSpace(contract.Spec.Reference.ScopeId)
	if scopeID == "" {
		return fmt.Errorf("contract spec data location scope ref is nil")
	}
	hash := strings.TrimSpace(contract.Spec.Reference.Hash)
	if hash == "" {
		return fmt.Errorf("contract spec data location ref hash is empty")
	}
	if _, err := base64.StdEncoding.DecodeString(hash); err != nil {
		return fmt.Errorf("ref hash is not base64 encoded: %w", err)
	}

	for _, p := range contract.Recitals.Parties {
		if _, err := sdk.AccAddressFromBech32(p.Address); err != nil {
			return err
		}
	}

	return nil
}

// GetSigners returns the required signers for a given contract
func (contract Contract) GetSigners() (signers []sdk.AccAddress) {
	signers = make([]sdk.AccAddress, 0, len(contract.Recitals.Parties))

	for i, p := range contract.Recitals.Parties {
		addr, err := sdk.AccAddressFromBech32(p.Address)
		if err != nil {
			panic(err)
		}
		signers[i] = addr
	}
	return
}
