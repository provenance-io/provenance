package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// GetAuthzMessageTypeURLs gets all msg type URLs that authz authorizations might
// be under for the provided msg type URL. It basicallly allows a single authorization
// be usable from multiple related endpoints. E.g. a MsgWriteScopeRequest authorization
// is usable for the MsgAddScopeDataAccessRequest endpoint as well.
func (k Keeper) GetAuthzMessageTypeURLs(msgTypeURL string) []string {
	urls := []string{}
	if len(msgTypeURL) > 0 {
		urls = append(urls, msgTypeURL)
	}
	switch msgTypeURL {
	case types.TypeURLMsgAddScopeDataAccessRequest, types.TypeURLMsgDeleteScopeDataAccessRequest,
		types.TypeURLMsgAddScopeOwnerRequest, types.TypeURLMsgDeleteScopeOwnerRequest:
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

// checkAuthzForMissing returns any of the provided addrs that have not been granted an authz authorization by one of the msg signers.
// An error is returned if there was a problem updating an authorization.
func (k Keeper) checkAuthzForMissing(
	ctx sdk.Context,
	addrs []string,
	msg types.MetadataMsg,
) ([]string, error) {
	stillMissing := []string{}
	if len(addrs) == 0 {
		return stillMissing, nil
	}

	signers := msg.GetSignersStr()
	signerAddrs := make([]sdk.AccAddress, 0, len(signers))
	for _, signer := range signers {
		signerAddr, err := sdk.AccAddressFromBech32(signer)
		// If it's not an address, there's no way there's an authorization for them.
		// This is mostly allowed for unit tests.
		// In actual usage, there's very little chance of it not being an address here.
		if err == nil {
			signerAddrs = append(signerAddrs, signerAddr)
		}
	}

	// return as a list this message type and its parent
	// type if it is a message belonging to a hierarchy
	msgTypeURLs := k.GetAuthzMessageTypeURLs(sdk.MsgTypeURL(msg))

	for _, addr := range addrs {
		granter, addrErr := sdk.AccAddressFromBech32(addr)
		found := false

		// if the addr wasn't an AccAddress, authz isn't going to help.
		// This is mostly allowed for unit tests.
		// In actual usage, there's very little chance of it not being an address here.
		if addrErr == nil {
			// loop through all the signers
			for _, grantee := range signerAddrs {
				for _, msgType := range msgTypeURLs {
					authorization, exp := k.authzKeeper.GetAuthorization(ctx, grantee, granter, msgType)
					if authorization != nil {
						resp, err := authorization.Accept(ctx, msg)
						if err == nil && resp.Accept {
							switch {
							case resp.Delete:
								err = k.authzKeeper.DeleteGrant(ctx, grantee, granter, msgType)
								if err != nil {
									return stillMissing, err
								}
							case resp.Updated != nil:
								if err = k.authzKeeper.SaveGrant(ctx, grantee, granter, resp.Updated, exp); err != nil {
									return stillMissing, err
								}
							}
							found = true
							break
						}
					}
				}
				if found {
					break
				}
			}
		}

		if !found {
			stillMissing = append(stillMissing, addr)
		}
	}

	return stillMissing, nil
}

// ValidateAllOwnersAreSignersWithAuthz makes sure that all entries in the existingOwners list
// are contained in the signers list and checks to see if any missing entries have an assigned grantee
func (k Keeper) ValidateAllOwnersAreSignersWithAuthz(
	ctx sdk.Context,
	existingOwners []string,
	msg types.MetadataMsg,
) error {
	signers := msg.GetSignersStr()
	missing := FindMissing(existingOwners, signers)
	stillMissing := missing
	var err error
	// Authz grants rights to address on specific message types.
	// If no message is provided, skip the Authz check.
	if msg != nil {
		stillMissing, err = k.checkAuthzForMissing(ctx, missing, msg)
		if err != nil {
			return fmt.Errorf("error validating signers: %w", err)
		}
	}

	switch len(stillMissing) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("missing signature from existing owner %s; required for update", stillMissing[0])
	default:
		return fmt.Errorf("missing signatures from existing owners %v; required for update", stillMissing)
	}
}

// ValidateAllPartiesAreSignersWithAuthz validate all parties are signers with authz module
func (k Keeper) ValidateAllPartiesAreSignersWithAuthz(ctx sdk.Context, parties []types.Party, msg types.MetadataMsg) error {
	addresses := make([]string, len(parties))
	for i, party := range parties {
		addresses[i] = party.Address
	}
	signers := msg.GetSignersStr()
	missing := FindMissing(addresses, signers)
	stillMissing := missing
	var err error
	// Authz grants rights to address on specific message types.
	// If no message is provided, skip the Authz check.
	if msg != nil {
		stillMissing, err = k.checkAuthzForMissing(ctx, missing, msg)
		if err != nil {
			return fmt.Errorf("error validating signers: %w", err)
		}
	}

	if len(stillMissing) > 0 {
		missingWithRoles := make([]string, len(missing))
		for i, addr := range stillMissing {
			for _, party := range parties {
				if addr == party.Address {
					missingWithRoles[i] = fmt.Sprintf("%s (%s)", addr, party.Role.String())
					break
				}
			}
		}
		return fmt.Errorf("missing signature%s from %v", pluralEnding(len(missing)), missingWithRoles)
	}

	return nil
}

// ValidatePartiesInvolved validate that all required parties are involved
func (k Keeper) ValidatePartiesInvolved(parties []types.Party, requiredParties []types.PartyType) error {
	partyRoles := make([]string, len(parties))
	reqRoles := make([]string, len(requiredParties))
	for i, party := range parties {
		partyRoles[i] = party.Role.String()
	}
	for i, req := range requiredParties {
		reqRoles[i] = req.String()
	}
	missing := FindMissing(reqRoles, partyRoles)
	if len(missing) > 0 {
		return fmt.Errorf("missing required party type%s %v from parties", pluralEnding(len(missing)), missing)
	}
	return nil
}

// ValidateScopeOwners is stateful validation for scope owners against a scope specification.
// This does NOT involve the Scope.ValidateOwnersBasic() function.
func (k Keeper) ValidateScopeOwners(owners []types.Party, spec types.ScopeSpecification) error {
	var missingPartyTypes []string
	for _, pt := range spec.PartiesInvolved {
		found := false
		for _, o := range owners {
			if o.Role == pt {
				found = true
				break
			}
		}
		if !found {
			// Get the party type without the "PARTY_TYPE_" prefix.
			missingPartyTypes = append(missingPartyTypes, pt.String()[11:])
		}
	}
	if len(missingPartyTypes) > 0 {
		return fmt.Errorf("missing party type%s required by spec: %v", pluralEnding(len(missingPartyTypes)), missingPartyTypes)
	}
	return nil
}

// FindMissing returns all elements of the required list that are not found in the entries list
// It is exported only so that it can be unit tested.
func FindMissing(required []string, entries []string) []string {
	retval := []string{}
	for _, req := range required {
		found := false
		for _, entry := range entries {
			if req == entry {
				found = true
				break
			}
		}
		if !found {
			retval = append(retval, req)
		}
	}
	return retval
}

// IsMarkerAndHasAuthority checks that the address is a marker addr and that one of the signers has the given role.
// First return boolean is whether the address is a marker.
// Second return boolean is whether one of the signers has the given role on that marker.
// If the first return boolean is false, they'll both be false.
func (k Keeper) IsMarkerAndHasAuthority(ctx sdk.Context, address string, signers []string, role markertypes.Access) (isMarker bool, hasAuth bool) {
	addr, err := sdk.AccAddressFromBech32(address)
	// if the address is invalid then it is not possible for it to be a marker.
	if err != nil {
		return false, false
	}

	acc := k.authKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return false, false
	}

	// Convert over to the actual underlying marker type, or not.
	marker, isMarker := acc.(*markertypes.MarkerAccount)
	if !isMarker {
		return false, false
	}

	// Check if any of the signers have the desired role.
	for _, signer := range signers {
		if marker.HasAccess(signer, role) {
			return true, true
		}
	}

	return true, false
}

// pluralEnding returns "" if i == 1, or "s" otherwise.
func pluralEnding(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
}
