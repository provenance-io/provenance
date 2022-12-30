package params

// Default simulation operation weights for messages and gov proposals
const (
	// Name
	DefaultWeightMsgBindName            int = 10
	DefaultWeightMsgDeleteName          int = 5
	DefaultWeightCreateRootNameProposal int = 5
	// Attribute
	DefaultWeightMsgAddAttribute            int = 15
	DefaultWeightMsgUpdateAttribute         int = 5
	DefaultWeightMsgDeleteAttribute         int = 5
	DefaultWeightMsgDeleteDistinctAttribute int = 5
	// Marker
	DefaultWeightAddMarkerProposalContent           int = 5
	DefaultWeightSupplyIncreaseProposalContent      int = 5
	DefaultWeightSupplyDecreaseProposalContent      int = 5
	DefaultWeightSetAdministratorProposalContent    int = 5
	DefaultWeightRemoveAdministratorProposalContent int = 5
	DefaultWeightChangeStatusProposalContent        int = 5
	DefaultWeightSetDenomMetadataProposalContent    int = 5
	DefaultWeightMsgAddMarker                       int = 100
	DefaultWeightMsgChangeStatus                    int = 10
	DefaultWeightMsgAddAccess                       int = 10
	DefaultWeightMsgAddFinalizeActivateMarker       int = 67
	DefaultWeightMsgMintMarker                      int = 67
	DefaultWeightMsgBurnMarker                      int = 67
	// MsgFees
	DefaultWeightAddMsgFeeProposalContent    int = 75
	DefaultWeightRemoveMsgFeeProposalContent int = 25
	// Rewards
	DefaultWeightSubmitCreateRewards int = 95
	DefaultWeightSubmitEndRewards    int = 5
)
