package types

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"fmt"

	"github.com/provenance-io/provenance/x/metadata/types/p8e"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	tmcrypt "github.com/tendermint/tendermint/crypto"
	tmcurve "github.com/tendermint/tendermint/crypto/secp256k1"
)

// ConvertP8eContractSpec converts a v39 ContractSpec to a v40 ContractSpecification
func ConvertP8eContractSpec(old *p8e.ContractSpec, owners []string) (
	newSpec ContractSpecification,
	newRecords []RecordSpecification,
	err error,
) {
	rawProtoOld, err := proto.Marshal(old)
	if err != nil {
		return newSpec, nil, err
	}
	sha512Old := sha512.Sum512(rawProtoOld)
	specUUID, err := uuid.FromBytes(sha512Old[0:16])
	if err != nil {
		return newSpec, nil, err
	}
	id := ContractSpecMetadataAddress(specUUID)

	parties := make([]PartyType, len(old.PartiesInvolved))
	for i := range old.PartiesInvolved {
		parties[i] = PartyType(int32(old.PartiesInvolved[i]))
	}

	newSpec = ContractSpecification{
		SpecificationId: id,
		Description: &Description{
			Name:        old.Definition.Name,
			Description: old.Definition.ResourceLocation.Classname,
		},
		PartiesInvolved: parties,
		OwnerAddresses:  owners,
		Source: &ContractSpecification_Hash{
			Hash: base64.StdEncoding.EncodeToString(sha512Old[:]),
		},
		ClassName: old.Definition.ResourceLocation.Classname,
	}
	err = newSpec.ValidateBasic()
	if err != nil {
		return ContractSpecification{}, nil, err
	}

	newRecords = make([]RecordSpecification, len(old.ConsiderationSpecs))
	for i := range old.ConsiderationSpecs {
		recordInputs, err := convertP8eInputSpecs(old.ConsiderationSpecs[i].InputSpecs)
		if err != nil {
			return newSpec, nil, err
		}
		specUUID, err := newSpec.SpecificationId.ContractSpecUUID()
		if err != nil {
			return newSpec, nil, err
		}
		recordSpecID := RecordSpecMetadataAddress(specUUID, old.ConsiderationSpecs[i].OutputSpec.Spec.Name)
		newRecords[i] = RecordSpecification{
			SpecificationId:    recordSpecID,
			Name:               old.ConsiderationSpecs[i].OutputSpec.Spec.Name,
			TypeName:           old.ConsiderationSpecs[i].OutputSpec.Spec.ResourceLocation.Classname,
			ResultType:         DefinitionType(old.ConsiderationSpecs[i].OutputSpec.Spec.Type),
			Inputs:             recordInputs,
			ResponsibleParties: []PartyType{PartyType(old.ConsiderationSpecs[i].ResponsibleParty)},
		}
		err = newRecords[i].ValidateBasic()
		if err != nil {
			return newSpec, nil, err
		}
	}

	return newSpec, newRecords, nil
}

// convertP8eInputSpecs a v39 DefinitionSpec used for inputs into a v40 InputSpecification
func convertP8eInputSpecs(old []*p8e.DefinitionSpec) (inputs []*InputSpecification, err error) {
	inputs = make([]*InputSpecification, len(old))
	for i, oldInput := range old {
		if oldInput.ResourceLocation.Ref.ScopeUuid != nil &&
			len(oldInput.ResourceLocation.Ref.ScopeUuid.Value) > 0 {
			scopeUUID, err := uuid.Parse(oldInput.ResourceLocation.Ref.ScopeUuid.Value)
			if err != nil {
				return nil, err
			}
			if len(oldInput.ResourceLocation.Ref.Name) < 1 {
				return nil, fmt.Errorf("must have a value for record name")
			}
			inputs[i] = &InputSpecification{
				Name:     oldInput.Name,
				TypeName: oldInput.ResourceLocation.Classname,
				Source: &InputSpecification_RecordId{
					RecordId: RecordMetadataAddress(scopeUUID, oldInput.ResourceLocation.Ref.Name),
				},
			}
		} else {
			inputs[i] = &InputSpecification{
				Name:     oldInput.Name,
				TypeName: oldInput.ResourceLocation.Classname,
				Source: &InputSpecification_Hash{
					Hash: oldInput.ResourceLocation.Ref.Hash,
				},
			}
		}
		err := inputs[i].ValidateBasic()
		if err != nil {
			return nil, err
		}
	}
	return inputs, nil
}

