// Code generated by protoc-gen-go. DO NOT EDIT.
// source: bot.proto

/*
Package botpb is a generated protocol buffer package.

It is generated from these files:
	bot.proto

It has these top-level messages:
	Message
	SendResponse
	Capability
	Registration
	CommandAuth
	Error
*/
package botpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// bffersize
// 0 ... 100
type CapabilityBufferconfig int32

const (
	// If unbuffered is chosen and the redshirt doesn't have an open stream, messages will fail
	Capability_UNBUFFERED CapabilityBufferconfig = 0
	// these are mesage queue depths both CommandStream and NextCommand will use it.
	Capability_SMALL  CapabilityBufferconfig = 10
	Capability_MEDIUM CapabilityBufferconfig = 50
	Capability_LARGE  CapabilityBufferconfig = 100
)

var CapabilityBufferconfig_name = map[int32]string{
	0:   "UNBUFFERED",
	10:  "SMALL",
	50:  "MEDIUM",
	100: "LARGE",
}
var CapabilityBufferconfig_value = map[string]int32{
	"UNBUFFERED": 0,
	"SMALL":      10,
	"MEDIUM":     50,
	"LARGE":      100,
}

func (x CapabilityBufferconfig) String() string {
	return proto.EnumName(CapabilityBufferconfig_name, int32(x))
}
func (CapabilityBufferconfig) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

