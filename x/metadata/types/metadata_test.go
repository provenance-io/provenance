package types

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MetadataTypesTestSuite struct {
	suite.Suite

	scopeUUIDStr        string
	sessionUUIDStr      string
	scopeSpecUUIDStr    string
	contractSpecUUIDStr string
	recordName          string

	scopeIDStr        string
	sessionIDStr      string
	recordIDStr       string
	scopeSpecIDStr    string
	contractSpecIDStr string
	recordSpecIDStr   string

	scopeUUID        uuid.UUID
	sessionUUID      uuid.UUID
	scopeSpecUUID    uuid.UUID
	contractSpecUUID uuid.UUID
	recordNameHash   []byte

	scopeID        MetadataAddress
	sessionID      MetadataAddress
	recordID       MetadataAddress
	scopeSpecID    MetadataAddress
	contractSpecID MetadataAddress
	recordSpecID   MetadataAddress
}

func (s *MetadataTypesTestSuite) SetupTest() {
	// Hard coded id components.
	s.scopeUUIDStr = "91978ba2-5f35-459a-86a7-feca1b0512e0"
	s.sessionUUIDStr = "5803f8bc-6067-4eb5-951f-2121671c2ec0"
	s.scopeSpecUUIDStr = "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"
	s.contractSpecUUIDStr = "def6bc0a-c9dd-4874-948f-5206e6060a84"
	s.recordName = "recordname"

	// Hard coded ids generated previously using the hard coded id components.
	s.scopeIDStr = "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"
	s.sessionIDStr = "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"
	s.recordIDStr = "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"
	s.scopeSpecIDStr = "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"
	s.contractSpecIDStr = "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"
	s.recordSpecIDStr = "recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"

	// Parsed uuid strings.
	s.scopeUUID = uuid.MustParse(s.scopeUUIDStr)
	s.sessionUUID = uuid.MustParse(s.sessionUUIDStr)
	s.scopeSpecUUID = uuid.MustParse(s.scopeSpecUUIDStr)
	s.contractSpecUUID = uuid.MustParse(s.contractSpecUUIDStr)

	// Hash the record name
	recordNameSum := sha256.Sum256([]byte(s.recordName))
	s.recordNameHash = recordNameSum[0:16]

	// Creating Metadata addresses from the components (that should equal the string versions above).
	s.scopeID = ScopeMetadataAddress(s.scopeUUID)
	s.sessionID = SessionMetadataAddress(s.scopeUUID, s.sessionUUID)
	s.recordID = RecordMetadataAddress(s.scopeUUID, s.recordName)
	s.scopeSpecID = ScopeSpecMetadataAddress(s.scopeSpecUUID)
	s.contractSpecID = ContractSpecMetadataAddress(s.contractSpecUUID)
	s.recordSpecID = RecordSpecMetadataAddress(s.contractSpecUUID, s.recordName)
}

func TestMetadataTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataTypesTestSuite))
}

func (s *MetadataTypesTestSuite) TestGetScopeIDInfo() {
	s.T().Run("empty address, empty object", func(t *testing.T) {
		info := GetScopeIDInfo(MetadataAddress{})
		require.NotNil(t, info, "GetScopeIDInfo result")
		assert.True(t, info.ScopeId.Empty(), "empty ScopeId")
		assert.Equal(t, 0, len(info.ScopeIdPrefix), "ScopeIdPrefix length")
		assert.Equal(t, 0, len(info.ScopeIdScopeUuid), "ScopeIdScopeUuid length")
		assert.Equal(t, "", info.ScopeAddr, "ScopeAddr")
		assert.Equal(t, "", info.ScopeUuid, "ScopeUuid")
	})

	s.T().Run("scope id, everything populated", func(t *testing.T) {
		info := GetScopeIDInfo(s.scopeID)
		require.NotNil(t, info, "GetScopeIDInfo result")
		assert.Equal(t, s.scopeID, info.ScopeId, "ScopeId")
		assert.Equal(t, ScopeKeyPrefix, info.ScopeIdPrefix, "ScopeIdPrefix")
		assert.Equal(t, s.scopeUUID[:], info.ScopeIdScopeUuid, "ScopeIdScopeUuid")
		assert.Equal(t, s.scopeIDStr, info.ScopeAddr, "ScopeAddr")
		assert.Equal(t, s.scopeUUIDStr, info.ScopeUuid, "ScopeUuid")
	})
}

