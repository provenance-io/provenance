package types

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name: "empty typeBytes",
			typeBytes: []byte{},
			hash: testHash,
			expectedAddr: "",
			expectedError: "empty typeCode bytes",
		},
		{
			name: "empty hash",
			typeBytes: ScopeKeyPrefix,
			hash: "",
			expectedAddr: "",
			expectedError: "empty hash string",
		},
		{
			name: "scope key prefix - valid",
			typeBytes: ScopeKeyPrefix,
			hash: testHash,
			expectedAddr: s.requireBech32String(ScopeKeyPrefix, testHashBytes[:16]),
			expectedError: "",
		},
		{
			name: "scope key prefix - too short",
			typeBytes: ScopeKeyPrefix,
			hash: testHash15,
			expectedAddr: "",
			expectedError: fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash15, 16, 15),
		},
		{
			name: "session key prefix - valid",
			typeBytes: SessionKeyPrefix,
			hash: testHash,
			expectedAddr: s.requireBech32String(SessionKeyPrefix, testHashBytes[:32]),
			expectedError: "",
		},
		{
			name: "session key prefix - too short",
			typeBytes: SessionKeyPrefix,
			hash: testHash31,
			expectedAddr: "",
			expectedError: fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash31, 32, 31),
		},
		{
			name: "record key prefix - valid",
			typeBytes: RecordKeyPrefix,
			hash: testHash,
			expectedAddr: s.requireBech32String(RecordKeyPrefix, testHashBytes[:32]),
			expectedError: "",
		},
		{
			name: "record key prefix - too short",
			typeBytes: RecordKeyPrefix,
			hash: testHash31,
			expectedAddr: "",
			expectedError: fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash31, 32, 31),
		},
		{
			name: "scope spec key prefix - valid",
			typeBytes: ScopeSpecificationKeyPrefix,
			hash: testHash,
			expectedAddr: s.requireBech32String(ScopeSpecificationKeyPrefix, testHashBytes[:16]),
			expectedError: "",
		},
		{
			name: "scope spec key prefix - too short",
			typeBytes: ScopeSpecificationKeyPrefix,
			hash: testHash15,
			expectedAddr: "",
			expectedError: fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash15, 16, 15),
		},
		{
			name: "contract spec key prefix - valid",
			typeBytes: ContractSpecificationKeyPrefix,
			hash: testHash,
			expectedAddr: s.requireBech32String(ContractSpecificationKeyPrefix, testHashBytes[:16]),
			expectedError: "",
		},
		{
			name: "contract spec key prefix - too short",
			typeBytes: ContractSpecificationKeyPrefix,
			hash: testHash15,
			expectedAddr: "",
			expectedError: fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash15, 16, 15),
		},
		{
			name: "record spec key prefix - valid",
			typeBytes: RecordSpecificationKeyPrefix,
			hash: testHash,
			expectedAddr: s.requireBech32String(RecordSpecificationKeyPrefix, testHashBytes[:32]),
			expectedError: "",
		},
		{
			name: "record spec key prefix - too short",
			typeBytes: RecordSpecificationKeyPrefix,
			hash: testHash31,
			expectedAddr: "",
			expectedError: fmt.Sprintf("invalid hash \"%s\" byte length, expected at least %d bytes, found %d",
				testHash31, 32, 31),
		},
		{
			name: "invalid type bytes",
			typeBytes: []byte{0x07},
			hash: testHash,
			expectedAddr: "",
			expectedError: "invalid address type code 0x07",
		},
		{
			name: "invalid hash",
			typeBytes: ScopeKeyPrefix,
			hash: "invalid hash",
			expectedAddr: "",
			expectedError: base64.CorruptInputError(7).Error(),
		},
		{
			name: "hash string decodes to empty",
			typeBytes: ScopeKeyPrefix,
			hash: "MA==",
			expectedAddr: "",
			expectedError: "invalid hash \"MA==\" byte length, expected at least 16 bytes, found 1",
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

func (s *AddressTestSuite) TestMetadataAddressWithInvalidData() {
	t := s.T()

	addr, addrErr := sdk.AccAddressFromBech32("cosmos1zgp4n2yvrtxkj5zl6rzcf6phqg0gfzuf3v08r4")
	require.NoError(t, addrErr, "address parsing error")

	_, err := VerifyMetadataAddressFormat(addr)
	require.EqualValues(t, fmt.Errorf("invalid metadata address type: %d", addr[0]), err)

	scopeID := ScopeMetadataAddress(s.scopeUUID)
	padded := make([]byte, 20)
	len, err := scopeID.MarshalTo(padded)
	require.EqualValues(t, 17, len)

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
	require.Equal(t, fmt.Sprintf("%s", s.scopeHex), fmt.Sprintf("%X", scopeID))

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
			name: "scope",
			id: ScopeMetadataAddress(uuid.New()),
			expected: [6]bool{true, false, false, false, false, false},
		},
		{
			name: "session",
			id: SessionMetadataAddress(uuid.New(), uuid.New()),
			expected: [6]bool{false, true, false, false, false, false},
		},
		{
			name: "record",
			id: RecordMetadataAddress(uuid.New(), "best ever"),
			expected: [6]bool{false, false, true, false, false, false},
		},
		{
			name: "scope specification",
			id: ScopeSpecMetadataAddress(uuid.New()),
			expected: [6]bool{false, false, false, true, false, false},
		},
		{
			name: "contract speficiation",
			id: ContractSpecMetadataAddress(uuid.New()),
			expected: [6]bool{false, false, false, false, true, false},
		},
		{
			name: "record specification",
			id: RecordSpecMetadataAddress(uuid.New(), "okayest dad"),
			expected: [6]bool{false, false, false, false, false, true},
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
			name: "scope",
			addr: ScopeMetadataAddress(uuid.New()),
			expectedValue: PrefixScope,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "scope without address",
			addr: MetadataAddress{ScopeKeyPrefix[0]},
			expectedValue: PrefixScope,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "session",
			addr: SessionMetadataAddress(uuid.New(), uuid.New()),
			expectedValue: PrefixSession,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "session without address",
			addr: MetadataAddress{SessionKeyPrefix[0]},
			expectedValue: PrefixSession,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "record",
			addr: RecordMetadataAddress(uuid.New(), "ronald"),
			expectedValue: PrefixRecord,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "record without address",
			addr: MetadataAddress{RecordKeyPrefix[0]},
			expectedValue: PrefixRecord,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "scope spec",
			addr: ScopeSpecMetadataAddress(uuid.New()),
			expectedValue: PrefixScopeSpecification,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "scope spec without address",
			addr: MetadataAddress{ScopeSpecificationKeyPrefix[0]},
			expectedValue: PrefixScopeSpecification,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "contract spec",
			addr: ContractSpecMetadataAddress(uuid.New()),
			expectedValue: PrefixContractSpecification,
			expectedError: "",
			expectAnyError: false,
		},
		{
		name: 	"contract spec without address",
			addr: MetadataAddress{ContractSpecificationKeyPrefix[0]},
			expectedValue: PrefixContractSpecification,
			expectedError: "",
			expectAnyError: false,
		},
		{
		name: 	"record spec",
			addr: RecordSpecMetadataAddress(uuid.New(), "george"),
			expectedValue: PrefixRecordSpecification,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "record spec without address",
			addr: MetadataAddress{RecordSpecificationKeyPrefix[0]},
			expectedValue: PrefixRecordSpecification,
			expectedError: "",
			expectAnyError: false,
		},
		{
			name: "empty",
			addr: MetadataAddress{},
			expectedValue: "",
			expectedError: "address is empty",
			expectAnyError: false,
		},
		{
			name: "bad",
			addr: MetadataAddress("don't do this"),
			expectedValue: "",
			expectedError: "",
			expectAnyError: true,
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
			name: "scope",
			id: ScopeMetadataAddress(scopePrimaryUUID),
			expectedValue: scopePrimaryUUID,
			expectedError: "",
		},
		{
			name: "session",
			id: SessionMetadataAddress(sessionPrimaryUUID, uuid.New()),
			expectedValue: sessionPrimaryUUID,
			expectedError: "",
		},
		{
			name: "record",
			id: RecordMetadataAddress(recordPrimaryUUID, "presence"),
			expectedValue: recordPrimaryUUID,
			expectedError: "",
		},
		{
			name: "scope spec",
			id: ScopeSpecMetadataAddress(scopeSpecPrimaryUUID),
			expectedValue: scopeSpecPrimaryUUID,
			expectedError: "",
		},
		{
			name: "contract spec",
			id: ContractSpecMetadataAddress(contractSpecPrimaryUUID),
			expectedValue: contractSpecPrimaryUUID,
			expectedError: "",
		},
		{
			name: "record spec",
			id: RecordSpecMetadataAddress(recordSpecPrimaryUUID, "profiled"),
			expectedValue: recordSpecPrimaryUUID,
			expectedError: "",
		},
		{
			name: "empty",
			id: MetadataAddress{},
			expectedValue: uuid.UUID{},
			expectedError: "address empty",
		},
		{
			name: "too short",
			id: MetadataAddress(ScopeMetadataAddress(scopePrimaryUUID).Bytes()[0:16]),
			expectedValue: uuid.UUID{},
			expectedError: "incorrect address length (must be at least 17, actual: 16)",
		},
		{
			name: "unknown type",
			id: MetadataAddress("This is not how this works."),
			expectedValue: uuid.UUID{},
			expectedError: "invalid address type out of valid range (got: 84)",
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
			name: "scope",
			id: ScopeMetadataAddress(uuid.New()),
			expectedValue: uuid.UUID{},
			expectedError: "invalid address type out of valid range (got: 0)",
		},
		{
			name: "session",
			id: SessionMetadataAddress(uuid.New(), sessionSecondaryUUID),
			expectedValue: sessionSecondaryUUID,
			expectedError: "",
		},
		{
			name: "record",
			id: RecordMetadataAddress(uuid.New(), "presence"),
			expectedValue: uuid.UUID{},
			expectedError: "invalid address type out of valid range (got: 2)",
		},
		{
			name: "scope spec",
			id: ScopeSpecMetadataAddress(uuid.New()),
			expectedValue: uuid.UUID{},
			expectedError: "invalid address type out of valid range (got: 4)",
		},
		{
			name: "contract spec",
			id: ContractSpecMetadataAddress(uuid.New()),
			expectedValue: uuid.UUID{},
			expectedError: "invalid address type out of valid range (got: 3)",
		},
		{
			name: "record spec",
			id: RecordSpecMetadataAddress(uuid.New(), "profiled"),
			expectedValue: uuid.UUID{},
			expectedError: "invalid address type out of valid range (got: 5)",
		},
		{
			name: "empty",
			id: MetadataAddress{},
			expectedValue: uuid.UUID{},
			expectedError: "address empty",
		},
		{
			name: "too short",
			id: MetadataAddress(SessionMetadataAddress(uuid.New(), sessionSecondaryUUID).Bytes()[0:32]),
			expectedValue: uuid.UUID{},
			expectedError: "incorrect address length (must be at least 33, actual: 32)",
		},
		{
			name: "unknown type",
			id: MetadataAddress("This is not how this works."),
			expectedValue: uuid.UUID{},
			expectedError: "invalid address type out of valid range (got: 84)",
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
			name: "scope",
			id: ScopeMetadataAddress(uuid.New()),
			expectedValue: []byte{},
			expectedError: "invalid address type out of valid range (got: 0)",
		},
		{
			name: "session",
			id: SessionMetadataAddress(uuid.New(), uuid.New()),
			expectedValue: []byte{},
			expectedError: "invalid address type out of valid range (got: 1)",
		},
		{
			name: "record",
			id: RecordMetadataAddress(uuid.New(), recordName),
			expectedValue: recordNameHash[0:16],
			expectedError: "",
		},
		{
			name: "scope spec",
			id: ScopeSpecMetadataAddress(uuid.New()),
			expectedValue: []byte{},
			expectedError: "invalid address type out of valid range (got: 4)",
		},
		{
			name: "contract spec",
			id: ContractSpecMetadataAddress(uuid.New()),
			expectedValue: []byte{},
			expectedError: "invalid address type out of valid range (got: 3)",
		},
		{
			name: "record spec",
			id: RecordSpecMetadataAddress(uuid.New(), recordSpecName),
			expectedValue: recordSpecNameHash[0:16],
			expectedError: "",
		},
		{
		name: 	"empty",
			id: MetadataAddress{},
			expectedValue: []byte{},
			expectedError: "address empty",
		},
		{
			name: "too short",
			id: MetadataAddress(RecordSpecMetadataAddress(uuid.New(), recordSpecName).Bytes()[0:32]),
			expectedValue: []byte{},
			expectedError: "incorrect address length (must be at least 33, actual: 32)",
		},
		{
			name: "unknown type",
			id: MetadataAddress("This is not how this works."),
			expectedValue: []byte{},
			expectedError: "invalid address type out of valid range (got: 84)",
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
			name: "scope id",
			baseID: scopeID,
			expectedID: scopeID,
			expectedError: "",
		},
		{
			name: "session id",
			baseID: sessionID,
			expectedID: scopeID,
			expectedError: "",
		},
		{
			name: "record id",
			baseID: recordID,
			expectedID: scopeID,
			expectedError: "",
		},
		{
			name: "scope spec id",
			baseID: scopeSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				scopeSpecID),
		},
		{
			name: "contract spec id",
			baseID: contractSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				contractSpecID),
		},
		{
			name: "record spec id",
			baseID: recordSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
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
			name: "scope id",
			baseID: scopeID,
			expectedID: sessionID,
			expectedError: "",
		},
		{
			name: "session id",
			baseID: sessionID,
			expectedID: sessionID,
			expectedError: "",
		},
		{
			name: "record id",
			baseID: recordID,
			expectedID: sessionID,
			expectedError: "",
		},
		{
			name: "scope spec id",
			baseID: scopeSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				scopeSpecID),
		},
		{
			name: "contract spec id",
			baseID: contractSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
				contractSpecID),
		},
		{
			name: "record spec id",
			baseID: recordSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
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
		name: 	"scope id",
		baseID: 	scopeID,
		expectedID: 	recordID,
		expectedError: 	"",
		},
		{
			name: "session id",
			baseID: sessionID,
			expectedID: recordID,
			expectedError: "",
		},
		{
			name: "record id",
			baseID: recordID,
			expectedID: recordID,
			expectedError: "",
		},
		{
			name: "scope spec id",
			baseID: scopeSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
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
			name: "record spec id",
			baseID: recordSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a scope uuid",
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
			name: "scope id",
			baseID: scopeID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeID),
		},
		{
			name: "session id",
			baseID: sessionID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				sessionID),
		},
		{
			name: "record id",
			baseID: recordID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				recordID),
		},
		{
			name: "scope spec id",
			baseID: scopeSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeSpecID),
		},
		{
			name: "contract spec id",
			baseID: contractSpecID,
			expectedID: recordSpecID,
			expectedError: "",
		},
		{
			name: "record spec id",
			baseID: recordSpecID,
			expectedID: recordSpecID,
			expectedError: "",
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
			name: "scope id",
			baseID: scopeID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeID),
		},
		{
			name: "session id",
			baseID: sessionID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				sessionID),
		},
		{
			name: "record id",
			baseID: recordID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				recordID),
		},
		{
			name: "scope spec id",
			baseID: scopeSpecID,
			expectedID: MetadataAddress{},
			expectedError: fmt.Sprintf("this metadata address (%s) does not contain a contract specification uuid",
				scopeSpecID),
		},
		{
			name: "contract spec id",
			baseID: contractSpecID,
			expectedID: contractSpecID,
			expectedError: "",
		},
		{
			name: "record spec id",
			baseID: recordSpecID,
			expectedID: contractSpecID,
			expectedError: "",
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

func (s *AddressTestSuite) TestFormat() {
	scopeID := ScopeMetadataAddress(s.scopeUUID)
	emptyID := MetadataAddress{}

	tests := []struct {
		name     string
		id       MetadataAddress
		format   string
		expected string
	}{
		{
			name : "format using %s",
			id : scopeID,
			format : "%s",
			expected : s.scopeBech32,
		},
		{
			// %p is for the address (in memory). Can't hard-code it.
			name : "format using %p",
			id : scopeID,
			format : "%p",
			expected : fmt.Sprintf("%p", scopeID),
		},
		{
			name : "format using %d - should use default %X",
			id : scopeID,
			format : "%d",
			expected : "008D80B25AC0894446956E5D08CFE3E1A5",
		},
		{
			name : "format using %v - should use default %X",
			id : scopeID,
			format : "%v",
			expected : "008D80B25AC0894446956E5D08CFE3E1A5",
		},
		{
			name : "format empty using %s",
			id : emptyID,
			format : "%s",
			expected : "",
		},
		{
			// %p is for the address (in memory). Can't hard-code it.
			name : "format using %p",
			id : emptyID,
			format : "%p",
			expected : fmt.Sprintf("%p", emptyID),
		},
		{
			name : "format empty using %d - should use default %X",
			id : emptyID,
			format : "%d",
			expected : "",
		},
		{
			name : "format empty using %v - should use default %X",
			id : emptyID,
			format : "%v",
			expected : "",
		},
		{
			name : "format %s is equal to .String() which fails on bad addresses",
			id : MetadataAddress("do not create MetadataAddresses this way"),
			format : "%s",
			expected : "%!s(PANIC=Format method: invalid metadata address type: 100)",
		},
	}

	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			actual := fmt.Sprintf(test.format, test.id)
			assert.Equal(t, test.expected, actual, test.name)
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
