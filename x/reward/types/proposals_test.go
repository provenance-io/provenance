package types

// import (
// 	"errors"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// 	"github.com/tendermint/tendermint/crypto/secp256k1"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// )

// type ProposalTypesTestSuite struct {
// 	suite.Suite
// 	accountAddr sdk.AccAddress
// }

// func TestProposalTypesTestSuite(t *testing.T) {
// 	suite.Run(t, new(ProposalTypesTestSuite))
// }

// func (s *ProposalTypesTestSuite) SetupSuite() {
// 	s.T().Parallel()
// 	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
// }

// func (s *ProposalTypesTestSuite) TestAddProposalValidateBasic() {
// 	tests := []struct {
// 		name     string
// 		proposal AddRewardProgramProposal
// 		err      error
// 	}{
// 		{
// 			"add reward proposal - reward program id is invalid",
// 			*NewAddRewardProgramProposal("title", "description",
// 				0,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(10)),
// 				"day",
// 				1,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				1,
// 				2,
// 			),
// 			errors.New("reward program id is invalid"),
// 		},
// 		{
// 			"add reward proposal - reward program epoch start offset is invalid",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(10)),
// 				"day",
// 				0,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				1,
// 				2,
// 			),
// 			errors.New("reward program epoch start offset is invalid"),
// 		},
// 		{
// 			"add reward proposal - invalid address for rewards program distribution from address",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				"invalid",
// 				sdk.NewCoin("nhash", sdk.NewInt(10)),
// 				"day",
// 				1,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				1,
// 				2,
// 			),
// 			errors.New("invalid address for rewards program distribution from address: decoding bech32 failed: invalid bech32 string length 7"),
// 		},
// 		{
// 			"add reward proposal - epoch id cannot be empty",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(10)),
// 				"",
// 				1,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				1,
// 				2,
// 			),
// 			errors.New("epoch id cannot be empty"),
// 		},
// 		{
// 			"add reward proposal - eligibility criteria is not valid",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(10)),
// 				"day",
// 				1,
// 				100,
// 				NewEligibilityCriteria("", &ActionDelegate{}),
// 				1,
// 				2,
// 			),
// 			errors.New("eligibility criteria is not valid: eligibility criteria must have a name"),
// 		},
// 		{
// 			"add reward proposal - reward program requires coins",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(0)),
// 				"day",
// 				1,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				1,
// 				2,
// 			),
// 			errors.New("reward program requires coins: 0nhash"),
// 		},
// 		{
// 			"add reward proposal - maximum must be larger than 0",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(1)),
// 				"day",
// 				1,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				1,
// 				0,
// 			),
// 			errors.New("maximum must be larger than 0"),
// 		},
// 		{
// 			"add reward proposal - minimum cannot be larger than the maximum",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(1)),
// 				"day",
// 				1,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				10,
// 				1,
// 			),
// 			errors.New("minimum (10) cannot be larger than the maximum (1)"),
// 		},
// 		{
// 			"add reward proposal - successful",
// 			*NewAddRewardProgramProposal("title", "description",
// 				1,
// 				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
// 				sdk.NewCoin("nhash", sdk.NewInt(1)),
// 				"day",
// 				1,
// 				100,
// 				NewEligibilityCriteria("delegation", &ActionDelegate{}),
// 				10,
// 				100,
// 			),
// 			nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		tt := tt
// 		s.T().Run(tt.name, func(t *testing.T) {
// 			err := tt.proposal.ValidateBasic()
// 			if tt.err != nil {
// 				require.Error(t, err)
// 				require.Equal(t, tt.err.Error(), err.Error())
// 			} else {
// 				require.NoError(t, err)
// 			}
// 		})
// 	}
// }
