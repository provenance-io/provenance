package keeper

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// ValidateSignersWithParties ensures the following:
//   - All optional=false reqParties have signed.
//   - All required roles are present in availableParties and are signers.
//   - All available parties with the PROVENANCE role are a smart contract account.
//   - All available parties with a smart contract account have the PROVENANCE role.
//   - All signers that are smart contracts are allowed to sign.
//
// The x/authz module is utilized to help facilitate signer checking.
//
//   - reqParties are the parties that might be required to sign, but might not
//     necessarily fulfill a required role. They can only fulfill a required
//     role if also provided in availableParties. Parties in reqParties with
//     optional=true, are ignored. Parties in reqParties with optional=false are
//     required to be in the msg signers.
//   - availableParties are the parties available to fulfill required roles.
//     Entries in here with optional=false are NOT required to sign (unless
//     they're in reqParties like that too).
//   - reqRoles are all the roles that are required.
//
// If a party is in both reqParties and availableParties, they are only optional
// if both have optional=true. Only parties in availableParties that are in the msg
// signers list are able to fulfill an entry in reqRoles, and each such party can
// only fulfill one required role entry.
//
// When parties and roles aren't involved, use ValidateSignersWithoutParties.
func (k Keeper) ValidateSignersWithParties(
	ctx sdk.Context,
	reqParties, availableParties []types.Party,
	reqRoles []types.PartyType,
	msg types.MetadataMsg,
) error {
	parties, err := k.validateAllRequiredPartiesSigned(ctx, reqParties, availableParties, reqRoles, msg)
	if err != nil {
		return err
	}
	if err = k.validateProvenanceRole(ctx, parties); err != nil {
		return err
	}
	return k.validateSmartContractSigners(ctx, GetUsedSigners(parties), msg)
}

// validateAllRequiredPartiesSigned ensures the following:
//   - All optional=false reqParties have signed.
//   - All required roles are present in availableParties and are signers.
//   - All parties with the PROVENANCE role are a smart contract account.
//   - All parties with a smart contract account have the PROVENANCE role.
//
// If you call this, you will probably also need to call validateSmartContractSigners on your own.
func (k Keeper) validateAllRequiredPartiesSigned(
	ctx sdk.Context,
	reqParties, availableParties []types.Party,
	reqRoles []types.PartyType,
	msg types.MetadataMsg,
) ([]*PartyDetails, error) {
	parties := BuildPartyDetails(reqParties, availableParties)
	signers := NewSignersWrapper(msg.GetSignerStrs())

	// Make sure all required parties are signers.
	associateSigners(parties, signers)
	if err := k.associateAuthorizations(ctx, findUnsignedRequired(parties), signers, msg, nil); err != nil {
		return nil, err
	}
	if missingReqParties := findUnsignedRequired(parties); len(missingReqParties) > 0 {
		missing := make([]string, len(missingReqParties))
		for i, party := range missingReqParties {
			missing[i] = fmt.Sprintf("%s (%s)", party.GetAddress(), party.GetRole().SimpleString())
		}
		return nil, fmt.Errorf("missing required signature%s: %s", pluralEnding(len(missing)), strings.Join(missing, ", "))
	}

	// Make sure all required roles are present as signers.
	missingRoles := associateRequiredRoles(parties, reqRoles)
	rolesAreMissing, err := k.associateAuthorizationsForRoles(ctx, missingRoles, parties, signers, msg)
	if err != nil {
		return nil, err
	}
	if rolesAreMissing {
		return nil, fmt.Errorf("missing signers for roles required by spec: %s", missingRolesString(parties, reqRoles))
	}

	return parties, nil
}

// associateSigners updates each PartyDetails to indicate there's a signer if its
// address is in the signers list.
func associateSigners(parties []*PartyDetails, signers *SignersWrapper) {
	if signers == nil {
		return
	}
	for _, party := range parties {
		partyAddress := party.GetAddress()
		for _, signer := range signers.Strings() {
			if partyAddress == signer {
				party.SetSigner(signer)
				break
			}
		}
	}
}

