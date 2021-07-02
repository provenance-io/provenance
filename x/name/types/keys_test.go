package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type NameKeyTestSuite struct {
	addr sdk.AccAddress
	suite.Suite
}

func TestNameKeySuite(t *testing.T) {
	s := new(NameKeyTestSuite)
	s.addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
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
			"name.domain",
			mustHexDecode("0569e54bac206cf0d1dc5a11c9ae404c5f15a1b75456327e2dba1a8182e507e23a"),
			false,
			"",
		},
		"valid single": {
			"domain",
			mustHexDecode("05f2ff83860a4dc203988ed1a22ba1f21237f04abdbd0c4c951103cfbed121de78"),
			false,
			"",
		},
		"valid multi-part": {
			"first.second.third.fourth.fifth.sixth.seventh.eighth.ninth.tenth",
			mustHexDecode("0519f8e4495c302135434642f853698172ef1f167400d04c12864f9cdf539fbaba"),
			false,
			"",
		},
		"invalid empty name": {
			"",
			[]byte(nil),
			true,
			fmt.Errorf("name can not be empty: %w", ErrNameInvalid).Error(),
		},
		"invalid empty name whitespace": {
			"   ",
			[]byte(nil),
			true,
			fmt.Errorf("name can not be empty: %w", ErrNameInvalid).Error(),
		},
		"invalid empty name segment": {
			"name..empty.segment",
			[]byte(nil),
			true,
			fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid).Error(),
		},
		"invalid empty name segment whitespace": {
			"name. .empty.segment",
			[]byte(nil),
			true,
			fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid).Error(),
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
			println(hex.EncodeToString(key))
			s.Equal(tc.key, key)
		})
	}
}

func (s *NameKeyTestSuite) TestAddressKeyPrefix() {
	key, err := GetAddressKeyPrefix(s.addr)
	s.Require().NoError(err)
	// check for address prefix
	s.Require().Equal("06", hex.EncodeToString(key[0:1]))
	s.Require().Equal(AddressKeyPrefix, key[0:1])
}

func mustHexDecode(h string) []byte {
	var err error
	var result []byte
	if result, err = hex.DecodeString(h); err != nil {
		panic(err)
	}
	return result
}