// P8EData contains entries converted from a MsgP8EMemorializeContractRequest.
type P8EData struct {
	Scope      *Scope
	Session    *Session
	RecordReqs []*RecordReq
	Signers    []string
}

type RecordReq struct {
	Record               *Record
	OriginalOutputHashes []string
}

// Migrate Converts a MsgP8EMemorializeContractRequest object into the new objects.
// The []string return parameter is a list of signer address strings.
func ConvertP8eMemorializeContractRequest(msg *MsgP8EMemorializeContractRequest) (P8EData, error) {
	p8EData := P8EData{
		Scope:      emptyScope(),
		Session:    emptySession(),
		RecordReqs: []*RecordReq{},
		Signers:    []string{},
	}
	var err error

	contractRecitalParties, err := convertParties(msg.Contract.Recitals)
	if err != nil {
		return p8EData, err
	}

	// Set the scope pieces.
	p8EData.Scope.ScopeId, err = parseScopeID(msg.ScopeId)
	if err != nil {
		return p8EData, err
	}
	p8EData.Scope.SpecificationId, err = parseScopeSpecificationID(msg.ScopeSpecificationId)
	if err != nil {
		return p8EData, err
	}
	p8EData.Scope.Owners = contractRecitalParties
	p8EData.Scope.DataAccess = partyAddresses(contractRecitalParties)
	p8EData.Scope.ValueOwnerAddress, err = getValueOwner(msg.Contract.Invoker, msg.Contract.Recitals)
	if err != nil {
		return p8EData, err
	}

	// Set the session pieces.
	p8EData.Session.SessionId, err = parseSessionID(p8EData.Scope.ScopeId, msg.GroupId)
	if err != nil {
		return p8EData, err
	}
	p8EData.Session.SpecificationId, err = getContractSpecID(msg.Contract)
	if err != nil {
		return p8EData, err
	}
	p8EData.Session.Parties = contractRecitalParties
	p8EData.Session.Name = msg.Contract.Definition.Name
	p8EData.Session.Context = msg.Contract.Context

	processID, pidErr := getProcessID(msg.Contract)
	if pidErr != nil {
		return p8EData, pidErr
	}

	facts := map[string]*p8e.Fact{}
	for i, f := range msg.Contract.Inputs {
		if len(f.Name) == 0 {
			return p8EData, fmt.Errorf("missing value in contract.input[%d].name", i)
		}
		facts[f.Name] = f
	}

	// Create the records.
	p8EData.RecordReqs = make([]*RecordReq, 0, len(msg.Contract.Considerations))
	for _, c := range msg.Contract.Considerations {
		if c != nil && c.Result != nil && c.Result.Output != nil && c.Result.Result != p8e.ExecutionResultType_RESULT_TYPE_SKIP {
			record := emptyRecord()
			record.Name = c.ConsiderationName
			record.SessionId = p8EData.Session.SessionId
			record.Process.ProcessId = processID
			record.Process.Name = c.Result.Output.Classname
			record.Process.Method = record.Name
			record.Inputs = make([]RecordInput, len(c.Inputs))
			for i, f := range c.Inputs {
				if len(f.Hash) == 0 && len(f.Name) == 0 {
					return p8EData, fmt.Errorf("consideration %s inputs[%d] name and hash cannot both be empty",
						c.ConsiderationName, i)
				}

				ri := RecordInput{
					Name:     f.Name,
					TypeName: f.Classname,
					Status:   RecordInputStatus_Unknown,
				}

				if len(f.Hash) > 0 {
					ri.Status = RecordInputStatus_Proposed
					ri.Source = &RecordInput_Hash{Hash: f.Hash}
				} else {
					ri.Status = RecordInputStatus_Record
					if facts[f.Name] == nil {
						return p8EData, fmt.Errorf("consideration %s inputs[%d] %s not found as contract input",
							c.ConsiderationName, i, f.Name)
					}
					if facts[f.Name].DataLocation == nil || facts[f.Name].DataLocation.Ref == nil {
						return p8EData, fmt.Errorf("contract input %s missing datalocation ref", f.Name)
					}
					if len(facts[f.Name].DataLocation.Ref.Hash) == 0 &&
						(facts[f.Name].DataLocation.Ref.ScopeUuid == nil ||
							len(facts[f.Name].DataLocation.Ref.ScopeUuid.Value) == 0) {
						return p8EData, fmt.Errorf("contract input %s datalocation ref must have either a hash or scope uuid",
							f.Name)
					}

					if len(facts[f.Name].DataLocation.Ref.Hash) > 0 {
						ri.Source = &RecordInput_Hash{Hash: facts[f.Name].DataLocation.Ref.Hash}
					} else {
						scopeUUID, scopeUUIDErr := uuid.Parse(facts[f.Name].DataLocation.Ref.ScopeUuid.Value)
						if scopeUUIDErr != nil {
							return p8EData, fmt.Errorf("invalid UUID in contract input %s: %w", f.Name, scopeUUIDErr)
						}
						ri.Source = &RecordInput_RecordId{RecordId: RecordMetadataAddress(scopeUUID, f.Name)}
					}
				}

				record.Inputs[i] = ri
			}
			record.Outputs = []RecordOutput{
				{
					Hash:   c.Result.Output.Hash,
					Status: ResultStatus(c.Result.Result),
				},
			}
			recReq := RecordReq{
				Record:               record,
				OriginalOutputHashes: []string{},
			}
			if c.Result.Output.Ancestor != nil && len(c.Result.Output.Ancestor.Hash) > 0 {
				recReq.OriginalOutputHashes = append(recReq.OriginalOutputHashes, c.Result.Output.Ancestor.Hash)
			}

			p8EData.RecordReqs = append(p8EData.RecordReqs, &recReq)
		}
	}

	// Get the signers.
	if msg.Signatures != nil {
		newSigners, e := convertSigners(msg.Signatures)
		if e != nil {
			return p8EData, e
		}
		p8EData.Signers = append(p8EData.Signers, newSigners...)
	}

	return p8EData, err
}

