
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

type PingRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	mi := &file_v1_common_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_common_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_v1_common_proto_rawDescGZIP(), []int{0}
}

func (x *PingRequest) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type PingResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	mi := &file_v1_common_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_common_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_v1_common_proto_rawDescGZIP(), []int{1}
}

func (x *PingResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type HeartbeatRequest struct {
	state     protoimpl.MessageState `protogen:"open.v1"`
	AgentId   string                 `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
	Timestamp int64                  `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	CpuUsage      float64 `protobuf:"fixed64,3,opt,name=cpu_usage,json=cpuUsage,proto3" json:"cpu_usage,omitempty"`
	MemUsage      float64 `protobuf:"fixed64,4,opt,name=mem_usage,json=memUsage,proto3" json:"mem_usage,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeartbeatRequest) Reset() {
	*x = HeartbeatRequest{}
	mi := &file_v1_common_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatRequest) ProtoMessage() {}

func (x *HeartbeatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_common_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HeartbeatRequest) Descriptor() ([]byte, []int) {
	return file_v1_common_proto_rawDescGZIP(), []int{2}
}

func (x *HeartbeatRequest) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

func (x *HeartbeatRequest) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *HeartbeatRequest) GetCpuUsage() float64 {
	if x != nil {
		return x.CpuUsage
	}
	return 0
}

func (x *HeartbeatRequest) GetMemUsage() float64 {
	if x != nil {
		return x.MemUsage
	}
	return 0
}

type HeartbeatResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	Ok    bool                   `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	ConfigStale   bool `protobuf:"varint,2,opt,name=config_stale,json=configStale,proto3" json:"config_stale,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeartbeatResponse) Reset() {
	*x = HeartbeatResponse{}
	mi := &file_v1_common_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeartbeatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeartbeatResponse) ProtoMessage() {}

func (x *HeartbeatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_common_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HeartbeatResponse) Descriptor() ([]byte, []int) {
	return file_v1_common_proto_rawDescGZIP(), []int{3}
}

func (x *HeartbeatResponse) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

func (x *HeartbeatResponse) GetConfigStale() bool {
	if x != nil {
		return x.ConfigStale
	}
	return false
}

var File_v1_common_proto protoreflect.FileDescriptor

const file_v1_common_proto_rawDesc = "" +
	"\n" +
	"\x0fv1/common.proto\x12\tgimpel.v1\"'\n" +
	"\vPingRequest\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\"(\n" +
	"\fPingResponse\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\"\x85\x01\n" +
	"\x10HeartbeatRequest\x12\x19\n" +
	"\bagent_id\x18\x01 \x01(\tR\aagentId\x12\x1c\n" +
	"\ttimestamp\x18\x02 \x01(\x03R\ttimestamp\x12\x1b\n" +
	"\tcpu_usage\x18\x03 \x01(\x01R\bcpuUsage\x12\x1b\n" +
	"\tmem_usage\x18\x04 \x01(\x01R\bmemUsage\"F\n" +
	"\x11HeartbeatResponse\x12\x0e\n" +
	"\x02ok\x18\x01 \x01(\bR\x02ok\x12!\n" +
	"\fconfig_stale\x18\x02 \x01(\bR\vconfigStale2\x90\x01\n" +
	"\rGimpelControl\x127\n" +
	"\x04Ping\x12\x16.gimpel.v1.PingRequest\x1a\x17.gimpel.v1.PingResponse\x12F\n" +
	"\tHeartbeat\x12\x1b.gimpel.v1.HeartbeatRequest\x1a\x1c.gimpel.v1.HeartbeatResponseB5Z3github.com/nohaxxjustlags/gimpel/api/go/v1;gimpelv1b\x06proto3"

var (
	file_v1_common_proto_rawDescOnce sync.Once
	file_v1_common_proto_rawDescData []byte
)

func file_v1_common_proto_rawDescGZIP() []byte {
	file_v1_common_proto_rawDescOnce.Do(func() {
		file_v1_common_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_v1_common_proto_rawDesc), len(file_v1_common_proto_rawDesc)))
	})
	return file_v1_common_proto_rawDescData
}

var file_v1_common_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_v1_common_proto_goTypes = []any{
	(*PingRequest)(nil),
	(*PingResponse)(nil),
	(*HeartbeatRequest)(nil),
	(*HeartbeatResponse)(nil),
}
var file_v1_common_proto_depIdxs = []int32{
	0,
	2,
	1,
	3,
	2,
	0,
	0,
	0,
	0,
}

func init() { file_v1_common_proto_init() }
func file_v1_common_proto_init() {
	if File_v1_common_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_v1_common_proto_rawDesc), len(file_v1_common_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_common_proto_goTypes,
		DependencyIndexes: file_v1_common_proto_depIdxs,
		MessageInfos:      file_v1_common_proto_msgTypes,
	}.Build()
	File_v1_common_proto = out.File
	file_v1_common_proto_goTypes = nil
	file_v1_common_proto_depIdxs = nil
}
