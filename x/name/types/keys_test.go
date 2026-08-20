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
			"name.domain",
			mustHexDecode("03e27733edfa985bcd3fdcfe8544741d602a379fd28464b3c15f483b7350e2dd20"),
			false,
			"",
		},
		"valid single": {
			"domain",
			mustHexDecode("03f2ff83860a4dc203988ed1a22ba1f21237f04abdbd0c4c951103cfbed121de78"),
			false,
			"",
		},
		"valid multi-part": {
			"first.second.third.fourth.fifth.sixth.seventh.eighth.ninth.tenth",
			mustHexDecode("031cccba52f948b1d9e2123bb38988a08d733a040e78ecb516c608e7454960bf01"),
			false,
			"",
		},
		"valid two-part with segment spacing": {
			" name . domain ",
			mustHexDecode("03e27733edfa985bcd3fdcfe8544741d602a379fd28464b3c15f483b7350e2dd20"),
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

func (s *NameKeyTestSuite) TestLegacyNameKeyPrefix() {
	cases := map[string]struct {
		name      string
		key       []byte
		expectErr bool
		errValue  string
	}{
		"valid two-part": {
			"name.domain",
			mustHexDecode("0369e54bac206cf0d1dc5a11c9ae404c5f15a1b75456327e2dba1a8182e507e23a"),
			false,
			"",
		},
		"valid single": {
			"domain",
			mustHexDecode("03f2ff83860a4dc203988ed1a22ba1f21237f04abdbd0c4c951103cfbed121de78"),
			false,
			"",
		},
		"valid multi-part": {
			"first.second.third.fourth.fifth.sixth.seventh.eighth.ninth.tenth",
			mustHexDecode("0319f8e4495c302135434642f853698172ef1f167400d04c12864f9cdf539fbaba"),
			false,
			"",
		},
		"invalid empty name": {
			"",
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
	}
	for n, tc := range cases {
		s.Run(n, func() {
			key, err := GetLegacyNameKeyPrefix(tc.name)
			if tc.expectErr {
				s.Error(err)
				s.Equal(tc.errValue, err.Error())
			} else {
				s.NoError(err)
			}
			s.Equal(tc.key, key)
		})
	}
}

func (s *NameKeyTestSuite) TestNameKeyPrefixSegmentCollisions() {
	cases := []struct {
		name  string
		name1 string
		name2 string
	}{
		{
			name:  "extra segment in the middle",
			name1: "leaf.mid.pb",
			name2: "midleaf.pb",
		},
		{
			name:  "extra segment at the front",
			name1: "kyc.identity.pb",
			name2: "identitykyc.pb",
		},
		{
			name:  "two segments vs one",
			name1: "aa.bb",
			name2: "bbaa",
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			key1, err := GetNameKeyPrefix(tc.name1)
			s.Require().NoError(err, "GetNameKeyPrefix(%q)", tc.name1)
			key2, err := GetNameKeyPrefix(tc.name2)
			s.Require().NoError(err, "GetNameKeyPrefix(%q)", tc.name2)
			s.Assert().NotEqual(hex.EncodeToString(key1), hex.EncodeToString(key2),
				"%q and %q are different names and must not share a store key", tc.name1, tc.name2)
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
