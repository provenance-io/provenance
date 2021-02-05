package types

import (
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	pubHex, _ = hex.DecodeString("85EA54E8598B27EC37EAEEEEA44F1E78A9B5E671")
	addr      = sdk.AccAddress(pubHex)
)

type scopeTestSuite struct {
	suite.Suite
}

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(scopeTestSuite))
}

func (s *scopeTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *scopeTestSuite) TestScopeValidateBasic() {
	tests := []struct {
		name    string
		scope   *Scope
		want    string
		wantErr bool
	}{
		{
			"valid scope one owner",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []string{addr.String()}, []string{}, ""),
			"",
			false,
		},
		{
			"valid scope one owner, one data access",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []string{addr.String()}, []string{addr.String()}, ""),
			"",
			false,
		},
		{
			"no owners",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []string{}, []string{}, ""),
			"scope must have at least one owner",
			true,
		},
		{
			"no owners, data accesss",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []string{}, []string{addr.String()}, ""),
			"scope must have at least one owner",
			true,
		},
		{
			"invalid scope id",
			NewScope(ScopeSpecMetadataAddress(uuid.New()), ScopeSpecMetadataAddress(uuid.New()), []string{}, []string{}, ""),
			"invalid scope identifier (expected: scope, got scopespec)",
			true,
		},
		{
			"invalid scope id - wrong address type",
			NewScope(MetadataAddress(addr), ScopeSpecMetadataAddress(uuid.New()), []string{}, []string{}, ""),
			"invalid metadata address type (must be 0-4, actual: 133)",
			true,
		},
		{
			"invalid spec id",
			NewScope(ScopeMetadataAddress(uuid.New()), ScopeMetadataAddress(uuid.New()), []string{}, []string{}, ""),
			"invalid scope specification identifier (expected: scopespec, got scope)",
			true,
		},
		{
			"invalid spec id - wrong address type",
			NewScope(ScopeMetadataAddress(uuid.New()), MetadataAddress(addr), []string{}, []string{}, ""),
			"invalid metadata address type (must be 0-4, actual: 133)",
			true,
		},
		{
			"invaid owner on scope",
			NewScope(ScopeMetadataAddress(
				uuid.New()),
				ScopeSpecMetadataAddress(uuid.New()),
				[]string{":invalid"},
				[]string{},
				"",
			),
			"invalid owner on scope: decoding bech32 failed: invalid index of 1",
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.scope.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("Scope ValidateBasic error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				require.Equal(t, tt.want, err.Error())
			}

		})
	}
}

func (s *scopeTestSuite) TestScopeString() {
	s.T().Run("scope string", func(t *testing.T) {
		scopeUuid = uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")
		groupUuid = uuid.MustParse("c25c7bd4-c639-4367-a842-f64fa5fccc19")
		scope := NewScope(ScopeMetadataAddress(
			scopeUuid), ScopeSpecMetadataAddress(groupUuid),
			[]string{addr.String()},
			[]string{},
			"")
		require.Equal(t, `scope_id: scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp
specification_id: scopespec1qnp9c775ccu5xeaggtmylf0uesvsqyrkq8
owneraddress:
- cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck
dataaccess: []
valueowneraddress: ""
`,
			scope.String())
	})
}
