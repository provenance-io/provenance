package types

import "github.com/golang/protobuf/proto"

type AuthenticatorI interface {
	proto.Message

	GetCredentialNumber() uint64
	SetCredentialNumber(uint64)
	SetEmail(string)
	Type() CredentialType
	Credential() *Credential
	Ephemeral() bool
}

// https://w3c.github.io/webauthn/#iface-pkcredential
// Attestation represents the attestation data structure
type Attestation struct {
	ID       string              `json:"id"`
	RawID    string              `json:"rawId"`
	Response AttestationResponse `json:"response"`
	Type     string              `json:"type"`
}

// AttestationResponse represents the attestation response structure
type AttestationResponse struct {
	AttestationObject string `json:"attestationObject"`
	ClientDataJSON    string `json:"clientDataJSON"`
}
