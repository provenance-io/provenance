// Manually written to match protoc-gen-gogo output for provenance/registry/v1/authorization.proto.
// Should be regenerated with protoc when the proto builder is available.

package types

import (
	"fmt"
	io "io"
	math_bits "math/bits"

	proto "github.com/cosmos/gogoproto/proto"
)

// SignatureType defines how signature requirements are evaluated.
type SignatureType int32

const (
	SignatureType_SIGNATURE_TYPE_UNSPECIFIED     SignatureType = 0
	SignatureType_SIGNATURE_TYPE_REQUIRED_ALL    SignatureType = 1
	SignatureType_SIGNATURE_TYPE_REQUIRED_ALL_IF_SET SignatureType = 2
	SignatureType_SIGNATURE_TYPE_REQUIRED_ANY    SignatureType = 3
	SignatureType_SIGNATURE_TYPE_REQUIRED_ANY_IF_SET SignatureType = 4
)

var SignatureType_name = map[int32]string{
	0: "SIGNATURE_TYPE_UNSPECIFIED",
	1: "SIGNATURE_TYPE_REQUIRED_ALL",
	2: "SIGNATURE_TYPE_REQUIRED_ALL_IF_SET",
	3: "SIGNATURE_TYPE_REQUIRED_ANY",
	4: "SIGNATURE_TYPE_REQUIRED_ANY_IF_SET",
}

var SignatureType_value = map[string]int32{
	"SIGNATURE_TYPE_UNSPECIFIED":         0,
	"SIGNATURE_TYPE_REQUIRED_ALL":        1,
	"SIGNATURE_TYPE_REQUIRED_ALL_IF_SET": 2,
	"SIGNATURE_TYPE_REQUIRED_ANY":        3,
	"SIGNATURE_TYPE_REQUIRED_ANY_IF_SET": 4,
}

func (x SignatureType) String() string {
	return proto.EnumName(SignatureType_name, int32(x))
}

// NftRole identifies roles managed outside the registry module.
type NftRole int32

const (
	NftRole_NFT_ROLE_UNSPECIFIED NftRole = 0
	NftRole_NFT_ROLE_NFT_OWNER   NftRole = 1
)

var NftRole_name = map[int32]string{
	0: "NFT_ROLE_UNSPECIFIED",
	1: "NFT_ROLE_NFT_OWNER",
}

var NftRole_value = map[string]int32{
	"NFT_ROLE_UNSPECIFIED": 0,
	"NFT_ROLE_NFT_OWNER":   1,
}

func (x NftRole) String() string {
	return proto.EnumName(NftRole_name, int32(x))
}

// Assignment describes which address(es) to resolve for a role.
type Assignment int32

const (
	Assignment_ASSIGNMENT_UNSPECIFIED Assignment = 0
	Assignment_ASSIGNMENT_CURRENT     Assignment = 1
	Assignment_ASSIGNMENT_CURRENT_ALL Assignment = 2
	Assignment_ASSIGNMENT_CURRENT_ANY Assignment = 3
	Assignment_ASSIGNMENT_NEW         Assignment = 4
	Assignment_ASSIGNMENT_NEW_ANY     Assignment = 5
	Assignment_ASSIGNMENT_NEW_ALL     Assignment = 6
)

var Assignment_name = map[int32]string{
	0: "ASSIGNMENT_UNSPECIFIED",
	1: "ASSIGNMENT_CURRENT",
	2: "ASSIGNMENT_CURRENT_ALL",
	3: "ASSIGNMENT_CURRENT_ANY",
	4: "ASSIGNMENT_NEW",
	5: "ASSIGNMENT_NEW_ANY",
	6: "ASSIGNMENT_NEW_ALL",
}

var Assignment_value = map[string]int32{
	"ASSIGNMENT_UNSPECIFIED": 0,
	"ASSIGNMENT_CURRENT":     1,
	"ASSIGNMENT_CURRENT_ALL": 2,
	"ASSIGNMENT_CURRENT_ANY": 3,
	"ASSIGNMENT_NEW":         4,
	"ASSIGNMENT_NEW_ANY":     5,
	"ASSIGNMENT_NEW_ALL":     6,
}

func (x Assignment) String() string {
	return proto.EnumName(Assignment_name, int32(x))
}

