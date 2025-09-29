package types

import (
	"crypto/sha256"
	"encoding/base64"
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
			mustHexDecode("0769e54bac206cf0d1dc5a11c9ae404c5f15a1b75456327e2dba1a8182e507e23a"),
			false,
			"",
		},
		"valid single": {
			"domain",
			mustHexDecode("07f2ff83860a4dc203988ed1a22ba1f21237f04abdbd0c4c951103cfbed121de78"),
			false,
			"",
		},
		"valid multi-part": {
			"first.second.third.fourth.fifth.sixth.seventh.eighth.ninth.tenth",
			mustHexDecode("0719f8e4495c302135434642f853698172ef1f167400d04c12864f9cdf539fbaba"),
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
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "single char", input: "a"},
		{name: "short string", input: "short"},
		{name: "domain style", input: "example.domain"},
		{name: "multi-level domain", input: "one.two.three.four"},
		{name: "long string", input: strings.Repeat("x", 100)},
		{name: "whitespace only", input: "   "},
		{name: "contains whitespace", input: "some domain.name"},
		{name: "special chars", input: "!@#$%^&*()_+{}|:\"<>?"},
	}

	for _, tc := range tests {
		tc := tc
		s.Run(tc.name, func() {
			expectedHash := codec.ComputeHash(tc.input)
			size := codec.Size(tc.input)
			s.Equal(sha256.Size, size, "Size mismatch for input: %q", tc.input)

			buffer := make([]byte, size)
			n, err := codec.Encode(buffer, tc.input)
			s.Require().NoError(err, "Encode error for input: %q", tc.input)
			s.Equal(size, n, "Encoded size mismatch for input: %q", tc.input)
			s.Equal(expectedHash, buffer[:n], "Hash mismatch after Encode for input: %q", tc.input)

			s.Run("Decode", func() {
				read, out, err := codec.Decode(buffer[:n])
				s.Require().NoError(err, "Decode error for input: %q", tc.input)
				s.Equal(n, read, "Decoded read length mismatch for input: %q", tc.input)
				s.NotEmpty(out, "Decode returned empty string for input: %q", tc.input)
				s.NotEqual(tc.input, out, "Decode should not return original input for: %q", tc.input)
			})

			s.Run("Encode/Decode NonTerminal", func() {
				nonTermSize := codec.SizeNonTerminal(tc.input)
				s.Equal(sha256.Size, nonTermSize, "NonTerminal size mismatch for input: %q", tc.input)

				nonTermBuf := make([]byte, nonTermSize)
				n2, err := codec.EncodeNonTerminal(nonTermBuf, tc.input)
				s.Require().NoError(err, "EncodeNonTerminal error for input: %q", tc.input)
				s.Equal(nonTermSize, n2, "EncodeNonTerminal size mismatch for input: %q", tc.input)
				s.Equal(expectedHash, nonTermBuf[:n2], "NonTerminal hash mismatch for input: %q", tc.input)

				read2, out2, err := codec.DecodeNonTerminal(nonTermBuf[:n2])
				s.Require().NoError(err, "DecodeNonTerminal error for input: %q", tc.input)
				s.Equal(n2, read2, "DecodeNonTerminal read mismatch for input: %q", tc.input)
				s.Equal(base64.StdEncoding.EncodeToString(expectedHash), out2, "DecodeNonTerminal base64 mismatch for input: %q", tc.input)
			})

			s.Run("JSON Encode/Decode", func() {
				jsonBytes, err := codec.EncodeJSON(tc.input)
				s.Require().NoError(err, "EncodeJSON error for input: %q", tc.input)

				outJSON, err := codec.DecodeJSON(jsonBytes)
				s.Require().NoError(err, "DecodeJSON error for input: %q", tc.input)
				s.Equal(tc.input, outJSON, "JSON round-trip mismatch for input: %q", tc.input)
			})

			s.Run("Stringify and KeyType", func() {
				s.Equal(tc.input, codec.Stringify(tc.input), "Stringify mismatch for input: %q", tc.input)
				s.Equal("hashedstring", codec.KeyType(), "KeyType mismatch for input: %q", tc.input)
			})

			s.Run("Hash is deterministic", func() {
				hash1 := codec.ComputeHash(tc.input)
				hash2 := codec.ComputeHash(tc.input)
				s.Equal(hash1, hash2, "ComputeHash not deterministic for input: %q", tc.input)
			})

		})
	}

	s.Run("Hash uniqueness (no collision)", func() {
		hashes := map[string][]byte{}
		seen := map[string]bool{}

		for _, tc := range tests {
			if strings.TrimSpace(tc.input) == "" {
				continue
			}
			hash := codec.ComputeHash(tc.input)
			hashStr := base64.StdEncoding.EncodeToString(hash)

			s.False(seen[hashStr], "Duplicate hash found for input: %q", tc.input)
			seen[hashStr] = true
			hashes[tc.input] = hash
		}
	})
}