// emptyScope creates a new empty Scope.
func emptyScope() *Scope {
	return &Scope{
		ScopeId:           MetadataAddress{},
		SpecificationId:   MetadataAddress{},
		Owners:            []Party{},
		DataAccess:        []string{},
		ValueOwnerAddress: "",
	}
}

// emptySession creates a new empty Session.
func emptySession() *Session {
	return &Session{
		SessionId:       MetadataAddress{},
		SpecificationId: MetadataAddress{},
		Parties:         []Party{},
		Name:            "",
		Audit:           nil,
		Context:         nil,
	}
}

// emptyRecord creates a new empty Record.
func emptyRecord() *Record {
	return &Record{
		Name:            "",
		SessionId:       MetadataAddress{},
		Process:         *emptyProcess(),
		Inputs:          []RecordInput{},
		Outputs:         []RecordOutput{},
		SpecificationId: MetadataAddress{},
	}
}

// emptyProcess creates a new empty Process.
func emptyProcess() *Process {
	return &Process{
		ProcessId: nil,
		Name:      "",
		Method:    "",
	}
}

func parseScopeID(input string) (MetadataAddress, error) {
	scopeID, maErr := MetadataAddressFromBech32(input)
	if maErr == nil {
		if !scopeID.IsScopeAddress() {
			return scopeID, fmt.Errorf("metadata address %s is not for a scope", scopeID)
		}
		return scopeID, nil
	}
	scopeUUID, uuidErr := uuid.Parse(input)
	if uuidErr == nil {
		return ScopeMetadataAddress(scopeUUID), nil
	}
	return MetadataAddress{}, fmt.Errorf("could not convert %s into either a scope metadata address (%s) or uuid (%s)",
		input, maErr.Error(), uuidErr.Error())
}

