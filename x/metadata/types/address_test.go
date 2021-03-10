package types

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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

func assertPanic(t *testing.T, testName string, expectedPanicMsg string, test func()) {
	// Recover from the panic if there is one
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, expectedPanicMsg, fmt.Sprintf("%s", r), "%s panic message", testName)
		}
	}()
	// Run the test that should cause a panic
	test()
	// If the function panicked, code execution doesn't get to this point.
	// And the deferred function at the top will recover from the panic,
	// allowing further testing to continue.
	// But if it didn't panic, then we need to indicate that the test failed.
	t.Errorf("%s: should have caused panic(\"%s\") but did not.", testName, expectedPanicMsg)
}

func assertDoesNotPanic(t *testing.T, testName string, test func()) {
	// If there's a panic, fail the test and recover from it
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: should not have caused panic but did: %v", testName, r)
		}
	}()
	// Run the test that should cause a panic
	test()
	// If the function panicked, code execution doesn't get to this point.
	// And the deferred function at the top will recover from the panic,
	// allowing further testing to continue.

}

func TestLegacySha512HashToAddress(t *testing.T) {
	testHash := sha512.Sum512([]byte("test"))
	hash := base64.StdEncoding.EncodeToString(testHash[:])
	specAddress, err := ConvertHashToAddress(ScopeSpecificationKeyPrefix, hash)
	require.NoError(t, err)
	require.NoError(t, specAddress.Validate())
	require.True(t, specAddress.IsScopeSpecificationAddress())
	// hash of "test" is consistent, this address should be too.
	require.Equal(t, "scopespec1qnhzdvxaftm7wjd2r28w8sg2axfqhzcsx0", specAddress.String())

	_, err = ConvertHashToAddress(RecordKeyPrefix, hash)
	require.Error(t, err)
	require.Equal(t, fmt.Errorf("invalid address type code 0x02, expected 0x00, 0x03, or 0x04"), err)

	_, err = ConvertHashToAddress(ScopeKeyPrefix, "invalid hash")
	require.Error(t, err)
	require.Equal(t, base64.CorruptInputError(7), err)

	_, err = ConvertHashToAddress(ScopeKeyPrefix, "MA==") // 0
	require.Error(t, err)
	require.Equal(t, fmt.Errorf("invalid specification identifier, expected at least 16 bytes, found 1"), err)
}

func TestMetadataAddressWithInvalidData(t *testing.T) {
	// TODO - this must be made static because the random bytes sometimes move past first error check breaking
	// the checks below.
	var addr = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	_, err := VerifyMetadataAddressFormat(addr)
	require.EqualValues(t, fmt.Errorf("invalid metadata address type: %d", addr[0]), err)
	_, err = VerifyMetadataAddressFormat(addr[0:12])
	require.EqualValues(t, fmt.Errorf("incorrect address length (must be at least 17, actual: %d)", 12), err)

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

// TODO: Tests on the Compare func

func TestMetadataAddressIteratorPrefix(t *testing.T) {
	var scopeID MetadataAddress
	bz, err := scopeID.ScopeSessionIteratorPrefix()
	require.NoError(t, err, "empty address must return the root iterator prefix ")
	require.Equal(t, SessionKeyPrefix, bz)
	bz, err = scopeID.ScopeRecordIteratorPrefix()
	require.NoError(t, err, "empty address must return the root iterator prefix")
	require.Equal(t, RecordKeyPrefix, bz)

	scopeID = ScopeSpecMetadataAddress(scopeUUID)
	bz, err = scopeID.ScopeRecordIteratorPrefix()
	require.Error(t, err, "not possible to iterate Records off a scope specification")
	require.Equal(t, []byte{}, bz)
	require.Equal(t, fmt.Errorf("this metadata address does not contain a scope uuid"), err)

	scopeID = ScopeMetadataAddress(scopeUUID)

	bz, err = scopeID.ScopeSessionIteratorPrefix()
	require.NoError(t, err)
	require.Equal(t, 17, len(bz), "length of iterator prefix is code plus scope uuid")
	require.Equal(t, SessionKeyPrefix[0], bz[0], "iterator prefix should start with session key id")

	bz, err = scopeID.ScopeRecordIteratorPrefix()
	require.NoError(t, err)
	require.Equal(t, 17, len(bz), "length of iterator prefix is code plus scope uuid")
	require.Equal(t, RecordKeyPrefix[0], bz[0], "iterator prefix should start with session key id")

	// TODO: ContractSpecRecordSpecIteratorPrefix test
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
	require.Equal(t, recordID, scopeID.AsRecordAddressE("test"))
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
	require.Equal(t, recordSpecID, contractSpecID.AsRecordSpecAddressE("myname"), "from contract spec id")

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
		assert.Equal(t, test.expected[0], test.id.IsScopeAddress(), fmt.Sprintf("%s: IsScopeAddress", test.name))
		assert.Equal(t, test.expected[1], test.id.IsSessionAddress(), fmt.Sprintf("%s: IsSessionAddress", test.name))
		assert.Equal(t, test.expected[2], test.id.IsRecordAddress(), fmt.Sprintf("%s: IsRecordAddress", test.name))
		assert.Equal(t, test.expected[3], test.id.IsScopeSpecificationAddress(), fmt.Sprintf("%s: IsScopeSpecificationAddress", test.name))
		assert.Equal(t, test.expected[4], test.id.IsContractSpecificationAddress(), fmt.Sprintf("%s: IsContractSpecificationAddress", test.name))
		assert.Equal(t, test.expected[5], test.id.IsRecordSpecificationAddress(), fmt.Sprintf("%s: IsRecordSpecificationAddress", test.name))
	}
}

