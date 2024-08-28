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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
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
		{id: invalidID, fmt: "%s", exp: "%!s(PANIC=Format method: invalid metadata address type: 100)"},
		{id: invalidID, fmt: "%q", exp: "%!q(PANIC=Format method: invalid metadata address type: 100)"},
		{id: invalidID, fmt: "%v", exp: "%!v(PANIC=Format method: invalid metadata address type: 100)"},
		{id: invalidID, fmt: "%#v", exp: "MetadataAddress{0x64, 0x6f, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x20, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x65, 0x73, 0x20, 0x74, 0x68, 0x69, 0x73, 0x20, 0x77, 0x61, 0x79}"},
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
			exp := "nft/" + tc
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

func (s *AddressTestSuite) TestAccMDLinks_GetAccAddrs() {
	newLink := func(accAddr sdk.AccAddress) *AccMDLink {
		return &AccMDLink{AccAddr: accAddr}
	}
	addrs := make([]sdk.AccAddress, 6)
	links := make(AccMDLinks, len(addrs))
	for i := range addrs {
		addrs[i] = sdk.AccAddress(fmt.Sprintf("addr[%d]_____________", i))
		links[i] = newLink(addrs[i])
	}

	tests := []struct {
		name  string
		links AccMDLinks
		exp   []sdk.AccAddress
	}{
		{
			name:  "nil links",
			links: nil,
			exp:   nil,
		},
		{
			name:  "empty links",
			links: AccMDLinks{},
			exp:   nil,
		},
		{
			name:  "one link is nil",
			links: AccMDLinks{nil},
			exp:   []sdk.AccAddress{},
		},
		{
			name:  "one link with acc",
			links: links[0:1],
			exp:   addrs[0:1],
		},
		{
			name:  "one link with nil acc",
			links: AccMDLinks{newLink(nil)},
			exp:   []sdk.AccAddress{},
		},
		{
			name:  "one link with empty acc",
			links: AccMDLinks{newLink(sdk.AccAddress{})},
			exp:   []sdk.AccAddress{},
		},
		{
			name: "two links no acc in either",
			links: AccMDLinks{
				newLink(nil),
				newLink(sdk.AccAddress{}),
			},
			exp: []sdk.AccAddress{},
		},
		{
			name: "two links first has acc second does not",
			links: AccMDLinks{
				links[0],
				newLink(nil),
			},
			exp: addrs[0:1],
		},
		{
			name: "two links second has acc first does not",
			links: AccMDLinks{
				newLink(nil),
				links[1],
			},
			exp: addrs[1:2],
		},
		{
			name: "two links both have same acc",
			links: AccMDLinks{
				links[0],
				newLink(addrs[0]),
			},
			exp: addrs[0:1],
		},
		{
			name:  "two links with different accs in sequential order",
			links: links[3:5],
			exp:   addrs[3:5],
		},
		{
			name:  "two links with different accs in reverse order",
			links: AccMDLinks{links[5], links[4]},
			exp:   []sdk.AccAddress{addrs[5], addrs[4]},
		},
		{
			name:  "three links all different",
			links: links[1:4],
			exp:   addrs[1:4],
		},
		{
			name: "three links all same",
			links: AccMDLinks{
				newLink(addrs[5]),
				newLink(addrs[5]),
				newLink(addrs[5]),
			},
			exp: addrs[5:6],
		},
		{
			name: "three links AAB",
			links: AccMDLinks{
				newLink(addrs[5]),
				newLink(addrs[5]),
				newLink(addrs[3]),
			},
			exp: []sdk.AccAddress{addrs[5], addrs[3]},
		},
		{
			name: "three links ABA",
			links: AccMDLinks{
				newLink(addrs[5]),
				newLink(addrs[3]),
				newLink(addrs[5]),
			},
			exp: []sdk.AccAddress{addrs[5], addrs[3]},
		},
		{
			name: "three links BAA",
			links: AccMDLinks{
				newLink(addrs[3]),
				newLink(addrs[5]),
				newLink(addrs[5]),
			},
			exp: []sdk.AccAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links ABnil",
			links: AccMDLinks{
				newLink(addrs[3]),
				newLink(addrs[5]),
				nil,
			},
			exp: []sdk.AccAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links AnilB",
			links: AccMDLinks{
				newLink(addrs[3]),
				nil,
				newLink(addrs[5]),
			},
			exp: []sdk.AccAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links nilAB",
			links: AccMDLinks{
				nil,
				newLink(addrs[3]),
				newLink(addrs[5]),
			},
			exp: []sdk.AccAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links AnilA",
			links: AccMDLinks{
				newLink(addrs[2]),
				nil,
				newLink(addrs[2]),
			},
			exp: []sdk.AccAddress{addrs[2]},
		},
		{
			name:  "six links each with different acc",
			links: links,
			exp:   addrs,
		},
		{
			name: "six links one nil one acc nil one acc empty two same",
			links: AccMDLinks{
				newLink(addrs[0]),
				nil,
				newLink(sdk.AccAddress{}),
				newLink(addrs[3]),
				newLink(nil),
				newLink(addrs[0]),
			},
			exp: []sdk.AccAddress{addrs[0], addrs[3]},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act []sdk.AccAddress
			testFunc := func() {
				act = tc.links.GetAccAddrs()
			}
			s.Require().NotPanics(testFunc, "GetAccAddrs")
			s.Assert().Equal(tc.exp, act, "GetAccAddrs")
		})
	}
}

