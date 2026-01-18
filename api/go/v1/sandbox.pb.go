
package gimpelv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateSessionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Image         string                 `protobuf:"bytes,2,opt,name=image,proto3" json:"image,omitempty"`
	Env           map[string]string      `protobuf:"bytes,3,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateSessionRequest) Reset() {
	*x = CreateSessionRequest{}
	mi := &file_v1_sandbox_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateSessionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSessionRequest) ProtoMessage() {}

func (x *CreateSessionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_sandbox_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CreateSessionRequest) Descriptor() ([]byte, []int) {
	return file_v1_sandbox_proto_rawDescGZIP(), []int{0}
}

func (x *CreateSessionRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *CreateSessionRequest) GetImage() string {
	if x != nil {
		return x.Image
	}
	return ""
}

func (x *CreateSessionRequest) GetEnv() map[string]string {
	if x != nil {
		return x.Env
	}
	return nil
}

type CreateSessionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Endpoint      string                 `protobuf:"bytes,1,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	TunnelKey     []byte                 `protobuf:"bytes,2,opt,name=tunnel_key,json=tunnelKey,proto3" json:"tunnel_key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateSessionResponse) Reset() {
	*x = CreateSessionResponse{}
	mi := &file_v1_sandbox_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateSessionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateSessionResponse) ProtoMessage() {}

func (x *CreateSessionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_sandbox_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*CreateSessionResponse) Descriptor() ([]byte, []int) {
	return file_v1_sandbox_proto_rawDescGZIP(), []int{1}
}

func (x *CreateSessionResponse) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *CreateSessionResponse) GetTunnelKey() []byte {
	if x != nil {
		return x.TunnelKey
	}
	return nil
}

type StopSessionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StopSessionRequest) Reset() {
	*x = StopSessionRequest{}
	mi := &file_v1_sandbox_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StopSessionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopSessionRequest) ProtoMessage() {}

func (x *StopSessionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_sandbox_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*StopSessionRequest) Descriptor() ([]byte, []int) {
	return file_v1_sandbox_proto_rawDescGZIP(), []int{2}
}

func (x *StopSessionRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

type StopSessionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *StopSessionResponse) Reset() {
	*x = StopSessionResponse{}
	mi := &file_v1_sandbox_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StopSessionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StopSessionResponse) ProtoMessage() {}

func (x *StopSessionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_sandbox_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*StopSessionResponse) Descriptor() ([]byte, []int) {
	return file_v1_sandbox_proto_rawDescGZIP(), []int{3}
}

func (x *StopSessionResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_v1_sandbox_proto protoreflect.FileDescriptor

const file_v1_sandbox_proto_rawDesc = "" +
	"\n" +
	"\x10v1/sandbox.proto\x12\tgimpel.v1\"\xbf\x01\n" +
	"\x14CreateSessionRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x14\n" +
	"\x05image\x18\x02 \x01(\tR\x05image\x12:\n" +
	"\x03env\x18\x03 \x03(\v2(.gimpel.v1.CreateSessionRequest.EnvEntryR\x03env\x1a6\n" +
	"\bEnvEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"R\n" +
	"\x15CreateSessionResponse\x12\x1a\n" +
	"\bendpoint\x18\x01 \x01(\tR\bendpoint\x12\x1d\n" +
	"\n" +
	"tunnel_key\x18\x02 \x01(\fR\ttunnelKey\"3\n" +
	"\x12StopSessionRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\"/\n" +
	"\x13StopSessionResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess2\xb2\x01\n" +
	"\x0eSandboxService\x12R\n" +
	"\rCreateSession\x12\x1f.gimpel.v1.CreateSessionRequest\x1a .gimpel.v1.CreateSessionResponse\x12L\n" +
	"\vStopSession\x12\x1d.gimpel.v1.StopSessionRequest\x1a\x1e.gimpel.v1.StopSessionResponseB5Z3github.com/nohaxxjustlags/gimpel/api/go/v1;gimpelv1b\x06proto3"

var (
	file_v1_sandbox_proto_rawDescOnce sync.Once
	file_v1_sandbox_proto_rawDescData []byte
)

func file_v1_sandbox_proto_rawDescGZIP() []byte {
	file_v1_sandbox_proto_rawDescOnce.Do(func() {
		file_v1_sandbox_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_v1_sandbox_proto_rawDesc), len(file_v1_sandbox_proto_rawDesc)))
	})
	return file_v1_sandbox_proto_rawDescData
}

var file_v1_sandbox_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_v1_sandbox_proto_goTypes = []any{
	(*CreateSessionRequest)(nil),
	(*CreateSessionResponse)(nil),
	(*StopSessionRequest)(nil),
	(*StopSessionResponse)(nil),
	nil,
}
var file_v1_sandbox_proto_depIdxs = []int32{
	4,
	0,
	2,
	1,
	3,
	3,
	1,
	1,
	1,
	0,
}

func init() { file_v1_sandbox_proto_init() }
func file_v1_sandbox_proto_init() {
	if File_v1_sandbox_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_v1_sandbox_proto_rawDesc), len(file_v1_sandbox_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_sandbox_proto_goTypes,
		DependencyIndexes: file_v1_sandbox_proto_depIdxs,
		MessageInfos:      file_v1_sandbox_proto_msgTypes,
	}.Build()
	File_v1_sandbox_proto = out.File
	file_v1_sandbox_proto_goTypes = nil
	file_v1_sandbox_proto_depIdxs = nil
}