func TestPrefix(t *testing.T) {
	actualScopePrefix, actualScopePrefixErr := ScopeMetadataAddress(uuid.New()).Prefix()
	assert.NoError(t, actualScopePrefixErr, "actualScopePrefixErr")
	assert.Equal(t, PrefixScope, actualScopePrefix, "actualScopePrefix")

	actualSessionPrefix, actualSessionPrefixErr := SessionMetadataAddress(uuid.New(), uuid.New()).Prefix()
	assert.NoError(t, actualSessionPrefixErr, "actualSessionPrefixErr")
	assert.Equal(t, PrefixSession, actualSessionPrefix, "actualSessionPrefix")

	actualRecordPrefix, actualRecordPrefixErr := RecordMetadataAddress(uuid.New(), "ronald").Prefix()
	assert.NoError(t, actualRecordPrefixErr, "actualRecordPrefixErr")
	assert.Equal(t, PrefixRecord, actualRecordPrefix, "actualRecordPrefix")

	actualScopeSpecPrefix, actualScopeSpecPrefixErr := ScopeSpecMetadataAddress(uuid.New()).Prefix()
	assert.NoError(t, actualScopeSpecPrefixErr, "actualScopeSpecPrefixErr")
	assert.Equal(t, PrefixScopeSpecification, actualScopeSpecPrefix, "actualScopeSpecPrefix")

	actualContractSpecPrefix, actualContractSpecPrefixErr := ContractSpecMetadataAddress(uuid.New()).Prefix()
	assert.NoError(t, actualContractSpecPrefixErr, "actualContractSpecPrefixErr")
	assert.Equal(t, PrefixContractSpecification, actualContractSpecPrefix, "actualContractSpecPrefix")

	actualRecordSpecPrefix, actualRecordSpecPrefixErr := RecordSpecMetadataAddress(uuid.New(), "george").Prefix()
	assert.NoError(t, actualRecordSpecPrefixErr, "actualRecordSpecPrefixErr")
	assert.Equal(t, PrefixRecordSpecification, actualRecordSpecPrefix, "actualRecordSpecPrefix")

	_, badPrefixErr := MetadataAddress("don't do this").Prefix()
	assert.Error(t, badPrefixErr, "badPrefixErr")
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
		primaryUUID, err := test.id.PrimaryUUID()
		if len(test.expectedError) == 0 {
			assert.NoError(t, err, fmt.Sprintf("%s: err", test.name))
			assert.Equal(t, test.expectedValue, primaryUUID, fmt.Sprintf("%s: value", test.name))
		} else {
			assert.EqualError(t, err, test.expectedError, fmt.Sprintf("%s: err", test.name))
		}
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
		secondaryUUID, err := test.id.SecondaryUUID()
		if len(test.expectedError) == 0 {
			assert.NoError(t, err, fmt.Sprintf("%s: err", test.name))
			assert.Equal(t, test.expectedValue, secondaryUUID, fmt.Sprintf("%s: value", test.name))
		} else {
			assert.EqualError(t, err, test.expectedError, fmt.Sprintf("%s: err", test.name))
		}
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
		nameHash, err := test.id.NameHash()
		if len(test.expectedError) == 0 {
			assert.NoError(t, err, fmt.Sprintf("%s: err", test.name))
			assert.Equal(t, test.expectedValue, nameHash, fmt.Sprintf("%s: value", test.name))
		} else {
			assert.EqualError(t, err, test.expectedError, fmt.Sprintf("%s: err", test.name))
		}
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
		// Test AsScopeAddress
		actualID, err := test.baseID.AsScopeAddress()
		if len(test.expectedError) == 0 {
			assert.NoError(t, err, "%s AsScopeAddress err", test.name)
			assert.Equal(t, test.expectedID, actualID, "%s AsScopeAddress value", test.name)
		} else {
			assert.EqualError(t, err, test.expectedError, "%s AsScopeAddress expected err", test.name)
		}

		// Test AsScopeAddressE
		if len(test.expectedError) == 0 {
			assertDoesNotPanic(t, fmt.Sprintf("%s AsScopeAddressE", test.name), func() {
				id := test.baseID.AsScopeAddressE()
				assert.Equal(t, test.expectedID, id, "%s AsScopeAddressE value", test.name)
			})
		} else {
			assertPanic(t, fmt.Sprintf("%s AsScopeAddressE", test.name), test.expectedError, func() {
				id := test.baseID.AsScopeAddressE()
				t.Errorf("%s AsScopeAddressE returned %s instead of panicking with message %s",
					test.name, id, test.expectedError)
			})
		}
	}
}

// TODO: AsRecordAddressE and AsRecordAddress tests
// TODO: AsRecordSpecAddressE and AsRecordSpecAddress tests
// TODO: AsContractSpecAddressE and AsContractSpecAddress tests

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
		actual := fmt.Sprintf(test.format, test.id)
		assert.Equal(t, test.expected, actual, test.name)
	}
}
