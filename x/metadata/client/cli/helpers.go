package cli

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// parseProcess parses a comma delimited string into a process.
// Expected format: "<Name>,<ProcessId>,<Method>"
// If <ProcessId> is not a metadata address, it is a hash.
func parseProcess(commaDelimitedString string) (types.Process, error) {
	var rv types.Process
	parts := strings.Split(commaDelimitedString, ",")
	if len(parts) != 3 {
		return rv, fmt.Errorf("invalid process %q: expected 3 parts, have: %d", commaDelimitedString, len(parts))
	}
	rv.Name = parts[0]
	_, err := types.MetadataAddressFromBech32(parts[1])
	if err != nil {
		rv.ProcessId = &types.Process_Hash{Hash: parts[1]}
	} else {
		rv.ProcessId = &types.Process_Address{Address: parts[1]}
	}
	rv.Method = parts[2]
	return rv, nil
}

// parseParties parses a semicolon delimited string into a list of parties.
// See also: parseParty.
func parseParties(semicolonDelimitedString string) ([]types.Party, error) {
	if len(semicolonDelimitedString) == 0 {
		return nil, nil
	}
	inputs := strings.Split(semicolonDelimitedString, ";")
	rv := make([]types.Party, len(inputs))
	var err error
	for i, input := range inputs {
		rv[i], err = parseParty(input)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

// parseParty parses a comma delimited string into a Party.
// Expected formats: "<address>" or "<address>,<party type>" or "<address>,<party type>,opt"
// If no <party type> is provided, OWNER is used.
// If "opt" is provided, the result has optional=true, otherwise optional=false.
// See also: parsePartyType.
func parseParty(commaDelimitedString string) (rv types.Party, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid party %q: %w", commaDelimitedString, err)
		}
		if rv.Role == types.PartyType_PARTY_TYPE_UNSPECIFIED {
			rv.Role = types.PartyType_PARTY_TYPE_OWNER
		}
	}()

	if len(commaDelimitedString) == 0 {
		return rv, fmt.Errorf("cannot be empty")
	}
	parts := strings.Split(commaDelimitedString, ",")
	if len(parts) > 3 {
		return rv, fmt.Errorf("expected 1, 2, or 3 parts, have: %d", len(parts))
	}
	_, err = sdk.AccAddressFromBech32(parts[0])
	if err != nil {
		return rv, fmt.Errorf("invalid address %q: %w", parts[0], err)
	}
	rv.Address = parts[0]
	if len(parts) >= 2 {
		// Don't change the role from default unless we know it parsed correctly.
		rv.Role, err = parsePartyType(parts[1])
		if err != nil && len(parts) != 2 {
			return rv, err
		}
		// allow for "<address>,opt"
		var optErr error
		rv.Optional, optErr = parseOptional(parts[1])
		if optErr != nil {
			// If the second (and only) thing wasn't an optional flag either, return the role error.
			return rv, err
		}
	}
	if len(parts) >= 3 {
		rv.Optional, err = parseOptional(parts[2])
		if err != nil {
			return rv, err
		}
	}
	return rv, nil
}

// parsePartyType parses a string to a specified PartyType.
func parsePartyType(input string) (types.PartyType, error) {
	rv := types.PartyType_value["PARTY_TYPE_"+strings.ToUpper(input)]
	if rv == 0 {
		rv = types.PartyType_value[strings.ToUpper(input)]
	}
	if rv == 0 {
		return 0, fmt.Errorf("unknown party type: %q", input)
	}
	return types.PartyType(rv), nil
}

// parseOptional parse a string into an optional/required boolean where true == optional.
func parseOptional(input string) (bool, error) {
	switch strings.ToLower(input) {
	case "o", "opt", "optional":
		return true, nil
	case "r", "req", "required":
		return false, nil
	default:
		return false, fmt.Errorf("unknown optional value: %q, expected \"opt\" or \"req\"", input)
	}
}

// parseRecordInputs parses a semicolon delimited string into a list of record inputs.
// See also: parseRecordInput.
func parseRecordInputs(semicolonDelimitedString string) ([]types.RecordInput, error) {
	if len(semicolonDelimitedString) == 0 {
		return nil, nil
	}
	inputs := strings.Split(semicolonDelimitedString, ";")
	rv := make([]types.RecordInput, len(inputs))
	var err error
	for i, input := range inputs {
		rv[i], err = parseRecordInput(input)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

// parseRecordInput parses a comma delimited string into a RecordInput.
// Expected format: "<Name>,<Source>,<TypeName>,<Status>"
// <Source> can be either a metadata address or hash.
// See also: parseRecordInputStatus.
func parseRecordInput(commaDelimitedString string) (rv types.RecordInput, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid record input %q: %w", commaDelimitedString, err)
		}
	}()

	parts := strings.Split(commaDelimitedString, ",")
	if len(parts) != 4 {
		return rv, fmt.Errorf("expected 4 parts, have %d", len(parts))
	}
	rv.Name = parts[0]
	recordID, idErr := types.MetadataAddressFromBech32(parts[1])
	if idErr != nil {
		rv.Source = &types.RecordInput_Hash{Hash: parts[1]}
	} else {
		rv.Source = &types.RecordInput_RecordId{RecordId: recordID}
	}
	rv.TypeName = parts[2]
	rv.Status, err = parseRecordInputStatus(parts[3])
	return rv, err
}

