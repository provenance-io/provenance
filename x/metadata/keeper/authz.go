package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

// checkAuthZForMissing checks to see if the missing types.Party have an assigned grantee that can sing on their behalf
func (k Keeper) checkAuthzForMissing(
	ctx sdk.Context,
	addrs []string,
	signers []string,
	msg sdk.Msg,
) ([]string, error) {
	stillMissing := []string{}
	// return as a list this message type and its parent
	// type if it is a message belonging to a hierarchy
	msgTypeURLs := k.GetAuthzMessageTypeURLs(sdk.MsgTypeURL(msg))

	for _, addr := range addrs {
		found := false
		granter := types.MustAccAddressFromBech32(addr)

		// loop through all the signers
		for _, signer := range signers {
			grantee := types.MustAccAddressFromBech32(signer)

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
	signers []string,
	msg sdk.Msg,
) error {
	missing := FindMissing(existingOwners, signers)
	stillMissing := missing
	var err error
	// Authz grants rights to address on specific message types.
	// If no message is provided, skip the Authz check.
	if msg != nil {
		stillMissing, err = k.checkAuthzForMissing(ctx, missing, signers, msg)
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
func (k Keeper) ValidateAllPartiesAreSignersWithAuthz(ctx sdk.Context, parties []types.Party, signers []string, msg sdk.Msg) error {
	addresses := make([]string, len(parties))
	for i, party := range parties {
		addresses[i] = party.Address
	}

	missing := FindMissing(addresses, signers)
	stillMissing := missing
	var err error
	// Authz grants rights to address on specific message types.
	// If no message is provided, skip the Authz check.
	if msg != nil {
		stillMissing, err = k.checkAuthzForMissing(ctx, missing, signers, msg)
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

// pluralEnding returns "" if i == 1, or "s" otherwise.
func pluralEnding(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
}