func init() {
	proto.RegisterEnum("provenance.registry.v1.SignatureType", SignatureType_name, SignatureType_value)
	proto.RegisterEnum("provenance.registry.v1.NftRole", NftRole_name, NftRole_value)
	proto.RegisterEnum("provenance.registry.v1.Assignment", Assignment_name, Assignment_value)
	proto.RegisterType((*RoleAuthorization)(nil), "provenance.registry.v1.RoleAuthorization")
	proto.RegisterType((*Authorization)(nil), "provenance.registry.v1.Authorization")
	proto.RegisterType((*SignatureRequirement)(nil), "provenance.registry.v1.SignatureRequirement")
	proto.RegisterType((*RoleAssignment)(nil), "provenance.registry.v1.RoleAssignment")
	proto.RegisterType((*RolePriority)(nil), "provenance.registry.v1.RolePriority")
	proto.RegisterType((*RolePriorityEntry)(nil), "provenance.registry.v1.RolePriorityEntry")
}

// -----------------------------------------------------------------------
// RoleAuthorization
// -----------------------------------------------------------------------

// RoleAuthorization configures who must sign to update a specific role.
type RoleAuthorization struct {
	// role is the registry role being controlled.
	Role RegistryRole `protobuf:"varint,1,opt,name=role,proto3,enum=provenance.registry.v1.RegistryRole" json:"role,omitempty"`
	// authorizations is a list of alternative approval paths.
	Authorizations []Authorization `protobuf:"bytes,2,rep,name=authorizations,proto3" json:"authorizations"`
}

func (m *RoleAuthorization) Reset()         { *m = RoleAuthorization{} }
func (m *RoleAuthorization) String() string { return proto.CompactTextString(m) }
func (*RoleAuthorization) ProtoMessage()    {}
func (m *RoleAuthorization) XXX_Unmarshal(b []byte) error { return m.Unmarshal(b) }
func (m *RoleAuthorization) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleAuthorization.Marshal(b, m, deterministic)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *RoleAuthorization) XXX_Merge(src proto.Message) { xxx_messageInfo_RoleAuthorization.Merge(m, src) }
func (m *RoleAuthorization) XXX_Size() int               { return m.Size() }
func (m *RoleAuthorization) XXX_DiscardUnknown()         { xxx_messageInfo_RoleAuthorization.DiscardUnknown(m) }

var xxx_messageInfo_RoleAuthorization proto.InternalMessageInfo

func (m *RoleAuthorization) GetRole() RegistryRole {
	if m != nil {
		return m.Role
	}
	return RegistryRole_REGISTRY_ROLE_UNSPECIFIED
}

func (m *RoleAuthorization) GetAuthorizations() []Authorization {
	if m != nil {
		return m.Authorizations
	}
	return nil
}