// used to represent a slack message between riker <-> client
type Message struct {
	// channel ID this came in on or should go out on :P
	Channel string `protobuf:"bytes,1,opt,name=channel" json:"channel,omitempty"`
	// WHen riker pushes a message the client needs the timestamp if it want to reply in a thread
	Timestamp string `protobuf:"bytes,2,opt,name=timestamp" json:"timestamp,omitempty"`
	// when client sends to riker, it needs to specify the thread_ts if its a reply.
	ThreadTs string `protobuf:"bytes,3,opt,name=thread_ts,json=threadTs" json:"thread_ts,omitempty"`
	// the slack nickname that sent the message
	Nickname string `protobuf:"bytes,4,opt,name=nickname" json:"nickname,omitempty"`
	// a list of slack groups that the user who sent the message is a member of
	Groups []string `protobuf:"bytes,5,rep,name=groups" json:"groups,omitempty"`
	// data is the text that was sent, or to be sent
	Payload string `protobuf:"bytes,20,opt,name=payload" json:"payload,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Message) GetChannel() string {
	if m != nil {
		return m.Channel
	}
	return ""
}

func (m *Message) GetTimestamp() string {
	if m != nil {
		return m.Timestamp
	}
	return ""
}

func (m *Message) GetThreadTs() string {
	if m != nil {
		return m.ThreadTs
	}
	return ""
}

func (m *Message) GetNickname() string {
	if m != nil {
		return m.Nickname
	}
	return ""
}

func (m *Message) GetGroups() []string {
	if m != nil {
		return m.Groups
	}
	return nil
}

func (m *Message) GetPayload() string {
	if m != nil {
		return m.Payload
	}
	return ""
}

type SendResponse struct {
	Ok      bool   `protobuf:"varint,1,opt,name=ok" json:"ok,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *SendResponse) Reset()                    { *m = SendResponse{} }
func (m *SendResponse) String() string            { return proto.CompactTextString(m) }
func (*SendResponse) ProtoMessage()               {}
func (*SendResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *SendResponse) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

func (m *SendResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type Capability struct {
	Name               string                 `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Usage              string                 `protobuf:"bytes,2,opt,name=usage" json:"usage,omitempty"`
	Description        string                 `protobuf:"bytes,3,opt,name=description" json:"description,omitempty"`
	Auth               *CommandAuth           `protobuf:"bytes,4,opt,name=auth" json:"auth,omitempty"`
	ForcedRegistration bool                   `protobuf:"varint,5,opt,name=forced_registration,json=forcedRegistration" json:"forced_registration,omitempty"`
	BufferSize         CapabilityBufferconfig `protobuf:"varint,6,opt,name=buffer_size,json=bufferSize,enum=botpb.CapabilityBufferconfig" json:"buffer_size,omitempty"`
}

func (m *Capability) Reset()                    { *m = Capability{} }
func (m *Capability) String() string            { return proto.CompactTextString(m) }
func (*Capability) ProtoMessage()               {}
func (*Capability) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Capability) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Capability) GetUsage() string {
	if m != nil {
		return m.Usage
	}
	return ""
}

func (m *Capability) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *Capability) GetAuth() *CommandAuth {
	if m != nil {
		return m.Auth
	}
	return nil
}

func (m *Capability) GetForcedRegistration() bool {
	if m != nil {
		return m.ForcedRegistration
	}
	return false
}

func (m *Capability) GetBufferSize() CapabilityBufferconfig {
	if m != nil {
		return m.BufferSize
	}
	return Capability_UNBUFFERED
}

// this is a simple response
type Registration struct {
	Name              string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	CapabilityApplied bool   `protobuf:"varint,2,opt,name=capability_applied,json=capabilityApplied" json:"capability_applied,omitempty"`
}

func (m *Registration) Reset()                    { *m = Registration{} }
func (m *Registration) String() string            { return proto.CompactTextString(m) }
func (*Registration) ProtoMessage()               {}
func (*Registration) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Registration) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Registration) GetCapabilityApplied() bool {
	if m != nil {
		return m.CapabilityApplied
	}
	return false
}

type CommandAuth struct {
	Users  []string `protobuf:"bytes,1,rep,name=users" json:"users,omitempty"`
	Groups []string `protobuf:"bytes,2,rep,name=groups" json:"groups,omitempty"`
}

func (m *CommandAuth) Reset()                    { *m = CommandAuth{} }
func (m *CommandAuth) String() string            { return proto.CompactTextString(m) }
func (*CommandAuth) ProtoMessage()               {}
func (*CommandAuth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *CommandAuth) GetUsers() []string {
	if m != nil {
		return m.Users
	}
	return nil
}

func (m *CommandAuth) GetGroups() []string {
	if m != nil {
		return m.Groups
	}
	return nil
}

type Error struct {
	Error string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
}

func (m *Error) Reset()                    { *m = Error{} }
func (m *Error) String() string            { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()               {}
func (*Error) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Error) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*Message)(nil), "botpb.Message")
	proto.RegisterType((*SendResponse)(nil), "botpb.SendResponse")
	proto.RegisterType((*Capability)(nil), "botpb.Capability")
	proto.RegisterType((*Registration)(nil), "botpb.Registration")
	proto.RegisterType((*CommandAuth)(nil), "botpb.CommandAuth")
	proto.RegisterType((*Error)(nil), "botpb.Error")
	proto.RegisterEnum("botpb.CapabilityBufferconfig", CapabilityBufferconfig_name, CapabilityBufferconfig_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Riker service

type RikerClient interface {
	// Register new RedShirt with riker
	NewRedShirt(ctx context.Context, in *Capability, opts ...grpc.CallOption) (*Registration, error)
	// Client pull from riker the next message available to process for the RedShirt
	NextCommand(ctx context.Context, in *Registration, opts ...grpc.CallOption) (*Message, error)
	//  Riker Push Stream to client
	CommandStream(ctx context.Context, in *Registration, opts ...grpc.CallOption) (Riker_CommandStreamClient, error)
	// client -> Riker Messaging
	SendStream(ctx context.Context, opts ...grpc.CallOption) (Riker_SendStreamClient, error)
	Send(ctx context.Context, in *Message, opts ...grpc.CallOption) (*SendResponse, error)
}

type rikerClient struct {
	cc *grpc.ClientConn
}

func NewRikerClient(cc *grpc.ClientConn) RikerClient {
	return &rikerClient{cc}
}

func (c *rikerClient) NewRedShirt(ctx context.Context, in *Capability, opts ...grpc.CallOption) (*Registration, error) {
	out := new(Registration)
	err := grpc.Invoke(ctx, "/botpb.Riker/NewRedShirt", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rikerClient) NextCommand(ctx context.Context, in *Registration, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := grpc.Invoke(ctx, "/botpb.Riker/NextCommand", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rikerClient) CommandStream(ctx context.Context, in *Registration, opts ...grpc.CallOption) (Riker_CommandStreamClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Riker_serviceDesc.Streams[0], c.cc, "/botpb.Riker/CommandStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &rikerCommandStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Riker_CommandStreamClient interface {
	Recv() (*Message, error)
	grpc.ClientStream
}

type rikerCommandStreamClient struct {
	grpc.ClientStream
}

func (x *rikerCommandStreamClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *rikerClient) SendStream(ctx context.Context, opts ...grpc.CallOption) (Riker_SendStreamClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Riker_serviceDesc.Streams[1], c.cc, "/botpb.Riker/SendStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &rikerSendStreamClient{stream}
	return x, nil
}

type Riker_SendStreamClient interface {
	Send(*Message) error
	CloseAndRecv() (*SendResponse, error)
	grpc.ClientStream
}

type rikerSendStreamClient struct {
	grpc.ClientStream
}

func (x *rikerSendStreamClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *rikerSendStreamClient) CloseAndRecv() (*SendResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(SendResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *rikerClient) Send(ctx context.Context, in *Message, opts ...grpc.CallOption) (*SendResponse, error) {
	out := new(SendResponse)
	err := grpc.Invoke(ctx, "/botpb.Riker/Send", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Riker service

type RikerServer interface {
	// Register new RedShirt with riker
	NewRedShirt(context.Context, *Capability) (*Registration, error)
	// Client pull from riker the next message available to process for the RedShirt
	NextCommand(context.Context, *Registration) (*Message, error)
	//  Riker Push Stream to client
	CommandStream(*Registration, Riker_CommandStreamServer) error
	// client -> Riker Messaging
	SendStream(Riker_SendStreamServer) error
	Send(context.Context, *Message) (*SendResponse, error)
}

func RegisterRikerServer(s *grpc.Server, srv RikerServer) {
	s.RegisterService(&_Riker_serviceDesc, srv)
}

func _Riker_NewRedShirt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Capability)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RikerServer).NewRedShirt(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/botpb.Riker/NewRedShirt",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RikerServer).NewRedShirt(ctx, req.(*Capability))
	}
	return interceptor(ctx, in, info, handler)
}

func _Riker_NextCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Registration)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RikerServer).NextCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/botpb.Riker/NextCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RikerServer).NextCommand(ctx, req.(*Registration))
	}
	return interceptor(ctx, in, info, handler)
}

func _Riker_CommandStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Registration)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RikerServer).CommandStream(m, &rikerCommandStreamServer{stream})
}

type Riker_CommandStreamServer interface {
	Send(*Message) error
	grpc.ServerStream
}

type rikerCommandStreamServer struct {
	grpc.ServerStream
}

func (x *rikerCommandStreamServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func _Riker_SendStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RikerServer).SendStream(&rikerSendStreamServer{stream})
}

type Riker_SendStreamServer interface {
	SendAndClose(*SendResponse) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type rikerSendStreamServer struct {
	grpc.ServerStream
}

func (x *rikerSendStreamServer) SendAndClose(m *SendResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *rikerSendStreamServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Riker_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RikerServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/botpb.Riker/Send",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RikerServer).Send(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

var _Riker_serviceDesc = grpc.ServiceDesc{
	ServiceName: "botpb.Riker",
	HandlerType: (*RikerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewRedShirt",
			Handler:    _Riker_NewRedShirt_Handler,
		},
		{
			MethodName: "NextCommand",
			Handler:    _Riker_NextCommand_Handler,
		},
		{
			MethodName: "Send",
			Handler:    _Riker_Send_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CommandStream",
			Handler:       _Riker_CommandStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SendStream",
			Handler:       _Riker_SendStream_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "bot.proto",
}

func init() { proto.RegisterFile("bot.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 536 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0x4f, 0x6b, 0xdb, 0x4e,
	0x14, 0xfc, 0x49, 0xb1, 0x1c, 0xeb, 0x39, 0x3f, 0xe3, 0xbc, 0x84, 0x22, 0xdc, 0x3f, 0x18, 0x1d,
	0x8a, 0xa1, 0xd4, 0x2d, 0x0e, 0x2d, 0x85, 0x1e, 0x5a, 0x37, 0x71, 0x4a, 0xc1, 0x0e, 0x74, 0x5d,
	0x9f, 0xcd, 0x4a, 0x5a, 0xdb, 0x8b, 0x2d, 0xad, 0xd8, 0x5d, 0xd3, 0x26, 0x9f, 0xa1, 0x1f, 0xa5,
	0xc7, 0x7e, 0xc0, 0xa2, 0x5d, 0xd9, 0x51, 0xda, 0x1c, 0x72, 0xd3, 0xcc, 0xbc, 0x27, 0x66, 0x86,
	0xb7, 0xe0, 0x47, 0x42, 0xf7, 0x73, 0x29, 0xb4, 0x40, 0x2f, 0x12, 0x3a, 0x8f, 0xc2, 0x5f, 0x0e,
	0x1c, 0x4e, 0x98, 0x52, 0x74, 0xc9, 0x30, 0x80, 0xc3, 0x78, 0x45, 0xb3, 0x8c, 0x6d, 0x02, 0xa7,
	0xeb, 0xf4, 0x7c, 0xb2, 0x83, 0xf8, 0x04, 0x7c, 0xcd, 0x53, 0xa6, 0x34, 0x4d, 0xf3, 0xc0, 0x35,
	0xda, 0x2d, 0x81, 0x8f, 0xc1, 0xd7, 0x2b, 0xc9, 0x68, 0x32, 0xd7, 0x2a, 0x38, 0x30, 0x6a, 0xc3,
	0x12, 0xdf, 0x14, 0x76, 0xa0, 0x91, 0xf1, 0x78, 0x9d, 0xd1, 0x94, 0x05, 0x35, 0xab, 0xed, 0x30,
	0x3e, 0x82, 0xfa, 0x52, 0x8a, 0x6d, 0xae, 0x02, 0xaf, 0x7b, 0xd0, 0xf3, 0x49, 0x89, 0x0a, 0x23,
	0x39, 0xbd, 0xde, 0x08, 0x9a, 0x04, 0xa7, 0xd6, 0x48, 0x09, 0xc3, 0x77, 0x70, 0x34, 0x65, 0x59,
	0x42, 0x98, 0xca, 0x45, 0xa6, 0x18, 0xb6, 0xc0, 0x15, 0x6b, 0xe3, 0xb6, 0x41, 0x5c, 0xb1, 0x2e,
	0x36, 0x53, 0x9b, 0xa6, 0xb4, 0xb9, 0x83, 0xe1, 0x6f, 0x17, 0xe0, 0x9c, 0xe6, 0x34, 0xe2, 0x1b,
	0xae, 0xaf, 0x11, 0xa1, 0x66, 0x2c, 0xd9, 0xa0, 0xe6, 0x1b, 0x4f, 0xc1, 0xdb, 0x56, 0x56, 0x2d,
	0xc0, 0x2e, 0x34, 0x13, 0xa6, 0x62, 0xc9, 0x73, 0xcd, 0x45, 0x56, 0xe6, 0xab, 0x52, 0xf8, 0x1c,
	0x6a, 0x74, 0xab, 0x57, 0x26, 0x5e, 0x73, 0x80, 0x7d, 0xd3, 0x6c, 0xff, 0x5c, 0xa4, 0x29, 0xcd,
	0x92, 0xe1, 0x56, 0xaf, 0x88, 0xd1, 0xf1, 0x15, 0x9c, 0x2c, 0x84, 0x8c, 0x59, 0x32, 0x97, 0x6c,
	0xc9, 0x95, 0x96, 0xd4, 0xfc, 0xd1, 0x33, 0xee, 0xd1, 0x4a, 0xa4, 0xa2, 0xe0, 0x07, 0x68, 0x46,
	0xdb, 0xc5, 0x82, 0xc9, 0xb9, 0xe2, 0x37, 0x2c, 0xa8, 0x77, 0x9d, 0x5e, 0x6b, 0xf0, 0x6c, 0xf7,
	0xff, 0x7d, 0x98, 0xbe, 0x1d, 0x8a, 0x45, 0xb6, 0xe0, 0x4b, 0x02, 0x16, 0x4d, 0xf9, 0x0d, 0x0b,
	0x3f, 0xc2, 0x51, 0x55, 0xc3, 0x16, 0xc0, 0xec, 0xea, 0xd3, 0xec, 0xf2, 0x72, 0x44, 0x46, 0x17,
	0xed, 0xff, 0xd0, 0x07, 0x6f, 0x3a, 0x19, 0x8e, 0xc7, 0x6d, 0x40, 0x80, 0xfa, 0x64, 0x74, 0xf1,
	0x65, 0x36, 0x69, 0x0f, 0x0a, 0x7a, 0x3c, 0x24, 0x9f, 0x47, 0xed, 0x24, 0xfc, 0x0a, 0x47, 0x77,
	0x2c, 0xdd, 0xd7, 0xdb, 0x4b, 0xc0, 0x78, 0x6f, 0x66, 0x4e, 0xf3, 0x7c, 0xc3, 0x59, 0x62, 0x4a,
	0x6c, 0x90, 0xe3, 0x5b, 0x65, 0x68, 0x85, 0xf0, 0x3d, 0x34, 0x2b, 0xdd, 0xd8, 0xd6, 0x99, 0x54,
	0x81, 0x63, 0x6e, 0xc0, 0x82, 0xca, 0x69, 0xb8, 0xd5, 0xd3, 0x08, 0x9f, 0x82, 0x37, 0x92, 0x52,
	0xc8, 0x62, 0x8d, 0x15, 0x1f, 0xa5, 0x13, 0x0b, 0x06, 0x3f, 0x5d, 0xf0, 0x08, 0x5f, 0x33, 0x89,
	0x6f, 0xa0, 0x79, 0xc5, 0xbe, 0x13, 0x96, 0x4c, 0x57, 0x5c, 0x6a, 0x3c, 0xfe, 0xa7, 0xb5, 0xce,
	0x49, 0x49, 0xdd, 0xc9, 0x37, 0x28, 0xd6, 0x7e, 0xe8, 0xd2, 0x20, 0xde, 0x37, 0xd3, 0x69, 0x95,
	0xe4, 0xee, 0xdd, 0xbc, 0x85, 0xff, 0xcb, 0xf9, 0xa9, 0x96, 0x8c, 0xa6, 0x0f, 0xda, 0x7a, 0xed,
	0xe0, 0x19, 0x40, 0x71, 0xcc, 0xe5, 0xd2, 0x5f, 0xfa, 0xde, 0x5e, 0xf5, 0xde, 0x7b, 0x0e, 0xbe,
	0x80, 0x5a, 0xc1, 0x3c, 0x68, 0x3c, 0xaa, 0x9b, 0xb7, 0x7e, 0xf6, 0x27, 0x00, 0x00, 0xff, 0xff,
	0xc8, 0xfa, 0x96, 0x23, 0xf8, 0x03, 0x00, 0x00,
}
