package types

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/testlog"
)

type AddressTestSuite struct {
	suite.Suite

	scopeUUIDstr   string
	scopeUUID      uuid.UUID
	sessionUUIDstr string
	sessionUUID    uuid.UUID

	scopeHex      string
	scopeBech32   string
	sessionBech32 string
	recordBech32  string
}

func (s *AddressTestSuite) SetupTest() {
	s.scopeUUIDstr = "8d80b25a-c089-4446-956e-5d08cfe3e1a5"
	s.scopeUUID = uuid.MustParse(s.scopeUUIDstr)
	s.sessionUUIDstr = "c25c7bd4-c639-4367-a842-f64fa5fccc19"
	s.sessionUUID = uuid.MustParse(s.sessionUUIDstr)

	s.scopeHex = "008D80B25AC0894446956E5D08CFE3E1A5"
	s.scopeBech32 = "scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp"
	s.sessionBech32 = "session1qxxcpvj6czy5g354dews3nlruxjuyhrm6nrrjsm84pp0vna9lnxpjewp6kf"
	s.recordBech32 = "record1q2xcpvj6czy5g354dews3nlruxjelpkssxyyclt9ngh74gx9ttgp27gt8kl"
}

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(AddressTestSuite))
}

func mapToStrings[S ~[]E, E fmt.Stringer](vals S) []string {
	if vals == nil {
		return nil
	}
	rv := make([]string, len(vals))
	for i, v := range vals {
		rv[i] = v.String()
	}
	return rv
}

// subSet creates a new slice containing the entries of vals that have the given indexes.
// The resulting entries will be in the order that the indexes are provided.
func subSet[S ~[]E, E any](vals S, indexes ...int) S {
	if len(indexes) == 0 {
		return nil
	}
	rv := make(S, 0, len(indexes))
	for _, i := range indexes {
		rv = append(rv, vals[i])
	}
	return rv
}

func (s *AddressTestSuite) requireBech32String(typeCode []byte, data []byte) string {
	hrp, err := MetadataAddress{typeCode[0]}.Prefix()
	s.Require().NoError(err, "getPrefix error")
	addr := append([]byte{typeCode[0]}, data...)
	bech32Addr, err := bech32.ConvertAndEncode(hrp, addr)
	s.Require().NoError(err, "bech32.ConvertAndEncode error")
	return bech32Addr
}

func (s *AddressTestSuite) TestLegacySha512HashToAddress() {
	testHashBytes := sha512.Sum512([]byte("test"))
	testHash := base64.StdEncoding.EncodeToString(testHashBytes[:])
	testHash15 := base64.StdEncoding.EncodeToString(testHashBytes[:15])
	testHash31 := base64.StdEncoding.EncodeToString(testHashBytes[:31])

	tests := []struct {
		name          string
		typeBytes     []byte
		hash          string
		expectedAddr  string
		expectedError string
	}{
		{
			"empty typeBytes",
			[]byte{},
			testHash,
			"",
			"empty typeCode bytes",
		},
		{
			"empty hash",
			ScopeKeyPrefix,
			"",
			"",
			"empty hash string",
		},
		{
			"scope key prefix - valid",
			ScopeKeyPrefix,
			testHash,
			s.requireBech32String(ScopeKeyPrefix, testHashBytes[:16]),
			"",
		},
		{
			"scope key prefix - too short",
			ScopeKeyPrefix,
			testHash15,
			"",
			fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash15, 16, 15),
		},
		{
			"session key prefix - valid",
			SessionKeyPrefix,
			testHash,
			s.requireBech32String(SessionKeyPrefix, testHashBytes[:32]),
			"",
		},
		{
			"session key prefix - too short",
			SessionKeyPrefix,
			testHash31,
			"",
			fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash31, 32, 31),
		},
		{
			"record key prefix - valid",
			RecordKeyPrefix,
			testHash,
			s.requireBech32String(RecordKeyPrefix, testHashBytes[:32]),
			"",
		},
		{
			"record key prefix - too short",
			RecordKeyPrefix,
			testHash31,
			"",
			fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash31, 32, 31),
		},
		{
			"scope spec key prefix - valid",
			ScopeSpecificationKeyPrefix,
			testHash,
			s.requireBech32String(ScopeSpecificationKeyPrefix, testHashBytes[:16]),
			"",
		},
		{
			"scope spec key prefix - too short",
			ScopeSpecificationKeyPrefix,
			testHash15,
			"",
			fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash15, 16, 15),
		},
		{
			"contract spec key prefix - valid",
			ContractSpecificationKeyPrefix,
			testHash,
			s.requireBech32String(ContractSpecificationKeyPrefix, testHashBytes[:16]),
			"",
		},
		{
			"contract spec key prefix - too short",
			ContractSpecificationKeyPrefix,
			testHash15,
			"",
			fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash15, 16, 15),
		},
		{
			"record spec key prefix - valid",
			RecordSpecificationKeyPrefix,
			testHash,
			s.requireBech32String(RecordSpecificationKeyPrefix, testHashBytes[:32]),
			"",
		},
		{
			"record spec key prefix - too short",
			RecordSpecificationKeyPrefix,
			testHash31,
			"",
			fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash31, 32, 31),
		},
		{
			"invalid type bytes",
			[]byte{0x07},
			testHash,
			"",
			"invalid address type code 0x07",
		},
		{
			"invalid hash",
			ScopeKeyPrefix,
			"invalid hash",
			"",
			base64.CorruptInputError(7).Error(),
		},
		{
			"hash string decodes to empty",
			ScopeKeyPrefix,
			"MA==",
			"",
			"invalid hash \"MA==\" byte length, expected at least 16 bytes, found 1",
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			addr, err := ConvertHashToAddress(tc.typeBytes, tc.hash)
			if len(tc.expectedError) == 0 {
				require.NoError(t, err, "ConvertHashToAddress error")
				require.NoError(t, addr.Validate(), "addr.Validate() error")
				require.Equal(t, tc.expectedAddr, addr.String(), "ConvertHashToAddress value (as string)")
			} else {
				require.EqualError(t, err, tc.expectedError, "ConvertHashToAddress error expected")
			}
		})
	}
}

func (s *AddressTestSuite) TestVerifyMetadataAddressFormat() {
	uuid0 := uuid.Nil
	uuid1 := uuid.MustParse("1D42DB43-FCF2-46F8-A4B6-974D73B6551E") // came from uuidgen
	uuid2 := uuid.MustParse("9713E8BB-8728-4CE9-8051-FCA03E7BD1D1") // came from uuidgen

	makeBz := func(parts ...[]byte) []byte {
		var rv []byte
		for _, part := range parts {
			rv = append(rv, part...)
		}
		return rv
	}

	type testCase struct {
		name   string
		bz     []byte
		expHRP string
		expErr string
	}

	tests := []testCase{
		{
			name:   "nil",
			bz:     nil,
			expErr: "address is empty",
		},
		{
			name:   "empty",
			bz:     []byte{},
			expErr: "address is empty",
		},
		{
			name:   "scope zeros",
			bz:     makeBz(ScopeKeyPrefix, uuid0[:]),
			expHRP: PrefixScope,
		},
		{
			name:   "scope normal",
			bz:     makeBz(ScopeKeyPrefix, uuid1[:]),
			expHRP: PrefixScope,
		},
		{
			name:   "scope too short",
			bz:     makeBz(ScopeKeyPrefix, uuid1[:15]),
			expHRP: PrefixScope,
			expErr: "incorrect address length (expected: 17, actual: 16)",
		},
		{
			name:   "scope too long",
			bz:     makeBz(ScopeKeyPrefix, uuid1[:], ScopeKeyPrefix),
			expHRP: PrefixScope,
			expErr: "incorrect address length (expected: 17, actual: 18)",
		},
		{
			name:   "session zeros",
			bz:     makeBz(SessionKeyPrefix, uuid0[:], uuid0[:]),
			expHRP: PrefixSession,
		},
		{
			name:   "session normal",
			bz:     makeBz(SessionKeyPrefix, uuid1[:], uuid2[:]),
			expHRP: PrefixSession,
		},
		{
			name:   "session too short",
			bz:     makeBz(SessionKeyPrefix, uuid1[:], uuid2[:15]),
			expHRP: PrefixSession,
			expErr: "incorrect address length (expected: 33, actual: 32)",
		},
		{
			name:   "session too long",
			bz:     makeBz(SessionKeyPrefix, uuid1[:], uuid2[:], SessionKeyPrefix),
			expHRP: PrefixSession,
			expErr: "incorrect address length (expected: 33, actual: 34)",
		},
		{
			name:   "record zeros",
			bz:     makeBz(RecordKeyPrefix, uuid0[:], uuid0[:]),
			expHRP: PrefixRecord,
		},
		{
			name:   "record normal",
			bz:     makeBz(RecordKeyPrefix, uuid1[:], uuid2[:]),
			expHRP: PrefixRecord,
		},
		{
			name:   "record too short",
			bz:     makeBz(RecordKeyPrefix, uuid1[:], uuid2[:15]),
			expHRP: PrefixRecord,
			expErr: "incorrect address length (expected: 33, actual: 32)",
		},
		{
			name:   "record too long",
			bz:     makeBz(RecordKeyPrefix, uuid1[:], uuid2[:], RecordKeyPrefix),
			expHRP: PrefixRecord,
			expErr: "incorrect address length (expected: 33, actual: 34)",
		},
		{
			name:   "scope spec zeros",
			bz:     makeBz(ScopeSpecificationKeyPrefix, uuid0[:]),
			expHRP: PrefixScopeSpecification,
		},
		{
			name:   "scope spec normal",
			bz:     makeBz(ScopeSpecificationKeyPrefix, uuid1[:]),
			expHRP: PrefixScopeSpecification,
		},
		{
			name:   "scope spec too short",
			bz:     makeBz(ScopeSpecificationKeyPrefix, uuid1[:15]),
			expHRP: PrefixScopeSpecification,
			expErr: "incorrect address length (expected: 17, actual: 16)",
		},
		{
			name:   "scope spec too long",
			bz:     makeBz(ScopeSpecificationKeyPrefix, uuid1[:], ScopeSpecificationKeyPrefix),
			expHRP: PrefixScopeSpecification,
			expErr: "incorrect address length (expected: 17, actual: 18)",
		},
		{
			name:   "contract spec zeros",
			bz:     makeBz(ContractSpecificationKeyPrefix, uuid0[:]),
			expHRP: PrefixContractSpecification,
		},
		{
			name:   "contract spec normal",
			bz:     makeBz(ContractSpecificationKeyPrefix, uuid1[:]),
			expHRP: PrefixContractSpecification,
		},
		{
			name:   "contract spec too short",
			bz:     makeBz(ContractSpecificationKeyPrefix, uuid1[:15]),
			expHRP: PrefixContractSpecification,
			expErr: "incorrect address length (expected: 17, actual: 16)",
		},
		{
			name:   "contract spec too long",
			bz:     makeBz(ContractSpecificationKeyPrefix, uuid1[:], ContractSpecificationKeyPrefix),
			expHRP: PrefixContractSpecification,
			expErr: "incorrect address length (expected: 17, actual: 18)",
		},
		{
			name:   "record spec zeros",
			bz:     makeBz(RecordSpecificationKeyPrefix, uuid0[:], uuid0[:]),
			expHRP: PrefixRecordSpecification,
		},
		{
			name:   "record spec normal",
			bz:     makeBz(RecordSpecificationKeyPrefix, uuid1[:], uuid2[:]),
			expHRP: PrefixRecordSpecification,
		},
		{
			name:   "record spec too short",
			bz:     makeBz(RecordSpecificationKeyPrefix, uuid1[:], uuid2[:15]),
			expHRP: PrefixRecordSpecification,
			expErr: "incorrect address length (expected: 33, actual: 32)",
		},
		{
			name:   "record spec too long",
			bz:     makeBz(RecordSpecificationKeyPrefix, uuid1[:], uuid2[:], RecordSpecificationKeyPrefix),
			expHRP: PrefixRecordSpecification,
			expErr: "incorrect address length (expected: 33, actual: 34)",
		},
		// Note: As of writing this, the only way uuid.FromBytes returns an error is if the
		// provided slice doesn't have length 16. But that's hard-coded to be the case in
		// VerifyMetadataAddressFormat, so there's no unit tests for those checks.
	}

	knownTypeBytes := []byte{
		ScopeKeyPrefix[0],
		SessionKeyPrefix[0],
		RecordKeyPrefix[0],
		ScopeSpecificationKeyPrefix[0],
		ContractSpecificationKeyPrefix[0],
		RecordSpecificationKeyPrefix[0],
	}
	isKnownTypeByte := func(b byte) bool {
		for _, kb := range knownTypeBytes {
			if b == kb {
				return true
			}
		}
		return false
	}

	for i := 0; i < 256; i++ {
		b := byte(i)
		if !isKnownTypeByte(b) {
			tests = append(tests, testCase{
				name:   fmt.Sprintf("type byte 0x%x", b),
				bz:     []byte{b},
				expErr: fmt.Sprintf("invalid metadata address type: %d", i),
			})
		}
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			hrp, err := VerifyMetadataAddressFormat(tc.bz)
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "VerifyMetadataAddressFormat error")
			} else {
				s.Assert().NoError(err, "VerifyMetadataAddressFormat error")
			}
			s.Assert().Equal(tc.expHRP, hrp, "VerifyMetadataAddressFormat hrp")
		})
	}
}

func (s *AddressTestSuite) TestGetNameForHRP() {
	tests := []struct {
		hrp string
		exp string
	}{
		{hrp: PrefixScope, exp: "scope"},
		{hrp: PrefixSession, exp: "session"},
		{hrp: PrefixRecord, exp: "record"},
		{hrp: PrefixScopeSpecification, exp: "scope specification"},
		{hrp: PrefixContractSpecification, exp: "contract specification"},
		{hrp: PrefixRecordSpecification, exp: "record specification"},
		{hrp: "", exp: `<"">`},
		{hrp: "unknown", exp: `<"unknown">`},
		{hrp: `I might be "evil"`, exp: `<"I might be \"evil\"">`},
		{hrp: PrefixScope + "1", exp: `<"scope1">`},
		{hrp: PrefixScope + "spe", exp: `<"scopespe">`},
		{hrp: PrefixScope + "spec ", exp: `<"scopespec ">`},
		{hrp: PrefixScope + "specification", exp: `<"scopespecification">`},
		{hrp: PrefixScope[1:], exp: `<"cope">`},
		{hrp: PrefixScope[:len(PrefixScope)-1], exp: `<"scop">`},
		{hrp: PrefixRecord + "spec", exp: `<"recordspec">`},
		{hrp: PrefixRecord + "specification", exp: `<"recordspecification">`},
	}

	for _, tc := range tests {
		name := tc.hrp
		if len(name) == 0 {
			name = "(empty)"
		}
		s.Run(name, func() {
			var act string
			testFunc := func() {
				act = getNameForHRP(tc.hrp)
			}
			s.Require().NotPanics(testFunc, "getNameForHRP(%q)", tc.hrp)
			s.Assert().Equal(tc.exp, act, "result from getNameForHRP(%q)", tc.hrp)
		})
	}
}

