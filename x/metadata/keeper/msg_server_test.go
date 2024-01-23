package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"

	simapp "github.com/provenance-io/provenance/app"
	metadatakeeper "github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type MsgServerTestSuite struct {
	suite.Suite

	app            *simapp.App
	ctx            sdk.Context
	msgServer      types.MsgServer
	blockStartTime time.Time

	privkey1   cryptotypes.PrivKey
	pubkey1    cryptotypes.PubKey
	owner1     string
	owner1Addr sdk.AccAddress
	acct1      authtypes.AccountI

	addresses []sdk.AccAddress
}

func (s *MsgServerTestSuite) SetupTest() {

	s.blockStartTime = time.Now()
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		Time: s.blockStartTime,
	})
	s.msgServer = metadatakeeper.NewMsgServerImpl(s.app.MetadataKeeper)

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
}
func TestMsgServerTestSuite(t *testing.T) {
	suite.Run(t, new(MsgServerTestSuite))
}

func (s *MsgServerTestSuite) TestAddNetAssetValue() {
	scopeSpecUUIDNF := uuid.New()
	scopeSpecIDNF := types.ScopeSpecMetadataAddress(scopeSpecUUIDNF)

	scopeUUID := uuid.New()
	scopeID := types.ScopeMetadataAddress(scopeUUID)
	scopeSpecUUID := uuid.New()
	scopeSpecID := types.ScopeSpecMetadataAddress(scopeSpecUUID)
	pubkey1 := secp256k1.GenPrivKey().PubKey()
	user1Addr := sdk.AccAddress(pubkey1.Address())
	user1 := user1Addr.String()
	// pubkey2 := secp256k1.GenPrivKey().PubKey()
	// user2Addr := sdk.AccAddress(pubkey2.Address())
	// user2 := user2Addr.String()

	ns := *types.NewScope(scopeID, scopeSpecID, ownerPartyList(user1), []string{user1}, user1, false)

	s.app.MetadataKeeper.SetScope(s.ctx, ns)

	testCases := []struct {
		name   string
		msg    types.MsgAddNetAssetValuesRequest
		expErr string
	}{
		{
			name: "no marker found",
			msg: types.MsgAddNetAssetValuesRequest{
				ScopeId: scopeSpecIDNF.String(),
				NetAssetValues: []types.NetAssetValue{
					{
						Price:  sdk.NewInt64Coin("navcoin", 1),
						Volume: 1,
					}},
				Signers: []string{user1},
			},
			expErr: fmt.Sprintf("scope not found: %v: not found", scopeSpecIDNF.String()),
		},
		{
			name: "value denom does not exist",
			msg: types.MsgAddNetAssetValuesRequest{
				ScopeId: scopeID.String(),
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin("hotdog", 100),
						Volume:             uint64(100),
						UpdatedBlockHeight: 1,
					},
				},
				Signers: []string{user1},
			},
			expErr: `net asset value denom does not exist: marker hotdog not found for address: cosmos1p6l3annxy35gm5mfm6m0jz2mdj8peheuzf9alh: invalid request`,
		},
		// {
		// 	name: "not authorize user",
		// 	msg: types.MsgAddNetAssetValuesRequest{
		// 		ScopeId: scopeID.String(),
		// 		NetAssetValues: []types.NetAssetValue{
		// 			{
		// 				Price:              sdk.NewInt64Coin(types.UsdDenom, 100),
		// 				Volume:             uint64(100),
		// 				UpdatedBlockHeight: 1,
		// 			},
		// 		},
		// 		Signers: []string{user2},
		// 	},
		// 	expErr: fmt.Sprintf(`signer %v does not have permission to add net asset value for "%v"`, user2, scopeID.String()),
		// },
		{
			name: "successfully set nav",
			msg: types.MsgAddNetAssetValuesRequest{
				ScopeId: scopeID.String(),
				NetAssetValues: []types.NetAssetValue{
					{
						Price:              sdk.NewInt64Coin(types.UsdDenom, 100),
						Volume:             uint64(100),
						UpdatedBlockHeight: 1,
					},
				},
				Signers: []string{user1},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res, err := s.msgServer.AddNetAssetValues(sdk.WrapSDKContext(s.ctx),
				&tc.msg)

			if len(tc.expErr) > 0 {
				s.Assert().Nil(res)
				s.Assert().EqualError(err, tc.expErr)

			} else {
				s.Assert().NoError(err)
				s.Assert().Equal(res, &types.MsgAddNetAssetValuesResponse{})
			}
		})
	}
}
