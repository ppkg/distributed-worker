// Code generated by protoc-gen-go. DO NOT EDIT.
// source: proto/task/task.proto

package task

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

type SubmitRequest struct {
	Id    int64 `protobuf:"varint,1,opt,name=Id,proto3" json:"Id"`
	JobId int64 `protobuf:"varint,2,opt,name=JobId,proto3" json:"JobId"`
	// task名称
	Name string `protobuf:"bytes,3,opt,name=Name,proto3" json:"Name"`
	// 任务处理插件
	Plugin string `protobuf:"bytes,4,opt,name=Plugin,proto3" json:"Plugin"`
	// task输入数据
	Data                 string   `protobuf:"bytes,5,opt,name=Data,proto3" json:"Data"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SubmitRequest) Reset()         { *m = SubmitRequest{} }
func (m *SubmitRequest) String() string { return proto.CompactTextString(m) }
func (*SubmitRequest) ProtoMessage()    {}
func (*SubmitRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_152e577c5c92a6d4, []int{0}
}

func (m *SubmitRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubmitRequest.Unmarshal(m, b)
}
func (m *SubmitRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubmitRequest.Marshal(b, m, deterministic)
}
func (m *SubmitRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubmitRequest.Merge(m, src)
}
func (m *SubmitRequest) XXX_Size() int {
	return xxx_messageInfo_SubmitRequest.Size(m)
}
func (m *SubmitRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SubmitRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SubmitRequest proto.InternalMessageInfo

func (m *SubmitRequest) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *SubmitRequest) GetJobId() int64 {
	if m != nil {
		return m.JobId
	}
	return 0
}

func (m *SubmitRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *SubmitRequest) GetPlugin() string {
	if m != nil {
		return m.Plugin
	}
	return ""
}

func (m *SubmitRequest) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

type SyncSubmitResponse struct {
	Id    int64 `protobuf:"varint,1,opt,name=Id,proto3" json:"Id"`
	JobId int64 `protobuf:"varint,2,opt,name=JobId,proto3" json:"JobId"`
	// task状态,1：执行中，2：执行完成，3：异常退出
	Status int32 `protobuf:"varint,3,opt,name=status,proto3" json:"status"`
	// 处理结果
	Result               string   `protobuf:"bytes,4,opt,name=Result,proto3" json:"Result"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SyncSubmitResponse) Reset()         { *m = SyncSubmitResponse{} }
func (m *SyncSubmitResponse) String() string { return proto.CompactTextString(m) }
func (*SyncSubmitResponse) ProtoMessage()    {}
func (*SyncSubmitResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_152e577c5c92a6d4, []int{1}
}

func (m *SyncSubmitResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SyncSubmitResponse.Unmarshal(m, b)
}
func (m *SyncSubmitResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SyncSubmitResponse.Marshal(b, m, deterministic)
}
func (m *SyncSubmitResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SyncSubmitResponse.Merge(m, src)
}
func (m *SyncSubmitResponse) XXX_Size() int {
	return xxx_messageInfo_SyncSubmitResponse.Size(m)
}
func (m *SyncSubmitResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SyncSubmitResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SyncSubmitResponse proto.InternalMessageInfo

func (m *SyncSubmitResponse) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *SyncSubmitResponse) GetJobId() int64 {
	if m != nil {
		return m.JobId
	}
	return 0
}

func (m *SyncSubmitResponse) GetStatus() int32 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *SyncSubmitResponse) GetResult() string {
	if m != nil {
		return m.Result
	}
	return ""
}

func init() {
	proto.RegisterType((*SubmitRequest)(nil), "task.SubmitRequest")
	proto.RegisterType((*SyncSubmitResponse)(nil), "task.SyncSubmitResponse")
}

func init() { proto.RegisterFile("proto/task/task.proto", fileDescriptor_152e577c5c92a6d4) }

