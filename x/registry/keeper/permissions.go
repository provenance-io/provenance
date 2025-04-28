package keeper

type RegistryPermission struct {
	CanUpdateServicer    bool
	CanUpdateSubservicer bool

	CanUpdateENoteController bool
	CanUpdateSecuredParty    bool

	CanPerformNFTUpdates bool

	CanRemoveFromRegistry        string
	DefaultUpdateOnLoanOnboarded string
	EncumberedWarehouseOutcome   string
	SoldToInvestorOutcome        string
}

type PermissionMap map[string]RegistryPermission

var P = PermissionMap{
	"REGISTRY_ROLE_SERVICER": {
		CanUpdateServicer:        true,
		CanUpdateSubservicer:     true,
		CanUpdateENoteController: false,
	},
	"REGISTRY_ROLE_SUBSERVICER": {
		CanUpdateServicer:        false,
		CanUpdateSubservicer:     true,
		CanUpdateENoteController: false,
	},

	"REGISTRY_ROLE_CONTROLLER": {
		CanUpdateServicer:        false,
		CanUpdateSubservicer:     false,
		CanUpdateENoteController: true,
	},

	"REGISTRY_ROLE_CUSTODIAN": {
		CanUpdateServicer:        true,
		CanUpdateSubservicer:     true,
		CanUpdateENoteController: true,
	},
	"REGISTRY_ROLE_BORROWER": {
		CanUpdateServicer:        false,
		CanUpdateSubservicer:     false,
		CanUpdateENoteController: false,
	},
	"REGISTRY_ROLE_ORIGINATOR": {
		CanUpdateServicer:        true,
		CanUpdateSubservicer:     true,
		CanUpdateENoteController: true,
	},
}
