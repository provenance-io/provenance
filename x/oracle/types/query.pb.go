// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: provenance/oracle/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	github_com_CosmWasm_wasmd_x_wasm_types "github.com/CosmWasm/wasmd/x/wasm/types"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// QueryOracleAddressRequest queries for the address of the oracle.
type QueryOracleAddressRequest struct {
}

func (m *QueryOracleAddressRequest) Reset()         { *m = QueryOracleAddressRequest{} }
func (m *QueryOracleAddressRequest) String() string { return proto.CompactTextString(m) }
func (*QueryOracleAddressRequest) ProtoMessage()    {}
func (*QueryOracleAddressRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_169907f611744c57, []int{0}
}
func (m *QueryOracleAddressRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryOracleAddressRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryOracleAddressRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryOracleAddressRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryOracleAddressRequest.Merge(m, src)
}
func (m *QueryOracleAddressRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryOracleAddressRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryOracleAddressRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryOracleAddressRequest proto.InternalMessageInfo

// QueryOracleAddressResponse contains the address of the oracle.
type QueryOracleAddressResponse struct {
	// The address of the oracle
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *QueryOracleAddressResponse) Reset()         { *m = QueryOracleAddressResponse{} }
func (m *QueryOracleAddressResponse) String() string { return proto.CompactTextString(m) }
func (*QueryOracleAddressResponse) ProtoMessage()    {}
func (*QueryOracleAddressResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_169907f611744c57, []int{1}
}
func (m *QueryOracleAddressResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryOracleAddressResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryOracleAddressResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryOracleAddressResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryOracleAddressResponse.Merge(m, src)
}
func (m *QueryOracleAddressResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryOracleAddressResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryOracleAddressResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryOracleAddressResponse proto.InternalMessageInfo

func (m *QueryOracleAddressResponse) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// QueryOracleRequest queries the module's oracle.
type QueryOracleRequest struct {
	// Query contains the query data passed to the oracle.
	Query github_com_CosmWasm_wasmd_x_wasm_types.RawContractMessage `protobuf:"bytes,1,opt,name=query,proto3,casttype=github.com/CosmWasm/wasmd/x/wasm/types.RawContractMessage" json:"query,omitempty"`
}

func (m *QueryOracleRequest) Reset()         { *m = QueryOracleRequest{} }
func (m *QueryOracleRequest) String() string { return proto.CompactTextString(m) }
func (*QueryOracleRequest) ProtoMessage()    {}
func (*QueryOracleRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_169907f611744c57, []int{2}
}
func (m *QueryOracleRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryOracleRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryOracleRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryOracleRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryOracleRequest.Merge(m, src)
}
func (m *QueryOracleRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryOracleRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryOracleRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryOracleRequest proto.InternalMessageInfo

func (m *QueryOracleRequest) GetQuery() github_com_CosmWasm_wasmd_x_wasm_types.RawContractMessage {
	if m != nil {
		return m.Query
	}
	return nil
}

// QueryOracleResponse contains the result of the query sent to the oracle.
type QueryOracleResponse struct {
	// Data contains the json data returned from the oracle.
	Data github_com_CosmWasm_wasmd_x_wasm_types.RawContractMessage `protobuf:"bytes,1,opt,name=data,proto3,casttype=github.com/CosmWasm/wasmd/x/wasm/types.RawContractMessage" json:"data,omitempty"`
}

func (m *QueryOracleResponse) Reset()         { *m = QueryOracleResponse{} }
func (m *QueryOracleResponse) String() string { return proto.CompactTextString(m) }
func (*QueryOracleResponse) ProtoMessage()    {}
func (*QueryOracleResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_169907f611744c57, []int{3}
}
func (m *QueryOracleResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryOracleResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryOracleResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryOracleResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryOracleResponse.Merge(m, src)
}
func (m *QueryOracleResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryOracleResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryOracleResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryOracleResponse proto.InternalMessageInfo

func (m *QueryOracleResponse) GetData() github_com_CosmWasm_wasmd_x_wasm_types.RawContractMessage {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryOracleAddressRequest)(nil), "provenance.oracle.v1.QueryOracleAddressRequest")
	proto.RegisterType((*QueryOracleAddressResponse)(nil), "provenance.oracle.v1.QueryOracleAddressResponse")
	proto.RegisterType((*QueryOracleRequest)(nil), "provenance.oracle.v1.QueryOracleRequest")
	proto.RegisterType((*QueryOracleResponse)(nil), "provenance.oracle.v1.QueryOracleResponse")
}

func init() { proto.RegisterFile("provenance/oracle/v1/query.proto", fileDescriptor_169907f611744c57) }

var fileDescriptor_169907f611744c57 = []byte{
	// 422 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x92, 0xbf, 0xae, 0xd3, 0x30,
	0x14, 0xc6, 0xeb, 0x2b, 0xee, 0x45, 0x58, 0xb0, 0x98, 0x4a, 0xdc, 0x1b, 0xaa, 0x50, 0x45, 0x15,
	0x2a, 0x12, 0x8d, 0x69, 0x99, 0x18, 0x18, 0x68, 0x67, 0x44, 0x9b, 0x0e, 0x48, 0x2c, 0x95, 0x9b,
	0x58, 0x6e, 0xa4, 0xc6, 0x27, 0x8d, 0xdd, 0x7f, 0x2b, 0xbc, 0x00, 0x12, 0x2f, 0xc0, 0x23, 0x30,
	0xf0, 0x10, 0x8c, 0x15, 0x2c, 0x4c, 0x08, 0xb5, 0x3c, 0x05, 0x13, 0xaa, 0x9d, 0xaa, 0xad, 0x14,
	0xa0, 0xc3, 0x9d, 0xe2, 0xf8, 0x7c, 0xfe, 0x7e, 0x9f, 0x7d, 0x0e, 0xae, 0xa6, 0x19, 0xcc, 0xb8,
	0x64, 0x32, 0xe4, 0x14, 0x32, 0x16, 0x8e, 0x39, 0x9d, 0x35, 0xe9, 0x64, 0xca, 0xb3, 0xa5, 0x9f,
	0x66, 0xa0, 0x81, 0x94, 0xf7, 0x0a, 0xdf, 0x2a, 0xfc, 0x59, 0xd3, 0x29, 0x0b, 0x10, 0x60, 0x04,
	0x74, 0xbb, 0xb2, 0x5a, 0xa7, 0x22, 0x00, 0xc4, 0x98, 0x53, 0x96, 0xc6, 0x94, 0x49, 0x09, 0x9a,
	0xe9, 0x18, 0xa4, 0xca, 0xab, 0x57, 0x21, 0xa8, 0x04, 0xd4, 0xc0, 0x1e, 0xb3, 0x3f, 0xb6, 0xe4,
	0xdd, 0xc7, 0x57, 0xbd, 0x2d, 0xf3, 0x95, 0x01, 0xbc, 0x88, 0xa2, 0x8c, 0x2b, 0x15, 0xf0, 0xc9,
	0x94, 0x2b, 0xed, 0x75, 0xb1, 0x53, 0x54, 0x54, 0x29, 0x48, 0xc5, 0x49, 0x0b, 0xdf, 0x64, 0x76,
	0xeb, 0x12, 0x55, 0x51, 0xfd, 0x56, 0xfb, 0xf2, 0xeb, 0xe7, 0x46, 0x39, 0x77, 0xcf, 0xc5, 0x7d,
	0x9d, 0xc5, 0x52, 0x04, 0x3b, 0xa1, 0x17, 0x63, 0x72, 0xe0, 0x98, 0x73, 0x48, 0x1f, 0x9f, 0x9b,
	0x8b, 0x1b, 0x9f, 0xdb, 0xed, 0xe7, 0xbf, 0x7f, 0x3c, 0x78, 0x26, 0x62, 0x3d, 0x9a, 0x0e, 0xfd,
	0x10, 0x12, 0xda, 0x01, 0x95, 0xbc, 0x66, 0x2a, 0xa1, 0x73, 0xa6, 0x92, 0x88, 0x2e, 0xcc, 0x97,
	0xea, 0x65, 0xca, 0x95, 0x1f, 0xb0, 0x79, 0x07, 0xa4, 0xce, 0x58, 0xa8, 0x5f, 0x72, 0xa5, 0x98,
	0xe0, 0x81, 0xf5, 0xf2, 0x46, 0xf8, 0xee, 0x11, 0x2a, 0x4f, 0xdd, 0xc3, 0x37, 0x22, 0xa6, 0xd9,
	0xf5, 0xa0, 0x8c, 0x55, 0xeb, 0xd3, 0x19, 0x3e, 0x37, 0x28, 0xf2, 0x11, 0xe1, 0x3b, 0x47, 0x8f,
	0x45, 0xa8, 0x5f, 0xd4, 0x45, 0xff, 0xaf, 0x6f, 0xee, 0x3c, 0x39, 0xfd, 0x80, 0xbd, 0x91, 0xf7,
	0xf8, 0xed, 0xb7, 0x5f, 0x1f, 0xce, 0x1e, 0x92, 0x1a, 0x2d, 0x1c, 0x29, 0xbb, 0x1a, 0xe4, 0x1d,
	0x20, 0xef, 0x10, 0xbe, 0xb0, 0x3e, 0xa4, 0xfe, 0x5f, 0xd4, 0x2e, 0xd4, 0xa3, 0x13, 0x94, 0x79,
	0x9a, 0x9a, 0x49, 0xe3, 0x92, 0xca, 0xbf, 0xd2, 0xb4, 0xc5, 0x97, 0xb5, 0x8b, 0x56, 0x6b, 0x17,
	0xfd, 0x5c, 0xbb, 0xe8, 0xfd, 0xc6, 0x2d, 0xad, 0x36, 0x6e, 0xe9, 0xfb, 0xc6, 0x2d, 0xe1, 0x7b,
	0x31, 0x14, 0xc2, 0xba, 0xe8, 0x4d, 0xeb, 0xa0, 0x51, 0x7b, 0x49, 0x23, 0x86, 0x43, 0xd4, 0x62,
	0x07, 0x33, 0x4d, 0x1b, 0x5e, 0x98, 0x31, 0x7f, 0xfa, 0x27, 0x00, 0x00, 0xff, 0xff, 0x4f, 0x78,
	0xa0, 0x70, 0x6f, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// OracleAddress returns the address of the oracle
	OracleAddress(ctx context.Context, in *QueryOracleAddressRequest, opts ...grpc.CallOption) (*QueryOracleAddressResponse, error)
	// Oracle forwards a query to the module's oracle
	Oracle(ctx context.Context, in *QueryOracleRequest, opts ...grpc.CallOption) (*QueryOracleResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) OracleAddress(ctx context.Context, in *QueryOracleAddressRequest, opts ...grpc.CallOption) (*QueryOracleAddressResponse, error) {
	out := new(QueryOracleAddressResponse)
	err := c.cc.Invoke(ctx, "/provenance.oracle.v1.Query/OracleAddress", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Oracle(ctx context.Context, in *QueryOracleRequest, opts ...grpc.CallOption) (*QueryOracleResponse, error) {
	out := new(QueryOracleResponse)
	err := c.cc.Invoke(ctx, "/provenance.oracle.v1.Query/Oracle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// OracleAddress returns the address of the oracle
	OracleAddress(context.Context, *QueryOracleAddressRequest) (*QueryOracleAddressResponse, error)
	// Oracle forwards a query to the module's oracle
	Oracle(context.Context, *QueryOracleRequest) (*QueryOracleResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) OracleAddress(ctx context.Context, req *QueryOracleAddressRequest) (*QueryOracleAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OracleAddress not implemented")
}
func (*UnimplementedQueryServer) Oracle(ctx context.Context, req *QueryOracleRequest) (*QueryOracleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Oracle not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_OracleAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryOracleAddressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).OracleAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/provenance.oracle.v1.Query/OracleAddress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).OracleAddress(ctx, req.(*QueryOracleAddressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Oracle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryOracleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Oracle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/provenance.oracle.v1.Query/Oracle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Oracle(ctx, req.(*QueryOracleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "provenance.oracle.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "OracleAddress",
			Handler:    _Query_OracleAddress_Handler,
		},
		{
			MethodName: "Oracle",
			Handler:    _Query_Oracle_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "provenance/oracle/v1/query.proto",
}

func (m *QueryOracleAddressRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryOracleAddressRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryOracleAddressRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryOracleAddressResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryOracleAddressResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryOracleAddressResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryOracleRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryOracleRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryOracleRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Query) > 0 {
		i -= len(m.Query)
		copy(dAtA[i:], m.Query)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Query)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryOracleResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryOracleResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryOracleResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryOracleAddressRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryOracleAddressResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryOracleRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Query)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryOracleResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryOracleAddressRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryOracleAddressRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryOracleAddressRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryOracleAddressResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryOracleAddressResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryOracleAddressResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryOracleRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryOracleRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryOracleRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Query", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Query = append(m.Query[:0], dAtA[iNdEx:postIndex]...)
			if m.Query == nil {
				m.Query = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryOracleResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryOracleResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryOracleResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
