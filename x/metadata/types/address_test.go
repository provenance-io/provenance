package types

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	scopeUUID   = uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")
	sessionUUID = uuid.MustParse("c25c7bd4-c639-4367-a842-f64fa5fccc19")

	scopeHex      = "008D80B25AC0894446956E5D08CFE3E1A5"
	scopeBech32   = "scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp"
	sessionBech32 = "session1qxxcpvj6czy5g354dews3nlruxjuyhrm6nrrjsm84pp0vna9lnxpjewp6kf"
	recordBech32  = "record1q2xcpvj6czy5g354dews3nlruxjelpkssxyyclt9ngh74gx9ttgp27gt8kl"
)

func requireBech32String (t *testing.T, typeCode []byte, data []byte) string {
	hrp, err := MetadataAddress{typeCode[0]}.Prefix()
	require.NoError(t, err, "getPrefix error")
	addr = append([]byte{typeCode[0]}, data...)
	bech32Addr, err := bech32.ConvertAndEncode(hrp, addr)
	require.NoError(t, err, "bech32.ConvertAndEncode error")
	return bech32Addr
}

func TestLegacySha512HashToAddress(t *testing.T) {
	testHashBytes := sha512.Sum512([]byte("test"))
	testHash := base64.StdEncoding.EncodeToString(testHashBytes[:])
	testHash15 := base64.StdEncoding.EncodeToString(testHashBytes[:15])
	testHash31 := base64.StdEncoding.EncodeToString(testHashBytes[:31])

	tests := []struct {
		name string
		typeBytes []byte
		hash string
		expectedAddr string
		expectedError string
	} {
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
			requireBech32String(t, ScopeKeyPrefix, testHashBytes[:16]),
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
			requireBech32String(t, SessionKeyPrefix, testHashBytes[:32]),
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
			requireBech32String(t, RecordKeyPrefix, testHashBytes[:32]),
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
			requireBech32String(t, ScopeSpecificationKeyPrefix, testHashBytes[:16]),
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
			requireBech32String(t, ContractSpecificationKeyPrefix, testHashBytes[:16]),
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
			requireBech32String(t, RecordSpecificationKeyPrefix, testHashBytes[:32]),
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
		t.Run(tc.name, func(t *testing.T) {
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

func TestMetadataAddressWithInvalidData(t *testing.T) {
	addr, addrErr := sdk.AccAddressFromBech32("cosmos1zgp4n2yvrtxkj5zl6rzcf6phqg0gfzuf3v08r4")
	require.NoError(t, addrErr, "address parsing error")

	_, err := VerifyMetadataAddressFormat(addr)
	require.EqualValues(t, fmt.Errorf("invalid metadata address type: %d", addr[0]), err)

	scopeID := ScopeMetadataAddress(scopeUUID)
	padded := make([]byte, 20)
	len, err := scopeID.MarshalTo(padded)
	require.EqualValues(t, 17, len)

	_, err = VerifyMetadataAddressFormat(padded)
	require.EqualValues(t, fmt.Errorf("incorrect address length (expected: %d, actual: %d)", 17, 20), err)

	_, err = MetadataAddressFromBech32("")
	require.EqualValues(t, errors.New("empty address string is not allowed"), err)

	_, err = MetadataAddressFromBech32("scope1qzxcpvj6czy5g354dews3nlruxjsahh")
	require.EqualValues(t, "decoding bech32 failed: checksum failed. Expected 57e9fl, got xjsahh.", err.Error())

	_, err = MetadataAddressFromHex("")
	require.EqualValues(t, errors.New("address decode failed: must provide an address"), err)

	_, err = MetadataAddressFromHex(scopeHex + "!!BAD")
	require.EqualValues(t, hex.InvalidByteError(0x21), err)

	var testMarshal MetadataAddress
	err = testMarshal.UnmarshalJSON([]byte(scopeBech32 + "{bad}{json}"))
	require.Error(t, err)
	err = testMarshal.UnmarshalYAML([]byte(scopeBech32 + "\n{badyaml}"))
	require.Error(t, err)
	err = testMarshal.Unmarshal([]byte{})
	require.NoError(t, err)
	err = testMarshal.UnmarshalJSON([]byte("\"\""))
	require.NoError(t, err)
	err = testMarshal.UnmarshalYAML([]byte("\"\""))
	require.NoError(t, err)
}

func TestMetadataAddressMarshal(t *testing.T) {
	var scopeID, newInstance MetadataAddress
	require.True(t, scopeID.Equals(newInstance), "two empty instances are equal")

	scopeID = ScopeMetadataAddress(scopeUUID)

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

func TestCompare(t *testing.T) {
	maEmpty := MetadataAddress{}
	ma1 := MetadataAddress("1")
	ma2 := MetadataAddress("2")
	ma11 := MetadataAddress("11")
	ma22 := MetadataAddress("22")

	tests := []struct{
		name string
		base MetadataAddress
		arg MetadataAddress
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
		t.Run(test.name, func(subtest *testing.T) {
			actual := test.base.Compare(test.arg)
			assert.Equal(subtest, test.expected, actual)
		})
	}
}

func TestMetadataAddressIteratorPrefix(t *testing.T) {
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

	scopeSpecID := ScopeSpecMetadataAddress(scopeUUID)
	bz, err = scopeSpecID.ScopeSessionIteratorPrefix()
	assert.EqualError(t, err,  "this metadata address does not contain a scope uuid", "scope spec id ScopeSessionIteratorPrefix error message")
	assert.Equal(t, []byte{}, bz, "scope spec id ScopeSessionIteratorPrefix value")
	bz, err = scopeSpecID.ScopeRecordIteratorPrefix()
	assert.EqualError(t, err,  "this metadata address does not contain a scope uuid", "scope spec id ScopeRecordIteratorPrefix error message")
	assert.Equal(t, []byte{}, bz, "scope spec id ScopeRecordIteratorPrefix value")

	scopeID := ScopeMetadataAddress(scopeUUID)
	bz, err = scopeID.ScopeSessionIteratorPrefix()
	require.NoError(t, err, "ScopeSessionIteratorPrefix error")
	require.Equal(t, 17, len(bz), "ScopeSessionIteratorPrefix length")
	require.Equal(t, SessionKeyPrefix[0], bz[0], "ScopeSessionIteratorPrefix first byte")
	bz, err = scopeID.ScopeRecordIteratorPrefix()
	require.NoError(t, err, "ScopeRecordIteratorPrefix err")
	require.Equal(t, 17, len(bz), "ScopeRecordIteratorPrefix length")
	require.Equal(t, RecordKeyPrefix[0], bz[0], "ScopeRecordIteratorPrefix first byte")

	contractSpecID := ContractSpecMetadataAddress(scopeUUID)
	bz, err = contractSpecID.ContractSpecRecordSpecIteratorPrefix()
	require.NoError(t, err, "ContractSpecRecordSpecIteratorPrefix error")
	require.Equal(t, 17, len(bz), "ContractSpecRecordSpecIteratorPrefix length")
	require.Equal(t, RecordSpecificationKeyPrefix[0], bz[0], "ContractSpecRecordSpecIteratorPrefix first byte")
}

func TestScopeMetadataAddress(t *testing.T) {
	// Make an address instance for a scope uuid
	scopeID := ScopeMetadataAddress(scopeUUID)
	require.NoError(t, scopeID.Validate())
	require.True(t, scopeID.IsScopeAddress())
	// Verify we can get a bech32 string for the scope
	require.Equal(t, scopeBech32, scopeID.String())

	// Verify we can get the uuid back
	scopeAddrUUID, err := scopeID.ScopeUUID()
	require.NoError(t, err)
	require.Equal(t, "8d80b25a-c089-4446-956e-5d08cfe3e1a5", scopeAddrUUID.String())

	_, err = scopeID.SessionUUID()
	require.Error(t, fmt.Errorf("this metadata addresss does not contain a session uuid"), err)

	// Check the string formatter for the scopeID
	require.Equal(t, scopeBech32, fmt.Sprintf("%s", scopeID))
	require.Equal(t, fmt.Sprintf("%s", scopeHex), fmt.Sprintf("%X", scopeID))

	// Ensure a second instance is equal to the first
	scopeID2 := ScopeMetadataAddress(scopeUUID)
	require.True(t, scopeID.Equals(scopeID2))
	require.False(t, scopeID.Equals(ScopeMetadataAddress(sessionUUID)))

	json, err := scopeID.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("\"%s\"", scopeBech32), string(json))

	yaml, err := scopeID.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, scopeBech32, yaml)

	var yamlAddress MetadataAddress
	yamlAddress.UnmarshalYAML([]byte(yaml.(string)))
	require.EqualValues(t, scopeID, yamlAddress)

	var jsonAddress MetadataAddress
	jsonAddress.UnmarshalJSON(json)
	require.EqualValues(t, scopeID, jsonAddress)
}

func TestSessionMetadataAddress(t *testing.T) {
	// Construct a composite key for a session within a scope
	sessionAddress := SessionMetadataAddress(scopeUUID, sessionUUID)
	require.NoError(t, sessionAddress.Validate(), "expect a valid MetadataAddress for a session")
	require.True(t, sessionAddress.IsSessionAddress())

	scopeUUIDFromSessionID, err := sessionAddress.ScopeUUID()
	require.NoError(t, err, "there should be no errors getting a scope uuid from the session address")
	require.Equal(t, "8d80b25a-c089-4446-956e-5d08cfe3e1a5", scopeUUIDFromSessionID.String())

	sessionUUID, err := sessionAddress.SessionUUID()
	require.NoError(t, err, "there should be no error getting the session uuid from the session address")
	require.Equal(t, "c25c7bd4-c639-4367-a842-f64fa5fccc19", sessionUUID.String(), "the session uuid should be recoverable")

	require.Equal(t, sessionBech32, sessionAddress.String())

	json, err := sessionAddress.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("\"%s\"", sessionBech32), string(json))

	yaml, err := sessionAddress.MarshalYAML()
	require.NoError(t, err)
	require.Equal(t, sessionBech32, yaml)

	var yamlAddress MetadataAddress
	require.True(t, yamlAddress.Empty())
	yamlAddress.UnmarshalYAML([]byte(yaml.(string)))
	require.EqualValues(t, sessionAddress, yamlAddress)

	var jsonAddress MetadataAddress
	require.True(t, jsonAddress.Empty())
	jsonAddress.UnmarshalJSON(json)
	require.EqualValues(t, sessionAddress, jsonAddress)
}

func TestRecordMetadataAddress(t *testing.T) {
	// Construct a composite key for a record within a scope
	scopeID := ScopeMetadataAddress(scopeUUID)
	recordID := RecordMetadataAddress(scopeUUID, "test")
	require.True(t, recordID.IsRecordAddress())
	require.Equal(t, recordBech32, recordID.String())
	recordAddress, err := MetadataAddressFromBech32(recordBech32)
	require.NoError(t, err)
	require.Equal(t, recordID, recordAddress)

	require.Equal(t, recordID, RecordMetadataAddress(scopeUUID, "tEst"))
	require.Equal(t, recordID, RecordMetadataAddress(scopeUUID, "TEST"))
	require.Equal(t, recordID, RecordMetadataAddress(scopeUUID, "   test   "))

	recAddrFromScopeID, err := scopeID.AsRecordAddress("test")
	require.NoError(t, err, "AsRecordAddress error")
	require.Equal(t, recordID, recAddrFromScopeID, "AsRecordAddress value")
}

func TestScopeSpecMetadataAddress(t *testing.T) {
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

func TestContractSpecMetadataAddress(t *testing.T) {
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

func TestRecordSpecMetadataAddress(t *testing.T) {
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

func TestMetadataAddressTypeTestFuncs(t *testing.T) {
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
		t.Run(test.name, func(subtest *testing.T) {
			assert.Equal(subtest, test.expected[0], test.id.IsScopeAddress(), fmt.Sprintf("%s: IsScopeAddress", test.name))
			assert.Equal(subtest, test.expected[1], test.id.IsSessionAddress(), fmt.Sprintf("%s: IsSessionAddress", test.name))
			assert.Equal(subtest, test.expected[2], test.id.IsRecordAddress(), fmt.Sprintf("%s: IsRecordAddress", test.name))
			assert.Equal(subtest, test.expected[3], test.id.IsScopeSpecificationAddress(), fmt.Sprintf("%s: IsScopeSpecificationAddress", test.name))
			assert.Equal(subtest, test.expected[4], test.id.IsContractSpecificationAddress(), fmt.Sprintf("%s: IsContractSpecificationAddress", test.name))
			assert.Equal(subtest, test.expected[5], test.id.IsRecordSpecificationAddress(), fmt.Sprintf("%s: IsRecordSpecificationAddress", test.name))
		})
	}
}

func TestPrefix(t *testing.T) {
	t.Run("scope", func(subtest *testing.T) {
		actualScopePrefix, actualScopePrefixErr := ScopeMetadataAddress(uuid.New()).Prefix()
		assert.NoError(subtest, actualScopePrefixErr, "actualScopePrefixErr")
		assert.Equal(subtest, PrefixScope, actualScopePrefix, "actualScopePrefix")
	})

	t.Run("scope without address", func(subtest *testing.T) {
		addr := MetadataAddress{ScopeKeyPrefix[0]}
		actual, err := addr.Prefix()
		assert.NoError(subtest, err, "Prefix error")
		assert.Equal(subtest, PrefixScope, actual, "Prefix value")
	})

	t.Run("session", func(subtest *testing.T) {
		actualSessionPrefix, actualSessionPrefixErr := SessionMetadataAddress(uuid.New(), uuid.New()).Prefix()
		assert.NoError(subtest, actualSessionPrefixErr, "actualSessionPrefixErr")
		assert.Equal(subtest, PrefixSession, actualSessionPrefix, "actualSessionPrefix")
	})

	t.Run("session without address", func(subtest *testing.T) {
		addr := MetadataAddress{SessionKeyPrefix[0]}
		actual, err := addr.Prefix()
		assert.NoError(subtest, err, "Prefix error")
		assert.Equal(subtest, PrefixSession, actual, "Prefix value")
	})

	t.Run("record", func(subtest *testing.T) {
		actualRecordPrefix, actualRecordPrefixErr := RecordMetadataAddress(uuid.New(), "ronald").Prefix()
		assert.NoError(subtest, actualRecordPrefixErr, "actualRecordPrefixErr")
		assert.Equal(subtest, PrefixRecord, actualRecordPrefix, "actualRecordPrefix")
	})

	t.Run("record without address", func(subtest *testing.T) {
		addr := MetadataAddress{RecordKeyPrefix[0]}
		actual, err := addr.Prefix()
		assert.NoError(subtest, err, "Prefix error")
		assert.Equal(subtest, PrefixRecord, actual, "Prefix value")
	})

	t.Run("scope spec", func(subtest *testing.T) {
		actualScopeSpecPrefix, actualScopeSpecPrefixErr := ScopeSpecMetadataAddress(uuid.New()).Prefix()
		assert.NoError(subtest, actualScopeSpecPrefixErr, "actualScopeSpecPrefixErr")
		assert.Equal(subtest, PrefixScopeSpecification, actualScopeSpecPrefix, "actualScopeSpecPrefix")
	})

	t.Run("scope spec without address", func(subtest *testing.T) {
		addr := MetadataAddress{ScopeSpecificationKeyPrefix[0]}
		actual, err := addr.Prefix()
		assert.NoError(subtest, err, "Prefix error")
		assert.Equal(subtest, PrefixScopeSpecification, actual, "Prefix value")
	})

	t.Run("contract spec", func(subtest *testing.T) {
		actualContractSpecPrefix, actualContractSpecPrefixErr := ContractSpecMetadataAddress(uuid.New()).Prefix()
		assert.NoError(subtest, actualContractSpecPrefixErr, "actualContractSpecPrefixErr")
		assert.Equal(subtest, PrefixContractSpecification, actualContractSpecPrefix, "actualContractSpecPrefix")
	})

	t.Run("contract spec without address", func(subtest *testing.T) {
		addr := MetadataAddress{ContractSpecificationKeyPrefix[0]}
		actual, err := addr.Prefix()
		assert.NoError(subtest, err, "Prefix error")
		assert.Equal(subtest, PrefixContractSpecification, actual, "Prefix value")
	})

	t.Run("record spec", func(subtest *testing.T) {
		actualRecordSpecPrefix, actualRecordSpecPrefixErr := RecordSpecMetadataAddress(uuid.New(), "george").Prefix()
		assert.NoError(subtest, actualRecordSpecPrefixErr, "actualRecordSpecPrefixErr")
		assert.Equal(subtest, PrefixRecordSpecification, actualRecordSpecPrefix, "actualRecordSpecPrefix")
	})

	t.Run("record spec without address", func(subtest *testing.T) {
		addr := MetadataAddress{RecordSpecificationKeyPrefix[0]}
		actual, err := addr.Prefix()
		assert.NoError(subtest, err, "Prefix error")
		assert.Equal(subtest, PrefixRecordSpecification, actual, "Prefix value")
	})

	t.Run("bad", func(subtest *testing.T) {
		_, badPrefixErr := MetadataAddress("don't do this").Prefix()
		assert.Error(subtest, badPrefixErr, "badPrefixErr")
	})
}

func TestPrimaryUUID(t *testing.T) {
	scopePrimaryUUID := uuid.New()
	sessionPrimaryUUID := uuid.New()
	recordPrimaryUUID := uuid.New()
	scopeSpecPrimaryUUID := uuid.New()
	contractSpecPrimaryUUID := uuid.New()
	recordSpecPrimaryUUID := uuid.New()

	tests := []struct {
		name string
		id MetadataAddress
		expectedValue uuid.UUID
		expectedError string
	} {
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
		t.Run(test.name, func(subtest *testing.T) {
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

func TestSecondaryUUID(t *testing.T) {
	sessionSecondaryUUID := uuid.New()

	tests := []struct {
		name string
		id MetadataAddress
		expectedValue uuid.UUID
		expectedError string
	} {
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
		t.Run(test.name, func(subtest *testing.T) {
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

func TestNameHash(t *testing.T) {
	recordName := "mothership"
	recordNameHash := sha256.Sum256([]byte(recordName))
	recordSpecName := "houses of the holy"
	recordSpecNameHash := sha256.Sum256([]byte(recordSpecName))

	tests := []struct {
		name string
		id MetadataAddress
		expectedValue []byte
		expectedError string
	} {
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
		t.Run(test.name, func(subtest *testing.T) {
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

func TestScopeAddressConverters(t *testing.T) {
	randomUUID := uuid.New()
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, uuid.New())
	recordID := RecordMetadataAddress(randomUUID, "year zero")
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, "the downard spiral")

	tests := []struct {
		name string
		baseID MetadataAddress
		expectedID MetadataAddress
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
		t.Run(fmt.Sprintf("%s AsScopeAddress", test.name), func(subtest *testing.T) {
			actualID, err := test.baseID.AsScopeAddress()
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, "%s AsScopeAddress err", test.name)
				assert.Equal(t, test.expectedID, actualID, "%s AsScopeAddress value", test.name)
			} else {
				assert.EqualError(t, err, test.expectedError, "%s AsScopeAddress expected err", test.name)
			}
		})
	}
}

func TestSessionAddressConverters(t *testing.T) {
	randomUUID := uuid.New()
	randomUUID2 := uuid.New()
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, randomUUID2)
	recordID := RecordMetadataAddress(randomUUID, "pet sounds")
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, "smile")

	tests := []struct {
		name string
		baseID MetadataAddress
		expectedID MetadataAddress
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
		t.Run(fmt.Sprintf("%s AsSessionAddress", test.name), func(subtest *testing.T) {
			actualID, err := test.baseID.AsSessionAddress(randomUUID2)
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, "%s AsSessionAddress err", test.name)
				assert.Equal(t, test.expectedID, actualID, "%s AsSessionAddress value", test.name)
			} else {
				assert.EqualError(t, err, test.expectedError, "%s AsSessionAddress expected err", test.name)
			}
		})
	}
}

func TestRecordAddressConverters(t *testing.T) {
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
		name string
		baseID MetadataAddress
		expectedID MetadataAddress
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
			t.Run(fmt.Sprintf("%s AsRecordAddress(\"%s\")", test.name, rName), func(subtest *testing.T) {
				actualID, err := test.baseID.AsRecordAddress(rName)
				if len(test.expectedError) == 0 {
					assert.NoError(t, err, "%s AsRecordAddress err", test.name)
					assert.Equal(t, test.expectedID, actualID, "%s AsRecordAddress value", test.name)
				} else {
					assert.EqualError(t, err, test.expectedError, "%s AsRecordAddress expected err", test.name)
				}
			})
		}
	}
}

func TestRecordSpecAddressConverters(t *testing.T) {
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
		name string
		baseID MetadataAddress
		expectedID MetadataAddress
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
			t.Run(fmt.Sprintf("%s AsRecordSpecAddress(\"%s\")", test.name, rName), func(subtest *testing.T) {
				actualID, err := test.baseID.AsRecordSpecAddress(rName)
				if len(test.expectedError) == 0 {
					assert.NoError(t, err, "%s AsRecordSpecAddress err", test.name)
					assert.Equal(t, test.expectedID, actualID, "%s AsRecordSpecAddress value", test.name)
				} else {
					assert.EqualError(t, err, test.expectedError, "%s AsRecordSpecAddress expected err", test.name)
				}
			})
		}
	}
}

func TestContractSpecAddressConverters(t *testing.T) {
	randomUUID := uuid.New()
	scopeID := ScopeMetadataAddress(randomUUID)
	sessionID := SessionMetadataAddress(randomUUID, uuid.New())
	recordID := RecordMetadataAddress(randomUUID, "pretty hate machine")
	scopeSpecID := ScopeSpecMetadataAddress(randomUUID)
	contractSpecID := ContractSpecMetadataAddress(randomUUID)
	recordSpecID := RecordSpecMetadataAddress(randomUUID, "with teeth")

	tests := []struct {
		name string
		baseID MetadataAddress
		expectedID MetadataAddress
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
		t.Run(fmt.Sprintf("%s AsContractSpecAddress", test.name), func(subtest *testing.T) {
			actualID, err := test.baseID.AsContractSpecAddress()
			if len(test.expectedError) == 0 {
				assert.NoError(t, err, "%s AsContractSpecAddress err", test.name)
				assert.Equal(t, test.expectedID, actualID, "%s AsContractSpecAddress value", test.name)
			} else {
				assert.EqualError(t, err, test.expectedError, "%s AsContractSpecAddress expected err", test.name)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	scopeID := ScopeMetadataAddress(scopeUUID)
	emptyID := MetadataAddress{}

	tests := []struct {
		name string
		id MetadataAddress
		format string
		expected string
	}{
		{
			"format using %s",
			scopeID,
			"%s",
			scopeBech32,
		},
		{
			// %p is for the address (in memory). Can't hard-code it.
			"format using %p",
			scopeID,
			"%p",
			fmt.Sprintf("%p", scopeID),
		},
		{
			"format using %d - should use default %X",
			scopeID,
			"%d",
			"008D80B25AC0894446956E5D08CFE3E1A5",
		},
		{
			"format using %v - should use default %X",
			scopeID,
			"%v",
			"008D80B25AC0894446956E5D08CFE3E1A5",
		},
		{
			"format empty using %s",
			emptyID,
			"%s",
			"",
		},
		{
			// %p is for the address (in memory). Can't hard-code it.
			"format using %p",
			emptyID,
			"%p",
			fmt.Sprintf("%p", emptyID),
		},
		{
			"format empty using %d - should use default %X",
			emptyID,
			"%d",
			"",
		},
		{
			"format empty using %v - should use default %X",
			emptyID,
			"%v",
			"",
		},
		{
			"format %s is equal to .String() which fails on bad addresses",
			MetadataAddress("do not create MetadataAddresses this way"),
			"%s",
			"%!s(PANIC=Format method: invalid metadata address type: 100)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(subtest *testing.T) {
			actual := fmt.Sprintf(test.format, test.id)
			assert.Equal(t, test.expected, actual, test.name)
		})
	}
}
