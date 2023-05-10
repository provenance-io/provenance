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
	// Marker
	DefaultWeightSupplyIncreaseProposalContent      int = 5
	DefaultWeightSupplyDecreaseProposalContent      int = 5
	DefaultWeightSetAdministratorProposalContent    int = 5
	DefaultWeightRemoveAdministratorProposalContent int = 5
	DefaultWeightChangeStatusProposalContent        int = 5
	DefaultWeightSetDenomMetadataProposalContent    int = 5
	// Adjusted marker operations to a cumulative weight of 100
	DefaultWeightMsgAddMarker                 int = 30
	DefaultWeightMsgChangeStatus              int = 10
	DefaultWeightMsgAddAccess                 int = 10
	DefaultWeightMsgAddFinalizeActivateMarker int = 10
	DefaultWeightMsgAddMarkerProposal         int = 40
	// MsgFees
	DefaultWeightAddMsgFeeProposalContent    int = 75
	DefaultWeightRemoveMsgFeeProposalContent int = 25
	// Rewards
	DefaultWeightSubmitCreateRewards int = 95
	DefaultWeightSubmitEndRewards    int = 5
)