// findUnsignedRequired returns a list of parties that are required (optional=false)
// and don't have a signer.
func findUnsignedRequired(parties []*PartyDetails) []*PartyDetails {
	var rv []*PartyDetails
	for _, party := range parties {
		if party.IsRequired() && !party.HasSigner() {
			rv = append(rv, party)
		}
	}
	return rv
}

// associateRequiredRoles goes through the required roles, marking parties as used
// when possible. Returns a list of required role entries that haven't yet been fulfilled.
//
// This is similar to validateRolesPresent except this requires a role to have a signer
// in order for it to fulfill a required role.
func associateRequiredRoles(parties []*PartyDetails, reqRoles []types.PartyType) []types.PartyType {
	var missingRoles []types.PartyType
reqRolesLoop:
	for _, role := range reqRoles {
		for _, party := range parties {
			if party.IsStillUsableAs(role) && party.HasSigner() {
				party.MarkAsUsed()
				continue reqRolesLoop
			}
		}
		missingRoles = append(missingRoles, role)
	}
	return missingRoles
}

// missingRolesString generates and returns an error message indicating that
// some required roles don't have signers.
func missingRolesString(parties []*PartyDetails, reqRoles []types.PartyType) string {
	// Get a count for each required role
	reqCountByRole := make(map[types.PartyType]int)
	for _, role := range reqRoles {
		reqCountByRole[role]++
	}

	// Get a count of each used party for each role.
	haveCountByRole := make(map[types.PartyType]int)
	for _, party := range parties {
		if party.IsUsed() {
			haveCountByRole[party.role]++
		}
	}

	// Generate the message strings for each role that is short.
	messageByRole := make(map[types.PartyType]string)
	var missingRoles []types.PartyType
	for role, reqCount := range reqCountByRole {
		if reqCount > haveCountByRole[role] {
			messageByRole[role] = fmt.Sprintf("%s need %d have %d",
				role.SimpleString(), reqCountByRole[role], haveCountByRole[role])
			missingRoles = append(missingRoles, role)
		}
	}
	// Sort the missing roles so that this result can be deterministic.
	sort.Slice(missingRoles, func(i, j int) bool {
		return missingRoles[i] < missingRoles[j]
	})

	// Order the messages for each of the missing roles.
	missing := make([]string, len(missingRoles))
	for i, role := range missingRoles {
		missing[i] = messageByRole[role]
	}

	return strings.Join(missing, ", ")
}

// getAuthzMessageTypeURLs gets all msg type URLs that authz authorizations might
// be under for the provided msg type URL. It basically allows a single authorization
// to be usable from multiple related endpoints. E.g. a MsgWriteScopeRequest authorization
// is usable for the MsgAddScopeDataAccessRequest endpoint as well.
func getAuthzMessageTypeURLs(msgTypeURL string) []string {
	urls := make([]string, 0, 2)
	if len(msgTypeURL) > 0 {
		urls = append(urls, msgTypeURL)
	}
	switch msgTypeURL {
	case types.TypeURLMsgAddScopeDataAccessRequest, types.TypeURLMsgDeleteScopeDataAccessRequest,
		types.TypeURLMsgAddScopeOwnerRequest, types.TypeURLMsgDeleteScopeOwnerRequest,
		types.TypeURLMsgUpdateValueOwnersRequest:
		urls = append(urls, types.TypeURLMsgWriteScopeRequest)
	case types.TypeURLMsgWriteRecordRequest:
		urls = append(urls, types.TypeURLMsgWriteSessionRequest)
	case types.TypeURLMsgAddContractSpecToScopeSpecRequest, types.TypeURLMsgDeleteContractSpecFromScopeSpecRequest:
		urls = append(urls, types.TypeURLMsgWriteScopeSpecificationRequest)
	case types.TypeURLMsgWriteRecordSpecificationRequest:
		urls = append(urls, types.TypeURLMsgWriteContractSpecificationRequest)
	case types.TypeURLMsgDeleteRecordSpecificationRequest:
		urls = append(urls, types.TypeURLMsgDeleteContractSpecificationRequest)
	}
	return urls
}