func (m *RoleAuthorization) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleAuthorization) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleAuthorization) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Authorizations) > 0 {
		for iNdEx := len(m.Authorizations) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Authorizations[iNdEx].MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintAuthorization(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Role != 0 {
		i = encodeVarintAuthorization(dAtA, i, uint64(m.Role))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *RoleAuthorization) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Role != 0 {
		n += 1 + sovAuthorization(uint64(m.Role))
	}
	for _, e := range m.Authorizations {
		l = e.Size()
		n += 1 + l + sovAuthorization(uint64(l))
	}
	return n
}

func (m *RoleAuthorization) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthorization
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RoleAuthorization: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RoleAuthorization: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Role", wireType)
			}
			m.Role = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Role |= RegistryRole(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authorizations", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuthorization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuthorization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authorizations = append(m.Authorizations, Authorization{})
			if err := m.Authorizations[len(m.Authorizations)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAuthorization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthorization
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// -----------------------------------------------------------------------
// Authorization
// -----------------------------------------------------------------------

// Authorization defines one complete approval path for a role update.
type Authorization struct {
	Description string               `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	Signatures  []SignatureRequirement `protobuf:"bytes,2,rep,name=signatures,proto3" json:"signatures"`
}

func (m *Authorization) Reset()         { *m = Authorization{} }
func (m *Authorization) String() string { return proto.CompactTextString(m) }
func (*Authorization) ProtoMessage()    {}
func (m *Authorization) XXX_Unmarshal(b []byte) error { return m.Unmarshal(b) }
func (m *Authorization) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Authorization.Marshal(b, m, deterministic)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *Authorization) XXX_Merge(src proto.Message) { xxx_messageInfo_Authorization.Merge(m, src) }
func (m *Authorization) XXX_Size() int               { return m.Size() }
func (m *Authorization) XXX_DiscardUnknown()         { xxx_messageInfo_Authorization.DiscardUnknown(m) }

var xxx_messageInfo_Authorization proto.InternalMessageInfo

func (m *Authorization) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Authorization) GetSignatures() []SignatureRequirement {
	if m != nil {
		return m.Signatures
	}
	return nil
}

func (m *Authorization) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Authorization) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Authorization) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Signatures) > 0 {
		for iNdEx := len(m.Signatures) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Signatures[iNdEx].MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintAuthorization(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Description) > 0 {
		i -= len(m.Description)
		copy(dAtA[i:], m.Description)
		i = encodeVarintAuthorization(dAtA, i, uint64(len(m.Description)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Authorization) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Description)
	if l > 0 {
		n += 1 + l + sovAuthorization(uint64(l))
	}
	for _, e := range m.Signatures {
		l = e.Size()
		n += 1 + l + sovAuthorization(uint64(l))
	}
	return n
}

func (m *Authorization) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthorization
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Authorization: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Authorization: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Description", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAuthorization
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAuthorization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Description = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signatures", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuthorization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuthorization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signatures = append(m.Signatures, SignatureRequirement{})
			if err := m.Signatures[len(m.Signatures)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAuthorization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthorization
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// -----------------------------------------------------------------------
// SignatureRequirement
// -----------------------------------------------------------------------

// SignatureRequirement defines a single signature check within an authorization path.
type SignatureRequirement struct {
	Type  SignatureType  `protobuf:"varint,1,opt,name=type,proto3,enum=provenance.registry.v1.SignatureType" json:"type,omitempty"`
	Roles []RoleAssignment `protobuf:"bytes,2,rep,name=roles,proto3" json:"roles"`
}

func (m *SignatureRequirement) Reset()         { *m = SignatureRequirement{} }
func (m *SignatureRequirement) String() string { return proto.CompactTextString(m) }
func (*SignatureRequirement) ProtoMessage()    {}
func (m *SignatureRequirement) XXX_Unmarshal(b []byte) error { return m.Unmarshal(b) }
func (m *SignatureRequirement) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SignatureRequirement.Marshal(b, m, deterministic)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *SignatureRequirement) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignatureRequirement.Merge(m, src)
}
func (m *SignatureRequirement) XXX_Size() int { return m.Size() }
func (m *SignatureRequirement) XXX_DiscardUnknown() {
	xxx_messageInfo_SignatureRequirement.DiscardUnknown(m)
}

var xxx_messageInfo_SignatureRequirement proto.InternalMessageInfo

func (m *SignatureRequirement) GetType() SignatureType {
	if m != nil {
		return m.Type
	}
	return SignatureType_SIGNATURE_TYPE_UNSPECIFIED
}

func (m *SignatureRequirement) GetRoles() []RoleAssignment {
	if m != nil {
		return m.Roles
	}
	return nil
}

func (m *SignatureRequirement) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SignatureRequirement) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SignatureRequirement) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Roles) > 0 {
		for iNdEx := len(m.Roles) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Roles[iNdEx].MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintAuthorization(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Type != 0 {
		i = encodeVarintAuthorization(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *SignatureRequirement) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovAuthorization(uint64(m.Type))
	}
	for _, e := range m.Roles {
		l = e.Size()
		n += 1 + l + sovAuthorization(uint64(l))
	}
	return n
}

func (m *SignatureRequirement) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthorization
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SignatureRequirement: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignatureRequirement: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= SignatureType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Roles", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuthorization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuthorization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Roles = append(m.Roles, RoleAssignment{})
			if err := m.Roles[len(m.Roles)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAuthorization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthorization
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// -----------------------------------------------------------------------
// RoleAssignment
// -----------------------------------------------------------------------

// RoleAssignment_RoleSelector is the oneof interface for RoleAssignment.
type isRoleAssignment_RoleSelector interface {
	isRoleAssignment_RoleSelector()
	MarshalTo([]byte) (int, error)
	Size() int
}

type RoleAssignment_RegistryRole struct {
	RegistryRole RegistryRole `protobuf:"varint,1,opt,name=registry_role,json=registryRole,proto3,enum=provenance.registry.v1.RegistryRole,oneof"`
}

type RoleAssignment_NftRole struct {
	NftRole NftRole `protobuf:"varint,2,opt,name=nft_role,json=nftRole,proto3,enum=provenance.registry.v1.NftRole,oneof"`
}

type RoleAssignment_RolePriority struct {
	RolePriority *RolePriority `protobuf:"bytes,3,opt,name=role_priority,json=rolePriority,proto3,oneof"`
}

func (*RoleAssignment_RegistryRole) isRoleAssignment_RoleSelector() {}
func (*RoleAssignment_NftRole) isRoleAssignment_RoleSelector()      {}
func (*RoleAssignment_RolePriority) isRoleAssignment_RoleSelector() {}

func (m *RoleAssignment_RegistryRole) MarshalTo(dAtA []byte) (int, error) {
	i := len(dAtA)
	i = encodeVarintAuthorization(dAtA, i, uint64(m.RegistryRole))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func (m *RoleAssignment_NftRole) MarshalTo(dAtA []byte) (int, error) {
	i := len(dAtA)
	i = encodeVarintAuthorization(dAtA, i, uint64(m.NftRole))
	i--
	dAtA[i] = 0x10
	return len(dAtA) - i, nil
}

func (m *RoleAssignment_RolePriority) MarshalTo(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.RolePriority != nil {
		size, err := m.RolePriority.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintAuthorization(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	return len(dAtA) - i, nil
}

func (m *RoleAssignment_RegistryRole) Size() int {
	return 1 + sovAuthorization(uint64(m.RegistryRole))
}

func (m *RoleAssignment_NftRole) Size() int {
	return 1 + sovAuthorization(uint64(m.NftRole))
}

func (m *RoleAssignment_RolePriority) Size() int {
	if m.RolePriority == nil {
		return 1 + sovAuthorization(0)
	}
	s := m.RolePriority.Size()
	return 1 + s + sovAuthorization(uint64(s))
}

// RoleAssignment specifies which role and assignment type to resolve for a signature check.
type RoleAssignment struct {
	// Types that are valid to be assigned to RoleSelector:
	//   *RoleAssignment_RegistryRole
	//   *RoleAssignment_NftRole
	//   *RoleAssignment_RolePriority
	RoleSelector isRoleAssignment_RoleSelector `protobuf_oneof:"role_selector"`
	// assignment specifies which address(es) to resolve for the role.
	Assignment Assignment `protobuf:"varint,4,opt,name=assignment,proto3,enum=provenance.registry.v1.Assignment" json:"assignment,omitempty"`
}

func (m *RoleAssignment) Reset()         { *m = RoleAssignment{} }
func (m *RoleAssignment) String() string { return proto.CompactTextString(m) }
func (*RoleAssignment) ProtoMessage()    {}
func (m *RoleAssignment) XXX_Unmarshal(b []byte) error { return m.Unmarshal(b) }
func (m *RoleAssignment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RoleAssignment.Marshal(b, m, deterministic)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *RoleAssignment) XXX_Merge(src proto.Message) { xxx_messageInfo_RoleAssignment.Merge(m, src) }
func (m *RoleAssignment) XXX_Size() int               { return m.Size() }
func (m *RoleAssignment) XXX_DiscardUnknown()         { xxx_messageInfo_RoleAssignment.DiscardUnknown(m) }

var xxx_messageInfo_RoleAssignment proto.InternalMessageInfo

func (m *RoleAssignment) GetRoleSelector() isRoleAssignment_RoleSelector {
	if m != nil {
		return m.RoleSelector
	}
	return nil
}

func (m *RoleAssignment) GetRegistryRole() RegistryRole {
	if x, ok := m.GetRoleSelector().(*RoleAssignment_RegistryRole); ok {
		return x.RegistryRole
	}
	return RegistryRole_REGISTRY_ROLE_UNSPECIFIED
}

func (m *RoleAssignment) GetNftRole() NftRole {
	if x, ok := m.GetRoleSelector().(*RoleAssignment_NftRole); ok {
		return x.NftRole
	}
	return NftRole_NFT_ROLE_UNSPECIFIED
}

func (m *RoleAssignment) GetRolePriority() *RolePriority {
	if x, ok := m.GetRoleSelector().(*RoleAssignment_RolePriority); ok {
		return x.RolePriority
	}
	return nil
}

func (m *RoleAssignment) GetAssignment() Assignment {
	if m != nil {
		return m.Assignment
	}
	return Assignment_ASSIGNMENT_UNSPECIFIED
}

func (m *RoleAssignment) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RoleAssignment) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RoleAssignment) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Assignment != 0 {
		i = encodeVarintAuthorization(dAtA, i, uint64(m.Assignment))
		i--
		dAtA[i] = 0x20
	}
	if m.RoleSelector != nil {
		n, err := m.RoleSelector.MarshalTo(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= n
	}
	return len(dAtA) - i, nil
}

func (m *RoleAssignment) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.RoleSelector != nil {
		n += m.RoleSelector.Size()
	}
	if m.Assignment != 0 {
		n += 1 + sovAuthorization(uint64(m.Assignment))
	}
	return n
}

func (m *RoleAssignment) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthorization
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RoleAssignment: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RoleAssignment: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RegistryRole", wireType)
			}
			var v RegistryRole
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= RegistryRole(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.RoleSelector = &RoleAssignment_RegistryRole{RegistryRole: v}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NftRole", wireType)
			}
			var v NftRole
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= NftRole(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.RoleSelector = &RoleAssignment_NftRole{NftRole: v}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RolePriority", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuthorization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuthorization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			v := &RolePriority{}
			if err := v.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			m.RoleSelector = &RoleAssignment_RolePriority{RolePriority: v}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Assignment", wireType)
			}
			m.Assignment = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Assignment |= Assignment(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipAuthorization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthorization
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// -----------------------------------------------------------------------
// RolePriority
// -----------------------------------------------------------------------

// RolePriority is an ordered list of role entries; the first existing role is used.
type RolePriority struct {
	Entries []RolePriorityEntry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries"`
}

func (m *RolePriority) Reset()         { *m = RolePriority{} }
func (m *RolePriority) String() string { return proto.CompactTextString(m) }
func (*RolePriority) ProtoMessage()    {}
func (m *RolePriority) XXX_Unmarshal(b []byte) error { return m.Unmarshal(b) }
func (m *RolePriority) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RolePriority.Marshal(b, m, deterministic)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *RolePriority) XXX_Merge(src proto.Message) { xxx_messageInfo_RolePriority.Merge(m, src) }
func (m *RolePriority) XXX_Size() int               { return m.Size() }
func (m *RolePriority) XXX_DiscardUnknown()         { xxx_messageInfo_RolePriority.DiscardUnknown(m) }

var xxx_messageInfo_RolePriority proto.InternalMessageInfo

func (m *RolePriority) GetEntries() []RolePriorityEntry {
	if m != nil {
		return m.Entries
	}
	return nil
}

func (m *RolePriority) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RolePriority) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RolePriority) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	for iNdEx := len(m.Entries) - 1; iNdEx >= 0; iNdEx-- {
		size, err := m.Entries[iNdEx].MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintAuthorization(dAtA, i, uint64(size))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *RolePriority) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	for _, e := range m.Entries {
		l = e.Size()
		n += 1 + l + sovAuthorization(uint64(l))
	}
	return n
}

