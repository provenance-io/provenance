package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

// SetAuthKeeper is a TEST ONLY setter for the keeper's authKeeper.
// It returns the previously defined AuthKeeper
func (k *Keeper) SetAuthKeeper(authKeeper AuthKeeper) AuthKeeper {
	rv := k.authKeeper
	k.authKeeper = authKeeper
	return rv
}

// SetAuthzKeeper is a TEST ONLY setter for the keeper's authzKeeper.
// It returns the previously defined AuthzKeeper
func (k *Keeper) SetAuthzKeeper(authzKeeper AuthzKeeper) AuthzKeeper {
	rv := k.authzKeeper
	k.authzKeeper = authzKeeper
	return rv
}

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

// Testable is a TEST ONLY function that converts a PartyDetails into a TestablePartyDetails.
func (p *PartyDetails) Testable() TestablePartyDetails {
	return TestablePartyDetails{
		Address:         p.address,
		Role:            p.role,
		Optional:        p.optional,
		Acc:             p.acc,
		Signer:          p.signer,
		SignerAcc:       p.signerAcc,
		CanBeUsedBySpec: p.canBeUsedBySpec,
		UsedBySpec:      p.usedBySpec,
	}
}

// Copy is a TEST ONLY function that copies a PartyDetails.
func (p *PartyDetails) Copy() *PartyDetails {
	if p == nil {
		return nil
	}
	rv := &PartyDetails{
		address:         p.address,
		role:            p.role,
		optional:        p.optional,
		acc:             nil,
		signer:          p.signer,
		signerAcc:       nil,
		canBeUsedBySpec: p.canBeUsedBySpec,
		usedBySpec:      p.usedBySpec,
	}
	if p.acc != nil {
		rv.acc = make(sdk.AccAddress, len(p.acc))
		copy(rv.acc, p.acc)
	}
	if p.signerAcc != nil {
		rv.signerAcc = make(sdk.AccAddress, len(p.signerAcc))
		copy(rv.signerAcc, p.signerAcc)
	}
	return rv
}

var (
	// AssociateSigners is a TEST ONLY exposure of associateSigners.
	AssociateSigners = associateSigners
	// FindUnsignedRequired is a TEST ONLY exposure of findUnsignedRequired.
	FindUnsignedRequired = findUnsignedRequired
	// AssociateRequiredRoles is a TEST ONLY exposure of associateRequiredRoles.
	AssociateRequiredRoles = associateRequiredRoles
	// MissingRolesString is a TEST ONLY exposure of missingRolesString.
	MissingRolesString = missingRolesString
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
	// ValidateRolesPresent is a TEST ONLY exposure of validateRolesPresent.
	ValidateRolesPresent = validateRolesPresent
	// ValidatePartiesArePresent is a TEST ONLY exposure of validatePartiesArePresent.
	ValidatePartiesArePresent = validatePartiesArePresent
	// FindMissing is a TEST ONLY exposure of findMissing.
	FindMissing = findMissing
	// FindMissingParties is a TEST ONLY exposure of findMissingParties.
	FindMissingParties = findMissingParties
)

// FindMissingComp is a TEST ONLY exposure of findMissingComp.
func FindMissingComp[R any, C any](required []R, toCheck []C, comp func(R, C) bool) []R {
	return findMissingComp(required, toCheck, comp)
}

var (
	// PluralEnding is a TEST ONLY exposure of pluralEnding.
	PluralEnding = pluralEnding
	// SafeBech32ToAccAddresses is a TEST ONLY exposure of safeBech32ToAccAddresses.
	SafeBech32ToAccAddresses = safeBech32ToAccAddresses
)
