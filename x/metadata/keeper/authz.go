package keeper

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// PartyDetails is a struct used to help process party and signer validation.
// Even though all the fields are public, you should usually use the Get/Set methods
// which handle automatic bech32 conversion when needed and reduce duplicated efforts.
type PartyDetails struct {
	address  string
	role     types.PartyType
	optional bool

	acc       sdk.AccAddress
	signer    string
	signerAcc sdk.AccAddress

	canBeUsedBySpec bool
	usedBySpec      bool
}

// WrapRequiredParty creates a PartyDetails from the provided Party.
func WrapRequiredParty(party types.Party) *PartyDetails {
	return &PartyDetails{
		address:  party.Address,
		role:     party.Role,
		optional: party.Optional,
	}
}

// WrapAvailableParty creates a PartyDetails from the provided Party and marks it as usable.
func WrapAvailableParty(party types.Party) *PartyDetails {
	return &PartyDetails{
		address:         party.Address,
		role:            party.Role,
		optional:        true, // An available party is optional unless something else says otherwise.
		canBeUsedBySpec: true,
	}
}

// BuildPartyDetails creates the list of PartyDetails to be used in party/signer/role validation.
func BuildPartyDetails(reqParties, availableParties []types.Party) []*PartyDetails {
	details := make([]*PartyDetails, 0, len(availableParties))

	// Start with creating details for each available party.
availablePartiesLoop:
	for _, party := range availableParties {
		for _, known := range details {
			if party.IsSameAs(known) {
				continue availablePartiesLoop
			}
		}
		details = append(details, WrapAvailableParty(party))
	}

	// Now update the details to include optional=false required parties.
	// If an equal party is already in the details, just update its optional flag
	// to false, otherwise, add it to the list.
reqPartiesLoop:
	for _, reqParty := range reqParties {
		if !reqParty.Optional {
			for _, party := range details {
				if reqParty.IsSameAs(party) {
					party.MakeRequired()
					continue reqPartiesLoop
				}
			}
			details = append(details, WrapRequiredParty(reqParty))
		}
	}

	return details
}

func (p *PartyDetails) SetAddress(address string) {
	if p.address != address {
		p.acc = nil
	}
	p.address = address
}

func (p *PartyDetails) GetAddress() string {
	if len(p.address) == 0 && len(p.acc) > 0 {
		p.address = p.acc.String()
	}
	return p.address
}

func (p *PartyDetails) SetAcc(addr sdk.AccAddress) {
	if !bytes.Equal(p.acc, addr) {
		p.address = ""
	}
	p.acc = addr
}

func (p *PartyDetails) GetAcc() sdk.AccAddress {
	if len(p.acc) == 0 && len(p.address) > 0 {
		p.acc, _ = sdk.AccAddressFromBech32(p.address)
	}
	return p.acc
}

func (p *PartyDetails) SetRole(role types.PartyType) {
	p.role = role
}

func (p *PartyDetails) GetRole() types.PartyType {
	return p.role
}

func (p *PartyDetails) SetOptional(optional bool) {
	p.optional = optional
}

func (p *PartyDetails) MakeRequired() {
	p.optional = false
}

func (p *PartyDetails) GetOptional() bool {
	return p.optional
}

func (p *PartyDetails) IsRequired() bool {
	return !p.optional
}

func (p *PartyDetails) SetSigner(signer string) {
	if p.signer != signer {
		p.signerAcc = nil
	}
	p.signer = signer
}

func (p *PartyDetails) GetSigner() string {
	if len(p.signer) == 0 && len(p.signerAcc) > 0 {
		p.signer = p.signerAcc.String()
	}
	return p.signer
}

func (p *PartyDetails) SetSignerAcc(signerAddr sdk.AccAddress) {
	if !bytes.Equal(p.signerAcc, signerAddr) {
		p.signer = ""
	}
	p.signerAcc = signerAddr
}

func (p *PartyDetails) GetSignerAcc() sdk.AccAddress {
	if len(p.signerAcc) == 0 && len(p.signer) > 0 {
		p.signerAcc, _ = sdk.AccAddressFromBech32(p.signer)
	}
	return p.signerAcc
}

func (p *PartyDetails) HasSigner() bool {
	return len(p.signer) > 0 || len(p.signerAcc) > 0
}

func (p *PartyDetails) CanBeUsed() bool {
	return p.canBeUsedBySpec
}

func (p *PartyDetails) MarkAsUsed() {
	p.usedBySpec = true
}

func (p *PartyDetails) IsUsed() bool {
	return p.usedBySpec
}

// IsStillUsableAs returns true if this party can be use, hasn't yet been used and has the provided role.
func (p *PartyDetails) IsStillUsableAs(role types.PartyType) bool {
	return p.CanBeUsed() && !p.IsUsed() && p.GetRole() == role
}

// IsSameAs returns true if this is the same as the provided Party or PartyDetails.
// Only the address and role are considered for this test.
func (p *PartyDetails) IsSameAs(p2 types.Partier) bool {
	return types.SamePartiers(p, p2)
}