// findAuthzGrantee checks authz for authorizations from the granter to each of the grantees.
// If one is found and accepted, the authorization is updated and that grantee is returned.
// An error is returned if there was a problem updating an authorization.
// If no authorization is found, and no error is encountered, then nil, nil is returned.
func (k Keeper) findAuthzGrantee(
	ctx sdk.Context,
	granter sdk.AccAddress,
	grantees []sdk.AccAddress,
	msg types.MetadataMsg,
) (sdk.AccAddress, error) {
	if len(granter) == 0 || len(grantees) == 0 {
		return nil, nil
	}
	cache := GetAuthzCache(ctx)
	msgTypes := getAuthzMessageTypeURLs(sdk.MsgTypeURL(msg))
	for _, grantee := range grantees {
		for _, msgType := range msgTypes {
			prevAuth := cache.GetAcceptable(grantee, granter, msgType)
			if prevAuth != nil {
				return grantee, nil
			}
			authorization, exp := k.authzKeeper.GetAuthorization(ctx, grantee, granter, msgType)
			if authorization != nil {
				// If Accept returns an error, we just ignore this authorization
				// and look for another that'll work.
				resp, err := authorization.Accept(ctx, msg)
				if err == nil && resp.Accept {
					switch {
					case resp.Delete:
						err = k.authzKeeper.DeleteGrant(ctx, grantee, granter, msgType)
						if err != nil {
							return nil, err
						}
					case resp.Updated != nil:
						if err = k.authzKeeper.SaveGrant(ctx, grantee, granter, resp.Updated, exp); err != nil {
							return nil, err
						}
					}
					cache.SetAcceptable(grantee, granter, msgType, authorization)
					return grantee, nil
				}
			}
		}
	}
	return nil, nil
}

// associateAuthorizations checks authz for authorizations from each party (the granters) to
// each signer (the grantees). If found, updates the party details to indicate there's a signer.
// The onAssociation function is called when a grantee is found; it should return whether
// to stop checking (i.e. true => stop now, false => keep checking the rest of the parties).
func (k Keeper) associateAuthorizations(
	ctx sdk.Context,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
	onAssociation func(party *PartyDetails) (stop bool),
) error {
	for _, party := range parties {
		if !party.HasSigner() {
			grantee, err := k.findAuthzGrantee(ctx, party.GetAcc(), signers.Accs(), msg)
			if err != nil {
				return err
			}
			if len(grantee) > 0 {
				party.SetSignerAcc(grantee)
				if onAssociation != nil && onAssociation(party) {
					break
				}
			}
		}
	}
	return nil
}

// associateAuthorizationsForRoles goes through each entry in roles and attempts
// to find an authorization from an appropriate party. When one is found, the
// party is marked as used.
// An error is returned if one is encountered while updating an authorization.
// True is returned if no usable party is found for one or more roles
// False is returned if all roles have been fulfilled.
//
// This assumes:
//   - Only roles that haven't yet been fulfilled are provided (e.g. roles = the result of associateRequiredRoles).
//   - If a party has a signer, it's already been considered (e.g. parties have been run through associateRequiredRoles).
func (k Keeper) associateAuthorizationsForRoles(
	ctx sdk.Context,
	roles []types.PartyType,
	parties []*PartyDetails,
	signers *SignersWrapper,
	msg types.MetadataMsg,
) (bool, error) {
	missingRoles := false
	for _, role := range roles {
		found := false
		var partiesToCheck []*PartyDetails
		for _, party := range parties {
			if party.IsStillUsableAs(role) && !party.HasSigner() {
				partiesToCheck = append(partiesToCheck, party)
			}
		}
		err := k.associateAuthorizations(ctx, partiesToCheck, signers, msg, func(party *PartyDetails) bool {
			party.MarkAsUsed()
			found = true
			return true
		})
		if err != nil {
			return true, err
		}
		if !found {
			missingRoles = true
			// We still want to process the rest so that the error message has the correct counts.
		}
	}

	return missingRoles, nil
}

