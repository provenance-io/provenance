package types

import (
	"bytes"
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

func (s *NameKeyTestSuite) TestRawBytesKeyCodec() {
	tests := []struct {
		name string
		key  []byte
	}{
		{"empty key", []byte{}},
		{"single byte", []byte{0x01}},
		{"short key", []byte("short")},
		{"20 byte key", bytes.Repeat([]byte{0xAB}, 20)},
		{"32 byte key", bytes.Repeat([]byte{0xCD}, 32)},
		{"random key", []byte{0x01, 0x02, 0x03, 0x04, 0x05}},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			buffer := make([]byte, RawBytesKey.Size(tc.key))
			n, err := RawBytesKey.Encode(buffer, tc.key)
			s.Require().NoError(err)
			s.Equal(len(tc.key), n, "encode returned length")
			s.True(bytes.Equal(tc.key, buffer[:n]), "encoded value should match input")

			// Decode back
			read, out, err := RawBytesKey.Decode(buffer[:n])
			s.Require().NoError(err)
			s.Equal(n, read, "decode should read full encoded length")
			s.True(bytes.Equal(tc.key, out), "decoded value should match input")

			// EncodeNonTerminal
			buffer = make([]byte, RawBytesKey.SizeNonTerminal(tc.key))
			n2, err := RawBytesKey.EncodeNonTerminal(buffer, tc.key)
			s.Require().NoError(err)
			s.Equal(n2, len(tc.key), "non-terminal encode length")
			s.True(bytes.Equal(tc.key, buffer[:n2]), "non-terminal encode matches input")

			read2, out2, err := RawBytesKey.DecodeNonTerminal(buffer[:n2])
			s.Require().NoError(err)
			s.Equal(read2, n2, "non-terminal decode read length")
			s.True(bytes.Equal(tc.key, out2), "non-terminal decode matches input")

			// JSON encode/decode
			jsonVal, err := RawBytesKey.EncodeJSON(tc.key)
			s.Require().NoError(err)

			outJSON, err := RawBytesKey.DecodeJSON(jsonVal)
			s.Require().NoError(err)
			s.True(bytes.Equal(tc.key, outJSON), "json roundtrip should match")

			// Stringify
			s.Equal(string(tc.key), RawBytesKey.Stringify(tc.key))
		})
	}
}