func (s *MetadataTypesTestSuite) TestGetSessionIDInfo() {
	s.T().Run("empty address, empty object", func(t *testing.T) {
		info := GetSessionIDInfo(MetadataAddress{})
		require.NotNil(t, info, "GetSessionIDInfo result")
		assert.True(t, info.SessionId.Empty(), "empty SessionId")
		assert.Equal(t, 0, len(info.SessionIdPrefix), "SessionIdPrefix length")
		assert.Equal(t, 0, len(info.SessionIdScopeUuid), "SessionIdScopeUuid length")
		assert.Equal(t, 0, len(info.SessionIdSessionUuid), "SessionIdSessionUuid length")
		assert.Equal(t, "", info.SessionAddr, "SessionAddr")
		assert.Equal(t, "", info.SessionUuid, "SessionUuid")
		require.NotNil(t, info.ScopeIdInfo, "info.ScopeIdInfo")
		assert.True(t, info.ScopeIdInfo.ScopeId.Empty(), "empty ScopeIdInfo.ScopeId")
		assert.Equal(t, 0, len(info.ScopeIdInfo.ScopeIdPrefix), "ScopeIdInfo.ScopeIdPrefix length")
		assert.Equal(t, 0, len(info.ScopeIdInfo.ScopeIdScopeUuid), "ScopeIdInfo.ScopeIdScopeUuid length")
		assert.Equal(t, "", info.ScopeIdInfo.ScopeAddr, "ScopeIdInfo.ScopeAddr")
		assert.Equal(t, "", info.ScopeIdInfo.ScopeUuid, "ScopeIdInfo.ScopeUuid")
	})

	s.T().Run("session id, everything populated", func(t *testing.T) {
		info := GetSessionIDInfo(s.sessionID)
		require.NotNil(t, info, "GetSessionIDInfo result")
		assert.Equal(t, s.sessionID, info.SessionId, "SessionId")
		assert.Equal(t, SessionKeyPrefix, info.SessionIdPrefix, "SessionIdPrefix")
		assert.Equal(t, s.scopeUUID[:], info.SessionIdScopeUuid, "SessionIdScopeUuid")
		assert.Equal(t, s.sessionUUID[:], info.SessionIdSessionUuid, "SessionIdSessionUuid")
		assert.Equal(t, s.sessionIDStr, info.SessionAddr, "SessionAddr")
		assert.Equal(t, s.sessionUUIDStr, info.SessionUuid, "SessionUuid")
		require.NotNil(t, info.ScopeIdInfo, "info.ScopeIdInfo")
		assert.Equal(t, s.scopeID, info.ScopeIdInfo.ScopeId, "ScopeIdInfo.ScopeId")
		assert.Equal(t, ScopeKeyPrefix, info.ScopeIdInfo.ScopeIdPrefix, "ScopeIdInfo.ScopeIdPrefix")
		assert.Equal(t, s.scopeUUID[:], info.ScopeIdInfo.ScopeIdScopeUuid, "ScopeIdInfo.ScopeIdScopeUuid")
		assert.Equal(t, s.scopeIDStr, info.ScopeIdInfo.ScopeAddr, "ScopeIdInfo.ScopeAddr")
		assert.Equal(t, s.scopeUUIDStr, info.ScopeIdInfo.ScopeUuid, "ScopeIdInfo.ScopeUuid")
	})
}

