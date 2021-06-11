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

	DefaultWeightCommunitySpendProposal int = 5
	DefaultWeightTextProposal           int = 5
	DefaultWeightParamChangeProposal    int = 5

	DefaultWeightMsgBindName            int = 10
	DefaultWeightMsgDeleteName          int = 5
	DefaultWeightCreateRootNameProposal int = 5

	DefaultWeightMsgAddAttribute            int = 15
	DefaultWeightMsgUpdateAttribute         int = 5
	DefaultWeightMsgDeleteAttribute         int = 5
	DefaultWeightMsgDeleteDistinctAttribute int = 5
)