func (s *AddressTestSuite) TestVerifyMetadataAddressHasType() {
	newUUID := func(i string) uuid.UUID {
		id := strings.ReplaceAll("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", "x", i)
		rv, err := uuid.Parse(id)
		s.Require().NoError(err, "uuid.Parse(%q)", id)
		return rv
	}
	type testCase struct {
		name   string
		ma     MetadataAddress
		hrp    string
		expErr string
	}

	tests := []testCase{
		{
			name:   "nil",
			ma:     nil,
			hrp:    "whatever",
			expErr: "invalid <\"whatever\"> metadata address MetadataAddress(nil): address is empty",
		},
		{
			name:   "empty",
			ma:     MetadataAddress{},
			hrp:    "thingy",
			expErr: "invalid <\"thingy\"> metadata address MetadataAddress{}: address is empty",
		},
		{
			name:   "invalid scope id",
			ma:     MetadataAddress{ScopeKeyPrefix[0], 0x1, 0x2},
			hrp:    PrefixScope,
			expErr: "invalid scope metadata address MetadataAddress{0x0, 0x1, 0x2}: incorrect address length (expected: 17, actual: 3)",
		},
		{
			name:   "invalid session id",
			ma:     MetadataAddress{SessionKeyPrefix[0], 0x3, 0x4},
			hrp:    PrefixSession,
			expErr: "invalid session metadata address MetadataAddress{0x1, 0x3, 0x4}: incorrect address length (expected: 33, actual: 3)",
		},
		{
			name:   "invalid record id",
			ma:     MetadataAddress{RecordKeyPrefix[0], 0x5, 0x6},
			hrp:    PrefixRecord,
			expErr: "invalid record metadata address MetadataAddress{0x2, 0x5, 0x6}: incorrect address length (expected: 33, actual: 3)",
		},
		{
			name:   "invalid scope spec id",
			ma:     MetadataAddress{ScopeSpecificationKeyPrefix[0], 0x7, 0x8},
			hrp:    PrefixScopeSpecification,
			expErr: "invalid scope specification metadata address MetadataAddress{0x4, 0x7, 0x8}: incorrect address length (expected: 17, actual: 3)",
		},
		{
			name:   "invalid contract spec id",
			ma:     MetadataAddress{ContractSpecificationKeyPrefix[0], 0x9, 0xa},
			hrp:    PrefixContractSpecification,
			expErr: "invalid contract specification metadata address MetadataAddress{0x3, 0x9, 0xa}: incorrect address length (expected: 17, actual: 3)",
		},
		{
			name:   "invalid record spec id",
			ma:     MetadataAddress{RecordSpecificationKeyPrefix[0], 0xb, 0xc},
			hrp:    PrefixRecordSpecification,
			expErr: "invalid record specification metadata address MetadataAddress{0x5, 0xb, 0xc}: incorrect address length (expected: 33, actual: 3)",
		},
	}

	validCases := []struct {
		hrp string
		ma  MetadataAddress
	}{
		{hrp: PrefixScope, ma: ScopeMetadataAddress(newUUID("1"))},
		{hrp: PrefixSession, ma: SessionMetadataAddress(newUUID("2"), newUUID("3"))},
		{hrp: PrefixRecord, ma: RecordMetadataAddress(newUUID("4"), newUUID("5").String())},
		{hrp: PrefixScopeSpecification, ma: ScopeSpecMetadataAddress(newUUID("6"))},
		{hrp: PrefixContractSpecification, ma: ContractSpecMetadataAddress(newUUID("7"))},
		{hrp: PrefixRecordSpecification, ma: RecordSpecMetadataAddress(newUUID("8"), newUUID("9").String())},
	}
	for _, vc := range validCases {
		for _, typeVC := range validCases {
			tc := testCase{
				name: fmt.Sprintf("valid %s: want %s", vc.hrp, typeVC.hrp),
				ma:   vc.ma,
				hrp:  typeVC.hrp,
			}
			if vc.hrp != typeVC.hrp {
				tc.expErr = fmt.Sprintf("invalid %s id \"%s\": wrong type", getNameForHRP(tc.hrp), tc.ma.String())
			}
			tests = append(tests, tc)
		}
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var err error
			testFunc := func() {
				err = VerifyMetadataAddressHasType(tc.ma, tc.hrp)
			}
			s.Require().NotPanics(testFunc, "VerifyMetadataAddressHasType(%q, %q)", tc.ma, tc.hrp)
			assertions.AssertErrorValue(s.T(), err, tc.expErr, "error from VerifyMetadataAddressHasType(%q, %q)", tc.ma, tc.hrp)
		})
	}
}

func (s *AddressTestSuite) TestMetadataAddressFromBech32() {
	notAScopeAddr := MetadataAddress{ScopeKeyPrefix[0], 1, 2, 3}
	notAScopeAddrStr, err := bech32.ConvertAndEncode(PrefixScope, notAScopeAddr)
	s.Require().NoError(err, "ConvertAndEncode notAScopeAddrStr")

	makeAddr := func(parts ...[]byte) MetadataAddress {
		var rv MetadataAddress
		for _, part := range parts {
			rv = append(rv, part...)
		}
		return rv
	}

	uuid1 := uuid.MustParse("1D42DB43-FCF2-46F8-A4B6-974D73B6551E") // came from uuidgen
	uuid2 := uuid.MustParse("9713E8BB-8728-4CE9-8051-FCA03E7BD1D1") // came from uuidgen
	scopeAddr := makeAddr(ScopeKeyPrefix, uuid1[:])
	sessionAddr := makeAddr(SessionKeyPrefix, uuid1[:], uuid2[:])
	recordAddr := makeAddr(RecordKeyPrefix, uuid1[:], uuid2[:])
	scopeSpecAddr := makeAddr(ScopeSpecificationKeyPrefix, uuid1[:])
	contractSpecAddr := makeAddr(ContractSpecificationKeyPrefix, uuid1[:])
	recordSpecAddr := makeAddr(RecordSpecificationKeyPrefix, uuid1[:], uuid2[:])

	hrpMismatchAddr, err := bech32.ConvertAndEncode(PrefixScopeSpecification, scopeAddr)
	s.Require().NoError(err, "ConvertAndEncode hrpMismatchAddr")

	tests := []struct {
		name    string
		input   string
		expAddr MetadataAddress
		expHRP  string
		expErr  string
	}{
		{
			name:    "empty",
			input:   "",
			expAddr: MetadataAddress{},
			expHRP:  "",
			expErr:  "empty address string is not allowed",
		},
		{
			name:    "white space",
			input:   "       ",
			expAddr: MetadataAddress{},
			expHRP:  "",
			expErr:  "empty address string is not allowed",
		},
		{
			name:    "invalid bech32",
			input:   "notvalid",
			expAddr: nil,
			expHRP:  "",
			expErr:  "decoding bech32 failed: invalid separator index -1",
		},
		{
			name:    "invalid metadata address bytes",
			input:   notAScopeAddrStr,
			expAddr: nil,
			expHRP:  "",
			expErr:  "incorrect address length (expected: 17, actual: 4)",
		},
		{
			name:    "hrp mismatch",
			input:   hrpMismatchAddr,
			expAddr: MetadataAddress{},
			expHRP:  "",
			expErr:  "invalid bech32 prefix; expected scope, got scopespec",
		},
		{
			name:    "scope",
			input:   scopeAddr.String(),
			expAddr: scopeAddr,
			expHRP:  PrefixScope,
		},
		{
			name:    "session",
			input:   sessionAddr.String(),
			expAddr: sessionAddr,
			expHRP:  PrefixSession,
		},
		{
			name:    "record",
			input:   recordAddr.String(),
			expAddr: recordAddr,
			expHRP:  PrefixRecord,
		},
		{
			name:    "scope spec",
			input:   scopeSpecAddr.String(),
			expAddr: scopeSpecAddr,
			expHRP:  PrefixScopeSpecification,
		},
		{
			name:    "contract spec",
			input:   contractSpecAddr.String(),
			expAddr: contractSpecAddr,
			expHRP:  PrefixContractSpecification,
		},
		{
			name:    "record spec",
			input:   recordSpecAddr.String(),
			expAddr: recordSpecAddr,
			expHRP:  PrefixRecordSpecification,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			addr, hrp, err := ParseMetadataAddressFromBech32(tc.input)
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "ParseMetadataAddressFromBech32 error")
			} else {
				s.Assert().NoError(err, "ParseMetadataAddressFromBech32 error")
			}
			s.Assert().Equal(tc.expAddr, addr, "ParseMetadataAddressFromBech32 address")
			s.Assert().Equal(tc.expHRP, hrp, "ParseMetadataAddressFromBech32 HRP")

			addr, err = MetadataAddressFromBech32(tc.input)
			if len(tc.expErr) > 0 {
				s.Assert().EqualError(err, tc.expErr, "MetadataAddressFromBech32 error")
			} else {
				s.Assert().NoError(err, "MetadataAddressFromBech32 error")
			}
			s.Assert().Equal(tc.expAddr, addr, "MetadataAddressFromBech32 address")
		})
	}
}

func (s *AddressTestSuite) TestMetadataAddressFromDenom() {
	newUUID := func(name string, i int) uuid.UUID {
		bz := []byte(fmt.Sprintf("%s[%d]________________", name, i))[:16]
		rv, err := uuid.FromBytes(bz)
		s.Require().NoError(err, "%s[%d]: uuid.FromBytes(%v)", name, i, bz)
		return rv
	}
	scopeID := ScopeMetadataAddress(newUUID("scope", 0))
	sessionID := SessionMetadataAddress(newUUID("session", 1), newUUID("session", 2))
	recordID := RecordMetadataAddress(newUUID("record", 3), "money1")
	scopeSpecID := ScopeSpecMetadataAddress(newUUID("scopespec", 4))
	contractSpecID := ContractSpecMetadataAddress(newUUID("contractspec", 5))
	recordSpecID := RecordSpecMetadataAddress(newUUID("recordspec", 6), "money2")

	tests := []struct {
		name    string
		denom   string
		expAddr MetadataAddress
		expErr  string
	}{
		{
			name:   "empty",
			denom:  "",
			expErr: "denom \"\" is not a MetadataAddress denom",
		},
		{
			name:   "non-medatadata denom",
			denom:  "nhash",
			expErr: "denom \"nhash\" is not a MetadataAddress denom",
		},
		{
			name:   "starts with nft without the slash",
			denom:  "nft" + scopeID.String(),
			expErr: "denom \"nft" + scopeID.String() + "\" is not a MetadataAddress denom",
		},
		{
			name:    "just the prefix",
			denom:   DenomPrefix,
			expAddr: nil,
			expErr:  "invalid metadata address in denom \"nft/\": empty address string is not allowed",
		},
		{
			name:   "invalid address",
			denom:  DenomPrefix + sdk.AccAddress("nope_nope_nope_nope_").String(),
			expErr: "invalid metadata address in denom \"nft/" + sdk.AccAddress("nope_nope_nope_nope_").String() + "\": invalid metadata address type: 110",
		},
		{
			name:   "just a scope id",
			denom:  scopeID.String(),
			expErr: "denom \"" + scopeID.String() + "\" is not a MetadataAddress denom",
		},
		{
			name:    "scope",
			denom:   scopeID.Denom(),
			expAddr: scopeID,
		},
		{
			name:    "session",
			denom:   sessionID.Denom(),
			expAddr: sessionID,
		},
		{
			name:    "record",
			denom:   recordID.Denom(),
			expAddr: recordID,
		},
		{
			name:    "scope spec",
			denom:   scopeSpecID.Denom(),
			expAddr: scopeSpecID,
		},
		{
			name:    "contract spec",
			denom:   contractSpecID.Denom(),
			expAddr: contractSpecID,
		},
		{
			name:    "record spec",
			denom:   recordSpecID.Denom(),
			expAddr: recordSpecID,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actAddr MetadataAddress
			var actErr error
			testFunc := func() {
				actAddr, actErr = MetadataAddressFromDenom(tc.denom)
			}
			s.Require().NotPanics(testFunc, "MetadataAddressFromDenom(%q)", tc.denom)
			assertions.AssertErrorValue(s.T(), actErr, tc.expErr, "error from MetadataAddressFromDenom(%q)", tc.denom)
			s.Assert().Equal(tc.expAddr, actAddr, "address from MetadataAddressFromDenom(%q)", tc.denom)
		})
	}
}

func (s *AddressTestSuite) TestMetadataAddressWithInvalidData() {
	t := s.T()

	addr, addrErr := sdk.AccAddressFromBech32("cosmos1zgp4n2yvrtxkj5zl6rzcf6phqg0gfzuf3v08r4")
	require.NoError(t, addrErr, "address parsing error")

	_, err := VerifyMetadataAddressFormat(addr)
	require.EqualValues(t, fmt.Errorf("invalid metadata address type: %d", addr[0]), err)

	scopeID := ScopeMetadataAddress(s.scopeUUID)
	padded := make([]byte, 20)
	length, err := scopeID.MarshalTo(padded)
	require.NoError(t, err, "must marshal to metadata address")
	require.EqualValues(t, 17, length)

	_, err = VerifyMetadataAddressFormat(padded)
	require.EqualValues(t, fmt.Errorf("incorrect address length (expected: %d, actual: %d)", 17, 20), err)

	_, err = MetadataAddressFromBech32("")
	require.EqualValues(t, errors.New("empty address string is not allowed"), err)

	_, err = MetadataAddressFromBech32("scope1qzxcpvj6czy5g354dews3nlruxjsahh")
	require.EqualValues(t, "decoding bech32 failed: invalid checksum (expected 57e9fl got xjsahh)", err.Error())

	_, err = MetadataAddressFromHex("")
	require.EqualValues(t, errors.New("address decode failed: must provide an address"), err)

	_, err = MetadataAddressFromHex(s.scopeHex + "!!BAD")
	require.EqualValues(t, hex.InvalidByteError(0x21), err)

	var testMarshal MetadataAddress
	err = testMarshal.UnmarshalJSON([]byte(s.scopeBech32 + "{bad}{json}"))
	require.Error(t, err)
	err = testMarshal.UnmarshalYAML([]byte(s.scopeBech32 + "\n{badyaml}"))
	require.Error(t, err)
	err = testMarshal.Unmarshal([]byte{})
	require.NoError(t, err)
	err = testMarshal.UnmarshalJSON([]byte("\"\""))
	require.NoError(t, err)
	err = testMarshal.UnmarshalYAML([]byte("\"\""))
	require.NoError(t, err)
}

func (s *AddressTestSuite) TestMetadataAddressMarshal() {
	t := s.T()

	var scopeID, newInstance MetadataAddress
	require.True(t, scopeID.Equals(newInstance), "two empty instances are equal")

	scopeID = ScopeMetadataAddress(s.scopeUUID)

	bz, err := scopeID.Marshal()
	require.NoError(t, err)
	require.Equal(t, 17, len(bz))
	require.EqualValues(t, bz, scopeID.Bytes())

	require.False(t, newInstance.IsScopeAddress())
	err = newInstance.Unmarshal(bz)
	require.NoError(t, err)
	require.True(t, newInstance.IsScopeAddress())

	require.EqualValues(t, scopeID, newInstance)
}