// validateProvenanceRole makes sure that:
//   - All parties with the address of a smart contract have the PROVENANCE role.
//   - All parties with the PROVENANCE role have the address of a smart contract.
func (k Keeper) validateProvenanceRole(ctx sdk.Context, parties []*PartyDetails) error {
	for _, party := range parties {
		if party.CanBeUsed() {
			// Using the party address here (instead of the signer) because it's
			// that address that needs to be the smart contract.
			addr := party.GetAcc()
			if len(addr) > 0 {
				isWasmAcct := k.isWasmAccount(ctx, party.GetAcc())
				isProvRole := party.GetRole() == types.PartyType_PARTY_TYPE_PROVENANCE
				if isWasmAcct && !isProvRole {
					return fmt.Errorf("account %q is a smart contract but does not have the PROVENANCE role", party.GetAddress())
				}
				if !isWasmAcct && isProvRole {
					return fmt.Errorf("account %q has role PROVENANCE but is not a smart contract", party.GetAddress())
				}
			}
		}
	}
	return nil
}

// isWasmAccount returns true if the provided addr is the address of a smart contract account.
// A smart contract account is a BaseAccount that exists, has a sequence of 0 and does not have a public key.
func (k Keeper) isWasmAccount(ctx sdk.Context, addr sdk.AccAddress) bool {
	if len(addr) == 0 {
		return false
	}
	account, isBaseAccount := k.authKeeper.GetAccount(ctx, addr).(*authtypes.BaseAccount)
	return account != nil && isBaseAccount && account.GetSequence() == uint64(0) && account.GetPubKey() == nil
}

// validateSmartContractSigners makes sure that any msg signers that are smart contracts
// are in the usedSigners map or are authorized by all signers after them.
// The usedSigners map has bech32 keys and value indicating whether that address was
// used as a signer in some capacity (e.g. they're a party).
func (k Keeper) validateSmartContractSigners(ctx sdk.Context, usedSigners UsedSignersMap, msg types.MetadataMsg) error {
	// When a smart contract is a signer, they must either be used as a signer
	// already, or must be authorized by all signers after it.
	// The wasm encoders (hopefully) put the smart contract as the first signer
	// followed by other signers. That's why we only check the signers after it.
	signerAccs := msg.GetSigners()
	canBeWasm := true
	for i, signer := range signerAccs {
		signerStr := signer.String()
		isWasm := k.isWasmAccount(ctx, signer)
		if isWasm && !canBeWasm {
			return fmt.Errorf("smart contract signer %s cannot follow non-smart-contract signer", signer)
		}
		if !isWasm {
			canBeWasm = false
			continue
		}
		if usedSigners.IsUsed(signerStr) {
			continue
		}
		// it's a wasm account, and it wasn't used yet.
		if i+1 >= len(signerAccs) {
			// Not fully accurate error message here, but close enough. And we'll probably never see it anyway.
			// A smart contract would be allowed to be the last signer if used, e.g. in a Party. We don't need
			// to tell people that though. But we need this in case, somehow, a smart contract is doing things
			// without any other signers, but it isn't supposed to be involved in what's going on.
			return fmt.Errorf("smart contract signer %s cannot be the last signer", signerStr)
		}
		// Make sure each of the remaining addresses have granted authorization to this smart contract.
		for _, granter := range signerAccs[i+1:] {
			grantee, err := k.findAuthzGrantee(ctx, granter, []sdk.AccAddress{signer}, msg)
			if err != nil {
				return err
			}
			if !signer.Equals(grantee) {
				return fmt.Errorf("smart contract signer %s is not authorized", signer)
			}
		}
	}
	return nil
}

// ValidateScopeValueOwnerUpdate verifies that it's okay for the msg signers to
// change a scope's value owner from existing to proposed.
// If some parties have already been validated (possibly utilizing authz), they
// can be provided in order to prevent an authorization from being used twice during
// a single Tx.
//
// If no error is returned, a map of bech32 strings to true is returned where each key
// is a signer that either has a signer in validatedParties, or is used directly in here.
func (k Keeper) ValidateScopeValueOwnerUpdate(
	ctx sdk.Context,
	existing,
	proposed string,
	msg types.MetadataMsg,
) (UsedSignersMap, error) {
	if existing == proposed {
		return NewUsedSignersMap(), nil
	}
	signers := NewSignersWrapper(msg.GetSignerStrs())

	usedSigners, err := k.validateScopeValueOwnerChangeFromExisting(ctx, existing, signers, msg)
	if err != nil {
		return nil, err
	}

	newUsedSigners, err := k.validateScopeValueOwnerChangeToProposed(ctx, proposed, signers)
	if err != nil {
		return nil, err
	}

	return usedSigners.AlsoUse(newUsedSigners), nil
}