func (s *MetadataTypesTestSuite) TestGetRecordIDInfo() {
	s.T().Run("empty address, empty object", func(t *testing.T) {
		info := GetRecordIDInfo(MetadataAddress{})
		require.NotNil(t, info, "GetRecordIDInfo result")
		assert.True(t, info.RecordId.Empty(), "empty RecordId")
		assert.Equal(t, 0, len(info.RecordIdPrefix), "RecordIdPrefix length")
		assert.Equal(t, 0, len(info.RecordIdScopeUuid), "RecordIdScopeUuid length")
		assert.Equal(t, 0, len(info.RecordIdHashedName), "RecordIdHashedName length")
		assert.Equal(t, "", info.RecordAddr, "RecordAddr")
		require.NotNil(t, info.ScopeIdInfo, "info.ScopeIdInfo")
		assert.True(t, info.ScopeIdInfo.ScopeId.Empty(), "empty ScopeIdInfo.ScopeId")
		assert.Equal(t, 0, len(info.ScopeIdInfo.ScopeIdPrefix), "ScopeIdInfo.ScopeIdPrefix length")
		assert.Equal(t, 0, len(info.ScopeIdInfo.ScopeIdScopeUuid), "ScopeIdInfo.ScopeIdScopeUuid length")
		assert.Equal(t, "", info.ScopeIdInfo.ScopeAddr, "ScopeIdInfo.ScopeAddr")
		assert.Equal(t, "", info.ScopeIdInfo.ScopeUuid, "ScopeIdInfo.ScopeUuid")
	})

	s.T().Run("record id, everything populated", func(t *testing.T) {
		info := GetRecordIDInfo(s.recordID)
		require.NotNil(t, info, "GetRecordIDInfo result")
		assert.Equal(t, s.recordID, info.RecordId, "RecordId")
		assert.Equal(t, RecordKeyPrefix, info.RecordIdPrefix, "RecordIdPrefix")
		assert.Equal(t, s.scopeUUID[:], info.RecordIdScopeUuid, "RecordIdScopeUuid")
		assert.Equal(t, s.recordNameHash, info.RecordIdHashedName, "RecordIdHashedName")
		assert.Equal(t, s.recordIDStr, info.RecordAddr, "RecordAddr")
		require.NotNil(t, info.ScopeIdInfo, "info.ScopeIdInfo")
		assert.Equal(t, s.scopeID, info.ScopeIdInfo.ScopeId, "ScopeIdInfo.ScopeId")
		assert.Equal(t, ScopeKeyPrefix, info.ScopeIdInfo.ScopeIdPrefix, "ScopeIdInfo.ScopeIdPrefix")
		assert.Equal(t, s.scopeUUID[:], info.ScopeIdInfo.ScopeIdScopeUuid, "ScopeIdInfo.ScopeIdScopeUuid")
		assert.Equal(t, s.scopeIDStr, info.ScopeIdInfo.ScopeAddr, "ScopeIdInfo.ScopeAddr")
		assert.Equal(t, s.scopeUUIDStr, info.ScopeIdInfo.ScopeUuid, "ScopeIdInfo.ScopeUuid")
	})
}

func (s *MetadataTypesTestSuite) TestGetScopeSpecIDInfo() {
	s.T().Run("empty address, empty object", func(t *testing.T) {
		info := GetScopeSpecIDInfo(MetadataAddress{})
		require.NotNil(t, info, "GetScopeSpecIDInfo result")
		assert.True(t, info.ScopeSpecId.Empty(), "empty ScopeSpecId")
		assert.Equal(t, 0, len(info.ScopeSpecIdPrefix), "ScopeSpecIdPrefix length")
		assert.Equal(t, 0, len(info.ScopeSpecIdScopeSpecUuid), "ScopeSpecIdScopeSpecUuid length")
		assert.Equal(t, "", info.ScopeSpecAddr, "ScopeSpecAddr")
		assert.Equal(t, "", info.ScopeSpecUuid, "ScopeSpecUuid")
	})

	s.T().Run("scope spec id, everything populated", func(t *testing.T) {
		info := GetScopeSpecIDInfo(s.scopeSpecID)
		require.NotNil(t, info, "GetScopeSpecIDInfo result")
		assert.Equal(t, s.scopeSpecID, info.ScopeSpecId, "ScopeSpecId")
		assert.Equal(t, ScopeSpecificationKeyPrefix, info.ScopeSpecIdPrefix, "ScopeSpecIdPrefix")
		assert.Equal(t, s.scopeSpecUUID[:], info.ScopeSpecIdScopeSpecUuid, "ScopeSpecIdScopeSpecUuid")
		assert.Equal(t, s.scopeSpecIDStr, info.ScopeSpecAddr, "ScopeSpecAddr")
		assert.Equal(t, s.scopeSpecUUIDStr, info.ScopeSpecUuid, "ScopeSpecUuid")
	})
}

