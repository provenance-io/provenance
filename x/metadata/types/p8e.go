package types

import (
	"bytes"
	"fmt"

	"github.com/provenance-io/provenance/x/metadata/types/p8e"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcrypt "github.com/tendermint/tendermint/crypto"
	tmcurve "github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/google/uuid"
)

// EmptyScope creates a new empty Scope.
func EmptyScope() *Scope {
	return &Scope{
		ScopeId:           MetadataAddress{},
		SpecificationId:   MetadataAddress{},
		Owners:            []Party{},
		DataAccess:        []string{},
		ValueOwnerAddress: "",
	}
}

// EmptySession creates a new empty Session.
func EmptySession() *Session {
	return &Session{
		SessionId:       MetadataAddress{},
		SpecificationId: MetadataAddress{},
		Parties:         []Party{},
		Name:            "",
		Audit:           nil,
	}
}

// EmptyRecord creates a new empty Record.
func EmptyRecord() *Record {
	return &Record{
		Name:      "",
		SessionId: MetadataAddress{},
		Process:   *EmptyProcess(),
		Inputs:    []RecordInput{},
		Outputs:   []RecordOutput{},
	}
}

// EmptyProcess creates a new empty Process.
func EmptyProcess() *Process {
	return &Process{
		ProcessId: nil,
		Name:      "",
		Method:    "",
	}
}

type P8EData struct {
	Scope   *Scope
	Session *Session
	Records []*Record
}

// Migrate Converts a MsgP8EMemorializeContractRequest object into the new objects.
func ConvertP8eMemorializeContractRequest(msg *MsgP8EMemorializeContractRequest) (P8EData, []string, error) {
	p8EData := P8EData{
		Scope:   EmptyScope(),
		Session: EmptySession(),
		Records: []*Record{},
	}
	signers := []string{}
	var err error

	contractRecitalParties, err := convertParties(msg.Contract.Recitals)
	if err != nil {
		return p8EData, signers, err
	}

	// Set the scope pieces.
	p8EData.Scope.ScopeId, err = parseScopeID(msg.ScopeId)
	if err != nil {
		return p8EData, signers, err
	}
	// TODO: Set scope.SpecificationId
	//       Not sure where to get this from.
	p8EData.Scope.SpecificationId = MetadataAddress{}
	p8EData.Scope.Owners = contractRecitalParties
	p8EData.Scope.DataAccess = partyAddresses(contractRecitalParties)
	p8EData.Scope.ValueOwnerAddress, err = getValueOwner(msg.Contract.Invoker, msg.Contract.Recitals)
	if err != nil {
		return p8EData, signers, err
	}

	// Set the session pieces.
	p8EData.Session.SpecificationId, err = parseSessionID(p8EData.Scope.ScopeId, msg.GroupId)
	if err != nil {
		return p8EData, signers, err
	}
	// TODO: Set session.SpecificationId
	//       old way comes from contract.Spec.DataLocation.Ref.Hash string
	//       Might need to communicate a value change here?
	// TODO: Add session.Parties.
	//       Same as the scope Owners.
	// TODO: Set session.Name
	//       Old way: From the contract spec, .Definition.ResourceLocation.Classname
	//       New way: From the contract spec, ClassName

	// Create the records.
	// TODO: Add records.
	//       Loop through the considerations.
	//       See old repo types/apply.go func considerationsAsRecords for clues.

	// Get the signers.
	if msg.Signatures != nil {
		for _, sig := range msg.Signatures.Signatures {
			if sig != nil && len(sig.Signature) > 0 {
				// TODO: verify that the sig.Signature value is what's desired here.
				//       other data piece: sig.Signer.SigningPublicKey.PublicKeyBytes []byte
				//       See old repo types/apply.go func OnChainRecitals for clues.
				signers = append(signers, sig.Signature)
			}
		}
	}
	return p8EData, signers, err
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
			if err == nil {
				err = e
			} else {
				err = fmt.Errorf("%s, %s", err.Error(), e.Error())
			}
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
	tmKey := tmcurve.PubKey{} // PubKeySecp256k1{}
	copy(tmKey[:], pk.SerializeCompressed())
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