func parseScopeSpecificationID(input string) (MetadataAddress, error) {
	scopeSpecID, maErr := MetadataAddressFromBech32(input)
	if maErr == nil {
		if !scopeSpecID.IsScopeSpecificationAddress() {
			return scopeSpecID, fmt.Errorf("metadata address %s is not for a scope specification", scopeSpecID)
		}
		return scopeSpecID, nil
	}
	scopeSpecUUID, uuidErr := uuid.Parse(input)
	if uuidErr == nil {
		return ScopeSpecMetadataAddress(scopeSpecUUID), nil
	}
	return MetadataAddress{}, fmt.Errorf("could not convert %s into either a scope specification metadata address (%s) or uuid (%s)",
		input, maErr.Error(), uuidErr.Error())
}

func parseSessionID(scopeID MetadataAddress, input string) (MetadataAddress, error) {
	sessionID, maErr := MetadataAddressFromBech32(input)
	if maErr == nil {
		if !sessionID.IsScopeAddress() {
			return sessionID, fmt.Errorf("metadata address %s is not for a session", sessionID)
		}
		return sessionID, nil
	}
	sessionUUID, uuidErr := uuid.Parse(input)
	if uuidErr == nil {
		return scopeID.AsSessionAddress(sessionUUID)
	}
	return MetadataAddress{}, fmt.Errorf("could not convert %s into either session metadata address (%s) or uuid (%s)",
		input, maErr.Error(), uuidErr.Error())
}

func convertParties(old []*p8e.Recital) (parties []Party, err error) {
	parties = make([]Party, len(old))
	err = nil
	for i, r := range old {
		p, e := convertParty(*r)
		if e != nil {
			err = appendError(err, e)
		} else {
			parties[i] = p
		}
	}
	return
}

func convertParty(old p8e.Recital) (Party, error) {
	party := Party{}
	if len(old.Address) > 0 {
		party.Address = sdk.AccAddress(old.Address).String()
	} else {
		addr, err := getAddressFromSigner(old.Signer)
		if err != nil {
			return party, err
		}
		party.Address = addr.String()
	}
	// All old party types map over by their values except for MARKER which no longer exists.
	if old.SignerRole == p8e.PartyType_PARTY_TYPE_MARKER {
		return party, fmt.Errorf("invalid signer role %s", old.SignerRole)
	}
	party.Role = PartyType(old.SignerRole)
	return party, nil
}

func convertSigners(ss *p8e.SignatureSet) ([]string, error) {
	signers := make([]string, len(ss.Signatures))
	var err error
	for i, s := range ss.Signatures {
		addr, e := getAddressFromSigner(s.Signer)
		if e != nil {
			err = appendError(err, e)
		} else {
			signers[i] = addr.String()
		}
	}
	return signers, err
}

func appendError(err1 error, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	return fmt.Errorf("%s, %s", err1.Error(), err2.Error())
}

func getAddressFromSigner(signer *p8e.SigningAndEncryptionPublicKeys) (sdk.AccAddress, error) {
	if signer == nil {
		return sdk.AccAddress{}, fmt.Errorf("nil signer")
	}
	return getAddressFromPublicKey(signer.SigningPublicKey)
}

func getAddressFromPublicKey(key *p8e.PublicKey) (sdk.AccAddress, error) {
	if key == nil {
		return sdk.AccAddress{}, fmt.Errorf("nil public key")
	}
	if key.Curve != p8e.PublicKeyCurve_SECP256K1 {
		return sdk.AccAddress{}, fmt.Errorf("address unavailable due to unsupported public key type %s", key.Curve)
	}
	_, addr, err := parsePublicKey(key.PublicKeyBytes)
	return addr, err
}

