package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	provenance "github.com/provenance-io/provenance/app"
)

type IntegrationTestSuite struct {
	suite.Suite

	app *provenance.App
	ctx sdk.Context
	k   namekeeper.Keeper

	accountAddr sdk.AccAddress
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.app = provenance.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.k = namekeeper.NewKeeper(s.app.AppCodec(), s.app.GetKey(nametypes.ModuleName), s.app.GetSubspace(nametypes.ModuleName))
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestNameProposals() {

	testCases := []struct {
		name    string
		prop    govtypes.Content
		wantErr bool
		err     error
	}{
		// ADD ROOT NAME PROPOSALS
		{
			"add marker - valid",
			nametypes.NewCreateRootNameProposal("title", "description", "root", s.accountAddr, false),
			false,
			nil,
		},
		{
			"add marker - valid full domain",
			nametypes.NewCreateRootNameProposal("title", "description", "example.provenance.io", s.accountAddr, false),
			false,
			nil,
		},
		{
			"add marker - valid new sub domain",
			nametypes.NewCreateRootNameProposal("title", "description", "another.provenance.io", s.accountAddr, false),
			false,
			nil,
		},
		{
			"add marker - invalid address",
			&nametypes.CreateRootNameProposal{Title: "title", Description: "description", Name: "badroot", Owner: "bad1address", Restricted: false},
			true,
			fmt.Errorf("decoding bech32 failed: invalid checksum (expected dpg8tu got ddress)"),
		},
		{
			"add marker - fails duplicate",
			nametypes.NewCreateRootNameProposal("title", "description", "root", s.accountAddr, false),
			true,
			fmt.Errorf("name is already bound to an address"),
		},
		{
			"add marker - fails duplicate sub domain",
			nametypes.NewCreateRootNameProposal("title", "description", "provenance.io", s.accountAddr, false),
			true,
			fmt.Errorf("name is already bound to an address"),
		},
		{
			"add marker - fails duplicate third level domain",
			nametypes.NewCreateRootNameProposal("title", "description", "example.provenance.io", s.accountAddr, false),
			true,
			fmt.Errorf("name is already bound to an address"),
		},
		{
			"add marker - fails another duplicate third level domain",
			nametypes.NewCreateRootNameProposal("title", "description", "another.provenance.io", s.accountAddr, false),
			true,
			fmt.Errorf("name is already bound to an address"),
		},
		{
			"add marker - fails invalid name",
			nametypes.NewCreateRootNameProposal("title", "description", "..badroot", s.accountAddr, false),
			true,
			fmt.Errorf("segment of name is too short"),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {

			var err error
			switch c := tc.prop.(type) {
			case *nametypes.CreateRootNameProposal:
				err = namekeeper.HandleCreateRootNameProposal(s.ctx, s.k, c)
			default:
				panic("invalid proposal type")
			}

			if tc.wantErr {
				s.Require().Error(err)
				s.Require().Equal(tc.err.Error(), err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