func (s *AddressTestSuite) TestCompare() {
	maEmpty := MetadataAddress{}
	ma1 := MetadataAddress("1")
	ma2 := MetadataAddress("2")
	ma11 := MetadataAddress("11")
	ma22 := MetadataAddress("22")

	tests := []struct {
		name     string
		base     MetadataAddress
		arg      MetadataAddress
		expected int
	}{
		{"maEmpty v maEmpty", maEmpty, maEmpty, 0},
		{"maEmpty v ma1", maEmpty, ma1, -1},
		{"maEmpty v ma2", maEmpty, ma2, -1},
		{"maEmpty v ma11", maEmpty, ma11, -1},
		{"maEmpty v ma22", maEmpty, ma22, -1},

		{"ma1 v maEmpty", ma1, maEmpty, 1},
		{"ma1 v ma1", ma1, ma1, 0},
		{"ma1 v ma2", ma1, ma2, -1},
		{"ma1 v ma11", ma1, ma11, -1},
		{"ma1 v ma22", ma1, ma22, -1},

		{"ma2 v maEmpty", ma2, maEmpty, 1},
		{"ma2 v ma1", ma2, ma1, 1},
		{"ma2 v ma2", ma2, ma2, 0},
		{"ma2 v ma11", ma2, ma11, 1},
		{"ma2 v ma22", ma2, ma22, -1},

		{"ma11 v maEmpty", ma11, maEmpty, 1},
		{"ma11 v ma1", ma11, ma1, 1},
		{"ma11 v ma2", ma11, ma2, -1},
		{"ma11 v ma11", ma11, ma11, 0},
		{"ma11 v ma22", ma11, ma22, -1},

		{"ma22 v maEmpty", ma22, maEmpty, 1},
		{"ma22 v ma1", ma22, ma1, 1},
		{"ma22 v ma2", ma22, ma2, 1},
		{"ma22 v ma11", ma22, ma11, 1},
		{"ma22 v ma22", ma22, ma22, 0},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			actual := test.base.Compare(test.arg)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func (s *AddressTestSuite) TestMetadataAddressIteratorPrefix() {
	t := s.T()

	var emptyID MetadataAddress
	bz, err := emptyID.ScopeSessionIteratorPrefix()
	assert.NoError(t, err, "empty address ScopeSessionIteratorPrefix error")
	assert.Equal(t, SessionKeyPrefix, bz, "empty address ScopeSessionIteratorPrefix value")
	bz, err = emptyID.ScopeRecordIteratorPrefix()
	assert.NoError(t, err, "empty address ScopeRecordIteratorPrefix error")
	assert.Equal(t, RecordKeyPrefix, bz, "empty address ScopeRecordIteratorPrefix value")
	bz, err = emptyID.ContractSpecRecordSpecIteratorPrefix()
	assert.NoError(t, err, "empty address ContractSpecRecordSpecIteratorPrefix error")
	assert.Equal(t, RecordSpecificationKeyPrefix, bz, "empty address ContractSpecRecordSpecIteratorPrefix value")

	scopeSpecID := ScopeSpecMetadataAddress(s.scopeUUID)
	bz, err = scopeSpecID.ScopeSessionIteratorPrefix()
	assert.EqualError(t, err, "this metadata address does not contain a scope uuid", "scope spec id ScopeSessionIteratorPrefix error message")
	assert.Equal(t, []byte{}, bz, "scope spec id ScopeSessionIteratorPrefix value")
	bz, err = scopeSpecID.ScopeRecordIteratorPrefix()
	assert.EqualError(t, err, "this metadata address does not contain a scope uuid", "scope spec id ScopeRecordIteratorPrefix error message")
	assert.Equal(t, []byte{}, bz, "scope spec id ScopeRecordIteratorPrefix value")

	scopeID := ScopeMetadataAddress(s.scopeUUID)
	bz, err = scopeID.ScopeSessionIteratorPrefix()
	require.NoError(t, err, "ScopeSessionIteratorPrefix error")
	require.Equal(t, 17, len(bz), "ScopeSessionIteratorPrefix length")
	require.Equal(t, SessionKeyPrefix[0], bz[0], "ScopeSessionIteratorPrefix first byte")
	bz, err = scopeID.ScopeRecordIteratorPrefix()
	require.NoError(t, err, "ScopeRecordIteratorPrefix err")
	require.Equal(t, 17, len(bz), "ScopeRecordIteratorPrefix length")
	require.Equal(t, RecordKeyPrefix[0], bz[0], "ScopeRecordIteratorPrefix first byte")

	contractSpecID := ContractSpecMetadataAddress(s.scopeUUID)
	bz, err = contractSpecID.ContractSpecRecordSpecIteratorPrefix()
	require.NoError(t, err, "ContractSpecRecordSpecIteratorPrefix error")
	require.Equal(t, 17, len(bz), "ContractSpecRecordSpecIteratorPrefix length")
	require.Equal(t, RecordSpecificationKeyPrefix[0], bz[0], "ContractSpecRecordSpecIteratorPrefix first byte")
}

func (s *AddressTestSuite) TestScopeMetadataAddress() {
	t := s.T()

	// Make an address instance for a scope uuid
	scopeID := ScopeMetadataAddress(s.scopeUUID)
	require.NoError(t, scopeID.Validate())
	require.True(t, scopeID.IsScopeAddress())
	// Verify we can get a bech32 string for the scope
	require.Equal(t, s.scopeBech32, scopeID.String())

	// Verify we can get the uuid back
	scopeAddrUUID, err := scopeID.ScopeUUID()
	require.NoError(t, err)
	require.Equal(t, "8d80b25a-c089-4446-956e-5d08cfe3e1a5", scopeAddrUUID.String())

	_, err = scopeID.SessionUUID()
	require.Error(t, fmt.Errorf("this metadata addresss does not contain a session uuid"), err)

	// Check the string formatter for the scopeID
	require.Equal(t, s.scopeBech32, fmt.Sprintf("%s", scopeID))
	require.Equal(t, s.scopeHex, fmt.Sprintf("%X", scopeID))

	// Ensure a second instance is equal to the first
	scopeID2 := ScopeMetadataAddress(s.scopeUUID)
	require.True(t, scopeID.Equals(scopeID2))
	require.False(t, scopeID.Equals(ScopeMetadataAddress(s.sessionUUID)))

	json, err := scopeID.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("\"%s\"", s.scopeBech32), string(json))

	yaml, err := scopeID.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, s.scopeBech32, yaml)

	var yamlAddress MetadataAddress
	yamlAddress.UnmarshalYAML([]byte(yaml.(string)))
	require.EqualValues(t, scopeID, yamlAddress)

	var jsonAddress MetadataAddress
	jsonAddress.UnmarshalJSON(json)
	require.EqualValues(t, scopeID, jsonAddress)
}

func (s *AddressTestSuite) TestSessionMetadataAddress() {
	t := s.T()

	// Construct a composite key for a session within a scope
	sessionAddress := SessionMetadataAddress(s.scopeUUID, s.sessionUUID)
	require.NoError(t, sessionAddress.Validate(), "expect a valid MetadataAddress for a session")
	require.True(t, sessionAddress.IsSessionAddress())

	scopeUUIDFromSessionID, err := sessionAddress.ScopeUUID()
	require.NoError(t, err, "there should be no errors getting a scope uuid from the session address")
	require.Equal(t, "8d80b25a-c089-4446-956e-5d08cfe3e1a5", scopeUUIDFromSessionID.String())

	sessionUUID, err := sessionAddress.SessionUUID()
	require.NoError(t, err, "there should be no error getting the session uuid from the session address")
	require.Equal(t, "c25c7bd4-c639-4367-a842-f64fa5fccc19", sessionUUID.String(), "the session uuid should be recoverable")

	require.Equal(t, s.sessionBech32, sessionAddress.String())

	json, err := sessionAddress.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("\"%s\"", s.sessionBech32), string(json))

	yaml, err := sessionAddress.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, s.sessionBech32, yaml)

	var yamlAddress MetadataAddress
	require.True(t, yamlAddress.Empty())
	yamlAddress.UnmarshalYAML([]byte(yaml.(string)))
	require.EqualValues(t, sessionAddress, yamlAddress)

	var jsonAddress MetadataAddress
	require.True(t, jsonAddress.Empty())
	jsonAddress.UnmarshalJSON(json)
	require.EqualValues(t, sessionAddress, jsonAddress)
}

func (s *AddressTestSuite) TestRecordMetadataAddress() {
	t := s.T()

	// Construct a composite key for a record within a scope
	scopeID := ScopeMetadataAddress(s.scopeUUID)
	recordID := RecordMetadataAddress(s.scopeUUID, "test")
	require.True(t, recordID.IsRecordAddress())
	require.Equal(t, s.recordBech32, recordID.String())
	recordAddress, err := MetadataAddressFromBech32(s.recordBech32)
	require.NoError(t, err)
	require.Equal(t, recordID, recordAddress)

	require.Equal(t, recordID, RecordMetadataAddress(s.scopeUUID, "tEst"))
	require.Equal(t, recordID, RecordMetadataAddress(s.scopeUUID, "TEST"))
	require.Equal(t, recordID, RecordMetadataAddress(s.scopeUUID, "   test   "))

	recAddrFromScopeID, err := scopeID.AsRecordAddress("test")
	require.NoError(t, err, "AsRecordAddress error")
	require.Equal(t, recordID, recAddrFromScopeID, "AsRecordAddress value")
}

func (s *AddressTestSuite) TestScopeSpecMetadataAddress() {
	t := s.T()

	scopeSpecUUID := uuid.New()
	scopeSpecID := ScopeSpecMetadataAddress(scopeSpecUUID)

	require.True(t, scopeSpecID.IsScopeSpecificationAddress(), "IsScopeSpecificationAddress")
	require.Equal(t, ScopeSpecificationKeyPrefix, scopeSpecID[0:1].Bytes(), "bytes[0]: the type bit")
	require.Equal(t, scopeSpecID[1:17].Bytes(), scopeSpecUUID[:], "bytes[1:17]: the scope spec uuid bytes")

	scopeSpecBech32 := scopeSpecID.String()
	scopeSpecIDFromBeck32, errBeck32 := MetadataAddressFromBech32(scopeSpecBech32)
	require.NoError(t, errBeck32, "error from MetadataAddressFromBech32")
	require.Equal(t, scopeSpecID, scopeSpecIDFromBeck32, "value from scopeSpecIDFromBeck32")

	scopeSpecUUIDFromScopeSpecId, errScopeSpecUUID := scopeSpecID.ScopeSpecUUID()
	require.NoError(t, errScopeSpecUUID, "error from ScopeSpecUUID")
	require.Equal(t, scopeSpecUUID, scopeSpecUUIDFromScopeSpecId, "value from ScopeSpecUUID")
}

func (s *AddressTestSuite) TestContractSpecMetadataAddress() {
	t := s.T()

	contractSpecUUID := uuid.New()
	contractSpecID := ContractSpecMetadataAddress(contractSpecUUID)

	require.True(t, contractSpecID.IsContractSpecificationAddress(), "IsContractSpecificationAddress")
	require.Equal(t, ContractSpecificationKeyPrefix, contractSpecID[0:1].Bytes(), "bytes[0]: the type bit")
	require.Equal(t, contractSpecID[1:17].Bytes(), contractSpecUUID[:], "bytes[1:17]: the contract spec uuid bytes")

	contractSpecBech32 := contractSpecID.String()
	contractSpecIDFromBeck32, errBeck32 := MetadataAddressFromBech32(contractSpecBech32)
	require.NoError(t, errBeck32, "error from MetadataAddressFromBech32")
	require.Equal(t, contractSpecID, contractSpecIDFromBeck32, "value from contractSpecIDFromBeck32")

	contractSpecUUIDFromContractSpecId, errContractSpecUUID := contractSpecID.ContractSpecUUID()
	require.NoError(t, errContractSpecUUID, "error from ContractSpecUUID")
	require.Equal(t, contractSpecUUID, contractSpecUUIDFromContractSpecId, "value from ContractSpecUUID")
}

func (s *AddressTestSuite) TestRecordSpecMetadataAddress() {
	t := s.T()

	contractSpecUUID := uuid.New()
	contractSpecID := ContractSpecMetadataAddress(contractSpecUUID)
	recordSpecID := RecordSpecMetadataAddress(contractSpecUUID, "myname")
	nameHash := sha256.Sum256([]byte("myname"))

	require.True(t, recordSpecID.IsRecordSpecificationAddress(), "IsRecordAddress")
	require.Equal(t, RecordSpecificationKeyPrefix, recordSpecID[0:1].Bytes(), "bytes[0]: the type bit")
	require.Equal(t, contractSpecID[1:17], recordSpecID[1:17], "bytes[1:17]: the contract spec id bytes")
	require.Equal(t, nameHash[0:16], recordSpecID[17:33].Bytes(), "bytes[17:33]: the hashed name")

	recordSpecBech32 := recordSpecID.String()
	recordSpecIDFromBeck32, errBeck32 := MetadataAddressFromBech32(recordSpecBech32)
	require.NoError(t, errBeck32, "error from MetadataAddressFromBech32")
	require.Equal(t, recordSpecID, recordSpecIDFromBeck32, "value from recordSpecIDFromBeck32")

	require.Equal(t, recordSpecID, RecordSpecMetadataAddress(contractSpecUUID, "MyName"), "camel case")
	require.Equal(t, recordSpecID, RecordSpecMetadataAddress(contractSpecUUID, "MYNAME"), "all caps")
	require.Equal(t, recordSpecID, RecordSpecMetadataAddress(contractSpecUUID, "   myname   "), "padded with spaces")

	recSpecIDFromContractSpec, recSpecIDFromContractSpecErr := contractSpecID.AsRecordSpecAddress("myname")
	require.NoError(t, recSpecIDFromContractSpecErr, "AsRecordSpecAddress error")
	require.Equal(t, recordSpecID, recSpecIDFromContractSpec, "AsRecordSpecAddress value")

	contractSpecUUIDFromRecordSpecId, errContractSpecUUID := recordSpecID.ContractSpecUUID()
	require.NoError(t, errContractSpecUUID, "error from ContractSpecUUID")
	require.Equal(t, contractSpecUUID, contractSpecUUIDFromRecordSpecId, "value from ContractSpecUUID")
}

func (s *AddressTestSuite) TestMetadataAddressTypeTestFuncs() {
	tests := []struct {
		name     string
		id       MetadataAddress
		expected [6]bool
	}{
		{
			"scope",
			ScopeMetadataAddress(uuid.New()),
			[6]bool{true, false, false, false, false, false},
		},
		{
			"session",
			SessionMetadataAddress(uuid.New(), uuid.New()),
			[6]bool{false, true, false, false, false, false},
		},
		{
			"record",
			RecordMetadataAddress(uuid.New(), "best ever"),
			[6]bool{false, false, true, false, false, false},
		},
		{
			"scope specification",
			ScopeSpecMetadataAddress(uuid.New()),
			[6]bool{false, false, false, true, false, false},
		},
		{
			"contract speficiation",
			ContractSpecMetadataAddress(uuid.New()),
			[6]bool{false, false, false, false, true, false},
		},
		{
			"record specification",
			RecordSpecMetadataAddress(uuid.New(), "okayest dad"),
			[6]bool{false, false, false, false, false, true},
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected[0], test.id.IsScopeAddress(), fmt.Sprintf("%s: IsScopeAddress", test.name))
			assert.Equal(t, test.expected[1], test.id.IsSessionAddress(), fmt.Sprintf("%s: IsSessionAddress", test.name))
			assert.Equal(t, test.expected[2], test.id.IsRecordAddress(), fmt.Sprintf("%s: IsRecordAddress", test.name))
			assert.Equal(t, test.expected[3], test.id.IsScopeSpecificationAddress(), fmt.Sprintf("%s: IsScopeSpecificationAddress", test.name))
			assert.Equal(t, test.expected[4], test.id.IsContractSpecificationAddress(), fmt.Sprintf("%s: IsContractSpecificationAddress", test.name))
			assert.Equal(t, test.expected[5], test.id.IsRecordSpecificationAddress(), fmt.Sprintf("%s: IsRecordSpecificationAddress", test.name))
		})
	}
}