// parsePublicKey parses a secp256k1 public key, calculates the account address, and returns both.
func parsePublicKey(data []byte) (tmcrypt.PubKey, sdk.AccAddress, error) {
	// Parse the secp256k1 public key.
	pk, err := btcec.ParsePubKey(data, btcec.S256())
	if err != nil {
		return nil, nil, err
	}
	// Create tendermint public key type and return with address.
	tmKey := tmcurve.PubKey(pk.SerializeCompressed()) // PubKeySecp256k1{}
	return tmKey, tmKey.Address().Bytes(), nil
}

// partyAddresses returns an array of addresses from an array of parties
func partyAddresses(parties []Party) (addresses []string) {
	addresses = make([]string, len(parties))
	for i, p := range parties {
		addresses[i] = p.Address
	}
	return
}

func addrString(addr sdk.AccAddress, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return addr.String(), nil
}

func getValueOwner(invoker *p8e.SigningAndEncryptionPublicKeys, recitals []*p8e.Recital) (string, error) {
	// If the contract.Invoker public key matches one in contract.Recitals, then use that.
	if invoker != nil && invoker.SigningPublicKey != nil && len(invoker.SigningPublicKey.PublicKeyBytes) > 0 {
		for _, r := range recitals {
			if r != nil && r.Signer != nil && r.Signer.SigningPublicKey != nil &&
				bytes.Equal(invoker.SigningPublicKey.PublicKeyBytes, r.Signer.SigningPublicKey.PublicKeyBytes) {
				return addrString(getAddressFromPublicKey(r.Signer.SigningPublicKey))
			}
		}
	}

	// Otherwise, use scope.Parties looking for roles in this order: Marker, Owner, Originator.
	roles := []p8e.PartyType{p8e.PartyType_PARTY_TYPE_MARKER, p8e.PartyType_PARTY_TYPE_OWNER, p8e.PartyType_PARTY_TYPE_ORIGINATOR}
	for _, role := range roles {
		if r := getFirstRecitalWithRole(recitals, role); r != nil {
			return addrString(getAddressFromSigner(r.Signer))
		}
	}

	// Otherwise, just use the first party.
	if len(recitals) > 0 {
		return addrString(getAddressFromSigner(recitals[0].Signer))
	}

	return "", fmt.Errorf("no suitable party found to be value owner")
}

func getFirstRecitalWithRole(recitals []*p8e.Recital, role p8e.PartyType) *p8e.Recital {
	for _, r := range recitals {
		if r != nil && r.SignerRole == role {
			return r
		}
	}
	return nil
}

func getContractSpecID(contract *p8e.Contract) (MetadataAddress, error) {
	if contract == nil || contract.Spec == nil || contract.Spec.DataLocation == nil ||
		contract.Spec.DataLocation.Ref == nil || len(contract.Spec.DataLocation.Ref.Hash) == 0 {
		return MetadataAddress{}, fmt.Errorf("no contract.spec.datalocation.ref.hash value")
	}
	hash := contract.Spec.DataLocation.Ref.Hash

	// First... just see if it's already a bech32 address. Maybe things are looking up!
	if addr, err := MetadataAddressFromBech32(hash); err == nil {
		if addr.IsContractSpecificationAddress() {
			return addr, nil
		}
		return addr, fmt.Errorf("metadata address is not for a contract spec: %s", hash)
	}

	// Okay, it's hopefully a hash...
	return ConvertHashToAddress(ContractSpecificationKeyPrefix, hash)
}

func getProcessID(contract *p8e.Contract) (isProcess_ProcessId, error) {
	if contract == nil || contract.Definition == nil || contract.Definition.ResourceLocation == nil ||
		contract.Definition.ResourceLocation.Ref == nil {
		return nil, nil
	}
	ref := contract.Definition.ResourceLocation.Ref

	if len(ref.Hash) > 0 {
		return &Process_Hash{Hash: ref.Hash}, nil
	}

	if ref.ScopeUuid != nil && len(ref.ScopeUuid.Value) > 0 && len(ref.Name) > 0 {
		scopeUUID, err := uuid.Parse(ref.ScopeUuid.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid contract definition resource location ref scope uuid value: %w", err)
		}
		return &Process_Address{Address: RecordMetadataAddress(scopeUUID, ref.Name).String()}, nil
	}

	return nil, nil
}
