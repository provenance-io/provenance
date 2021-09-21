package provenance

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type MetadataAddressTestSuite struct {
	suite.Suite

	// Pre-selected UUID strings that go with ID strings generated from the Go code.
	scopeUUIDStr          string
	sessionUUIDStr        string
	scopeSpecUUIDStr      string
	contractSpecUUIDStr   string
	recordName            string
	recordNameHashedBytes []byte

	// Pre-generated ID strings created using Go code and providing the above strings.
	scopeIDStr        string
	sessionIDStr      string
	recordIDStr       string
	scopeSpecIDStr    string
	contractSpecIDStr string
	recordSpecIDStr   string

	// UUID versions of the UUID strings.
	scopeUUID        uuid.UUID
	sessionUUID      uuid.UUID
	scopeSpecUUID    uuid.UUID
	contractSpecUUID uuid.UUID
}

func (s *MetadataAddressTestSuite) SetupTest() {
	// These strings come from the output of x/metadata/types/address_test.go TestGenerateExamples().

	s.scopeUUIDStr = "91978ba2-5f35-459a-86a7-feca1b0512e0"
	s.sessionUUIDStr = "5803f8bc-6067-4eb5-951f-2121671c2ec0"
	s.scopeSpecUUIDStr = "dc83ea70-eacd-40fe-9adf-1cf6148bf8a2"
	s.contractSpecUUIDStr = "def6bc0a-c9dd-4874-948f-5206e6060a84"
	s.recordName = "recordname"
	s.recordNameHashedBytes = []byte{234, 169, 160, 84, 154, 205, 183, 162, 227, 133, 142, 181, 183, 185, 209, 190}

	s.scopeIDStr = "scope1qzge0zaztu65tx5x5llv5xc9ztsqxlkwel"
	s.sessionIDStr = "session1qxge0zaztu65tx5x5llv5xc9zts9sqlch3sxwn44j50jzgt8rshvqyfrjcr"
	s.recordIDStr = "record1q2ge0zaztu65tx5x5llv5xc9ztsw42dq2jdvmdazuwzcaddhh8gmu3mcze3"
	s.scopeSpecIDStr = "scopespec1qnwg86nsatx5pl56muw0v9ytlz3qu3jx6m"
	s.contractSpecIDStr = "contractspec1q000d0q2e8w5say53afqdesxp2zqzkr4fn"
	s.recordSpecIDStr = "recspec1qh00d0q2e8w5say53afqdesxp2zw42dq2jdvmdazuwzcaddhh8gmuqhez44"

	s.scopeUUID = uuid.MustParse(s.scopeUUIDStr)
	s.sessionUUID = uuid.MustParse(s.sessionUUIDStr)
	s.scopeSpecUUID = uuid.MustParse(s.scopeSpecUUIDStr)
	s.contractSpecUUID = uuid.MustParse(s.contractSpecUUIDStr)
}

func TestMetadataAddressTestSuite(t *testing.T) {
	suite.Run(t, new(MetadataAddressTestSuite))
}

func mustGetMetadataAddressFromBech32(str string) MetadataAddress {
	retval, err := MetadataAddressFromBech32(str)
	if err != nil {
		panic(err)
	}
	return retval
}

func mustGetMetadataAddressFromBytes(bz []byte) MetadataAddress {
	retval, err := MetadataAddressFromBytes(bz)
	if err != nil {
		panic(err)
	}
	return retval
}