func (s *AddressTestSuite) TestPrefix() {
	tests := []struct {
		name           string
		addr           MetadataAddress
		expectedValue  string
		expectedError  string
		expectAnyError bool // set to true if you expect an error, but don't care what the error string is.
	}{
		{
			"scope",
			ScopeMetadataAddress(uuid.New()),
			PrefixScope,
			"",
			false,
		},
		{
			"scope without address",
			MetadataAddress{ScopeKeyPrefix[0]},
			PrefixScope,
			"",
			false,
		},
		{
			"session",
			SessionMetadataAddress(uuid.New(), uuid.New()),
			PrefixSession,
			"",
			false,
		},
		{
			"session without address",
			MetadataAddress{SessionKeyPrefix[0]},
			PrefixSession,
			"",
			false,
		},
		{
			"record",
			RecordMetadataAddress(uuid.New(), "ronald"),
			PrefixRecord,
			"",
			false,
		},
		{
			"record without address",
			MetadataAddress{RecordKeyPrefix[0]},
			PrefixRecord,
			"",
			false,
		},
		{
			"scope spec",
			ScopeSpecMetadataAddress(uuid.New()),
			PrefixScopeSpecification,
			"",
			false,
		},
		{
			"scope spec without address",
			MetadataAddress{ScopeSpecificationKeyPrefix[0]},
			PrefixScopeSpecification,
			"",
			false,
		},
		{
			"contract spec",
			ContractSpecMetadataAddress(uuid.New()),
			PrefixContractSpecification,
			"",
			false,
		},
		{
			"contract spec without address",
			MetadataAddress{ContractSpecificationKeyPrefix[0]},
			PrefixContractSpecification,
			"",
			false,
		},
		{
			"record spec",
			RecordSpecMetadataAddress(uuid.New(), "george"),
			PrefixRecordSpecification,
			"",
			false,
		},
		{
			"record spec without address",
			MetadataAddress{RecordSpecificationKeyPrefix[0]},
			PrefixRecordSpecification,
			"",
			false,
		},
		{
			"empty",
			MetadataAddress{},
			"",
			"address is empty",
			false,
		},
		{
			"bad",
			MetadataAddress("don't do this"),
			"",
			"",
			true,
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			actual, err := tc.addr.Prefix()
			assert.Equal(t, tc.expectedValue, actual, "Prefix value")
			if len(tc.expectedError) > 0 {
				assert.EqualError(t, err, tc.expectedError, "Prefix error string")
			} else if tc.expectAnyError {
				assert.Error(t, err, "Prefix error (any)")
			} else {
				assert.NoError(t, err, "Prefix error")
			}
		})
	}
}

func (s *AddressTestSuite) TestPrimaryUUID() {
	scopePrimaryUUID := uuid.New()
	sessionPrimaryUUID := uuid.New()
	recordPrimaryUUID := uuid.New()
	scopeSpecPrimaryUUID := uuid.New()
	contractSpecPrimaryUUID := uuid.New()
	recordSpecPrimaryUUID := uuid.New()

	tests := []struct {
		name          string
		id            MetadataAddress
		expectedValue uuid.UUID
		expectedError string
	}{
		{
			"scope",
			ScopeMetadataAddress(scopePrimaryUUID),
			scopePrimaryUUID,
			"",
		},
		{
			"session",
			SessionMetadataAddress(sessionPrimaryUUID, uuid.New()),
			sessionPrimaryUUID,
			"",
		},
		{
			"record",
			RecordMetadataAddress(recordPrimaryUUID, "presence"),
			recordPrimaryUUID,
			"",
		},
		{
			"scope spec",
			ScopeSpecMetadataAddress(scopeSpecPrimaryUUID),
			scopeSpecPrimaryUUID,
			"",
		},
		{
			"contract spec",
			ContractSpecMetadataAddress(contractSpecPrimaryUUID),
			contractSpecPrimaryUUID,
			"",
		},
		{
			"record spec",
			RecordSpecMetadataAddress(recordSpecPrimaryUUID, "profiled"),
			recordSpecPrimaryUUID,
			"",
		},
		{
			"empty",
			MetadataAddress{},
			uuid.UUID{},
			"address empty",
		},
		{
			"too short",
			MetadataAddress(ScopeMetadataAddress(scopePrimaryUUID).Bytes()[0:16]),
			uuid.UUID{},
			"incorrect address length (must be at least 17, actual: 16)",
		},
		{
			"unknown type",
			MetadataAddress("This is not how this works."),
			uuid.UUID{},
			"invalid address type out of valid range (got: 84)",
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			primaryUUID, err := test.id.PrimaryUUID()
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, fmt.Sprintf("%s: err", test.name))
				assert.Equal(t, test.expectedValue, primaryUUID, fmt.Sprintf("%s: value", test.name))
			} else {
				assert.EqualError(t, err, test.expectedError, fmt.Sprintf("%s: err", test.name))
			}
		})
	}
}

func (s *AddressTestSuite) TestSecondaryUUID() {
	sessionSecondaryUUID := uuid.New()

	tests := []struct {
		name          string
		id            MetadataAddress
		expectedValue uuid.UUID
		expectedError string
	}{
		{
			"scope",
			ScopeMetadataAddress(uuid.New()),
			uuid.UUID{},
			"invalid address type out of valid range (got: 0)",
		},
		{
			"session",
			SessionMetadataAddress(uuid.New(), sessionSecondaryUUID),
			sessionSecondaryUUID,
			"",
		},
		{
			"record",
			RecordMetadataAddress(uuid.New(), "presence"),
			uuid.UUID{},
			"invalid address type out of valid range (got: 2)",
		},
		{
			"scope spec",
			ScopeSpecMetadataAddress(uuid.New()),
			uuid.UUID{},
			"invalid address type out of valid range (got: 4)",
		},
		{
			"contract spec",
			ContractSpecMetadataAddress(uuid.New()),
			uuid.UUID{},
			"invalid address type out of valid range (got: 3)",
		},
		{
			"record spec",
			RecordSpecMetadataAddress(uuid.New(), "profiled"),
			uuid.UUID{},
			"invalid address type out of valid range (got: 5)",
		},
		{
			"empty",
			MetadataAddress{},
			uuid.UUID{},
			"address empty",
		},
		{
			"too short",
			MetadataAddress(SessionMetadataAddress(uuid.New(), sessionSecondaryUUID).Bytes()[0:32]),
			uuid.UUID{},
			"incorrect address length (must be at least 33, actual: 32)",
		},
		{
			"unknown type",
			MetadataAddress("This is not how this works."),
			uuid.UUID{},
			"invalid address type out of valid range (got: 84)",
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			secondaryUUID, err := test.id.SecondaryUUID()
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, fmt.Sprintf("%s: err", test.name))
				assert.Equal(t, test.expectedValue, secondaryUUID, fmt.Sprintf("%s: value", test.name))
			} else {
				assert.EqualError(t, err, test.expectedError, fmt.Sprintf("%s: err", test.name))
			}
		})
	}
}

func (s *AddressTestSuite) TestNameHash() {
	recordName := "mothership"
	recordNameHash := sha256.Sum256([]byte(recordName))
	recordSpecName := "houses of the holy"
	recordSpecNameHash := sha256.Sum256([]byte(recordSpecName))

	tests := []struct {
		name          string
		id            MetadataAddress
		expectedValue []byte
		expectedError string
	}{
		{
			"scope",
			ScopeMetadataAddress(uuid.New()),
			[]byte{},
			"invalid address type out of valid range (got: 0)",
		},
		{
			"session",
			SessionMetadataAddress(uuid.New(), uuid.New()),
			[]byte{},
			"invalid address type out of valid range (got: 1)",
		},
		{
			"record",
			RecordMetadataAddress(uuid.New(), recordName),
			recordNameHash[0:16],
			"",
		},
		{
			"scope spec",
			ScopeSpecMetadataAddress(uuid.New()),
			[]byte{},
			"invalid address type out of valid range (got: 4)",
		},
		{
			"contract spec",
			ContractSpecMetadataAddress(uuid.New()),
			[]byte{},
			"invalid address type out of valid range (got: 3)",
		},
		{
			"record spec",
			RecordSpecMetadataAddress(uuid.New(), recordSpecName),
			recordSpecNameHash[0:16],
			"",
		},
		{
			"empty",
			MetadataAddress{},
			[]byte{},
			"address empty",
		},
		{
			"too short",
			MetadataAddress(RecordSpecMetadataAddress(uuid.New(), recordSpecName).Bytes()[0:32]),
			[]byte{},
			"incorrect address length (must be at least 33, actual: 32)",
		},
		{
			"unknown type",
			MetadataAddress("This is not how this works."),
			[]byte{},
			"invalid address type out of valid range (got: 84)",
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			nameHash, err := test.id.NameHash()
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, fmt.Sprintf("%s: err", test.name))
				assert.Equal(t, test.expectedValue, nameHash, fmt.Sprintf("%s: value", test.name))
			} else {
				assert.EqualError(t, err, test.expectedError, fmt.Sprintf("%s: err", test.name))
			}
		})
	}
}

func (s *AddressTestSuite) TestScopeAddressConverters() {
	randomUUID := uuid.New()
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, uuid.New())
	recordID := RecordMetadataAddress(randomUUID, "year zero")
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, "the downard spiral")

	tests := []struct {
		name          string
		baseID        MetadataAddress
		expectedID    MetadataAddress
		expectedError string
	}{
		{
			"scope id",
			scopeID,
			scopeID,
			"",
		},
		{
			"session id",
			sessionID,
			scopeID,
			"",
		},
		{
			"record id",
			recordID,
			scopeID,
			"",
		},
		{
			"scope spec id",
			scopeSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				scopeSpecID),
		},
		{
			"contract spec id",
			contractSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				contractSpecID),
		},
		{
			"record spec id",
			recordSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				recordSpecID),
		},
	}

	for _, test := range tests {
		s.T().Run(fmt.Sprintf("%s AsScopeAddress", test.name), func(t *testing.T) {
			actualID, err := test.baseID.AsScopeAddress()
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, "%s AsScopeAddress err", test.name)
				assert.Equal(t, test.expectedID, actualID, "%s AsScopeAddress value", test.name)
			} else {
				assert.EqualError(t, err, test.expectedError, "%s AsScopeAddress expected err", test.name)
			}
		})
		s.T().Run(fmt.Sprintf("%s MustGetAsScopeAddress", test.name), func(t *testing.T) {
			if len(test.expectedError) == 0 {
				assert.NotPanics(t, func() {
					actualID := test.baseID.MustGetAsScopeAddress()
					assert.Equal(t, test.expectedID, actualID, "%s MustGetAsScopeAddress value", test.name)
				}, "%s MustGetAsScopeAddress unexpected panic", test.name)
			} else {
				assert.PanicsWithError(t, test.expectedError, func() {
					_ = test.baseID.MustGetAsScopeAddress()
				}, "%s MustGetAsScopeAddress expected panic", test.name)
			}
		})
	}
}

func (s *AddressTestSuite) TestSessionAddressConverters() {
	randomUUID := uuid.New()
	randomUUID2 := uuid.New()
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, randomUUID2)
	recordID := RecordMetadataAddress(randomUUID, "pet sounds")
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, "smile")

	tests := []struct {
		name          string
		baseID        MetadataAddress
		expectedID    MetadataAddress
		expectedError string
	}{
		{
			"scope id",
			scopeID,
			sessionID,
			"",
		},
		{
			"session id",
			sessionID,
			sessionID,
			"",
		},
		{
			"record id",
			recordID,
			sessionID,
			"",
		},
		{
			"scope spec id",
			scopeSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				scopeSpecID),
		},
		{
			"contract spec id",
			contractSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				contractSpecID),
		},
		{
			"record spec id",
			recordSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				recordSpecID),
		},
	}

	for _, test := range tests {
		s.T().Run(fmt.Sprintf("%s AsSessionAddress", test.name), func(t *testing.T) {
			actualID, err := test.baseID.AsSessionAddress(randomUUID2)
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, "%s AsSessionAddress err", test.name)
				assert.Equal(t, test.expectedID, actualID, "%s AsSessionAddress value", test.name)
			} else {
				assert.EqualError(t, err, test.expectedError, "%s AsSessionAddress expected err", test.name)
			}
		})
		s.T().Run(fmt.Sprintf("%s MustGetAsSessionAddress", test.name), func(t *testing.T) {
			if len(test.expectedError) == 0 {
				assert.NotPanics(t, func() {
					actualID := test.baseID.MustGetAsSessionAddress(randomUUID2)
					assert.Equal(t, test.expectedID, actualID, "%s MustGetAsSessionAddress value", test.name)
				}, "%s MustGetAsSessionAddress unexpected panic", test.name)
			} else {
				assert.PanicsWithError(t, test.expectedError, func() {
					_ = test.baseID.MustGetAsSessionAddress(randomUUID2)
				}, "%s MustGetAsSessionAddress expected panic", test.name)
			}
		})
	}
}

func (s *AddressTestSuite) TestRecordAddressConverters() {
	randomUUID := uuid.New()
	recordName := "the fragile"
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, uuid.New())
	recordID := RecordMetadataAddress(randomUUID, recordName)
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, recordName)

	recordNameVersions := []string{
		recordName,
		"  " + recordName,
		recordName + "  ",
		"  " + recordName + "  ",
	}

	tests := []struct {
		name          string
		baseID        MetadataAddress
		expectedID    MetadataAddress
		expectedError string
	}{
		{
			"scope id",
			scopeID,
			recordID,
			"",
		},
		{
			"session id",
			sessionID,
			recordID,
			"",
		},
		{
			"record id",
			recordID,
			recordID,
			"",
		},
		{
			"scope spec id",
			scopeSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				scopeSpecID),
		},
		{
			"contract spec id",
			contractSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				contractSpecID),
		},
		{
			"record spec id",
			recordSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				recordSpecID),
		},
	}

	for _, test := range tests {
		for _, rName := range recordNameVersions {
			s.T().Run(fmt.Sprintf("%s AsRecordAddress(\"%s\")", test.name, rName), func(t *testing.T) {
				actualID, err := test.baseID.AsRecordAddress(rName)
				if len(test.expectedError) == 0 {
					assert.NoError(t, err, "%s AsRecordAddress err", test.name)
					assert.Equal(t, test.expectedID, actualID, "%s AsRecordAddress value", test.name)
				} else {
					assert.EqualError(t, err, test.expectedError, "%s AsRecordAddress expected err", test.name)
				}
			})
			s.T().Run(fmt.Sprintf("%s MustGetAsRecordAddress(\"%s\")", test.name, rName), func(t *testing.T) {
				if len(test.expectedError) == 0 {
					assert.NotPanics(t, func() {
						actualID := test.baseID.MustGetAsRecordAddress(rName)
						assert.Equal(t, test.expectedID, actualID, "%s MustGetAsRecordAddress value", test.name)
					}, "%s MustGetAsRecordAddress unexpected panic", test.name)
				} else {
					assert.PanicsWithError(t, test.expectedError, func() {
						_ = test.baseID.MustGetAsRecordAddress(rName)
					}, "%s MustGetAsRecordAddress expected panic", test.name)
				}
			})
		}
	}
}

func (s *AddressTestSuite) TestRecordSpecAddressConverters() {
	randomUUID := uuid.New()
	recordName := "bad witch"
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, uuid.New())
	recordID := RecordMetadataAddress(randomUUID, recordName)
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, recordName)

	recordNameVersions := []string{
		recordName,
		"  " + recordName,
		recordName + "  ",
		"  " + recordName + "  ",
	}

	tests := []struct {
		name          string
		baseID        MetadataAddress
		expectedID    MetadataAddress
		expectedError string
	}{
		{
			"scope id",
			scopeID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeID),
		},
		{
			"session id",
			sessionID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				sessionID),
		},
		{
			"record id",
			recordID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				recordID),
		},
		{
			"scope spec id",
			scopeSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeSpecID),
		},
		{
			"contract spec id",
			contractSpecID,
			recordSpecID,
			"",
		},
		{
			"record spec id",
			recordSpecID,
			recordSpecID,
			"",
		},
	}

	for _, test := range tests {
		for _, rName := range recordNameVersions {
			s.T().Run(fmt.Sprintf("%s AsRecordSpecAddress(\"%s\")", test.name, rName), func(t *testing.T) {
				actualID, err := test.baseID.AsRecordSpecAddress(rName)
				if len(test.expectedError) == 0 {
					assert.NoError(t, err, "%s AsRecordSpecAddress err", test.name)
					assert.Equal(t, test.expectedID, actualID, "%s AsRecordSpecAddress value", test.name)
				} else {
					assert.EqualError(t, err, test.expectedError, "%s AsRecordSpecAddress expected err", test.name)
				}
			})
			s.T().Run(fmt.Sprintf("%s MustGetAsRecordSpecAddress(\"%s\")", test.name, rName), func(t *testing.T) {
				if len(test.expectedError) == 0 {
					assert.NotPanics(t, func() {
						actualID := test.baseID.MustGetAsRecordSpecAddress(rName)
						assert.Equal(t, test.expectedID, actualID, "%s MustGetAsRecordSpecAddress value", test.name)
					}, "%s MustGetAsRecordSpecAddress unexpected panic", test.name)
				} else {
					assert.PanicsWithError(t, test.expectedError, func() {
						_ = test.baseID.MustGetAsRecordSpecAddress(rName)
					}, "%s MustGetAsRecordSpecAddress expected panic", test.name)
				}
			})
		}
	}
}