func (m *RolePriority) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthorization
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RolePriority: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RolePriority: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Entries", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAuthorization
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAuthorization
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Entries = append(m.Entries, RolePriorityEntry{})
			if err := m.Entries[len(m.Entries)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAuthorization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthorization
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// -----------------------------------------------------------------------
// RolePriorityEntry
// -----------------------------------------------------------------------

type isRolePriorityEntry_Role interface {
	isRolePriorityEntry_Role()
	MarshalTo([]byte) (int, error)
	Size() int
}

type RolePriorityEntry_RegistryRole struct {
	RegistryRole RegistryRole `protobuf:"varint,1,opt,name=registry_role,json=registryRole,proto3,enum=provenance.registry.v1.RegistryRole,oneof"`
}

type RolePriorityEntry_NftRole struct {
	NftRole NftRole `protobuf:"varint,2,opt,name=nft_role,json=nftRole,proto3,enum=provenance.registry.v1.NftRole,oneof"`
}

func (*RolePriorityEntry_RegistryRole) isRolePriorityEntry_Role() {}
func (*RolePriorityEntry_NftRole) isRolePriorityEntry_Role()      {}

func (m *RolePriorityEntry_RegistryRole) MarshalTo(dAtA []byte) (int, error) {
	i := len(dAtA)
	i = encodeVarintAuthorization(dAtA, i, uint64(m.RegistryRole))
	i--
	dAtA[i] = 0x8
	return len(dAtA) - i, nil
}

func (m *RolePriorityEntry_NftRole) MarshalTo(dAtA []byte) (int, error) {
	i := len(dAtA)
	i = encodeVarintAuthorization(dAtA, i, uint64(m.NftRole))
	i--
	dAtA[i] = 0x10
	return len(dAtA) - i, nil
}

func (m *RolePriorityEntry_RegistryRole) Size() int {
	return 1 + sovAuthorization(uint64(m.RegistryRole))
}

func (m *RolePriorityEntry_NftRole) Size() int {
	return 1 + sovAuthorization(uint64(m.NftRole))
}

// RolePriorityEntry is a single role in a RolePriority list.
type RolePriorityEntry struct {
	// Types that are valid to be assigned to Role:
	//   *RolePriorityEntry_RegistryRole
	//   *RolePriorityEntry_NftRole
	Role isRolePriorityEntry_Role `protobuf_oneof:"role"`
}

func (m *RolePriorityEntry) Reset()         { *m = RolePriorityEntry{} }
func (m *RolePriorityEntry) String() string { return proto.CompactTextString(m) }
func (*RolePriorityEntry) ProtoMessage()    {}
func (m *RolePriorityEntry) XXX_Unmarshal(b []byte) error { return m.Unmarshal(b) }
func (m *RolePriorityEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RolePriorityEntry.Marshal(b, m, deterministic)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}
func (m *RolePriorityEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RolePriorityEntry.Merge(m, src)
}
func (m *RolePriorityEntry) XXX_Size() int { return m.Size() }
func (m *RolePriorityEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_RolePriorityEntry.DiscardUnknown(m)
}

var xxx_messageInfo_RolePriorityEntry proto.InternalMessageInfo

func (m *RolePriorityEntry) GetRole() isRolePriorityEntry_Role {
	if m != nil {
		return m.Role
	}
	return nil
}

func (m *RolePriorityEntry) GetRegistryRole() RegistryRole {
	if x, ok := m.GetRole().(*RolePriorityEntry_RegistryRole); ok {
		return x.RegistryRole
	}
	return RegistryRole_REGISTRY_ROLE_UNSPECIFIED
}

func (m *RolePriorityEntry) GetNftRole() NftRole {
	if x, ok := m.GetRole().(*RolePriorityEntry_NftRole); ok {
		return x.NftRole
	}
	return NftRole_NFT_ROLE_UNSPECIFIED
}

func (m *RolePriorityEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RolePriorityEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RolePriorityEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Role != nil {
		n, err := m.Role.MarshalTo(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= n
	}
	return len(dAtA) - i, nil
}

func (m *RolePriorityEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Role != nil {
		n += m.Role.Size()
	}
	return n
}

func (m *RolePriorityEntry) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAuthorization
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RolePriorityEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RolePriorityEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RegistryRole", wireType)
			}
			var v RegistryRole
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= RegistryRole(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Role = &RolePriorityEntry_RegistryRole{RegistryRole: v}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NftRole", wireType)
			}
			var v NftRole
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= NftRole(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Role = &RolePriorityEntry_NftRole{NftRole: v}
		default:
			iNdEx = preIndex
			skippy, err := skipAuthorization(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAuthorization
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func encodeVarintAuthorization(dAtA []byte, offset int, v uint64) int {
	offset -= sovAuthorization(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func sovAuthorization(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func skipAuthorization(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAuthorization
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAuthorization
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthAuthorization
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAuthorization
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAuthorization
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAuthorization        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAuthorization          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAuthorization = fmt.Errorf("proto: unexpected end of group")
)
