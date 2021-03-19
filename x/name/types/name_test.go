package types

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type NameRecordTestSuite struct {
	addr sdk.AccAddress
	suite.Suite
}

func TestNameRecordSuite(t *testing.T) {
	s := new(NameRecordTestSuite)
	s.addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	suite.Run(t, s)
}
func (s *NameRecordTestSuite) TestNameRecordString() {
	nr := NewNameRecord("example", s.addr, true)
	s.Require().Equal(fmt.Sprintf("example: %s [restricted]", s.addr.String()), nr.String())
	nr = NewNameRecord("example", s.addr, false)
	s.Require().Equal(fmt.Sprintf("example: %s", s.addr.String()), nr.String())
}

func (s *NameRecordTestSuite) TestNameRecordValidateBasic() {
	cases := map[string]struct {
		name      NameRecord
		expectErr bool
		errValue  string
	}{
		"valid name": {
			NewNameRecord("example", s.addr, true),
			false,
			"",
		},
		"should fail to validate basic empty name": {
			NewNameRecord("", s.addr, true),
			true,
			"segment of name is too short",
		},
		"should fail to validate basic empty addr": {
			NewNameRecord("example", sdk.AccAddress{}, true),
			true,
			"invalid account address",
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := tc.name.ValidateBasic()
			if tc.expectErr {
				s.Error(err)
				if s != nil {
					s.Equal(tc.errValue, err.Error())
				}
			} else {
				s.NoError(err)
			}

		})
	}
}
