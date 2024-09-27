package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

// GetStoreKey is a TEST ONLY getter for the keeper's storekey.
func (k *Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

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

// SetBankKeeper is a TEST ONLY setter for the keeper's bank keeper.
// It returns the previously defined BankKeeper
func (k *Keeper) SetBankKeeper(bankKeeper BankKeeper) BankKeeper {
	rv := k.bankKeeper
	k.bankKeeper = bankKeeper
	return rv
}

// SetBankKeeper is a TEST ONLY setter for the keeper's marker keeper.
// It returns the previously defined MarkerKeeper
func (k *Keeper) SetMarkerKeeper(markerKeeper MarkerKeeper) MarkerKeeper {
	rv := k.markerKeeper
	k.markerKeeper = markerKeeper
	return rv
}

// WriteScopeToState is a TEST ONLY exposure of writeScopeToState.
func (k *Keeper) WriteScopeToState(ctx sdk.Context, scope types.Scope) {
	k.writeScopeToState(ctx, scope)
}

// ValidateAllRequiredPartiesSigned is a TEST ONLY exposure of validateAllRequiredPartiesSigned.
func (k Keeper) ValidateAllRequiredPartiesSigned(
	ctx sdk.Context,
	reqParties, availableParties []types.Party,
	reqRoles []types.PartyType,
	msg types.MetadataMsg,
) ([]*types.PartyDetails, error) {
	return k.validateAllRequiredPartiesSigned(ctx, reqParties, availableParties, reqRoles, msg)
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
	parties []*types.PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
	onAssociation func(party *types.PartyDetails) (stop bool),
) error {
	return k.associateAuthorizations(ctx, parties, signers, msg, onAssociation)
}

// AssociateAuthorizationsForRoles is a TEST ONLY exposure of associateAuthorizationsForRoles.
func (k Keeper) AssociateAuthorizationsForRoles(
	ctx sdk.Context,
	roles []types.PartyType,
	parties []*types.PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
) (bool, error) {
	return k.associateAuthorizationsForRoles(ctx, roles, parties, signers, msg)
}

// ValidateProvenanceRole is a TEST ONLY exposure of validateProvenanceRole.
func (k Keeper) ValidateProvenanceRole(ctx sdk.Context, parties []*types.PartyDetails) error {
	return k.validateProvenanceRole(ctx, parties)
}

// IsWasmAccount is a TEST ONLY exposure of isWasmAccount.
func (k Keeper) IsWasmAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	return k.isWasmAccount(ctx, addr)
}

// ValidateAllRequiredSigned is a TEST ONLY exposure of validateAllRequiredSigned.
func (k Keeper) ValidateAllRequiredSigned(ctx sdk.Context, required []string, msg types.MetadataMsg) ([]*types.PartyDetails, error) {
	return k.validateAllRequiredSigned(ctx, required, msg)
}

// ValidateSmartContractSigners is a TEST ONLY exposure of validateSmartContractSigners.
func (k Keeper) ValidateSmartContractSigners(ctx sdk.Context, usedSigners types.UsedSignersMap, msg types.MetadataMsg) error {
	return k.validateSmartContractSigners(ctx, usedSigners, msg)
}

var (
	// ValidateRolesPresent is a TEST ONLY exposure of validateRolesPresent.
	ValidateRolesPresent = validateRolesPresent
	// ValidatePartiesArePresent is a TEST ONLY exposure of validatePartiesArePresent.
	ValidatePartiesArePresent = validatePartiesArePresent
)

var (
	// SafeBech32ToAccAddresses is a TEST ONLY exposure of safeBech32ToAccAddresses.
	SafeBech32ToAccAddresses = safeBech32ToAccAddresses
)

var (
	// NewKeeper3To4 is a TEST ONLY exposure of newKeeper3To4.
	NewKeeper3To4 = newKeeper3To4
	// MigrateValueOwners is a TEST ONLY exposure of migrateValueOwners.
	MigrateValueOwners = migrateValueOwners
	// MigrateValueOwnerToBank is a TEST ONLY exposure of migrateValueOwnerToBank.
	MigrateValueOwnerToBank = migrateValueOwnerToBank
	// DeleteValueOwnerIndexEntries is a TEST ONLY exposure of deleteValueOwnerIndexEntries.
	DeleteValueOwnerIndexEntries = deleteValueOwnerIndexEntries
)

// Keeper3To4 is a TEST ONLY exposure of keeper3To4.
type Keeper3To4 = keeper3To4