func (s MetadataAddressTestSuite) TestScopeID() {
	expectedAddr := mustGetMetadataAddressFromBech32(s.scopeIDStr)
	expectedID := s.scopeIDStr
	expectedKey := KeyScope
	expectedPrefix := PrefixScope
	expectedPrimaryUUID := s.scopeUUID
	expectedSecondaryBytes := []byte{}

	actualAddr := MetadataAddressForScope(s.scopeUUID)
	actualId := actualAddr.String()
	actualKey := actualAddr.GetKey()
	actualPrefix := actualAddr.GetPrefix()
	actualPrimaryUuid := actualAddr.GetPrimaryUUID()
	actualSecondaryBytes := actualAddr.GetSecondaryBytes()

	addrFromBytes := mustGetMetadataAddressFromBytes(actualAddr.Bytes())

	s.Assert().Equal(expectedKey, actualKey, "key")
	s.Assert().Equal(expectedPrefix, actualPrefix, "prefix")
	s.Assert().Equal(expectedPrimaryUUID, actualPrimaryUuid, "primary uuid")
	s.Assert().Equal(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
	s.Assert().Equal(expectedID, actualId, "as bech32 strings")
	s.Assert().Equal(expectedAddr, actualAddr, "whole metadata address")
	s.Assert().Equal(expectedAddr, addrFromBytes, "address from bytes")
	s.Assert().True(expectedAddr.Equals(actualAddr), "%s.Equals(%s)", expectedAddr, actualAddr)
}

func (s MetadataAddressTestSuite) TestSessionID() {
	expectedAddr := mustGetMetadataAddressFromBech32(s.sessionIDStr)
	expectedID := s.sessionIDStr
	expectedKey := KeySession
	expectedPrefix := PrefixSession
	expectedPrimaryUUID := s.scopeUUID
	expectedSecondaryBytes, _ := s.sessionUUID.MarshalBinary()

	actualAddr := MetadataAddressForSession(s.scopeUUID, s.sessionUUID)
	actualId := actualAddr.String()
	actualKey := actualAddr.GetKey()
	actualPrefix := actualAddr.GetPrefix()
	actualPrimaryUuid := actualAddr.GetPrimaryUUID()
	actualSecondaryBytes := actualAddr.GetSecondaryBytes()

	addrFromBytes := mustGetMetadataAddressFromBytes(actualAddr.Bytes())

	s.Assert().Equal(expectedKey, actualKey, "key")
	s.Assert().Equal(expectedPrefix, actualPrefix, "prefix")
	s.Assert().Equal(expectedPrimaryUUID, actualPrimaryUuid, "primary uuid")
	s.Assert().Equal(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
	s.Assert().Equal(expectedID, actualId, "as bech32 strings")
	s.Assert().Equal(expectedAddr, actualAddr, "whole metadata address")
	s.Assert().Equal(expectedAddr, addrFromBytes, "address from bytes")
	s.Assert().True(expectedAddr.Equals(actualAddr), "%s.Equals(%s)", expectedAddr, actualAddr)
}

func (s MetadataAddressTestSuite) TestRecordID() {
	expectedAddr := mustGetMetadataAddressFromBech32(s.recordIDStr)
	expectedID := s.recordIDStr
	expectedKey := KeyRecord
	expectedPrefix := PrefixRecord
	expectedPrimaryUUID := s.scopeUUID
	expectedSecondaryBytes := s.recordNameHashedBytes

	actualAddr := MetadataAddressForRecord(s.scopeUUID, s.recordName)
	actualId := actualAddr.String()
	actualKey := actualAddr.GetKey()
	actualPrefix := actualAddr.GetPrefix()
	actualPrimaryUuid := actualAddr.GetPrimaryUUID()
	actualSecondaryBytes := actualAddr.GetSecondaryBytes()

	addrFromBytes := mustGetMetadataAddressFromBytes(actualAddr.Bytes())

	s.Assert().Equal(expectedKey, actualKey, "key")
	s.Assert().Equal(expectedPrefix, actualPrefix, "prefix")
	s.Assert().Equal(expectedPrimaryUUID, actualPrimaryUuid, "primary uuid")
	s.Assert().Equal(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
	s.Assert().Equal(expectedID, actualId, "as bech32 strings")
	s.Assert().Equal(expectedAddr, actualAddr, "whole metadata address")
	s.Assert().Equal(expectedAddr, addrFromBytes, "address from bytes")
	s.Assert().True(expectedAddr.Equals(actualAddr), "%s.Equals(%s)", expectedAddr, actualAddr)
}

func (s MetadataAddressTestSuite) TestScopeSpecID() {
	expectedAddr := mustGetMetadataAddressFromBech32(s.scopeSpecIDStr)
	expectedID := s.scopeSpecIDStr
	expectedKey := KeyScopeSpecification
	expectedPrefix := PrefixScopeSpecification
	expectedPrimaryUUID := s.scopeSpecUUID
	expectedSecondaryBytes := []byte{}

	actualAddr := MetadataAddressForScopeSpecification(s.scopeSpecUUID)
	actualId := actualAddr.String()
	actualKey := actualAddr.GetKey()
	actualPrefix := actualAddr.GetPrefix()
	actualPrimaryUuid := actualAddr.GetPrimaryUUID()
	actualSecondaryBytes := actualAddr.GetSecondaryBytes()

	addrFromBytes := mustGetMetadataAddressFromBytes(actualAddr.Bytes())

	s.Assert().Equal(expectedKey, actualKey, "key")
	s.Assert().Equal(expectedPrefix, actualPrefix, "prefix")
	s.Assert().Equal(expectedPrimaryUUID, actualPrimaryUuid, "primary uuid")
	s.Assert().Equal(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
	s.Assert().Equal(expectedID, actualId, "as bech32 strings")
	s.Assert().Equal(expectedAddr, actualAddr, "whole metadata address")
	s.Assert().Equal(expectedAddr, addrFromBytes, "address from bytes")
	s.Assert().True(expectedAddr.Equals(actualAddr), "%s.Equals(%s)", expectedAddr, actualAddr)
}

func (s MetadataAddressTestSuite) TestContractSpecID() {
	expectedAddr := mustGetMetadataAddressFromBech32(s.contractSpecIDStr)
	expectedID := s.contractSpecIDStr
	expectedKey := KeyContractSpecification
	expectedPrefix := PrefixContractSpecification
	expectedPrimaryUUID := s.contractSpecUUID
	expectedSecondaryBytes := []byte{}

	actualAddr := MetadataAddressForContractSpecification(s.contractSpecUUID)
	actualId := actualAddr.String()
	actualKey := actualAddr.GetKey()
	actualPrefix := actualAddr.GetPrefix()
	actualPrimaryUuid := actualAddr.GetPrimaryUUID()
	actualSecondaryBytes := actualAddr.GetSecondaryBytes()

	addrFromBytes := mustGetMetadataAddressFromBytes(actualAddr.Bytes())

	s.Assert().Equal(expectedKey, actualKey, "key")
	s.Assert().Equal(expectedPrefix, actualPrefix, "prefix")
	s.Assert().Equal(expectedPrimaryUUID, actualPrimaryUuid, "primary uuid")
	s.Assert().Equal(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
	s.Assert().Equal(expectedID, actualId, "as bech32 strings")
	s.Assert().Equal(expectedAddr, actualAddr, "whole metadata address")
	s.Assert().Equal(expectedAddr, addrFromBytes, "address from bytes")
	s.Assert().True(expectedAddr.Equals(actualAddr), "%s.Equals(%s)", expectedAddr, actualAddr)
}

func (s MetadataAddressTestSuite) TestRecordSpecID() {
	expectedAddr := mustGetMetadataAddressFromBech32(s.recordSpecIDStr)
	expectedID := s.recordSpecIDStr
	expectedKey := KeyRecordSpecification
	expectedPrefix := PrefixRecordSpecification
	expectedPrimaryUUID := s.contractSpecUUID
	expectedSecondaryBytes := s.recordNameHashedBytes

	actualAddr := MetadataAddressForRecordSpecification(s.contractSpecUUID, s.recordName)
	actualId := actualAddr.String()
	actualKey := actualAddr.GetKey()
	actualPrefix := actualAddr.GetPrefix()
	actualPrimaryUuid := actualAddr.GetPrimaryUUID()
	actualSecondaryBytes := actualAddr.GetSecondaryBytes()

	addrFromBytes := mustGetMetadataAddressFromBytes(actualAddr.Bytes())

	s.Assert().Equal(expectedKey, actualKey, "key")
	s.Assert().Equal(expectedPrefix, actualPrefix, "prefix")
	s.Assert().Equal(expectedPrimaryUUID, actualPrimaryUuid, "primary uuid")
	s.Assert().Equal(expectedSecondaryBytes, actualSecondaryBytes, "secondary bytes")
	s.Assert().Equal(expectedID, actualId, "as bech32 strings")
	s.Assert().Equal(expectedAddr, actualAddr, "whole metadata address")
	s.Assert().Equal(expectedAddr, addrFromBytes, "address from bytes")
	s.Assert().True(expectedAddr.Equals(actualAddr), "%s.Equals(%s)", expectedAddr, actualAddr)
}
