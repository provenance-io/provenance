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
			expectedHash := codec.ComputeHash(tc.input)
			size := codec.Size(tc.input)
			s.Equal(sha256.Size, size, "codec.Size should match sha256 size")

			buffer := make([]byte, size)
			n, err := codec.Encode(buffer, tc.input)

			if tc.expectErr {
				s.Error(err)
				return
			}

			s.Require().NoError(err)
			s.Equal(size, n)
			s.Equal(expectedHash, buffer[:n], "computed hash and encoded hash should match")

			read, out, err := codec.Decode(buffer[:n])
			s.Require().NoError(err)
			s.Equal(n, read)
			s.NotEmpty(out)
			s.NotEqual(tc.input, out, "decoded string should be base64 of hash, not the original input")

			nonTermSize := codec.SizeNonTerminal(tc.input)
			s.Equal(sha256.Size, nonTermSize, "non-terminal size must also match sha256 size")

			nonTermBuf := make([]byte, nonTermSize)
			n2, err := codec.EncodeNonTerminal(nonTermBuf, tc.input)
			s.Require().NoError(err)
			s.Equal(nonTermSize, n2)
			s.Equal(expectedHash, nonTermBuf[:n2], "non-terminal hash must match terminal hash")

			read2, out2, err := codec.DecodeNonTerminal(nonTermBuf[:n2])
			s.Require().NoError(err)
			s.Equal(n2, read2)
			s.Equal(out, out2, "non-terminal decode must match regular decode")

			jsonBytes, err := codec.EncodeJSON(tc.input)
			s.Require().NoError(err)

			outJSON, err := codec.DecodeJSON(jsonBytes)
			s.Require().NoError(err)
			s.Equal(tc.input, outJSON)

			s.Equal(tc.input, codec.Stringify(tc.input))

			s.Equal("hashedstring", codec.KeyType())

			hash1 := codec.ComputeHash(tc.input)
			hash2 := codec.ComputeHash(tc.input)
			s.Equal(hash1, hash2, "ComputeHash must be deterministic")

			s.T().Logf("Input: %q → Hash: %x (Base64: %s)", tc.input, hash1, base64.StdEncoding.EncodeToString(hash1))
		})
	}

	s.Run("Hash uniqueness (no collision)", func() {
		hashes := map[string][]byte{}
		seen := map[string]bool{}
		for _, tc := range tests {
			hash := codec.ComputeHash(tc.input)
			hashStr := base64.StdEncoding.EncodeToString(hash)
			if strings.TrimSpace(tc.input) == "" {
				continue
			}
			s.False(seen[hashStr], "duplicate hash found for input: %q", tc.input)
			seen[hashStr] = true
			hashes[tc.input] = hash
		}
	})
}
