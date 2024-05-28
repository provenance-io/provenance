package params

// Default simulation operation weights for messages and gov proposals
const (
	// Name
	DefaultWeightMsgBindName            int = 15
	DefaultWeightMsgDeleteName          int = 10
	DefaultWeightMsgModifyName          int = 5
	DefaultWeightCreateRootNameProposal int = 10
	// Attribute
	DefaultWeightMsgAddAttribute            int = 15
	DefaultWeightMsgUpdateAttribute         int = 5
	DefaultWeightMsgDeleteAttribute         int = 5
	DefaultWeightMsgDeleteDistinctAttribute int = 5
	DefaultWeightMsgSetAccountData          int = 10
	// Marker
	DefaultWeightMsgAddMarker                 int = 30
	DefaultWeightMsgChangeStatus              int = 10
	DefaultWeightMsgFinalize                  int = 10
	DefaultWeightMsgAddAccess                 int = 10
	DefaultWeightMsgAddFinalizeActivateMarker int = 10
	DefaultWeightMsgAddMarkerProposal         int = 40
	DefaultWeightMsgUpdateDenySendList        int = 10
	// Trigger
	DefaultWeightSubmitCreateTrigger  int = 95
	DefaultWeightSubmitDestroyTrigger int = 5
	// Oracle
	DefaultWeightUpdateOracle    int = 25
	DefaultWeightSendOracleQuery int = 75
	// Ibc Rate Limiter
	DefaultWeightIBCRLUpdateParams int = 100
)
