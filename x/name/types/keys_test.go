package types

import (
	"bytes"
	"encoding/hex"
	"slices"
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

func mustHexDecode(h string) []byte {
	var err error
	var result []byte
	if result, err = hex.DecodeString(h); err != nil {
		panic(err)
	}
	return result
}

func (s *NameKeyTestSuite) TestReverseName() {
	tests := []struct {
		name string
		exp  string
	}{
		{name: "", exp: ""},
		{name: "domain", exp: "domain"},
		{name: "name.domain", exp: "domain.name"},
		{name: "ab.cd.ef", exp: "ef.cd.ab"},
		{name: "one.two.three.four.five", exp: "five.four.three.two.one"},
		{name: "AB.Cd.eF", exp: "ef.cd.ab"},
		{name: " ab . cd .ef ", exp: "ef.cd.ab"},
		{name: "name..domain", exp: "domain..name"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Assert().Equal(tc.exp, ReverseName(tc.name), "ReverseName(%q)", tc.name)
		})
	}
}

func (s *NameKeyTestSuite) TestReversedNameKeyCodec() {
	codec := ReversedNameKeyCodec{}

	tests := []struct {
		name    string
		input   string
		expKey  string // The expected encoded key bytes (as a string).
		expName string // The expected result of decoding the key.
	}{
		{name: "single segment", input: "domain", expKey: "domain", expName: "domain"},
		{name: "two segments", input: "name.domain", expKey: "domain.name", expName: "name.domain"},
		{name: "three segments", input: "ab.cd.ef", expKey: "ef.cd.ab", expName: "ab.cd.ef"},
		{name: "not normalized", input: " AB.Cd . ef", expKey: "ef.cd.ab", expName: "ab.cd.ef"},
		{name: "uuid segment", input: "91978ba2-5f35-459a-86a7-feca1b0512e0.pb", expKey: "pb.91978ba2-5f35-459a-86a7-feca1b0512e0", expName: "91978ba2-5f35-459a-86a7-feca1b0512e0.pb"},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.Run("Encode/Decode", func() {
				size := codec.Size(tc.input)
				s.Require().Equal(len(tc.expKey), size, "Size(%q)", tc.input)

				buffer := make([]byte, size)
				n, err := codec.Encode(buffer, tc.input)
				s.Require().NoError(err, "Encode(%q)", tc.input)
				s.Require().Equal(size, n, "Encode(%q) bytes written", tc.input)
				s.Assert().Equal(tc.expKey, string(buffer), "Encode(%q) key", tc.input)

				read, name, err := codec.Decode(buffer)
				s.Require().NoError(err, "Decode(%q)", buffer)
				s.Assert().Equal(n, read, "Decode(%q) bytes read", buffer)
				s.Assert().Equal(tc.expName, name, "Decode(%q) name", buffer)
			})

			s.Run("EncodeNonTerminal/DecodeNonTerminal", func() {
				size := codec.SizeNonTerminal(tc.input)
				s.Require().Equal(len(tc.expKey)+1, size, "SizeNonTerminal(%q)", tc.input)

				buffer := make([]byte, size)
				n, err := codec.EncodeNonTerminal(buffer, tc.input)
				s.Require().NoError(err, "EncodeNonTerminal(%q)", tc.input)
				s.Require().Equal(size, n, "EncodeNonTerminal(%q) bytes written", tc.input)

				// Put something after it to make sure DecodeNonTerminal only reads its part.
				buffer = append(buffer, []byte("extra")...)
				read, name, err := codec.DecodeNonTerminal(buffer)
				s.Require().NoError(err, "DecodeNonTerminal(%q)", buffer)
				s.Assert().Equal(n, read, "DecodeNonTerminal(%q) bytes read", buffer)
				s.Assert().Equal(tc.expName, name, "DecodeNonTerminal(%q) name", buffer)
			})

			s.Run("JSON", func() {
				bz, err := codec.EncodeJSON(tc.input)
				s.Require().NoError(err, "EncodeJSON(%q)", tc.input)
				name, err := codec.DecodeJSON(bz)
				s.Require().NoError(err, "DecodeJSON(%q)", bz)
				s.Assert().Equal(tc.input, name, "DecodeJSON(%q)", bz)
			})

			s.Run("Stringify and KeyType", func() {
				s.Assert().Equal(tc.input, codec.Stringify(tc.input), "Stringify(%q)", tc.input)
				s.Assert().Equal("reversedname", codec.KeyType(), "KeyType()")
			})
		})
	}
}

// TestReversedNameKeyOrdering makes sure that a name's sub-names are stored right after it, and share its key prefix.
func (s *NameKeyTestSuite) TestReversedNameKeyOrdering() {
	codec := ReversedNameKeyCodec{}
	encode := func(name string) []byte {
		buffer := make([]byte, codec.Size(name))
		_, err := codec.Encode(buffer, name)
		s.Require().NoError(err, "Encode(%q)", name)
		return buffer
	}

	names := []string{"ab.cd.ef", "gh.ef", "ef", "cdx.ef", "cd.ef", "xy.ab.cd.ef", "cd.gh"}
	keys := make([][]byte, len(names))
	for i, name := range names {
		keys[i] = encode(name)
	}
	slices.SortFunc(keys, bytes.Compare)

	actual := make([]string, len(keys))
	for i, key := range keys {
		actual[i] = string(key)
	}
	expected := []string{"ef", "ef.cd", "ef.cd.ab", "ef.cd.ab.xy", "ef.cdx", "ef.gh", "gh.cd"}
	s.Assert().Equal(expected, actual, "sorted keys")

	// Everything under cd.ef has a key starting with "ef.cd." (note the trailing dot, so cdx.ef isn't included).
	childPrefix := append(encode("cd.ef"), '.')
	var children []string
	for _, name := range names {
		if bytes.HasPrefix(encode(name), childPrefix) {
			children = append(children, name)
		}
	}
	s.Assert().ElementsMatch([]string{"ab.cd.ef", "xy.ab.cd.ef"}, children, "names under cd.ef")
}