func (s *AddressTestSuite) TestContractSpecAddressConverters() {
	randomUUID := uuid.New()
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, uuid.New())
	recordID := RecordMetadataAddress(randomUUID, "pretty hate machine")
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, "with teeth")

	tests := []struct {
		name          string
		baseID        MetadataAddress
		expectedID    MetadataAddress
		expectedError string
	}{
		{
			"scope id",
			scopeID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeID),
		},
		{
			"session id",
			sessionID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				sessionID),
		},
		{
			"record id",
			recordID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				recordID),
		},
		{
			"scope spec id",
			scopeSpecID,
			MetadataAddress{},
			fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeSpecID),
		},
		{
			"contract spec id",
			contractSpecID,
			contractSpecID,
			"",
		},
		{
			"record spec id",
			recordSpecID,
			contractSpecID,
			"",
		},
	}

	for _, test := range tests {
		s.T().Run(fmt.Sprintf("%s AsContractSpecAddress", test.name), func(t *testing.T) {
			actualID, err := test.baseID.AsContractSpecAddress()
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, "%s AsContractSpecAddress err", test.name)
				assert.Equal(t, test.expectedID, actualID, "%s AsContractSpecAddress value", test.name)
			} else {
				assert.EqualError(t, err, test.expectedError, "%s AsContractSpecAddress expected err", test.name)
			}
		})
		s.T().Run(fmt.Sprintf("%s MustGetAsContractSpecAddress", test.name), func(t *testing.T) {
			if len(test.expectedError) == 0 {
				assert.NotPanics(t, func() {
					actualID := test.baseID.MustGetAsContractSpecAddress()
					assert.Equal(t, test.expectedID, actualID, "%s MustGetAsContractSpecAddress value", test.name)
				}, "%s MustGetAsContractSpecAddress unexpected panic", test.name)
			} else {
				assert.PanicsWithError(t, test.expectedError, func() {
					_ = test.baseID.MustGetAsContractSpecAddress()
				}, "%s MustGetAsContractSpecAddress expected panic", test.name)
			}
		})
	}
}

// mockState satisfies the fmt.State interface, but always returns an error from Write, and doesn't do anything else.
type mockState struct {
	err string
}

var _ fmt.State = (*mockState)(nil)

func (s mockState) Write(b []byte) (n int, err error) {
	return 0, errors.New(s.err)
}

func (s mockState) Width() (int, bool) {
	return 0, false
}

func (s mockState) Precision() (int, bool) {
	return 0, false
}

func (s mockState) Flag(c int) bool {
	return false
}

func (s *AddressTestSuite) TestFormat() {
	someUUIDStr := "97263339-CFAA-41D9-809E-82CD78C84F02"
	someUUID, err := uuid.Parse(someUUIDStr)
	s.Require().NoError(err, "uuid.Parse(%q)", someUUIDStr)

	type namedMetadataAddress struct {
		name string
		id   MetadataAddress
	}

	scopeID := namedMetadataAddress{name: "scope", id: ScopeMetadataAddress(someUUID)}
	contractSpecID := namedMetadataAddress{name: "contract spec", id: ContractSpecMetadataAddress(someUUID)}
	emptyID := namedMetadataAddress{name: "empty", id: MetadataAddress{}}
	nilID := namedMetadataAddress{name: "nil", id: nil}
	invalidID := namedMetadataAddress{name: "invalid", id: MetadataAddress("do not create MetadataAddresses this way")}
	expInvID := "MetadataAddress{0x64, 0x6f, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x20, " +
		"0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x20, " +
		"0x74, 0x68, 0x69, 0x73, 0x20, 0x77, 0x61, 0x79}"

	tests := []struct {
		id  namedMetadataAddress
		fmt string
		exp string
	}{
		{id: scopeID, fmt: "%s", exp: "scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"},
		{id: scopeID, fmt: "%20s", exp: "scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"},
		{id: scopeID, fmt: "%-20s", exp: "scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"},
		{id: scopeID, fmt: "%50s", exp: "          scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"},
		{id: scopeID, fmt: "%-50s", exp: "scope1qztjvveee74yrkvqn6pv67xgfupqyumx55          "},
		{id: scopeID, fmt: "%q", exp: `"scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"`},
		{id: scopeID, fmt: "%20q", exp: `"scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"`},
		{id: scopeID, fmt: "%-20q", exp: `"scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"`},
		{id: scopeID, fmt: "%50q", exp: `        "scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"`},
		{id: scopeID, fmt: "%-50q", exp: `"scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"        `},
		{id: scopeID, fmt: "%v", exp: "scope1qztjvveee74yrkvqn6pv67xgfupqyumx55"},
		{id: scopeID, fmt: "%#v", exp: "MetadataAddress{0x0, 0x97, 0x26, 0x33, 0x39, 0xcf, 0xaa, 0x41, 0xd9, 0x80, 0x9e, 0x82, 0xcd, 0x78, 0xc8, 0x4f, 0x2}"},
		{id: scopeID, fmt: "%p", exp: fmt.Sprintf("%p", []byte(scopeID.id))},   // e.g. 0x14000d95818
		{id: scopeID, fmt: "%#p", exp: fmt.Sprintf("%#p", []byte(scopeID.id))}, // e.g. 14000d95818
		{id: scopeID, fmt: "%T", exp: "types.MetadataAddress"},
		{id: scopeID, fmt: "%d", exp: "[0 151 38 51 57 207 170 65 217 128 158 130 205 120 200 79 2]"},
		{id: scopeID, fmt: "%x", exp: "0097263339cfaa41d9809e82cd78c84f02"},
		{id: scopeID, fmt: "%#x", exp: "0x0097263339cfaa41d9809e82cd78c84f02"},
		{id: scopeID, fmt: "%X", exp: "0097263339CFAA41D9809E82CD78C84F02"},
		{id: scopeID, fmt: "%#X", exp: "0X0097263339CFAA41D9809E82CD78C84F02"},
		{id: contractSpecID, fmt: "%s", exp: "contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"},
		{id: contractSpecID, fmt: "%20s", exp: "contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"},
		{id: contractSpecID, fmt: "%-20s", exp: "contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"},
		{id: contractSpecID, fmt: "%50s", exp: "   contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"},
		{id: contractSpecID, fmt: "%-50s", exp: "contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh   "},
		{id: contractSpecID, fmt: "%q", exp: `"contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"`},
		{id: contractSpecID, fmt: "%20q", exp: `"contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"`},
		{id: contractSpecID, fmt: "%-20q", exp: `"contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"`},
		{id: contractSpecID, fmt: "%50q", exp: ` "contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"`},
		{id: contractSpecID, fmt: "%-50q", exp: `"contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh" `},
		{id: contractSpecID, fmt: "%v", exp: "contractspec1qwtjvveee74yrkvqn6pv67xgfupqghravh"},
		{id: contractSpecID, fmt: "%#v", exp: "MetadataAddress{0x3, 0x97, 0x26, 0x33, 0x39, 0xcf, 0xaa, 0x41, 0xd9, 0x80, 0x9e, 0x82, 0xcd, 0x78, 0xc8, 0x4f, 0x2}"},
		{id: contractSpecID, fmt: "%p", exp: fmt.Sprintf("%p", []byte(contractSpecID.id))},   // e.g. 0x14000d95818
		{id: contractSpecID, fmt: "%#p", exp: fmt.Sprintf("%#p", []byte(contractSpecID.id))}, // e.g. 14000d95818
		{id: contractSpecID, fmt: "%T", exp: "types.MetadataAddress"},
		{id: contractSpecID, fmt: "%d", exp: "[3 151 38 51 57 207 170 65 217 128 158 130 205 120 200 79 2]"},
		{id: contractSpecID, fmt: "%x", exp: "0397263339cfaa41d9809e82cd78c84f02"},
		{id: contractSpecID, fmt: "%#x", exp: "0x0397263339cfaa41d9809e82cd78c84f02"},
		{id: contractSpecID, fmt: "%X", exp: "0397263339CFAA41D9809E82CD78C84F02"},
		{id: contractSpecID, fmt: "%#X", exp: "0X0397263339CFAA41D9809E82CD78C84F02"},
		{id: emptyID, fmt: "%s", exp: ""},
		{id: emptyID, fmt: "%q", exp: `""`},
		{id: emptyID, fmt: "%v", exp: ""},
		{id: emptyID, fmt: "%#v", exp: "MetadataAddress{}"},
		{id: emptyID, fmt: "%T", exp: "types.MetadataAddress"},
		{id: emptyID, fmt: "%x", exp: ""},
		{id: nilID, fmt: "%s", exp: ""},
		{id: nilID, fmt: "%q", exp: `""`},
		{id: nilID, fmt: "%v", exp: ""},
		{id: nilID, fmt: "%#v", exp: "MetadataAddress(nil)"},
		{id: nilID, fmt: "%T", exp: "types.MetadataAddress"},
		{id: nilID, fmt: "%x", exp: ""},
		{id: invalidID, fmt: "%s", exp: expInvID},
		{id: invalidID, fmt: "%q", exp: `"` + expInvID + `"`},
		{id: invalidID, fmt: "%v", exp: expInvID},
		{id: invalidID, fmt: "%#v", exp: expInvID},
		{id: invalidID, fmt: "%T", exp: "types.MetadataAddress"},
		{id: invalidID, fmt: "%x", exp: "646f206e6f7420637265617465204d65746164617461416464726573736573207468697320776179"},
	}

	for _, test := range tests {
		s.Run(test.id.name+" "+test.fmt, func() {
			var actual string
			testFunc := func() {
				actual = fmt.Sprintf(test.fmt, test.id.id)
			}
			s.Require().NotPanics(testFunc, "Sprintf(%q, ...)", test.fmt)
			s.Assert().Equal(test.exp, actual)
		})
	}

	s.Run("write error", func() {
		expPanic := "injected write error"
		state := &mockState{err: expPanic}
		verb := 's'
		addr := ScopeMetadataAddress(s.scopeUUID)
		testFunc := func() {
			addr.Format(state, verb)
		}
		s.Require().PanicsWithError(expPanic, testFunc, "Format")
	})
}

func (s *AddressTestSuite) TestValidateIsTypeAddressFuncs() {
	newUUID := func(name string, i int) uuid.UUID {
		bz := []byte(fmt.Sprintf("%s[%d]________________", name, i))[:16]
		rv, err := uuid.FromBytes(bz)
		s.Require().NoError(err, "%s[%d]: uuid.FromBytes(%q)", name, i, bz)
		return rv
	}

	funcsToTest := []struct {
		name      string
		hrp       string
		validator func(ma MetadataAddress) error
	}{
		{
			name:      "ValidateIsScopeAddress",
			hrp:       PrefixScope,
			validator: MetadataAddress.ValidateIsScopeAddress,
		},
		{
			name:      "ValidateIsScopeSpecificationAddress",
			hrp:       PrefixScopeSpecification,
			validator: MetadataAddress.ValidateIsScopeSpecificationAddress,
		},
	}

	addrsToTest := []struct {
		name       string
		ma         MetadataAddress
		expInvalid string
	}{
		{
			name:       "nil",
			ma:         nil,
			expInvalid: "address is empty",
		},
		{
			name:       "empty",
			ma:         MetadataAddress{},
			expInvalid: "address is empty",
		},
		{
			name:       "unknown type",
			ma:         MetadataAddress{0xa0, 0x1, 0x2, 0xff},
			expInvalid: "invalid metadata address type: 160",
		},
		{
			name:       "invalid scope",
			ma:         MetadataAddress{ScopeKeyPrefix[0], 0x1, 0x2, 0xff},
			expInvalid: "incorrect address length (expected: 17, actual: 4)",
		},
		{name: PrefixScope, ma: ScopeMetadataAddress(newUUID("scope", 1))},
		{name: PrefixSession, ma: SessionMetadataAddress(newUUID("session", 2), newUUID("session", 3))},
		{name: PrefixRecord, ma: RecordMetadataAddress(newUUID("record", 4), "bananas")},
		{name: PrefixScopeSpecification, ma: ScopeSpecMetadataAddress(newUUID("scopespec", 5))},
		{name: PrefixContractSpecification, ma: ContractSpecMetadataAddress(newUUID("scopespec", 6))},
		{name: PrefixRecordSpecification, ma: RecordSpecMetadataAddress(newUUID("recordspec", 7), "alsobananas")},
	}

	for _, funcDef := range funcsToTest {
		typeName := getNameForHRP(funcDef.hrp)
		s.Run(funcDef.name, func() {
			for _, addrDef := range addrsToTest {
				s.Run(addrDef.name, func() {
					var expErr string
					switch {
					case len(addrDef.expInvalid) > 0:
						expErr = fmt.Sprintf("invalid %s metadata address %#v: %s", typeName, addrDef.ma, addrDef.expInvalid)
					case funcDef.hrp != addrDef.name:
						expErr = fmt.Sprintf("invalid %s id %q: wrong type", typeName, addrDef.ma)
					}

					var err error
					testFunc := func() {
						err = funcDef.validator(addrDef.ma)
					}
					s.Require().NotPanics(testFunc, "%#v.%s()", addrDef.ma, funcDef.name)
					assertions.AssertErrorValue(s.T(), err, expErr, "%#v.%s()", addrDef.ma, funcDef.name)
				})
			}
		})
	}
}

func (s *AddressTestSuite) TestGenerateExamples() {
	// This "test" doesn't actually test anything. It just generates some output that's helpful
	// for validating other MetadataAddress implementations.

	scopeUUIDString := "91978ba2-5f35-459a-86a7-feca1b0512e0"
	sessionUUIDString := "5803f8bc-6067-4eb5-951f-2121671c2ec0"
	scopeSpecUUIDString := "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"
	contractSpecUUIDString := "def6bc0a-c9dd-4874-948f-5206e6060a84"
	recordName := "recordname"

	scopeUUID := uuid.MustParse(scopeUUIDString)
	sessionUUID := uuid.MustParse(sessionUUIDString)
	scopeSpecUUID := uuid.MustParse(scopeSpecUUIDString)
	contractSpecUUID := uuid.MustParse(contractSpecUUIDString)

	scopeId := ScopeMetadataAddress(scopeUUID)
	sessionId := SessionMetadataAddress(scopeUUID, sessionUUID)
	recordId := RecordMetadataAddress(scopeUUID, recordName)
	scopeSpecId := ScopeSpecMetadataAddress(scopeSpecUUID)
	contractSpecId := ContractSpecMetadataAddress(contractSpecUUID)
	recordSpecId := RecordSpecMetadataAddress(contractSpecUUID, recordName)
	recordNameBytes, _ := recordId.NameHash()

	fmt.Printf("        scope uuid: \"%s\"\n", scopeUUIDString)
	fmt.Printf("      session uuid: \"%s\"\n", sessionUUIDString)
	fmt.Printf("   scope spec uuid: \"%s\"\n", scopeSpecUUIDString)
	fmt.Printf("contract spec uuid: \"%s\"\n", contractSpecUUIDString)
	fmt.Printf("       record name: \"%s\"\n", recordName)
	fmt.Printf("hashed record name bytes: %v\n", recordNameBytes)
	fmt.Printf("       scope id: \"%s\"\n", scopeId)
	fmt.Printf("     session id: \"%s\"\n", sessionId)
	fmt.Printf("      record id: \"%s\"\n", recordId)
	fmt.Printf("  scope spec id: \"%s\"\n", scopeSpecId)
	fmt.Printf("contrat spec id: \"%s\"\n", contractSpecId)
	fmt.Printf(" record spec id: \"%s\"\n", recordSpecId)

	s.Assert().True(true)
}

