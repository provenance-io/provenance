package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type NameKeyTestSuite struct {
	addr1 sdk.AccAddress
	addr2 sdk.AccAddress
	suite.Suite
}

func TestNameKeySuite(t *testing.T) {
	s := new(NameKeyTestSuite)
	s.addr1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	privKey, _ := secp256r1.GenPrivKey()
	s.addr2 = sdk.AccAddress(privKey.PubKey().Address())
	suite.Run(t, s)
}

func (s *NameKeyTestSuite) TestNameKeyPrefix() {
	cases := map[string]struct {
		name      string
		key       []byte
		expectErr bool
		errValue  string
	}{
		"valid two-part": {
		name: 	"name.domain",
		key: 	mustHexDecode("0369e54bac206cf0d1dc5a11c9ae404c5f15a1b75456327e2dba1a8182e507e23a"),
		expectErr: 	false,
		errValue: 	"",
		},
		"valid single": {
		    name: 	"domain",
		    key: 	mustHexDecode("03f2ff83860a4dc203988ed1a22ba1f21237f04abdbd0c4c951103cfbed121de78"),
		    expectErr: 	false,
		    errValue: 	"",
		},
		"valid multi-part": {
			name: "first.second.third.fourth.fifth.sixth.seventh.eighth.ninth.tenth",
			key: mustHexDecode("0319f8e4495c302135434642f853698172ef1f167400d04c12864f9cdf539fbaba"),
			expectErr: false,
			errValue: "",
		},
		"invalid empty name": {
			name: "",
			key: []byte(nil),
			expectErr: true,
			errValue: fmt.Errorf("name can not be empty: %w", ErrNameInvalid).Error(),
		},
		"invalid empty name whitespace": {
			name: "   ",
			key: []byte(nil),
			expectErr: true,
			errValue: fmt.Errorf("name can not be empty: %w", ErrNameInvalid).Error(),
		},
		"invalid empty name segment": {
			name: "name..empty.segment",
			key: []byte(nil),
			expectErr: true,
			errValue: fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid).Error(),
		},
		"invalid empty name segment whitespace": {
			name: "name. .empty.segment",
			key: []byte(nil),
			expectErr: true,
			errValue: fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid).Error(),
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			key, err := GetNameKeyPrefix(tc.name)
			if tc.expectErr {
				s.Error(err)
				if s != nil {
					s.Equal(tc.errValue, err.Error())
				}
			} else {
				s.NoError(err)
			}
			s.Equal(tc.key, key)
		})
	}
}

func (s *NameKeyTestSuite) TestAddressKeyPrefix() {
	key, err := GetAddressKeyPrefix(s.addr1)
	s.Assert().NoError(err)
	// check for address prefix
	s.Assert().Equal("05", hex.EncodeToString(key[0:1]))
	s.Assert().Equal(byte(20), key[1:2][0], "should be the length of key 20 for secp256k1")
	s.Assert().Equal(AddressKeyPrefix, key[0:1])

	key, err = GetAddressKeyPrefix(s.addr2)
	s.Assert().NoError(err)
	// check for address prefix
	s.Assert().Equal("05", hex.EncodeToString(key[0:1]))
	s.Assert().Equal(byte(32), key[1:2][0], "should be the length of key 32 for secp256r1")
	s.Assert().Equal(AddressKeyPrefix, key[0:1])
}

func mustHexDecode(h string) []byte {
	var err error
	var result []byte
	if result, err = hex.DecodeString(h); err != nil {
		panic(err)
	}
	return result
}