// validateScopeValueOwnerChangeFromExisting validates that the provided signers
// are allowed to change the existing value owner.
func (k Keeper) validateScopeValueOwnerChangeFromExisting(
	ctx sdk.Context,
	existing string,
	signers *SignersWrapper,
	msg types.MetadataMsg,
) (UsedSignersMap, error) {
	usedSigners := NewUsedSignersMap()

	// Nothing to check (in here) if the existing is empty.
	if len(existing) == 0 {
		return usedSigners, nil
	}

	// If the existing is a marker, make sure a signer has withdraw authority on it.
	marker, hasAuth, accWithAccess := k.GetMarkerAndCheckAuthority(ctx, existing, signers.Strings(), markertypes.Access_Withdraw)
	if marker != nil {
		if !hasAuth {
			return nil, fmt.Errorf("missing signature for %s (%s) with authority to withdraw/remove it as scope value owner", existing, marker.GetDenom())
		}
		return usedSigners.Use(accWithAccess), nil
	}

	// If the existing isn't a marker, make sure they're one of the signers or
	// have an authorization grant for one of the signers.
	for _, signer := range signers.Strings() {
		if existing == signer {
			return usedSigners.Use(signer), nil
		}
	}

	// Not a signer. Check with authz for help.
	// If existing isn't a bech32, we just skip the authz check. Should only happen in unit tests.
	granter, err := sdk.AccAddressFromBech32(existing)
	if err == nil {
		// For the value owner address, we only check authz for non smart-contract signers
		// This prevents Alice from using a smart contract to update Bob's
		// scope when both have authorized the smart contract to WriteScope.
		// But it allows Bob to authorize Alice and then Alice can update Bob's scope regardless
		// of whether it's by means of a smart contract.
		var grantees []sdk.AccAddress
		for _, signer := range signers.Accs() {
			if !k.isWasmAccount(ctx, signer) {
				grantees = append(grantees, signer)
			}
		}
		grantee, err := k.findAuthzGrantee(ctx, granter, grantees, msg)
		if err != nil {
			return nil, fmt.Errorf("authz error with existing value owner %q: %w", existing, err)
		}
		if len(grantee) > 0 {
			return usedSigners.Use(grantee.String()), nil
		}
	}

	return nil, fmt.Errorf("missing signature from existing value owner %s", existing)
}

// validateScopeValueOwnerChangeToProposed validates that the provided signers
// are allowed to set the value owner to the proposed value.
func (k Keeper) validateScopeValueOwnerChangeToProposed(
	ctx sdk.Context,
	proposed string,
	signers *SignersWrapper,
) (UsedSignersMap, error) {
	usedSigners := NewUsedSignersMap()

	// Nothing to check if the proposed is empty.
	if len(proposed) == 0 {
		return usedSigners, nil
	}

	// If the proposed is a marker, make sure a signer has deposit authority on it.
	marker, hasAuth, accWithAccess := k.GetMarkerAndCheckAuthority(ctx, proposed, signers.Strings(), markertypes.Access_Deposit)
	if marker != nil {
		if !hasAuth {
			return nil, fmt.Errorf("missing signature for %s (%s) with authority to deposit/add it as scope value owner", proposed, marker.GetDenom())
		}
		return usedSigners.Use(accWithAccess), nil
	}

	// If the proposed isn't a marker, we don't really care what it's being set to and no one needs to sign.
	return usedSigners, nil
}