func (s *MetadataTypesTestSuite) TestGetContractSpecIDInfo() {
	s.T().Run("empty address, empty object", func(t *testing.T) {
		info := GetContractSpecIDInfo(MetadataAddress{})
		require.NotNil(t, info, "GetContractSpecIDInfo result")
		assert.True(t, info.ContractSpecId.Empty(), "empty ContractSpecId")
		assert.Equal(t, 0, len(info.ContractSpecIdPrefix), "ContractSpecIdPrefix length")
		assert.Equal(t, 0, len(info.ContractSpecIdContractSpecUuid), "ContractSpecIdContractSpecUuid length")
		assert.Equal(t, "", info.ContractSpecAddr, "ContractSpecAddr")
		assert.Equal(t, "", info.ContractSpecUuid, "ContractSpecUuid")
	})

	s.T().Run("contract spec id, everything populated", func(t *testing.T) {
		info := GetContractSpecIDInfo(s.contractSpecID)
		require.NotNil(t, info, "GetContractSpecIDInfo result")
		assert.Equal(t, s.contractSpecID, info.ContractSpecId, "ContractSpecId")
		assert.Equal(t, ContractSpecificationKeyPrefix, info.ContractSpecIdPrefix, "ContractSpecIdPrefix")
		assert.Equal(t, s.contractSpecUUID[:], info.ContractSpecIdContractSpecUuid, "ContractSpecIdContractSpecUuid")
		assert.Equal(t, s.contractSpecIDStr, info.ContractSpecAddr, "ContractSpecAddr")
		assert.Equal(t, s.contractSpecUUIDStr, info.ContractSpecUuid, "ContractSpecUuid")
	})
}

func (s *MetadataTypesTestSuite) GetRecordSpecIDInfo() {
	s.T().Run("empty address, empty object", func(t *testing.T) {
		info := GetRecordSpecIDInfo(MetadataAddress{})
		require.NotNil(t, info, "GetRecordSpecIDInfo result")
		assert.True(t, info.RecordSpecId.Empty(), "empty RecordSpecId")
		assert.Equal(t, 0, len(info.RecordSpecIdPrefix), "RecordSpecIdPrefix length")
		assert.Equal(t, 0, len(info.RecordSpecIdContractSpecUuid), "RecordSpecIdContractSpecUuid length")
		assert.Equal(t, 0, len(info.RecordSpecIdHashedName), "RecordSpecIdHashedName length")
		assert.Equal(t, "", info.RecordSpecAddr, "RecordSpecAddr")
		require.NotNil(t, info.ContractSpecIdInfo, "info.ContractSpecIdInfo")
		assert.True(t, info.ContractSpecIdInfo.ContractSpecId.Empty(), "empty ContractSpecIdInfo.ContractSpecId")
		assert.Equal(t, 0, len(info.ContractSpecIdInfo.ContractSpecIdPrefix), "ContractSpecIdInfo.ContractSpecIdPrefix length")
		assert.Equal(t, 0, len(info.ContractSpecIdInfo.ContractSpecIdContractSpecUuid), "ContractSpecIdInfo.ContractSpecIdContractSpecUuid length")
		assert.Equal(t, "", info.ContractSpecIdInfo.ContractSpecAddr, "ContractSpecIdInfo.ContractSpecAddr")
		assert.Equal(t, "", info.ContractSpecIdInfo.ContractSpecUuid, "ContractSpecIdInfo.ContractSpecUuid")
	})

	s.T().Run("record spec id, everything populated", func(t *testing.T) {
		info := GetRecordSpecIDInfo(s.recordSpecID)
		require.NotNil(t, info, "GetRecordSpecIDInfo result")
		assert.Equal(t, s.recordSpecID, info.RecordSpecId, "RecordSpecId")
		assert.Equal(t, RecordSpecificationKeyPrefix, info.RecordSpecIdPrefix, "RecordSpecIdPrefix")
		assert.Equal(t, s.contractSpecUUID[:], info.RecordSpecIdContractSpecUuid, "RecordSpecIdContractSpecUuid")
		assert.Equal(t, s.recordNameHash, info.RecordSpecIdHashedName, "RecordSpecIdHashedName")
		assert.Equal(t, s.recordSpecIDStr, info.RecordSpecAddr, "RecordSpecAddr")
		require.NotNil(t, info.ContractSpecIdInfo, "info.ContractSpecIdInfo")
		assert.Equal(t, s.contractSpecID, info.ContractSpecIdInfo.ContractSpecId, "ContractSpecIdInfo.ContractSpecId")
		assert.Equal(t, ContractSpecificationKeyPrefix, info.ContractSpecIdInfo.ContractSpecIdPrefix, "ContractSpecIdInfo.ContractSpecIdPrefix")
		assert.Equal(t, s.contractSpecUUID[:], info.ContractSpecIdInfo.ContractSpecIdContractSpecUuid, "ContractSpecIdInfo.ContractSpecIdContractSpecUuid")
		assert.Equal(t, s.contractSpecIDStr, info.ContractSpecIdInfo.ContractSpecAddr, "ContractSpecIdInfo.ContractSpecAddr")
		assert.Equal(t, s.contractSpecUUIDStr, info.ContractSpecIdInfo.ContractSpecUuid, "ContractSpecIdInfo.ContractSpecUuid")
	})
}