func (s *AddressTestSuite) TestAccMDLinks_GetMetadataAddrs() {
	newLink := func(mdAddr MetadataAddress) *AccMDLink {
		return &AccMDLink{MDAddr: mdAddr}
	}
	newUUID := func(b byte) uuid.UUID {
		bz := bytes.Repeat([]byte{b}, 16)
		rv, err := uuid.FromBytes(bz)
		s.Require().NoError(err, "uuid.FromBytes(%v)", bz)
		return rv
	}
	addrs := []MetadataAddress{
		ScopeMetadataAddress(newUUID('0')),
		SessionMetadataAddress(newUUID('1'), newUUID('1')),
		RecordMetadataAddress(newUUID('2'), strings.Repeat("2", 2)),
		ScopeSpecMetadataAddress(newUUID('3')),
		ContractSpecMetadataAddress(newUUID('4')),
		RecordSpecMetadataAddress(newUUID('5'), strings.Repeat("5", 5)),
	}
	links := make(AccMDLinks, len(addrs))
	for i := range addrs {
		links[i] = &AccMDLink{MDAddr: addrs[i]}
	}

	tests := []struct {
		name  string
		links AccMDLinks
		exp   []MetadataAddress
	}{
		{
			name:  "nil links",
			links: nil,
			exp:   nil,
		},
		{
			name:  "empty links",
			links: AccMDLinks{},
			exp:   nil,
		},
		{
			name:  "one link is nil",
			links: AccMDLinks{nil},
			exp:   []MetadataAddress{},
		},
		{
			name:  "one link with md",
			links: links[0:1],
			exp:   addrs[0:1],
		},
		{
			name:  "one link with nil md",
			links: AccMDLinks{newLink(nil)},
			exp:   []MetadataAddress{},
		},
		{
			name:  "one link with empty md",
			links: AccMDLinks{newLink(MetadataAddress{})},
			exp:   []MetadataAddress{},
		},
		{
			name: "two links no md in either",
			links: AccMDLinks{
				newLink(nil),
				newLink(MetadataAddress{}),
			},
			exp: []MetadataAddress{},
		},
		{
			name: "two links first has md second does not",
			links: AccMDLinks{
				links[0],
				newLink(nil),
			},
			exp: addrs[0:1],
		},
		{
			name: "two links second has md first does not",
			links: AccMDLinks{
				newLink(nil),
				links[1],
			},
			exp: addrs[1:2],
		},
		{
			name: "two links both have same md",
			links: AccMDLinks{
				links[0],
				newLink(addrs[0]),
			},
			exp: addrs[0:1],
		},
		{
			name:  "two links with different mds in sequential order",
			links: links[3:5],
			exp:   addrs[3:5],
		},
		{
			name:  "two links with different mds in reverse order",
			links: AccMDLinks{links[5], links[4]},
			exp:   []MetadataAddress{addrs[5], addrs[4]},
		},
		{
			name:  "three links all different",
			links: links[1:4],
			exp:   addrs[1:4],
		},
		{
			name: "three links all same",
			links: AccMDLinks{
				newLink(addrs[5]),
				newLink(addrs[5]),
				newLink(addrs[5]),
			},
			exp: addrs[5:6],
		},
		{
			name: "three links AAB",
			links: AccMDLinks{
				newLink(addrs[5]),
				newLink(addrs[5]),
				newLink(addrs[3]),
			},
			exp: []MetadataAddress{addrs[5], addrs[3]},
		},
		{
			name: "three links ABA",
			links: AccMDLinks{
				newLink(addrs[5]),
				newLink(addrs[3]),
				newLink(addrs[5]),
			},
			exp: []MetadataAddress{addrs[5], addrs[3]},
		},
		{
			name: "three links BAA",
			links: AccMDLinks{
				newLink(addrs[3]),
				newLink(addrs[5]),
				newLink(addrs[5]),
			},
			exp: []MetadataAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links ABnil",
			links: AccMDLinks{
				newLink(addrs[3]),
				newLink(addrs[5]),
				nil,
			},
			exp: []MetadataAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links AnilB",
			links: AccMDLinks{
				newLink(addrs[3]),
				nil,
				newLink(addrs[5]),
			},
			exp: []MetadataAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links nilAB",
			links: AccMDLinks{
				nil,
				newLink(addrs[3]),
				newLink(addrs[5]),
			},
			exp: []MetadataAddress{addrs[3], addrs[5]},
		},
		{
			name: "three links AnilA",
			links: AccMDLinks{
				newLink(addrs[2]),
				nil,
				newLink(addrs[2]),
			},
			exp: []MetadataAddress{addrs[2]},
		},
		{
			name:  "six links each with different md",
			links: links,
			exp:   addrs,
		},
		{
			name: "six links one nil one md nil one md empty two same",
			links: AccMDLinks{
				newLink(addrs[0]),
				nil,
				newLink(MetadataAddress{}),
				newLink(addrs[3]),
				newLink(nil),
				newLink(addrs[0]),
			},
			exp: []MetadataAddress{addrs[0], addrs[3]},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var act []MetadataAddress
			testFunc := func() {
				act = tc.links.GetMetadataAddrs()
			}
			s.Require().NotPanics(testFunc, "GetMetadataAddrs")
			s.Assert().Equal(tc.exp, act, "GetMetadataAddrs")
		})
	}
}
