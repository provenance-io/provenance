package wasm

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WasmTestSuite struct {
	suite.Suite

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	pubkey3   cryptotypes.PubKey
	user3     string
	user3Addr sdk.AccAddress

	scopeID types.MetadataAddress
	specID  types.MetadataAddress
}

func (s *WasmTestSuite) SetupTest() {
	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.pubkey3 = secp256k1.GenPrivKey().PubKey()
	s.user3Addr = sdk.AccAddress(s.pubkey3.Address())
	s.user3 = s.user3Addr.String()

	s.specID = types.ScopeSpecMetadataAddress(uuid.New())
	s.scopeID = types.ScopeMetadataAddress(uuid.New())
}

func TestWasmTestSuite(t *testing.T) {
	suite.Run(t, new(WasmTestSuite))
}

func (s *WasmTestSuite) TestEncode() {

	cases := map[string]struct {
		contract sdk.AccAddress
		messages []sdk.Msg
		scope    WriteScope
		wantErr  bool
		errorMsg string
	}{
		"valid, 1 address and is contract address": {
			contract: s.user1Addr,
			messages: []sdk.Msg{
				types.NewMsgWriteScopeRequest(types.Scope{
					ScopeId:         s.scopeID,
					SpecificationId: s.specID,
					Owners: []types.Party{{
						Address: s.user1,
						Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
					}},
					DataAccess:        []string{s.user1},
					ValueOwnerAddress: s.user1,
				}, []string{s.user1},
				),
			},
			scope: WriteScope{
				Scope: Scope{
					ScopeID:         s.scopeID.String(),
					SpecificationID: s.specID.String(),
					DataAccess:      []string{s.user1},
					Owners: []*Party{{
						Address: s.user1,
						Role:    "originator",
					}},
					ValueOwnerAddress: s.user1,
				},
				Signers: []string{s.user1},
			},
			wantErr:  false,
			errorMsg: "",
		},
		"valid, multi address and has contract address": {
			contract: s.user1Addr,
			messages: []sdk.Msg{
				types.NewMsgWriteScopeRequest(types.Scope{
					ScopeId:         s.scopeID,
					SpecificationId: s.specID,
					Owners: []types.Party{{
						Address: s.user1,
						Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
					}},
					DataAccess:        []string{s.user1},
					ValueOwnerAddress: s.user1,
				}, []string{s.user1, s.user2},
				),
			},
			scope: WriteScope{
				Scope: Scope{
					ScopeID:         s.scopeID.String(),
					SpecificationID: s.specID.String(),
					DataAccess:      []string{s.user1},
					Owners: []*Party{{
						Address: s.user1,
						Role:    "originator",
					}},
					ValueOwnerAddress: s.user1,
				},
				Signers: []string{s.user2, s.user1},
			},
			wantErr:  false,
			errorMsg: "",
		},
		"valid, 1 address and is not contract address": {
			contract: s.user2Addr,
			messages: []sdk.Msg{
				types.NewMsgWriteScopeRequest(types.Scope{
					ScopeId:         s.scopeID,
					SpecificationId: s.specID,
					Owners: []types.Party{{
						Address: s.user1,
						Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
					}},
					DataAccess:        []string{s.user1},
					ValueOwnerAddress: s.user1,
				}, []string{s.user2, s.user1},
				),
			},
			scope: WriteScope{
				Scope: Scope{
					ScopeID:         s.scopeID.String(),
					SpecificationID: s.specID.String(),
					DataAccess:      []string{s.user1},
					Owners: []*Party{{
						Address: s.user1,
						Role:    "originator",
					}},
					ValueOwnerAddress: s.user1,
				},
				Signers: []string{s.user1},
			},
			wantErr:  false,
			errorMsg: "",
		},
		"valid, multi address and does not have contract address": {
			contract: s.user3Addr,
			messages: []sdk.Msg{
				types.NewMsgWriteScopeRequest(types.Scope{
					ScopeId:         s.scopeID,
					SpecificationId: s.specID,
					Owners: []types.Party{{
						Address: s.user1,
						Role:    types.PartyType_PARTY_TYPE_ORIGINATOR,
					}},
					DataAccess:        []string{s.user1},
					ValueOwnerAddress: s.user1,
				}, []string{s.user3, s.user2, s.user1},
				),
			},
			scope: WriteScope{
				Scope: Scope{
					ScopeID:         s.scopeID.String(),
					SpecificationID: s.specID.String(),
					DataAccess:      []string{s.user1},
					Owners: []*Party{{
						Address: s.user1,
						Role:    "originator",
					}},
					ValueOwnerAddress: s.user1,
				},
				Signers: []string{s.user2, s.user1},
			},
			wantErr:  false,
			errorMsg: "",
		},
		"invalid, bad scope": {
			contract: s.user3Addr,
			messages: []sdk.Msg{},
			scope: WriteScope{
				Scope: Scope{
					ScopeID:         "bad_id",
					SpecificationID: s.specID.String(),
					DataAccess:      []string{s.user1},
					Owners: []*Party{{
						Address: s.user1,
						Role:    "originator",
					}},
					ValueOwnerAddress: s.user1,
				},
				Signers: []string{s.user3, s.user1},
			},
			wantErr:  true,
			errorMsg: "wasm: invalid 'scope id': decoding bech32 failed: invalid bech32 string length 6",
		},
		"invalid, bad signer": {
			contract: s.user1Addr,
			messages: []sdk.Msg{},
			scope: WriteScope{
				Scope: Scope{
					ScopeID:         s.scopeID.String(),
					SpecificationID: s.specID.String(),
					DataAccess:      []string{s.user1},
					Owners: []*Party{{
						Address: s.user1,
						Role:    "originator",
					}},
					ValueOwnerAddress: s.user1,
				},
				Signers: []string{"bad_signer"},
			},
			wantErr:  true,
			errorMsg: "wasm: signer address must be a Bech32 string: decoding bech32 failed: invalid separator index -1",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.T().Run(n, func(t *testing.T) {
			msgs, err := tc.scope.Encode(tc.contract)
			if tc.wantErr {
				assert.EqualError(t, err, tc.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.messages, msgs)
			}
		})
	}
}