// TODO: GetDetails tests.

func (s *AddressTestSuite) TestDenom() {
	// As of writing this, the only metadata type that we should be making denoms for are scopes.
	// However, I figured that restriction would be better left higher up which allows the Denom() method
	// to be simpler (and wouldn't have to either return an error or panic).
	// That's why I've included test cases for all the metadata types.

	// These were all generated in a terminal by running commands like this:
	// $ provenanced metaaddress encode scope `uuidgen`
	// I used `uuidgen` for the record names too because it was an easy way to get a random string.
	tests := []string{
		"scope1qrrlhs80yxy5dwyp0tgmpvd49kmqn6cwx2",
		"scope1qz4x0pmzt505ae4nmjfjq6ngq82q3g203c",
		"session1q9xqvqjv73m5x245uwp3ut6jc6ykhs7xnvea7jxenx0pwtmu6cta6u48vr3",
		"session1q93ef6tuvha5359suwj9svp0g2yqd8t72ky7vsfpnqx95aqc3rtnw57g9kn",
		"record1qge4f7nh68tyvy473dlekp369lcfwax7ysuhsm3x7kql8mxxrcn85jjqy4m",
		"record1q2r24x62aze5gxyws04qvxc9ey59kj7u6gsgcrljgcczkcuflc865ae0wzx",
		"scopespec1qjjwrht7ne25cq9tua7tn0vtezcqrsneea",
		"scopespec1qjzk96c9mjzy9z9zpjkuhvptujqsjhm5lz",
		"contractspec1qw88nm7astdy9ay7vh8hc3jpur7qr27mh8",
		"contractspec1qdez57m4vp2y4wd9qckuuhcp0lls0xx3nz",
		"recspec1q48tu2exkajyafvn979hhe36ls5pv9jj9c08ldh7wtas4k0uj9xnzvgyg6e",
		"recspec1q42c5g4a9k25zqyg6v95hp85sp2hy85aelha0l05yy635l2cp44mg3ca006",
	}

	for i, tc := range tests {
		s.Run(fmt.Sprintf("[%d]%s...%s", i, tc[:strings.Index(tc, "1")], tc[len(tc)-2:]), func() {
			exp := DenomPrefix + tc
			addr, err := MetadataAddressFromBech32(tc)
			s.Require().NoError(err, "MetadataAddressFromBech32(%q)", tc)

			var act string
			testFunc := func() {
				act = addr.Denom()
			}
			s.Require().NotPanics(testFunc, "%#v.Denom()", addr)
			s.Assert().Equal(exp, act, "%#v.Denom()", addr)
		})
	}
}

func (s *AddressTestSuite) TestCoin() {
	// Just like the Denom tests, I'm testing all the types even though only scopes should really only ever be used.

	// These were copied from TestDenom.
	tests := []string{
		"scope1qrrlhs80yxy5dwyp0tgmpvd49kmqn6cwx2",
		"scope1qz4x0pmzt505ae4nmjfjq6ngq82q3g203c",
		"session1q9xqvqjv73m5x245uwp3ut6jc6ykhs7xnvea7jxenx0pwtmu6cta6u48vr3",
		"session1q93ef6tuvha5359suwj9svp0g2yqd8t72ky7vsfpnqx95aqc3rtnw57g9kn",
		"record1qge4f7nh68tyvy473dlekp369lcfwax7ysuhsm3x7kql8mxxrcn85jjqy4m",
		"record1q2r24x62aze5gxyws04qvxc9ey59kj7u6gsgcrljgcczkcuflc865ae0wzx",
		"scopespec1qjjwrht7ne25cq9tua7tn0vtezcqrsneea",
		"scopespec1qjzk96c9mjzy9z9zpjkuhvptujqsjhm5lz",
		"contractspec1qw88nm7astdy9ay7vh8hc3jpur7qr27mh8",
		"contractspec1qdez57m4vp2y4wd9qckuuhcp0lls0xx3nz",
		"recspec1q48tu2exkajyafvn979hhe36ls5pv9jj9c08ldh7wtas4k0uj9xnzvgyg6e",
		"recspec1q42c5g4a9k25zqyg6v95hp85sp2hy85aelha0l05yy635l2cp44mg3ca006",
	}

	for i, tc := range tests {
		s.Run(fmt.Sprintf("[%d]%s...%s", i, tc[:strings.Index(tc, "1")], tc[len(tc)-2:]), func() {
			expCoin := sdk.Coin{Denom: DenomPrefix + tc, Amount: sdkmath.OneInt()}
			expCoins := sdk.Coins{expCoin}
			addr, err := MetadataAddressFromBech32(tc)
			s.Require().NoError(err, "MetadataAddressFromBech32(%q)", tc)

			var actCoin sdk.Coin
			testCoin := func() {
				actCoin = addr.Coin()
			}
			if s.Assert().NotPanics(testCoin, "%#v.Coin()", addr) {
				s.Assert().Equal(expCoin, actCoin, "%#v.Coin()", addr)
			}

			var actCoins sdk.Coins
			testCoins := func() {
				actCoins = addr.Coins()
			}
			if s.Assert().NotPanics(testCoins, "%#v.Coins()", addr) {
				s.Assert().Equal(expCoins, actCoins, "%#v.Coins()", addr)
			}
		})
	}
}

func (s *AddressTestSuite) TestAccMDLink_String() {
	newUUID := func(b byte) uuid.UUID {
		bz := bytes.Repeat([]byte{b}, 16)
		rv, err := uuid.FromBytes(bz)
		s.Require().NoError(err, "uuid.FromBytes(%v)", bz)
		return rv
	}
	makeExp := func(accStr, mdStr string) string {
		return accStr + ":" + mdStr
	}
	accAddr := sdk.AccAddress("accAddr_____________")
	scopeAddr := ScopeMetadataAddress(newUUID('0'))
	sessionAddr := SessionMetadataAddress(newUUID('1'), newUUID('1'))
	recordAddr := RecordMetadataAddress(newUUID('2'), strings.Repeat("2", 2))
	sSpecAddr := ScopeSpecMetadataAddress(newUUID('3'))
	cSpecAddr := ContractSpecMetadataAddress(newUUID('4'))
	rSpecAddr := RecordSpecMetadataAddress(newUUID('5'), strings.Repeat("5", 5))

	tests := []struct {
		name string
		link *AccMDLink
		exp  string
	}{
		{
			name: "nil link",
			link: nil,
			exp:  nilStr,
		},
		{
			name: "nil + nil",
			link: NewAccMDLink(nil, nil),
			exp:  makeExp(nilStr, nilStr),
		},
		{
			name: "nil + empty",
			link: NewAccMDLink(nil, MetadataAddress{}),
			exp:  makeExp(nilStr, emptyStr),
		},
		{
			name: "empty + nil",
			link: NewAccMDLink(sdk.AccAddress{}, nil),
			exp:  makeExp(emptyStr, nilStr),
		},
		{
			name: "empty + empty",
			link: NewAccMDLink(sdk.AccAddress{}, MetadataAddress{}),
			exp:  makeExp(emptyStr, emptyStr),
		},
		{
			name: "nil + scope",
			link: NewAccMDLink(nil, scopeAddr),
			exp:  makeExp(nilStr, scopeAddr.String()),
		},
		{
			name: "empty + scope",
			link: NewAccMDLink(sdk.AccAddress{}, scopeAddr),
			exp:  makeExp(emptyStr, scopeAddr.String()),
		},
		{
			name: "addr + nil",
			link: NewAccMDLink(accAddr, nil),
			exp:  makeExp(accAddr.String(), nilStr),
		},
		{
			name: "addr + empty",
			link: NewAccMDLink(accAddr, MetadataAddress{}),
			exp:  makeExp(accAddr.String(), emptyStr),
		},
		{
			name: "addr + scope",
			link: NewAccMDLink(accAddr, scopeAddr),
			exp:  makeExp(accAddr.String(), scopeAddr.String()),
		},
		{
			name: "addr + session",
			link: NewAccMDLink(accAddr, sessionAddr),
			exp:  makeExp(accAddr.String(), sessionAddr.String()),
		},
		{
			name: "addr + record",
			link: NewAccMDLink(accAddr, recordAddr),
			exp:  makeExp(accAddr.String(), recordAddr.String()),
		},
		{
			name: "addr + scope spec",
			link: NewAccMDLink(accAddr, sSpecAddr),
			exp:  makeExp(accAddr.String(), sSpecAddr.String()),
		},
		{
			name: "addr + contract spec",
			link: NewAccMDLink(accAddr, cSpecAddr),
			exp:  makeExp(accAddr.String(), cSpecAddr.String()),
		},
		{
			name: "addr + record spec",
			link: NewAccMDLink(accAddr, rSpecAddr),
			exp:  makeExp(accAddr.String(), rSpecAddr.String()),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act string
			testFunc := func() {
				act = tc.link.String()
			}
			s.Require().NotPanics(testFunc, "%v.String()", tc.link)
			s.Assert().Equal(tc.exp, act, "$v.String()", tc.link)
		})
	}
}

func (s *AddressTestSuite) TestAccMDLinks_String() {
	newUUID := func(b byte) uuid.UUID {
		bz := bytes.Repeat([]byte{b}, 16)
		rv, err := uuid.FromBytes(bz)
		s.Require().NoError(err, "uuid.FromBytes(%v)", bz)
		return rv
	}
	accAddr1 := sdk.AccAddress("accAddr1____________")
	accAddr2 := sdk.AccAddress("accAddr2____________")
	accAddr3 := sdk.AccAddress("accAddr3____________")
	scopeAddr := ScopeMetadataAddress(newUUID('0'))
	sessionAddr := SessionMetadataAddress(newUUID('1'), newUUID('1'))
	recordAddr := RecordMetadataAddress(newUUID('2'), strings.Repeat("2", 2))
	sSpecAddr := ScopeSpecMetadataAddress(newUUID('3'))
	cSpecAddr := ContractSpecMetadataAddress(newUUID('4'))
	rSpecAddr := RecordSpecMetadataAddress(newUUID('5'), strings.Repeat("5", 5))
	makeExp := func(entries ...string) string {
		return "[" + strings.Join(entries, ", ") + "]"
	}
	tests := []struct {
		name  string
		links AccMDLinks
		exp   string
	}{
		{
			name:  "nil",
			links: nil,
			exp:   nilStr,
		},
		{
			name:  "empty",
			links: AccMDLinks{},
			exp:   emptyStr,
		},
		{
			name:  "one nil entry",
			links: AccMDLinks{nil},
			exp:   makeExp(nilStr),
		},
		{
			name:  "one nil nil entry",
			links: AccMDLinks{{}},
			exp:   makeExp(NewAccMDLink(nil, nil).String()),
		},
		{
			name:  "one empty empty entry",
			links: AccMDLinks{NewAccMDLink(sdk.AccAddress{}, MetadataAddress{})},
			exp:   makeExp(NewAccMDLink(sdk.AccAddress{}, MetadataAddress{}).String()),
		},
		{
			name:  "one normal entry",
			links: AccMDLinks{NewAccMDLink(accAddr1, scopeAddr)},
			exp:   makeExp(NewAccMDLink(accAddr1, scopeAddr).String()),
		},
		{
			name: "many entries",
			links: AccMDLinks{
				NewAccMDLink(accAddr1, scopeAddr),
				NewAccMDLink(accAddr1, sessionAddr),
				NewAccMDLink(accAddr2, recordAddr),
				nil,
				NewAccMDLink(accAddr2, sSpecAddr),
				NewAccMDLink(accAddr3, cSpecAddr),
				NewAccMDLink(accAddr3, rSpecAddr),
				NewAccMDLink(accAddr1, nil),
				NewAccMDLink(sdk.AccAddress{}, scopeAddr),
			},
			exp: makeExp(
				NewAccMDLink(accAddr1, scopeAddr).String(),
				NewAccMDLink(accAddr1, sessionAddr).String(),
				NewAccMDLink(accAddr2, recordAddr).String(),
				nilStr,
				NewAccMDLink(accAddr2, sSpecAddr).String(),
				NewAccMDLink(accAddr3, cSpecAddr).String(),
				NewAccMDLink(accAddr3, rSpecAddr).String(),
				NewAccMDLink(accAddr1, nil).String(),
				NewAccMDLink(sdk.AccAddress{}, scopeAddr).String(),
			),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act string
			testFunc := func() {
				act = tc.links.String()
			}
			s.Require().NotPanics(testFunc, "%v.String()", tc.links)
			s.Assert().Equal(tc.exp, act, "$v.String()", tc.links)
		})
	}
}

