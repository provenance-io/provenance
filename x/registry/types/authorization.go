package types

// DefaultRoleAuthorizations returns the static role authorization policies.
// These policies encode who must sign to change each participant role.
// In a future phase, these will be replaced by per-RegistryClass dynamic policies.
func DefaultRoleAuthorizations() []RoleAuthorization {
	return []RoleAuthorization{
		controllerRoleAuthorization(),
	}
}

// RoleAuthorizationMap returns a map of RegistryRole → RoleAuthorization for fast lookup.
func RoleAuthorizationMap() map[RegistryRole]RoleAuthorization {
	auths := DefaultRoleAuthorizations()
	m := make(map[RegistryRole]RoleAuthorization, len(auths))
	for _, a := range auths {
		m[a.Role] = a
	}
	return m
}

// controllerRoleAuthorization defines the CONTROLLER role authorization policy.
//
// A Controller update requires all of the following to sign:
//   - the current controller (or the NFT owner if no controller is set);
//   - the current Secured Party for eNote, if one is set;
//   - the incoming new controller.
//
// The current controller's signature is always required: there is no path that lets any other
// party assume control unilaterally.
func controllerRoleAuthorization() RoleAuthorization {
	return RoleAuthorization{
		Role: RegistryRole_REGISTRY_ROLE_CONTROLLER,
		Authorizations: []*Authorization{
			{
				Description: "Transfer requiring current controller approval",
				Signatures: []*SignatureRequirement{
					{
						// The current authority (controller if set, otherwise NFT owner) must sign.
						Type: SignatureType_SIGNATURE_TYPE_REQUIRED_ALL,
						Roles: []*RoleAssignment{
							{
								RoleSelector: &RoleAssignment_RolePriority{
									RolePriority: &RolePriority{
										Entries: []*RolePriorityEntry{
											{Role: &RolePriorityEntry_RegistryRole{RegistryRole: RegistryRole_REGISTRY_ROLE_CONTROLLER}},
											{Role: &RolePriorityEntry_NftRole{NftRole: NftRole_NFT_ROLE_NFT_OWNER}},
										},
									},
								},
								Assignment: Assignment_ASSIGNMENT_CURRENT,
							},
						},
					},
					{
						// The current Secured Party for eNote must sign (if set).
						// The new/incoming controller must sign (if being assigned).
						Type: SignatureType_SIGNATURE_TYPE_REQUIRED_ALL_IF_SET,
						Roles: []*RoleAssignment{
							{
								RoleSelector: &RoleAssignment_RegistryRole{RegistryRole: RegistryRole_REGISTRY_ROLE_SECURED_PARTY_FOR_ENOTE},
								Assignment:   Assignment_ASSIGNMENT_CURRENT,
							},
							{
								RoleSelector: &RoleAssignment_RegistryRole{RegistryRole: RegistryRole_REGISTRY_ROLE_CONTROLLER},
								Assignment:   Assignment_ASSIGNMENT_NEW,
							},
						},
					},
				},
			},
		},
	}
}
