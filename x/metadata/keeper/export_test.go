package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

// TestablePartyDetails is the same as PartyDetails, but with
// public fields so that they can be created in unit tests as needed.
// Use the Real() method to convert it to a PartyDetails.
// I went this way instead of a NewTestPartyDetails constructor due to the
// number of arguments that one would need. Having named parameters (e.g. when
// defining a struct) is much easier to read and maintain.
type TestablePartyDetails struct {
	Address  string
	Role     types.PartyType
	Optional bool

	Acc       sdk.AccAddress
	Signer    string
	SignerAcc sdk.AccAddress

	CanBeUsedBySpec bool
	UsedBySpec      bool
}

// Real returns the PartyDetails version of this.
func (p TestablePartyDetails) Real() *PartyDetails {
	return &PartyDetails{
		address:         p.Address,
		role:            p.Role,
		optional:        p.Optional,
		acc:             p.Acc,
		signer:          p.Signer,
		signerAcc:       p.SignerAcc,
		canBeUsedBySpec: p.CanBeUsedBySpec,
		usedBySpec:      p.UsedBySpec,
	}
}

var (
	// AssociateSigners is a TEST ONLY exposure of associateSigners.
	AssociateSigners = associateSigners
	// FindMissingRequired is a TEST ONLY exposure of findMissingRequired.
	FindMissingRequired = findMissingRequired
	// AssociateRequiredRoles is a TEST ONLY exposure of associateRequiredRoles.
	AssociateRequiredRoles = associateRequiredRoles
	// MissingRolesError is a TEST ONLY exposure of missingRolesError.
	MissingRolesError = missingRolesError
	// GetAuthzMessageTypeURLs is a TEST ONLY exposure of getAuthzMessageTypeURLs.
	GetAuthzMessageTypeURLs = getAuthzMessageTypeURLs
)

// FindAuthzGrantee is a TEST ONLY exposure of findAuthzGrantee.
func (k Keeper) FindAuthzGrantee(
	ctx sdk.Context,
	granter sdk.AccAddress,
	grantees []sdk.AccAddress,
	msg types.MetadataMsg,
) (sdk.AccAddress, error) {
	return k.findAuthzGrantee(ctx, granter, grantees, msg)
}

// AssociateAuthorizations is a TEST ONLY exposure of associateAuthorizations.
func (k Keeper) AssociateAuthorizations(
	ctx sdk.Context,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
	onAssociation func(party *PartyDetails) (stop bool),
) error {
	return k.associateAuthorizations(ctx, parties, signers, msg, onAssociation)
}

// AssociateAuthorizationsForRoles is a TEST ONLY exposure of associateAuthorizationsForRoles.
func (k Keeper) AssociateAuthorizationsForRoles(
	ctx sdk.Context,
	roles []types.PartyType,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
) (bool, error) {
	return k.associateAuthorizationsForRoles(ctx, roles, parties, signers, msg)
}

// ValidateProvenanceRole is a TEST ONLY exposure of validateProvenanceRole.
func (k Keeper) ValidateProvenanceRole(ctx sdk.Context, parties []*PartyDetails) error {
	return k.validateProvenanceRole(ctx, parties)
}

var (
	// FindMissing is a TEST ONLY exposure of findMissing.
	FindMissing = findMissing
	// SafeBech32ToAccAddresses is a TEST ONLY exposure of safeBech32ToAccAddresses.
	SafeBech32ToAccAddresses = safeBech32ToAccAddresses
)