// ValidateSignersWithoutParties makes sure that each entry in the required list are either signers of the msg,
// or have granted an authz authorization to one of the signers.
// It then makes sure that any signers that are smart contracts are allowed to sign.
//
// When parties (and/or roles) are involved, use ValidateSignersWithParties.
func (k Keeper) ValidateSignersWithoutParties(
	ctx sdk.Context,
	required []string,
	msg types.MetadataMsg,
) error {
	parties, err := k.validateAllRequiredSigned(ctx, required, msg)
	if err != nil {
		return err
	}
	return k.validateSmartContractSigners(ctx, GetUsedSigners(parties), msg)
}

// validateAllRequiredSigned ensures that all required addresses are either in the msg signers,
// or have granted an authorization to someone in the signers.
//
// If you call this, you will probably also need to call validateSmartContractSigners on your own.
func (k Keeper) validateAllRequiredSigned(ctx sdk.Context, required []string, msg types.MetadataMsg) ([]*PartyDetails, error) {
	details := make([]*PartyDetails, len(required))
	for i, addr := range required {
		details[i] = &PartyDetails{
			address:  addr,
			role:     types.PartyType_PARTY_TYPE_UNSPECIFIED,
			optional: false,
		}
	}

	signers := NewSignersWrapper(msg.GetSignerStrs())

	// First pass: without authz.
	associateSigners(details, signers)
	missingReqParties := findUnsignedRequired(details)

	// Second pass: Check authz for any authorizations on missing signers.
	if len(missingReqParties) > 0 {
		if err := k.associateAuthorizations(ctx, missingReqParties, signers, msg, nil); err != nil {
			return nil, err
		}
		missingReqParties = findUnsignedRequired(details)
	}

	if len(missingReqParties) > 0 {
		missing := make([]string, len(missingReqParties))
		for i, party := range missingReqParties {
			missing[i] = party.GetAddress()
		}
		return nil, fmt.Errorf("missing signature%s: %s", pluralEnding(len(missing)), strings.Join(missing, ", "))
	}

	return details, nil
}

// validateRolesPresent returns an error if one or more required roles are not present in the parties.
//
// This is similar to associateRequiredRoles, except this one doesn't require the party to have a signer.
func validateRolesPresent(parties []types.Party, reqRoles []types.PartyType) error {
	details := BuildPartyDetails(nil, parties)
	roleMissing := false
reqRolesLoop:
	for _, role := range reqRoles {
		for _, party := range details {
			if party.IsStillUsableAs(role) {
				party.MarkAsUsed()
				continue reqRolesLoop
			}
		}
		roleMissing = true
	}
	if roleMissing {
		return fmt.Errorf("missing roles required by spec: %s", missingRolesString(details, reqRoles))
	}
	return nil
}

// validatePartiesArePresent returns an error if there are any parties in required that are not in available.
func validatePartiesArePresent(required, available []types.Party) error {
	missing := findMissingParties(required, available)
	if len(missing) == 0 {
		return nil
	}
	parts := make([]string, len(missing))
	for i, party := range missing {
		parts[i] = fmt.Sprintf("%s (%s)", party.Address, party.Role.SimpleString())
	}
	word := "party"
	if len(missing) != 1 {
		word = "parties"
	}
	return fmt.Errorf("missing %s: %s", word, strings.Join(parts, ", "))
}

// GetMarkerAndCheckAuthority gets a marker by address and checks if one of the signers has the provided role.
// If the address isn't a marker, nil, false is returned.
// The signer that has the requested permission is also returned.
func (k Keeper) GetMarkerAndCheckAuthority(
	ctx sdk.Context,
	address string,
	signers []string,
	role markertypes.Access,
) (markertypes.MarkerAccountI, bool, string) {
	addr, err := sdk.AccAddressFromBech32(address)
	// if the address is invalid then it is not possible for it to be a marker.
	if err != nil {
		return nil, false, ""
	}

	acc := k.authKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, false, ""
	}

	// Convert over to the actual underlying marker type, or not.
	marker, isMarker := acc.(*markertypes.MarkerAccount)
	if !isMarker {
		return nil, false, ""
	}

	// Check if any of the signers have the desired role.
	for _, signer := range signers {
		if marker.HasAccess(signer, role) {
			return marker, true, signer
		}
	}

	return marker, false, ""
}
