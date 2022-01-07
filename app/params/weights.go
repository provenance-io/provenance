package params

// Default simulation operation weights for messages and gov proposals
const (
	DefaultWeightMsgSend                        int = 100
	DefaultWeightMsgMultiSend                   int = 10
	DefaultWeightMsgSetWithdrawAddress          int = 50
	DefaultWeightMsgWithdrawDelegationReward    int = 50
	DefaultWeightMsgWithdrawValidatorCommission int = 50
	DefaultWeightMsgFundCommunityPool           int = 50
	DefaultWeightMsgDeposit                     int = 100
	DefaultWeightMsgVote                        int = 67
	DefaultWeightMsgUnjail                      int = 100
	DefaultWeightMsgCreateValidator             int = 100
	DefaultWeightMsgEditValidator               int = 5
	DefaultWeightMsgDelegate                    int = 100
	DefaultWeightMsgUndelegate                  int = 100
	DefaultWeightMsgBeginRedelegate             int = 100
	DefaultWeightCommunitySpendProposal         int = 5
	DefaultWeightTextProposal                   int = 5
	DefaultWeightParamChangeProposal            int = 5
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
	DefaultWeightMsgMintMarker                      int = 67
	DefaultWeightMsgBurnMarker                      int = 67
	// MsgFees
	DefaultWeightAddMsgFeeProposalContent    int = 75
	DefaultWeightRemoveMsgFeeProposalContent int = 25
)