var fileDescriptor_152e577c5c92a6d4 = []byte{
	// 220 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0x3f, 0x4b, 0xc4, 0x40,
	0x10, 0xc5, 0x49, 0x2e, 0x09, 0x38, 0xa2, 0xc5, 0xa8, 0xc7, 0x62, 0x75, 0x5c, 0x75, 0xd5, 0x09,
	0x5a, 0xda, 0xda, 0xe4, 0x0a, 0x91, 0x8d, 0x5f, 0x60, 0x93, 0x2c, 0x12, 0xf3, 0x67, 0x63, 0x66,
	0x56, 0xf0, 0xdb, 0xcb, 0x4e, 0x12, 0x44, 0x6c, 0x6c, 0x96, 0xf7, 0x7b, 0x0c, 0xfc, 0x96, 0x07,
	0x37, 0xe3, 0xe4, 0xd8, 0xdd, 0xb1, 0xa1, 0x56, 0x9e, 0xa3, 0x30, 0x26, 0x21, 0xef, 0x3d, 0x5c,
	0x14, 0xbe, 0xec, 0x1b, 0xd6, 0xf6, 0xc3, 0x5b, 0x62, 0xbc, 0x84, 0x38, 0xaf, 0x55, 0xb4, 0x8b,
	0x0e, 0x1b, 0x1d, 0xe7, 0x35, 0x5e, 0x43, 0x7a, 0x72, 0x65, 0x5e, 0xab, 0x58, 0xaa, 0x19, 0x10,
	0x21, 0x79, 0x36, 0xbd, 0x55, 0x9b, 0x5d, 0x74, 0x38, 0xd3, 0x92, 0x71, 0x0b, 0xd9, 0x4b, 0xe7,
	0xdf, 0x9a, 0x41, 0x25, 0xd2, 0x2e, 0x14, 0x6e, 0x9f, 0x0c, 0x1b, 0x95, 0xce, 0xb7, 0x21, 0xef,
	0xdf, 0x01, 0x8b, 0xaf, 0xa1, 0x5a, 0xd5, 0x34, 0xba, 0x81, 0xec, 0x3f, 0xdd, 0x5b, 0xc8, 0x88,
	0x0d, 0x7b, 0x12, 0x7b, 0xaa, 0x17, 0x0a, 0xbd, 0xb6, 0xe4, 0x3b, 0x5e, 0xfd, 0x33, 0xdd, 0x9f,
	0xe0, 0xfc, 0xd5, 0x50, 0x5b, 0xd8, 0xe9, 0xb3, 0xa9, 0x2c, 0x3e, 0x02, 0xfc, 0xa8, 0xf1, 0xea,
	0x28, 0x93, 0xfc, 0xda, 0xe0, 0x56, 0x2d, 0xe5, 0x9f, 0x1f, 0x96, 0x99, 0x6c, 0xf7, 0xf0, 0x1d,
	0x00, 0x00, 0xff, 0xff, 0x66, 0xa0, 0xd4, 0xc3, 0x54, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TaskServiceClient is the client API for TaskService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TaskServiceClient interface {
	// 同步提交task
	SyncSubmit(ctx context.Context, in *SubmitRequest, opts ...grpc.CallOption) (*SyncSubmitResponse, error)
}

type taskServiceClient struct {
	cc *grpc.ClientConn
}

func NewTaskServiceClient(cc *grpc.ClientConn) TaskServiceClient {
	return &taskServiceClient{cc}
}

func (c *taskServiceClient) SyncSubmit(ctx context.Context, in *SubmitRequest, opts ...grpc.CallOption) (*SyncSubmitResponse, error) {
	out := new(SyncSubmitResponse)
	err := c.cc.Invoke(ctx, "/task.TaskService/SyncSubmit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TaskServiceServer is the server API for TaskService service.
type TaskServiceServer interface {
	// 同步提交task
	SyncSubmit(context.Context, *SubmitRequest) (*SyncSubmitResponse, error)
}

// UnimplementedTaskServiceServer can be embedded to have forward compatible implementations.
type UnimplementedTaskServiceServer struct {
}

func (*UnimplementedTaskServiceServer) SyncSubmit(ctx context.Context, req *SubmitRequest) (*SyncSubmitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncSubmit not implemented")
}

func RegisterTaskServiceServer(s *grpc.Server, srv TaskServiceServer) {
	s.RegisterService(&_TaskService_serviceDesc, srv)
}

func _TaskService_SyncSubmit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TaskServiceServer).SyncSubmit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/task.TaskService/SyncSubmit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TaskServiceServer).SyncSubmit(ctx, req.(*SubmitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TaskService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "task.TaskService",
	HandlerType: (*TaskServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SyncSubmit",
			Handler:    _TaskService_SyncSubmit_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/task/task.proto",
}
