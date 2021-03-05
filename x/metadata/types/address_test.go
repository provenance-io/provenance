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
	"github.com/stretchr/testify/require"
)

var (
	scopeUUID   = uuid.MustParse("8d80b25a-c089-4446-956e-5d08cfe3e1a5")
	sessionUUID = uuid.MustParse("c25c7bd4-c639-4367-a842-f64fa5fccc19")

	scopeHex      = "008D80B25AC0894446956E5D08CFE3E1A5"
	scopeBech32   = "scope1qzxcpvj6czy5g354dews3nlruxjsahhnsp"
	sessionBech32 = "session1qxxcpvj6czy5g354dews3nlruxjuyhrm6nrrjsm84pp0vna9lnxpjewp6kf"
	recordBech32  = "record1q2xcpvj6czy5g354dews3nlruxjelpkssxyyclt9ngh74gx9ttgptgalfudjkzuz9ng46mq4krcq5zqsttnk0"
)

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

	scopeUUID, err := sessionAddress.ScopeUUID()
	require.NoError(t, err, "there should be no errors getting a scope uuid from the session address")
	require.Equal(t, "8d80b25a-c089-4446-956e-5d08cfe3e1a5", scopeUUID.String())

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
	require.Equal(t, recordID, scopeID.GetRecordAddress("test"))
}

func TestRecordSpecMetadataAddress(t *testing.T) {
	contractSpecUUID := uuid.New()
	contractSpecID := ContractSpecMetadataAddress(contractSpecUUID)
	recordSpecID := RecordSpecMetadataAddress(contractSpecUUID, "myname")
	nameHash := sha256.Sum256([]byte("myname"))

	require.True(t, recordSpecID.IsRecordSpecificationAddress(), "IsRecordAddress")
	require.Equal(t, RecordSpecificationKeyPrefix, recordSpecID[0:1].Bytes(), "bytes[0]: the type bit")
	require.Equal(t, contractSpecID[1:17], recordSpecID[1:17], "bytes[1:17]: the contract spec id bytes")
	require.Equal(t, nameHash[:], recordSpecID[17:49].Bytes(), "bytes[17:49]: the hashed name")

	recordSpecBech32 := recordSpecID.String()
	recordSpecIDFromBeck32, errBeck32 := MetadataAddressFromBech32(recordSpecBech32)
	require.NoError(t, errBeck32, "error from MetadataAddressFromBech32")
	require.Equal(t, recordSpecID, recordSpecIDFromBeck32, "value from recordSpecIDFromBeck32")

	require.Equal(t, recordSpecID, RecordSpecMetadataAddress(contractSpecUUID, "MyName"), "camel case")
	require.Equal(t, recordSpecID, RecordSpecMetadataAddress(contractSpecUUID, "MYNAME"), "all caps")
	require.Equal(t, recordSpecID, RecordSpecMetadataAddress(contractSpecUUID, "   myname   "), "padded with spaces")
	require.Equal(t, recordSpecID, contractSpecID.GetRecordSpecAddress("myname"), "from contract spec id")

	contractSpecUUIDFromRecordSpecId, errContractSpecUUID := recordSpecID.ContractSpecUUID()
	require.NoError(t, errContractSpecUUID, "error from ContractSpecUUID")
	require.Equal(t, contractSpecUUID, contractSpecUUIDFromRecordSpecId, "value from ContractSpecUUID")
}
