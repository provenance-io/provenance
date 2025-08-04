package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
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
			fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid).Error(),
		},
		"invalid empty name whitespace": {
			"   ",
			[]byte(nil),
			true,
			fmt.Errorf("name segment cannot be empty: %w", ErrNameInvalid).Error(),
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
			key, err := GetNameKeyBytes(tc.name)
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

func (s *NameKeyTestSuite) TestHashedStringKeyCodec() {
	codec := HashedStringKeyCodec{}

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{"empty string", "", false},
		{"single char", "a", false},
		{"short string", "short", false},
		{"domain style", "example.domain", false},
		{"multi-level domain", "one.two.three.four", false},
		{"long string", strings.Repeat("x", 100), false},
		{"whitespace only", "   ", false},
		{"contains whitespace", "some domain.name", false},
		{"special chars", "!@#$%^&*()_+{}|:\"<>?", false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Encode
			size := codec.Size(tc.input)
			buffer := make([]byte, size)
			n, err := codec.Encode(buffer, tc.input)

			if tc.expectErr {
				s.Error(err)
				return
			}

			s.Require().NoError(err)
			s.Equal(sha256.Size, n, "encoded hash size must be 32 bytes")

			// Decode: will return base64 of the hash (not the original string)
			read, out, err := codec.Decode(buffer[:n])
			s.Require().NoError(err)
			s.Equal(n, read)
			s.NotEmpty(out)
			s.NotEqual(tc.input, out, "decoded value must not match original (it's a hash)")

			// EncodeNonTerminal / DecodeNonTerminal
			nonTermSize := codec.SizeNonTerminal(tc.input)
			nonTermBuf := make([]byte, nonTermSize)
			n2, err := codec.EncodeNonTerminal(nonTermBuf, tc.input)
			s.Require().NoError(err)
			s.Equal(nonTermSize, n2, "non-terminal encode length")

			read2, out2, err := codec.DecodeNonTerminal(nonTermBuf[:n2])
			s.Require().NoError(err)
			s.Equal(n2, read2, "non-terminal decode should match encoded length")
			s.Equal(tc.input, out2, "non-terminal decode should match original")

			// JSON encode/decode
			jsonBytes, err := codec.EncodeJSON(tc.input)
			s.Require().NoError(err)

			outJSON, err := codec.DecodeJSON(jsonBytes)
			s.Require().NoError(err)
			s.Equal(tc.input, outJSON)

			// Stringify
			s.Equal(tc.input, codec.Stringify(tc.input))
		})
	}
}
