package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

var (
	AssociateSigners       = associateSigners
	FindMissingRequired    = findMissingRequired
	AssociateRequiredRoles = associateRequiredRoles
	MissingRolesError      = missingRolesError
)

// GetAuthzMessageTypeURLs ONLY FOR UNIT TESTING exposes the getAuthzMessageTypeURLs keeper function.
func (k Keeper) GetAuthzMessageTypeURLs(msgTypeURL string) []string {
	return k.getAuthzMessageTypeURLs(msgTypeURL)
}

func (k Keeper) FindAuthzGrantee(
	ctx sdk.Context,
	granter sdk.AccAddress,
	grantees []sdk.AccAddress,
	msg types.MetadataMsg,
) (sdk.AccAddress, error) {
	return k.findAuthzGrantee(ctx, granter, grantees, msg)
}

func (k Keeper) AssociateAuthorizations(
	ctx sdk.Context,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
	onAssociation func(party *PartyDetails) (stop bool),
) error {
	return k.associateAuthorizations(ctx, parties, signers, msg, onAssociation)
}

func (k Keeper) AssociateAuthorizationsForRoles(
	ctx sdk.Context,
	roles []types.PartyType,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
) (bool, error) {
	return k.associateAuthorizationsForRoles(ctx, roles, parties, signers, msg)
}

func (k Keeper) ValidateProvenanceRole(ctx sdk.Context, parties []*PartyDetails) error {
	return k.validateProvenanceRole(ctx, parties)
}

// ValidateAllOwnersAreSigners ONLY FOR UNIT TESTING exposes the validateAllOwnersAreSigners keeper function.
func (k Keeper) ValidateAllOwnersAreSigners(
	ctx sdk.Context,
	existingOwners []string,
	msg types.MetadataMsg,
) error {
	return k.validateAllOwnersAreSigners(ctx, existingOwners, msg)
}

var FindMissing = findMissing
