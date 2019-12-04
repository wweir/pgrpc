// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ping.proto

package main

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type PingMsg struct {
	Msg                  string   `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingMsg) Reset()         { *m = PingMsg{} }
func (m *PingMsg) String() string { return proto.CompactTextString(m) }
func (*PingMsg) ProtoMessage()    {}
func (*PingMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_6d51d96c3ad891f5, []int{0}
}

func (m *PingMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingMsg.Unmarshal(m, b)
}
func (m *PingMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingMsg.Marshal(b, m, deterministic)
}
func (m *PingMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingMsg.Merge(m, src)
}
func (m *PingMsg) XXX_Size() int {
	return xxx_messageInfo_PingMsg.Size(m)
}
func (m *PingMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_PingMsg.DiscardUnknown(m)
}

var xxx_messageInfo_PingMsg proto.InternalMessageInfo

func (m *PingMsg) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func init() {
	proto.RegisterType((*PingMsg)(nil), "main.PingMsg")
}

func init() { proto.RegisterFile("ping.proto", fileDescriptor_6d51d96c3ad891f5) }

var fileDescriptor_6d51d96c3ad891f5 = []byte{
	// 95 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2a, 0xc8, 0xcc, 0x4b,
	0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0xc9, 0x4d, 0xcc, 0xcc, 0x53, 0x92, 0xe6, 0x62,
	0x0f, 0xc8, 0xcc, 0x4b, 0xf7, 0x2d, 0x4e, 0x17, 0x12, 0xe0, 0x62, 0xce, 0x2d, 0x4e, 0x97, 0x60,
	0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x02, 0x31, 0x8d, 0xf4, 0xb8, 0x58, 0x40, 0x92, 0x42, 0x6a, 0x50,
	0x9a, 0x57, 0x0f, 0xa4, 0x47, 0x0f, 0xaa, 0x41, 0x0a, 0x95, 0xab, 0xc4, 0x90, 0xc4, 0x06, 0x36,
	0xd9, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x81, 0x4c, 0xd3, 0x32, 0x67, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PingClient is the client API for Ping service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PingClient interface {
	Ping(ctx context.Context, in *PingMsg, opts ...grpc.CallOption) (*PingMsg, error)
}

type pingClient struct {
	cc *grpc.ClientConn
}

func NewPingClient(cc *grpc.ClientConn) PingClient {
	return &pingClient{cc}
}

func (c *pingClient) Ping(ctx context.Context, in *PingMsg, opts ...grpc.CallOption) (*PingMsg, error) {
	out := new(PingMsg)
	err := c.cc.Invoke(ctx, "/main.Ping/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PingServer is the server API for Ping service.
type PingServer interface {
	Ping(context.Context, *PingMsg) (*PingMsg, error)
}

// UnimplementedPingServer can be embedded to have forward compatible implementations.
type UnimplementedPingServer struct {
}

func (*UnimplementedPingServer) Ping(ctx context.Context, req *PingMsg) (*PingMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func RegisterPingServer(s *grpc.Server, srv PingServer) {
	s.RegisterService(&_Ping_serviceDesc, srv)
}

func _Ping_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingMsg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.Ping/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingServer).Ping(ctx, req.(*PingMsg))
	}
	return interceptor(ctx, in, info, handler)
}

var _Ping_serviceDesc = grpc.ServiceDesc{
	ServiceName: "main.Ping",
	HandlerType: (*PingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Ping_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ping.proto",
}