func (s *AddressTestSuite) TestAccMDLinks_ValidateForScopes() {
	newUUID := func(name string, i int) uuid.UUID {
		bz := []byte(fmt.Sprintf("%s[%d]________________", name, i))[:16]
		rv, err := uuid.FromBytes(bz)
		s.Require().NoError(err, "%s[%d]: uuid.FromBytes(%v)", name, i, bz)
		return rv
	}
	scopeIDs := make([]MetadataAddress, 6)
	for i := range scopeIDs {
		scopeIDs[i] = ScopeMetadataAddress(newUUID("scopeIDs", i))
	}
	addrs := make([]sdk.AccAddress, len(scopeIDs))
	for i := range addrs {
		addrs[i] = sdk.AccAddress(fmt.Sprintf("addr[%d]_____________", i))
	}

	tests := []struct {
		name  string
		links AccMDLinks
		exp   string
	}{
		{
			name:  "nil links",
			links: nil,
			exp:   "",
		},
		{
			name:  "empty links",
			links: nil,
			exp:   "",
		},
		{
			name:  "one link: nil",
			links: AccMDLinks{nil},
			exp:   "nil entry not allowed",
		},
		{
			name:  "one link: empty",
			links: AccMDLinks{{}},
			exp:   "invalid scope metadata address MetadataAddress(nil): address is empty",
		},
		{
			name:  "one link: nil md addr",
			links: AccMDLinks{{MDAddr: nil, AccAddr: addrs[0]}},
			exp:   "invalid scope metadata address MetadataAddress(nil): address is empty",
		},
		{
			name:  "one link: empty md addr",
			links: AccMDLinks{{MDAddr: MetadataAddress{}, AccAddr: addrs[0]}},
			exp:   "invalid scope metadata address MetadataAddress{}: address is empty",
		},
		{
			name:  "one link: scope",
			links: AccMDLinks{{MDAddr: scopeIDs[0], AccAddr: addrs[0]}},
			exp:   "",
		},
		{
			name:  "one link: nil acc addr",
			links: AccMDLinks{{MDAddr: scopeIDs[0], AccAddr: nil}},
			exp:   fmt.Sprintf("no account address associated with metadata address %q", scopeIDs[0]),
		},
		{
			name:  "one link: nil empty addr",
			links: AccMDLinks{{MDAddr: scopeIDs[0], AccAddr: sdk.AccAddress{}}},
			exp:   fmt.Sprintf("no account address associated with metadata address %q", scopeIDs[0]),
		},
		{
			name:  "one link: session",
			links: AccMDLinks{{MDAddr: SessionMetadataAddress(newUUID("session", 0), newUUID("session", 1)), AccAddr: addrs[0]}},
			exp:   fmt.Sprintf("invalid scope id %q: wrong type", SessionMetadataAddress(newUUID("session", 0), newUUID("session", 1))),
		},
		{
			name:  "one link: record",
			links: AccMDLinks{{MDAddr: RecordMetadataAddress(newUUID("record", 0), "recordname"), AccAddr: addrs[0]}},
			exp:   fmt.Sprintf("invalid scope id %q: wrong type", RecordMetadataAddress(newUUID("record", 0), "recordname")),
		},
		{
			name:  "one link: scope spec",
			links: AccMDLinks{{MDAddr: ScopeSpecMetadataAddress(newUUID("scopespec", 0)), AccAddr: addrs[0]}},
			exp:   fmt.Sprintf("invalid scope id %q: wrong type", ScopeSpecMetadataAddress(newUUID("scopespec", 0))),
		},
		{
			name:  "one link: contract spec",
			links: AccMDLinks{{MDAddr: ContractSpecMetadataAddress(newUUID("contractspec", 0)), AccAddr: addrs[0]}},
			exp:   fmt.Sprintf("invalid scope id %q: wrong type", ContractSpecMetadataAddress(newUUID("contractspec", 0))),
		},
		{
			name:  "one link: record spec",
			links: AccMDLinks{{MDAddr: RecordSpecMetadataAddress(newUUID("contractspec", 0), "recordname"), AccAddr: addrs[0]}},
			exp:   fmt.Sprintf("invalid scope id %q: wrong type", RecordSpecMetadataAddress(newUUID("contractspec", 0), "recordname")),
		},
		{
			name:  "one link: unknown mdaddr type",
			links: AccMDLinks{{MDAddr: MetadataAddress{0xa0, 0x6e, 0x6f, 0x70, 0x65}, AccAddr: addrs[0]}},
			exp:   "invalid scope metadata address MetadataAddress{0xa0, 0x6e, 0x6f, 0x70, 0x65}: invalid metadata address type: 160",
		},
		{
			name:  "one link: scope type byte but invalid",
			links: AccMDLinks{{MDAddr: MetadataAddress{ScopeKeyPrefix[0], 0x6e, 0x6f, 0x70, 0x65}, AccAddr: addrs[0]}},
			exp:   "invalid scope metadata address MetadataAddress{0x0, 0x6e, 0x6f, 0x70, 0x65}: incorrect address length (expected: 17, actual: 5)",
		},
		{
			name:  "two links: first nil",
			links: AccMDLinks{nil, {MDAddr: scopeIDs[1], AccAddr: addrs[1]}},
			exp:   "nil entry not allowed",
		},
		{
			name:  "two links: first empty",
			links: AccMDLinks{{}, {MDAddr: scopeIDs[1], AccAddr: addrs[1]}},
			exp:   "invalid scope metadata address MetadataAddress(nil): address is empty",
		},
		{
			name:  "two links: second nil",
			links: AccMDLinks{{MDAddr: scopeIDs[0], AccAddr: addrs[0]}, nil},
			exp:   "nil entry not allowed",
		},
		{
			name:  "two links: second empty",
			links: AccMDLinks{{MDAddr: scopeIDs[0], AccAddr: addrs[0]}, {}},
			exp:   "invalid scope metadata address MetadataAddress(nil): address is empty",
		},
		{
			name: "two links: fully different",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[1], AccAddr: addrs[1]},
			},
			exp: "",
		},
		{
			name: "two links: same scopes different acc addrs",
			links: AccMDLinks{
				{MDAddr: scopeIDs[2], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[2], AccAddr: addrs[1]},
			},
			exp: fmt.Sprintf("duplicate metadata address %q not allowed", scopeIDs[2]),
		},
		{
			name: "two links: same acc addrs different md addrs",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[1], AccAddr: addrs[0]},
			},
			exp: "",
		},
		{
			name: "two links: invalid first md addr",
			links: AccMDLinks{
				{MDAddr: ScopeSpecMetadataAddress(newUUID("scopespec", 1)), AccAddr: addrs[0]},
				{MDAddr: scopeIDs[1], AccAddr: addrs[1]},
			},
			exp: fmt.Sprintf("invalid scope id %q: wrong type", ScopeSpecMetadataAddress(newUUID("scopespec", 1))),
		},
		{
			name: "two links: invalid second md addr",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]},
				{MDAddr: MetadataAddress{0xa0, 0x6e, 0x6f, 0x70, 0x65}, AccAddr: addrs[1]},
			},
			exp: "invalid scope metadata address MetadataAddress{0xa0, 0x6e, 0x6f, 0x70, 0x65}: invalid metadata address type: 160",
		},
		{
			name: "two links: first missing acc addr",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: nil},
				{MDAddr: scopeIDs[1], AccAddr: addrs[1]},
			},
			exp: fmt.Sprintf("no account address associated with metadata address %q", scopeIDs[0]),
		},
		{
			name: "two links: second missing acc addr",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[1], AccAddr: nil},
			},
			exp: fmt.Sprintf("no account address associated with metadata address %q", scopeIDs[1]),
		},
		{
			name: "six links: all valid and fully different",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]}, {MDAddr: scopeIDs[1], AccAddr: addrs[1]},
				{MDAddr: scopeIDs[2], AccAddr: addrs[2]}, {MDAddr: scopeIDs[3], AccAddr: addrs[3]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, {MDAddr: scopeIDs[5], AccAddr: addrs[5]},
			},
			exp: "",
		},
		{
			name: "six links: different scopes but same acc addrs",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[2]}, {MDAddr: scopeIDs[1], AccAddr: addrs[2]},
				{MDAddr: scopeIDs[2], AccAddr: addrs[2]}, {MDAddr: scopeIDs[3], AccAddr: addrs[2]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[2]}, {MDAddr: scopeIDs[5], AccAddr: addrs[2]},
			},
			exp: "",
		},
		{
			name: "six links: same scopes but different acc addrs",
			links: AccMDLinks{
				{MDAddr: scopeIDs[4], AccAddr: addrs[0]}, {MDAddr: scopeIDs[4], AccAddr: addrs[1]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[2]}, {MDAddr: scopeIDs[4], AccAddr: addrs[3]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, {MDAddr: scopeIDs[4], AccAddr: addrs[5]},
			},
			exp: fmt.Sprintf("duplicate metadata address %q not allowed", scopeIDs[4]),
		},
		{
			name: "six links: all same",
			links: AccMDLinks{
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, {MDAddr: scopeIDs[4], AccAddr: addrs[4]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, {MDAddr: scopeIDs[4], AccAddr: addrs[4]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, {MDAddr: scopeIDs[4], AccAddr: addrs[4]},
			},
			exp: fmt.Sprintf("duplicate metadata address %q not allowed", scopeIDs[4]),
		},
		{
			name: "six links: last is invalid md addr",
			links: AccMDLinks{
				{MDAddr: scopeIDs[5], AccAddr: addrs[5]}, {MDAddr: scopeIDs[4], AccAddr: addrs[4]},
				{MDAddr: scopeIDs[3], AccAddr: addrs[3]}, {MDAddr: scopeIDs[2], AccAddr: addrs[2]},
				{MDAddr: scopeIDs[1], AccAddr: addrs[1]}, {MDAddr: MetadataAddress{0xa0, 0x6e, 0x6f, 0x70, 0x65}, AccAddr: addrs[0]},
			},
			exp: "invalid scope metadata address MetadataAddress{0xa0, 0x6e, 0x6f, 0x70, 0x65}: invalid metadata address type: 160",
		},
		{
			name: "six links: last is missing acc addr",
			links: AccMDLinks{
				{MDAddr: scopeIDs[1], AccAddr: addrs[1]}, {MDAddr: scopeIDs[0], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[2], AccAddr: addrs[2]}, {MDAddr: scopeIDs[5], AccAddr: addrs[5]},
				{MDAddr: scopeIDs[3], AccAddr: addrs[3]}, {MDAddr: scopeIDs[4], AccAddr: nil},
			},
			exp: fmt.Sprintf("no account address associated with metadata address %q", scopeIDs[4]),
		},
		{
			name: "six links: last is dup scope",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]}, {MDAddr: scopeIDs[1], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[2], AccAddr: addrs[2]}, {MDAddr: scopeIDs[3], AccAddr: addrs[3]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, {MDAddr: scopeIDs[3], AccAddr: addrs[5]},
			},
			exp: fmt.Sprintf("duplicate metadata address %q not allowed", scopeIDs[3]),
		},
		{
			name: "six links: last is nil",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]}, {MDAddr: scopeIDs[1], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[2], AccAddr: addrs[2]}, {MDAddr: scopeIDs[3], AccAddr: addrs[3]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, nil,
			},
			exp: "nil entry not allowed",
		},
		{
			name: "six links: last is empty",
			links: AccMDLinks{
				{MDAddr: scopeIDs[0], AccAddr: addrs[0]}, {MDAddr: scopeIDs[1], AccAddr: addrs[0]},
				{MDAddr: scopeIDs[2], AccAddr: addrs[2]}, {MDAddr: scopeIDs[3], AccAddr: addrs[3]},
				{MDAddr: scopeIDs[4], AccAddr: addrs[4]}, {},
			},
			exp: "invalid scope metadata address MetadataAddress(nil): address is empty",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var err error
			testFunc := func() {
				err = tc.links.ValidateForScopes()
			}
			s.Require().NotPanics(testFunc, "ValidateForScopes")
			assertions.AssertErrorValue(s.T(), err, tc.exp, "ValidateForScopes")
		})
	}
}

func (s *AddressTestSuite) TestAccMDLinks_GetAccAddrs() {
	addr1 := sdk.AccAddress("1addr_______________") // cosmos1x9skgerjta047h6lta047h6lta047h6l4429yc
	addr2 := sdk.AccAddress("2addr_______________") // cosmos1xfskgerjta047h6lta047h6lta047h6lh0rr9a
	addr3 := sdk.AccAddress("3addr_______________") // cosmos1xdskgerjta047h6lta047h6lta047h6lw7ypa7

	tests := []struct {
		name  string
		links AccMDLinks
		exp   []sdk.AccAddress
	}{
		{
			name:  "nil",
			links: nil,
			exp:   nil,
		},
		{
			name:  "empty",
			links: AccMDLinks{},
			exp:   nil,
		},
		{
			name:  "one nil entry",
			links: AccMDLinks{nil},
			exp:   nil,
		},
		{
			name:  "one entry: nil addr",
			links: AccMDLinks{{AccAddr: nil}},
			exp:   nil,
		},
		{
			name:  "one entry: empty addr",
			links: AccMDLinks{{AccAddr: sdk.AccAddress{}}},
			exp:   nil,
		},
		{
			name:  "one entry: ok addr",
			links: AccMDLinks{{AccAddr: addr1}},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two nil entries",
			links: AccMDLinks{nil, nil},
			exp:   nil,
		},
		{
			name:  "two entries: addrs: nil nil",
			links: AccMDLinks{{AccAddr: nil}, {AccAddr: nil}},
			exp:   nil,
		},
		{
			name:  "two entries: addrs: nil empty",
			links: AccMDLinks{{AccAddr: nil}, {AccAddr: sdk.AccAddress{}}},
			exp:   nil,
		},
		{
			name:  "two entries: addrs: nil ok",
			links: AccMDLinks{{AccAddr: nil}, {AccAddr: addr1}},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two entries: nil entry, entry with ok address",
			links: AccMDLinks{nil, {AccAddr: addr1}},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two entries: addrs: empty nil",
			links: AccMDLinks{{AccAddr: sdk.AccAddress{}}, {AccAddr: nil}},
			exp:   nil,
		},
		{
			name:  "two entries: addrs: empty empty",
			links: AccMDLinks{{AccAddr: sdk.AccAddress{}}, {AccAddr: sdk.AccAddress{}}},
			exp:   nil,
		},
		{
			name:  "two entries: addrs: empty ok",
			links: AccMDLinks{{AccAddr: sdk.AccAddress{}}, {AccAddr: addr1}},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two entries: addrs: ok nil",
			links: AccMDLinks{{AccAddr: addr1}, {AccAddr: nil}},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two entries: ok addr then nil entry",
			links: AccMDLinks{{AccAddr: addr1}, nil},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two entries: addrs: ok empty",
			links: AccMDLinks{{AccAddr: addr1}, {AccAddr: sdk.AccAddress{}}},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two entries: addrs: same",
			links: AccMDLinks{{AccAddr: addr1}, {AccAddr: addr1}},
			exp:   []sdk.AccAddress{addr1},
		},
		{
			name:  "two entries: addrs: different",
			links: AccMDLinks{{AccAddr: addr1}, {AccAddr: addr2}},
			exp:   []sdk.AccAddress{addr1, addr2},
		},
		{
			name:  "two entries: addrs: different opposite order",
			links: AccMDLinks{{AccAddr: addr2}, {AccAddr: addr1}},
			exp:   []sdk.AccAddress{addr2, addr1},
		},
		{
			name: "three different addrs with duplicates",
			links: AccMDLinks{
				{AccAddr: addr2}, {AccAddr: addr2}, {AccAddr: addr1}, {AccAddr: addr2},
				{AccAddr: addr1}, {AccAddr: addr1}, {AccAddr: addr3}, {AccAddr: addr1},
				{AccAddr: addr2}, {AccAddr: addr1},
			},
			exp: []sdk.AccAddress{addr2, addr1, addr3},
		},
		{
			name: "a bit of everything",
			links: AccMDLinks{
				nil, nil, {AccAddr: nil}, {AccAddr: addr1},
				{AccAddr: addr1}, {AccAddr: sdk.AccAddress{}}, nil, {AccAddr: addr2},
				{AccAddr: addr1}, {AccAddr: nil}, {AccAddr: addr3}, {AccAddr: sdk.AccAddress{}},
				{AccAddr: addr3}, nil, {AccAddr: addr1}, nil,
			},
			exp: []sdk.AccAddress{addr1, addr2, addr3},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act []sdk.AccAddress
			testFunc := func() {
				act = tc.links.GetAccAddrs()
			}
			s.Require().NotPanics(testFunc, "GetAccAddrs() on %s", tc.links.String())
			if !s.Assert().Equal(tc.exp, act, "result of GetAccAddrs()") {
				expStrs := mapToStrings(tc.exp)
				actStrs := mapToStrings(act)
				s.Assert().Equal(expStrs, actStrs, "strings of the result of GetAccAddrs()")
			}
		})
	}
}