// parseRecordInputStatus parses a string into a specified RecordInputStatus.
func parseRecordInputStatus(input string) (types.RecordInputStatus, error) {
	rv := types.RecordInputStatus_value["RECORD_INPUT_STATUS_"+strings.ToUpper(input)]
	if rv == 0 {
		rv = types.RecordInputStatus_value[strings.ToUpper(input)]
	}
	if rv == 0 {
		return 0, fmt.Errorf("unknown record input status: %q", input)
	}
	return types.RecordInputStatus(rv), nil
}

// parseRecordOutputs parses a semicolon delimited string into a list of record outputs.
// See also: parseRecordOutput.
func parseRecordOutputs(semicolonDelimitedString string) ([]types.RecordOutput, error) {
	if len(semicolonDelimitedString) == 0 {
		return nil, nil
	}
	inputs := strings.Split(semicolonDelimitedString, ";")
	rv := make([]types.RecordOutput, len(inputs))
	var err error
	for i, input := range inputs {
		rv[i], err = parseRecordOutput(input)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

// parseRecordOutput parses a comma delimited string into a RecordOutput.
// Expected format: "<Hash>,<Status>"
// See also: parseResultStatus.
func parseRecordOutput(commaDelimitedString string) (rv types.RecordOutput, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid record output %q: %w", commaDelimitedString, err)
		}
	}()

	parts := strings.Split(commaDelimitedString, ",")
	if len(parts) != 2 {
		return rv, fmt.Errorf("expected 2 parts, have %d", len(parts))
	}
	rv.Hash = parts[0]
	rv.Status, err = parseResultStatus(parts[1])
	return rv, err
}

// parseResultStatus parses a string into a specified ResultStatus.
func parseResultStatus(input string) (types.ResultStatus, error) {
	rv := types.ResultStatus_value["RESULT_STATUS_"+strings.ToUpper(input)]
	if rv == 0 {
		rv = types.ResultStatus_value[strings.ToUpper(input)]
	}
	if rv == 0 {
		return 0, fmt.Errorf("unknown result status: %q", input)
	}
	return types.ResultStatus(rv), nil
}

// parseInputSpecifications parses a semicolon delimited string into a list of input specifications.
// See also: parseInputSpecification.
func parseInputSpecifications(semicolonDelimitedString string) ([]*types.InputSpecification, error) {
	if len(semicolonDelimitedString) == 0 {
		return nil, nil
	}
	inputs := strings.Split(semicolonDelimitedString, ";")
	rv := make([]*types.InputSpecification, len(inputs))
	var err error
	for i, input := range inputs {
		rv[i], err = parseInputSpecification(input)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

// parseInputSpecification parses a comma delimited string into an InputSpecification.
// Expected format: "<Name>,<TypeName>,<Source>"
// <Source> can be either a metadata address or hash.
func parseInputSpecification(commaDelimitedString string) (*types.InputSpecification, error) {
	parts := strings.Split(commaDelimitedString, ",")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid input specification %q: expected 3 parts, have %d", commaDelimitedString, len(parts))
	}
	rv := &types.InputSpecification{
		Name:     parts[0],
		TypeName: parts[1],
	}
	recordID, err := types.MetadataAddressFromBech32(parts[2])
	if err != nil {
		rv.Source = &types.InputSpecification_Hash{Hash: parts[2]}
	} else {
		rv.Source = &types.InputSpecification_RecordId{RecordId: recordID}
	}
	return rv, nil
}

// parsePartyTypes parses a comma delimited string into a list of party types.
// See also: parsePartyType.
func parsePartyTypes(commaDelimitedString string) ([]types.PartyType, error) {
	if len(commaDelimitedString) == 0 {
		return nil, nil
	}
	parties := strings.Split(commaDelimitedString, ",")
	rv := make([]types.PartyType, len(parties))
	var err error
	for i, party := range parties {
		rv[i], err = parsePartyType(party)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

// parseDescription parses a slice of args into a Description.
// Expected args: [<Name>, [<Description>, [<WebsiteUrl>, [<IconUrl>]]]]
// If no args are provided, returns nil.
func parseDescription(args []string) *types.Description {
	if len(args) == 0 {
		return nil
	}
	description := &types.Description{}
	description.Name = args[0]
	if len(args) >= 2 {
		description.Description = args[1]
	}
	if len(args) >= 3 {
		description.WebsiteUrl = args[2]
	}
	if len(args) >= 4 {
		description.IconUrl = args[3]
	}
	return description
}