// SignersWrapper is a thing that holds the signers as strings and acc addresses.
// One is created by providing the strings. They are then converted to acc addresses
// if they're needed that way.
type SignersWrapper struct {
	signers    []string
	signerAccs []sdk.AccAddress
	converted  bool
}

func NewSignersWrapper(signers []string) *SignersWrapper {
	return &SignersWrapper{signers: signers}
}

// Strings gets the string versions of the signers.
func (s *SignersWrapper) Strings() []string {
	return s.signers
}

// Accs gets the sdk.AccAddress versions of the signers.
// Conversion happens if it hasn't already been done yet.
// Any strings that fail to convert are simply ignored.
func (s *SignersWrapper) Accs() []sdk.AccAddress {
	if !s.converted {
		s.signerAccs = safeBech32ToAccAddresses(s.signers)
		s.converted = true
	}
	return s.signerAccs
}

// ValidateSignersWithParties ensures the following:
//   - All optional=false reqParties have signed.
//   - All required roles are present in availableParties and are signers.
//   - All parties with the PROVENANCE role are a smart contract account.
//   - All parties with a smart contract account have the PROVENANCE role.
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
// Party details are returned containing information on which parties were signers and
// which were used to fulfill required roles.
//
// When parties and roles aren't involved, use ValidateSignersWithoutParties.
func (k Keeper) ValidateSignersWithParties(
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

	// Make sure all smart contract accounts have the PROVENANCE role,
	// and all parties with the PROVENANCE role have smart contract accounts.
	if err = k.validateProvenanceRole(ctx, parties); err != nil {
		return nil, err
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
	msgTypes := getAuthzMessageTypeURLs(sdk.MsgTypeURL(msg))
	for _, grantee := range grantees {
		for _, msgType := range msgTypes {
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
// * All parties with the address of a smart contract have the PROVENANCE role.
// * All parties with the PROVENANCE role have the address of a smart contract.
func (k Keeper) validateProvenanceRole(ctx sdk.Context, parties []*PartyDetails) error {
	for _, party := range parties {
		if party.CanBeUsed() {
			// Using the party address here (instead of the signer) because it's
			// that address that needs to be the smart contract.
			account := k.GetAccount(ctx, party.GetAcc())
			isWasmAcct := account != nil && account.GetSequence() == uint64(0) && account.GetPubKey() == nil
			isProvRole := party.role == types.PartyType_PARTY_TYPE_PROVENANCE
			if isWasmAcct && !isProvRole {
				return fmt.Errorf("account %q is a smart contract but does not have the PROVENANCE role", party.GetAddress())
			}
			if !isWasmAcct && isProvRole {
				return fmt.Errorf("account %q has role PROVENANCE but is not a smart contract", party.GetAddress())
			}
		}
	}
	return nil
}

// ValidateScopeValueOwnerUpdate verifies that it's okay for the msg signers to
// to change a scope's value owner from existing to proposed.
// If some parties have already been validated (possibly utilizaing authz), they
// can be provided in order to prevent an authorization from being used twice during
// a single Tx.
func (k Keeper) ValidateScopeValueOwnerUpdate(
	ctx sdk.Context,
	existing,
	proposed string,
	validatedParties []*PartyDetails,
	msg types.MetadataMsg,
) error {
	if existing == proposed {
		return nil
	}
	signers := msg.GetSignerStrs()
	if len(existing) > 0 {
		marker, hasAuth := k.GetMarkerAndCheckAuthority(ctx, existing, signers, markertypes.Access_Withdraw)
		if marker != nil {
			// If the existing is a marker, make sure a signer has withdraw authority on it.
			if !hasAuth {
				return fmt.Errorf("missing signature for %s (%s) with authority to withdraw/remove it as scope value owner", existing, marker.GetDenom())
			}
		} else {
			// If the existing isn't a marker, make sure they're one of the signers or
			// have an authorization grant for one of the signers.
			found := false

			// First just check the list of signers.
			for _, signer := range signers {
				if existing == signer {
					found = true
					break
				}
			}

			// If not yet found, check the validated parties for one with this address
			// that also has an associated signer.
			// This way, if it does exist, the authorization is only used once during a single call.
			if !found {
				for _, party := range validatedParties {
					if party.GetAddress() == existing && party.HasSigner() {
						found = true
						break
					}
				}
			}

			// If still not found, directly check with authz for help.
			if !found {
				// If existing isn't a bech32, we just skip the authz check. Should only happen in unit tests.
				granter, err := sdk.AccAddressFromBech32(existing)
				if err == nil {
					grantees := safeBech32ToAccAddresses(signers)
					grantee, err := k.findAuthzGrantee(ctx, granter, grantees, msg)
					if err != nil {
						return fmt.Errorf("authz error with existing value owner %q: %w", existing, err)
					}
					if len(grantee) > 0 {
						found = true
					}
				}
			}

			if !found {
				return fmt.Errorf("missing signature from existing value owner %s", existing)
			}
		}
	}

	if len(proposed) > 0 {
		// If the proposed is a marker, make sure a signer has deposit authority on it.
		marker, hasAuth := k.GetMarkerAndCheckAuthority(ctx, proposed, signers, markertypes.Access_Deposit)
		if marker != nil && !hasAuth {
			return fmt.Errorf("missing signature for %s (%s) with authority to deposit/add it as scope value owner", proposed, marker.GetDenom())
		}
		// If it's not a marker, we don't really care what it's being set to.
	}

	return nil
}

// TODELETEcheckAuthzForMissing returns any of the provided addrs that have not been granted an authz authorization by one of the msg signers.
// An error is returned if there was a problem updating an authorization.
// This is replaced by findAuthzGrantee.
// It hasn't been deleted yet because I wanted test cases for the new func.
// TODO[1438]: Delete TODELETEcheckAuthzForMissing
func (k Keeper) TODELETEcheckAuthzForMissing(
	ctx sdk.Context,
	addrs []string,
	msg types.MetadataMsg,
) ([]string, error) {
	stillMissing := []string{}
	if len(addrs) == 0 {
		return stillMissing, nil
	}

	signers := msg.GetSignerStrs()
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
	msgTypeURLs := getAuthzMessageTypeURLs(sdk.MsgTypeURL(msg))

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

// ValidateSignersWithoutParties makes sure that each entry in the required list are either signers of the msg,
// or have granted an authz authorization to one of the signers.
//
// Party details are returned containing information on which addresses had signers.
// All roles in these details are UNSPECIFIED.
//
// When parties (and/or roles) are involved, use ValidateSignersWithParties.
func (k Keeper) ValidateSignersWithoutParties(
	ctx sdk.Context,
	required []string,
	msg types.MetadataMsg,
) ([]*PartyDetails, error) {
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

// TODELETEValidateAllPartiesAreSignersWithAuthz validate all parties are signers with authz module
// This is replaced by ValidateSignersWithParties.
// It hasn't been deleted yet because I wanted the test cases for the new func.
// TODO[1438]: Delete TODELETEValidateAllPartiesAreSignersWithAuthz
func (k Keeper) TODELETEValidateAllPartiesAreSignersWithAuthz(ctx sdk.Context, parties []types.Party, msg types.MetadataMsg) error {
	addresses := make([]string, len(parties))
	for i, party := range parties {
		addresses[i] = party.Address
	}
	signers := msg.GetSignerStrs()
	missing := findMissing(addresses, signers)
	stillMissing := missing
	var err error
	// Authz grants rights to address on specific message types.
	// If no message is provided, skip the Authz check.
	if msg != nil {
		stillMissing, err = k.TODELETEcheckAuthzForMissing(ctx, missing, msg)
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

// findMissing returns all elements of the required list that are not found in the entries list.
//
// See also: findMissingComp
func findMissing(required, toCheck []string) []string {
	return findMissingComp(required, toCheck, func(r, c string) bool { return r == c })
}

// findMissingParties returns all parties from the required list that don't have a same party in the toCheck list.
//
// See also: findMissingComp
func findMissingParties(required, toCheck []types.Party) []types.Party {
	return findMissingComp(required, toCheck, func(r, c types.Party) bool { return types.SamePartiers(&r, &c) })
}

// findMissingComp returns all entries in required where an entry does not exist in toCheck
// such that the provided comp function returns true.
// Duplicate entries in required do not require duplicate entries in toCheck.
// E.g. findMissingComp([a, b, a], [a]) => [b], and findMissingComp([a, b, a], [b]) => [a, a].
func findMissingComp[R any, C any](required []R, toCheck []C, comp func(R, C) bool) []R {
	var rv []R
reqLoop:
	for _, req := range required {
		for _, entry := range toCheck {
			if comp(req, entry) {
				continue reqLoop
			}
		}
		rv = append(rv, req)
	}
	return rv
}

// GetMarkerAndCheckAuthority gets a marker by address and checks if one of the signers has the provided role.
// If the address isn't a marker, nil, false is returned.
func (k Keeper) GetMarkerAndCheckAuthority(
	ctx sdk.Context,
	address string,
	signers []string,
	role markertypes.Access,
) (markertypes.MarkerAccountI, bool) {
	addr, err := sdk.AccAddressFromBech32(address)
	// if the address is invalid then it is not possible for it to be a marker.
	if err != nil {
		return nil, false
	}

	acc := k.authKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return nil, false
	}

	// Convert over to the actual underlying marker type, or not.
	marker, isMarker := acc.(*markertypes.MarkerAccount)
	if !isMarker {
		return nil, false
	}

	// Check if any of the signers have the desired role.
	for _, signer := range signers {
		if marker.HasAccess(signer, role) {
			return marker, true
		}
	}

	return marker, false
}

// pluralEnding returns "" if i == 1, or "s" otherwise.
func pluralEnding(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
}

// safeBech32ToAccAddresses attempts to convert all provided strings to AccAddresses.
// Any that fail to convert are ignored.
func safeBech32ToAccAddresses(bech32s []string) []sdk.AccAddress {
	rv := make([]sdk.AccAddress, 0, len(bech32s))
	for _, bech32 := range bech32s {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err == nil {
			rv = append(rv, addr)
		}
	}
	return rv
}