func (s *AddressTestSuite) TestAccMDLinks_GetPrimaryUUIDs() {
	newUUIDStr := func(i int) string {
		s.Require().LessOrEqual(0, i, "arg provided to newUUID(%d)", i)
		s.Require().GreaterOrEqual(15, i, "arg provided to newUUID(%d)", i)
		h := '0' + byte(i)
		if i >= 10 {
			h = 'a' - 10 + byte(i)
		}
		return strings.ReplaceAll("xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", "x", string(h))
	}
	newUUID := func(i int) uuid.UUID {
		str := newUUIDStr(i)
		rv, err := uuid.Parse(str)
		s.Require().NoError(err, "uuid.Parse(%q)", str)
		return rv
	}

	tests := []struct {
		name  string
		links AccMDLinks
		exp   []string
	}{
		{
			name:  "nil links",
			links: nil,
			exp:   nil,
		},
		{
			name:  "empty links",
			links: AccMDLinks{},
			exp:   []string{},
		},
		{
			name:  "one link: nil",
			links: AccMDLinks{nil},
			exp:   []string{""},
		},
		{
			name:  "one link: nil addr",
			links: AccMDLinks{{MDAddr: nil}},
			exp:   []string{""},
		},
		{
			name:  "one link: empty addr",
			links: AccMDLinks{{MDAddr: MetadataAddress{}}},
			exp:   []string{""},
		},
		{
			name:  "one link: unknown type in addr",
			links: AccMDLinks{{MDAddr: MetadataAddress{0xA0, 0x1, 0x2}}},
			exp:   []string{""},
		},
		{
			name:  "one link: addr too short",
			links: AccMDLinks{{MDAddr: ScopeMetadataAddress(newUUID(0))[:16]}},
			exp:   []string{""},
		},
		{
			name:  "one link: scope id",
			links: AccMDLinks{{MDAddr: ScopeMetadataAddress(newUUID(1))}},
			exp:   []string{newUUIDStr(1)},
		},
		{
			name:  "one link: session id",
			links: AccMDLinks{{MDAddr: SessionMetadataAddress(newUUID(2), newUUID(8))}},
			exp:   []string{newUUIDStr(2)},
		},
		{
			name:  "one link: record id",
			links: AccMDLinks{{MDAddr: RecordMetadataAddress(newUUID(3), newUUIDStr(4))}},
			exp:   []string{newUUIDStr(3)},
		},
		{
			name:  "one link: scope spec id",
			links: AccMDLinks{{MDAddr: ScopeSpecMetadataAddress(newUUID(5))}},
			exp:   []string{newUUIDStr(5)},
		},
		{
			name:  "one link: contract spec id",
			links: AccMDLinks{{MDAddr: ContractSpecMetadataAddress(newUUID(6))}},
			exp:   []string{newUUIDStr(6)},
		},
		{
			name:  "one link: record spec id",
			links: AccMDLinks{{MDAddr: RecordSpecMetadataAddress(newUUID(7), newUUIDStr(9))}},
			exp:   []string{newUUIDStr(7)},
		},
		{
			name: "six links: one of each type",
			links: AccMDLinks{
				{MDAddr: ScopeMetadataAddress(newUUID(11))},
				{MDAddr: SessionMetadataAddress(newUUID(10), newUUID(3))},
				{MDAddr: RecordMetadataAddress(newUUID(15), newUUIDStr(2))},
				{MDAddr: ScopeSpecMetadataAddress(newUUID(13))},
				{MDAddr: ContractSpecMetadataAddress(newUUID(12))},
				{MDAddr: RecordSpecMetadataAddress(newUUID(14), newUUIDStr(1))},
			},
			exp: []string{
				newUUIDStr(11), newUUIDStr(10), newUUIDStr(15),
				newUUIDStr(13), newUUIDStr(12), newUUIDStr(14),
			},
		},
		{
			name: "six links: all different scopes",
			links: AccMDLinks{
				{MDAddr: ScopeMetadataAddress(newUUID(7))},
				{MDAddr: ScopeMetadataAddress(newUUID(1))},
				{MDAddr: ScopeMetadataAddress(newUUID(0))},
				{MDAddr: ScopeMetadataAddress(newUUID(8))},
				{MDAddr: ScopeMetadataAddress(newUUID(14))},
				{MDAddr: ScopeMetadataAddress(newUUID(2))},
			},
			exp: []string{
				newUUIDStr(7), newUUIDStr(1), newUUIDStr(0),
				newUUIDStr(8), newUUIDStr(14), newUUIDStr(2),
			},
		},
		{
			name: "six links: mix of different, same and invalid scopes",
			links: AccMDLinks{
				{MDAddr: ScopeMetadataAddress(newUUID(14))},
				{MDAddr: ScopeMetadataAddress(newUUID(12))},
				{MDAddr: ScopeMetadataAddress(newUUID(8))},
				{MDAddr: ScopeMetadataAddress(newUUID(10))[:16]},
				{MDAddr: ScopeMetadataAddress(newUUID(11))},
				{MDAddr: ScopeMetadataAddress(newUUID(12))},
			},
			exp: []string{
				newUUIDStr(14), newUUIDStr(12), newUUIDStr(8),
				"", newUUIDStr(11), newUUIDStr(12),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act []string
			testFunc := func() {
				act = tc.links.GetPrimaryUUIDs()
			}
			s.Require().NotPanics(testFunc, "GetPrimaryUUIDs")
			s.Assert().Equal(tc.exp, act, "result from GetPrimaryUUIDs")
		})
	}
}

func (s *AddressTestSuite) TestAccMDLinks_GetMDAddrsForAccAddr() {
	newUUID := func(name string) uuid.UUID {
		s.Require().LessOrEqual(len(name), 16, "newUUID(%q): name too long")
		if len(name) < 16 {
			name = name + strings.Repeat("_", 16-len(name))
		}
		rv, err := uuid.FromBytes([]byte(name))
		s.Require().NoError(err, "uuid.FromBytes([]byte(%q))", name)
		return rv
	}
	newAddr := func(name string) sdk.AccAddress {
		switch {
		case len(name) < 20:
			name = name + strings.Repeat("_", 20-len(name))
		case len(name) > 20 && len(name) < 32:
			name = name + strings.Repeat("_", 32-len(name))
		}
		return sdk.AccAddress(name)
	}

	accAddrs := []sdk.AccAddress{
		newAddr("0_addr"), // cosmos1xp0kzerywf047h6lta047h6lta047h6ln9q3ue
		newAddr("1_addr"), // cosmos1x90kzerywf047h6lta047h6lta047h6l258ny6
		newAddr("2_addr"), // cosmos1xf0kzerywf047h6lta047h6lta047h6lgww49l
		newAddr("3_addr"), // cosmos1xd0kzerywf047h6lta047h6lta047h6l3lfhau
		newAddr("4_addr"), // cosmos1x30kzerywf047h6lta047h6lta047h6lvnue84
	}
	testlog.WriteSlice(s.T(), "accAddrs", accAddrs)

	mdAddrs := []MetadataAddress{
		ScopeMetadataAddress(newUUID("0_scope")),                                  // scope1qqc97umrdacx2h6lta047h6lta0s4e2vmr
		ScopeMetadataAddress(newUUID("1_scope")),                                  // scope1qqc47umrdacx2h6lta047h6lta0sfyvr90
		ScopeMetadataAddress(newUUID("2_scope")),                                  // scope1qqe97umrdacx2h6lta047h6lta0sk6uj0g
		ScopeMetadataAddress(newUUID("3_scope")),                                  // scope1qqe47umrdacx2h6lta047h6lta0s286a3y
		ScopeMetadataAddress(newUUID("4_scope")),                                  // scope1qq697umrdacx2h6lta047h6lta0snl0e64
		SessionMetadataAddress(newUUID("5_session_1"), newUUID("5_session_2")),    // session1qy647um9wdekjmmwtuc47h6lta0n2hmnv4ehx6t0de0nyh6lta047fgqzjx
		RecordMetadataAddress(newUUID("6_record"), "6_record_name"),               // record1qgm97un9vdhhyezlta047h6lta05sqnucqwnxlr6pxatcmq9sf0f5u999l2
		ScopeSpecMetadataAddress(newUUID("7_scope_spec")),                         // scopespec1qsm47umrdacx2hmnwpjkxh6lta0sz95anh
		ContractSpecMetadataAddress(newUUID("8_contract_spec")),                   // contractspec1qvu97cm0de68yctrw30hxur9vd0smvmt09
		RecordSpecMetadataAddress(newUUID("9_record_spec"), "9_record_spec_name"), // recspec1q5u47un9vdhhyezlwdcx2c6lta0julrf7q442a5js2y8sm4gcnx8u98dcxq
	}
	testlog.WriteSlice(s.T(), "mdAddrs", mdAddrs)

	tests := []struct {
		name  string
		links AccMDLinks
		addr  sdk.AccAddress
		exp   []MetadataAddress
	}{
		{
			name:  "nil links",
			links: nil,
			addr:  accAddrs[0],
			exp:   nil,
		},
		{
			name:  "empty links",
			links: make(AccMDLinks, 0),
			addr:  accAddrs[0],
			exp:   nil,
		},
		{
			name:  "one link: nil",
			links: AccMDLinks{nil},
			addr:  nil,
			exp:   nil,
		},
		{
			name:  "one link: nil AccAddr",
			links: AccMDLinks{NewAccMDLink(nil, mdAddrs[0])},
			addr:  accAddrs[0],
			exp:   nil,
		},
		{
			name:  "one link: empty AccAddr",
			links: AccMDLinks{NewAccMDLink(make(sdk.AccAddress, 0), mdAddrs[0])},
			addr:  accAddrs[0],
			exp:   nil,
		},
		{
			name:  "one link: other addr",
			links: AccMDLinks{NewAccMDLink(accAddrs[0], mdAddrs[0])},
			addr:  accAddrs[1],
			exp:   nil,
		},
		{
			name:  "one link: same addr",
			links: AccMDLinks{NewAccMDLink(accAddrs[0], mdAddrs[0])},
			addr:  accAddrs[0],
			exp:   subSet(mdAddrs, 0),
		},
		{
			name:  "two links with same AccAddr: other addr",
			links: AccMDLinks{NewAccMDLink(accAddrs[2], mdAddrs[0]), NewAccMDLink(accAddrs[2], mdAddrs[1])},
			addr:  accAddrs[3],
			exp:   nil,
		},
		{
			name:  "two links with same AccAddr: that addr",
			links: AccMDLinks{NewAccMDLink(accAddrs[2], mdAddrs[0]), NewAccMDLink(accAddrs[2], mdAddrs[1])},
			addr:  accAddrs[2],
			exp:   subSet(mdAddrs, 0, 1),
		},
		{
			name:  "two links with same AccAddr: that addr, opposite order",
			links: AccMDLinks{NewAccMDLink(accAddrs[2], mdAddrs[1]), NewAccMDLink(accAddrs[2], mdAddrs[0])},
			addr:  accAddrs[2],
			exp:   subSet(mdAddrs, 1, 0),
		},
		{
			name: "three links with diff AccAddr: get none",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[0]),
				NewAccMDLink(accAddrs[1], mdAddrs[1]),
				NewAccMDLink(accAddrs[2], mdAddrs[2]),
			},
			addr: accAddrs[3],
			exp:  nil,
		},
		{
			name: "three links with diff AccAddr: get first",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[0]),
				NewAccMDLink(accAddrs[1], mdAddrs[1]),
				NewAccMDLink(accAddrs[2], mdAddrs[2]),
			},
			addr: accAddrs[0],
			exp:  subSet(mdAddrs, 0),
		},
		{
			name: "three links with diff AccAddr: get second",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[0]),
				NewAccMDLink(accAddrs[1], mdAddrs[1]),
				NewAccMDLink(accAddrs[2], mdAddrs[2]),
			},
			addr: accAddrs[1],
			exp:  subSet(mdAddrs, 1),
		},
		{
			name: "three links with diff AccAddr: get third",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[0]),
				NewAccMDLink(accAddrs[1], mdAddrs[1]),
				NewAccMDLink(accAddrs[2], mdAddrs[2]),
			},
			addr: accAddrs[2],
			exp:  subSet(mdAddrs, 2),
		},
		{
			name: "three links two with same AccAddr: get first and second",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[3], mdAddrs[1]),
				NewAccMDLink(accAddrs[3], mdAddrs[2]),
				NewAccMDLink(accAddrs[0], mdAddrs[3]),
			},
			addr: accAddrs[3],
			exp:  subSet(mdAddrs, 1, 2),
		},
		{
			name: "three links two with same AccAddr: get third",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[3], mdAddrs[1]),
				NewAccMDLink(accAddrs[3], mdAddrs[2]),
				NewAccMDLink(accAddrs[0], mdAddrs[3]),
			},
			addr: accAddrs[0],
			exp:  subSet(mdAddrs, 3),
		},
		{
			name: "three links two with same AccAddr: get first and third",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[3], mdAddrs[1]),
				NewAccMDLink(accAddrs[0], mdAddrs[2]),
				NewAccMDLink(accAddrs[3], mdAddrs[3]),
			},
			addr: accAddrs[3],
			exp:  subSet(mdAddrs, 1, 3),
		},
		{
			name: "three links two with same AccAddr: get second",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[3], mdAddrs[1]),
				NewAccMDLink(accAddrs[0], mdAddrs[2]),
				NewAccMDLink(accAddrs[3], mdAddrs[3]),
			},
			addr: accAddrs[0],
			exp:  subSet(mdAddrs, 2),
		},
		{
			name: "three links two with same AccAddr: get second and third",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[1]),
				NewAccMDLink(accAddrs[3], mdAddrs[2]),
				NewAccMDLink(accAddrs[3], mdAddrs[3]),
			},
			addr: accAddrs[3],
			exp:  subSet(mdAddrs, 2, 3),
		},
		{
			name: "three links two with same AccAddr: get first",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[1]),
				NewAccMDLink(accAddrs[3], mdAddrs[2]),
				NewAccMDLink(accAddrs[3], mdAddrs[3]),
			},
			addr: accAddrs[0],
			exp:  subSet(mdAddrs, 1),
		},
		{
			name: "three links all with same AccAddr: get none",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[4], mdAddrs[1]),
				NewAccMDLink(accAddrs[4], mdAddrs[2]),
				NewAccMDLink(accAddrs[4], mdAddrs[3]),
			},
			addr: accAddrs[3],
			exp:  nil,
		},
		{
			name: "three links all with same AccAddr: get all",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[4], mdAddrs[1]),
				NewAccMDLink(accAddrs[4], mdAddrs[2]),
				NewAccMDLink(accAddrs[4], mdAddrs[3]),
			},
			addr: accAddrs[4],
			exp:  subSet(mdAddrs, 1, 2, 3),
		},
		{
			name: "five links: get two",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[0]),
				NewAccMDLink(accAddrs[1], mdAddrs[1]),
				NewAccMDLink(accAddrs[2], mdAddrs[2]),
				NewAccMDLink(accAddrs[1], mdAddrs[3]),
				NewAccMDLink(accAddrs[4], mdAddrs[4]),
			},
			addr: accAddrs[1],
			exp:  subSet(mdAddrs, 1, 3),
		},
		{
			name: "six links with diff MDAddr types: get three",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[4]),
				NewAccMDLink(accAddrs[1], mdAddrs[5]),
				NewAccMDLink(accAddrs[0], mdAddrs[6]),
				NewAccMDLink(accAddrs[2], mdAddrs[7]),
				NewAccMDLink(accAddrs[2], mdAddrs[8]),
				NewAccMDLink(accAddrs[0], mdAddrs[9]),
			},
			addr: accAddrs[0],
			exp:  subSet(mdAddrs, 4, 6, 9),
		},
		{
			name: "six links with diff MDAddr types: get all",
			links: AccMDLinks{
				NewAccMDLink(accAddrs[0], mdAddrs[4]),
				NewAccMDLink(accAddrs[0], mdAddrs[5]),
				NewAccMDLink(accAddrs[0], mdAddrs[6]),
				NewAccMDLink(accAddrs[0], mdAddrs[7]),
				NewAccMDLink(accAddrs[0], mdAddrs[8]),
				NewAccMDLink(accAddrs[0], mdAddrs[9]),
			},
			addr: accAddrs[0],
			exp:  subSet(mdAddrs, 4, 5, 6, 7, 8, 9),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act []MetadataAddress
			testFunc := func() {
				act = tc.links.GetMDAddrsForAccAddr(tc.addr.String())
			}
			s.Require().NotPanics(testFunc, "GetMDAddrsForAccAddr")
			// Compare them as strings first since that failure message is probably easier to understand.
			expStrs := mapToStrings(tc.exp)
			actStrs := mapToStrings(act)
			if s.Assert().Equal(expStrs, actStrs, "result of GetMDAddrsForAccAddr (as strings)") {
				// The strings are equal, make sure that means they're actually equal.
				s.Assert().Equal(tc.exp, act, "result of GetMDAddrsForAccAddr")
			}
		})
	}
}
