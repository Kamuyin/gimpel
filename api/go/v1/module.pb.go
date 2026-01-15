
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

type ConnectionInfo struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ConnectionId  string                 `protobuf:"bytes,1,opt,name=connection_id,json=connectionId,proto3" json:"connection_id,omitempty"`
	SourceIp      string                 `protobuf:"bytes,2,opt,name=source_ip,json=sourceIp,proto3" json:"source_ip,omitempty"`
	SourcePort    uint32                 `protobuf:"varint,3,opt,name=source_port,json=sourcePort,proto3" json:"source_port,omitempty"`
	DestIp        string                 `protobuf:"bytes,4,opt,name=dest_ip,json=destIp,proto3" json:"dest_ip,omitempty"`
	DestPort      uint32                 `protobuf:"varint,5,opt,name=dest_port,json=destPort,proto3" json:"dest_port,omitempty"`
	Protocol      string                 `protobuf:"bytes,6,opt,name=protocol,proto3" json:"protocol,omitempty"`
	TimestampNs   int64                  `protobuf:"varint,7,opt,name=timestamp_ns,json=timestampNs,proto3" json:"timestamp_ns,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConnectionInfo) Reset() {
	*x = ConnectionInfo{}
	mi := &file_v1_module_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConnectionInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConnectionInfo) ProtoMessage() {}

func (x *ConnectionInfo) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ConnectionInfo) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{0}
}

func (x *ConnectionInfo) GetConnectionId() string {
	if x != nil {
		return x.ConnectionId
	}
	return ""
}

func (x *ConnectionInfo) GetSourceIp() string {
	if x != nil {
		return x.SourceIp
	}
	return ""
}

func (x *ConnectionInfo) GetSourcePort() uint32 {
	if x != nil {
		return x.SourcePort
	}
	return 0
}

func (x *ConnectionInfo) GetDestIp() string {
	if x != nil {
		return x.DestIp
	}
	return ""
}

func (x *ConnectionInfo) GetDestPort() uint32 {
	if x != nil {
		return x.DestPort
	}
	return 0
}

func (x *ConnectionInfo) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *ConnectionInfo) GetTimestampNs() int64 {
	if x != nil {
		return x.TimestampNs
	}
	return 0
}

type HandleConnectionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Connection    *ConnectionInfo        `protobuf:"bytes,1,opt,name=connection,proto3" json:"connection,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandleConnectionRequest) Reset() {
	*x = HandleConnectionRequest{}
	mi := &file_v1_module_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandleConnectionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandleConnectionRequest) ProtoMessage() {}

func (x *HandleConnectionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HandleConnectionRequest) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{1}
}

func (x *HandleConnectionRequest) GetConnection() *ConnectionInfo {
	if x != nil {
		return x.Connection
	}
	return nil
}

type HandleConnectionResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Accepted      bool                   `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
	DataPort      int32                  `protobuf:"varint,2,opt,name=data_port,json=dataPort,proto3" json:"data_port,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HandleConnectionResponse) Reset() {
	*x = HandleConnectionResponse{}
	mi := &file_v1_module_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HandleConnectionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandleConnectionResponse) ProtoMessage() {}

func (x *HandleConnectionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HandleConnectionResponse) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{2}
}

func (x *HandleConnectionResponse) GetAccepted() bool {
	if x != nil {
		return x.Accepted
	}
	return false
}

func (x *HandleConnectionResponse) GetDataPort() int32 {
	if x != nil {
		return x.DataPort
	}
	return 0
}

type HealthCheckRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HealthCheckRequest) Reset() {
	*x = HealthCheckRequest{}
	mi := &file_v1_module_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HealthCheckRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckRequest) ProtoMessage() {}

func (x *HealthCheckRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HealthCheckRequest) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{3}
}

type HealthCheckResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Healthy       bool                   `protobuf:"varint,1,opt,name=healthy,proto3" json:"healthy,omitempty"`
	Status        string                 `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Metadata      map[string]string      `protobuf:"bytes,3,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HealthCheckResponse) Reset() {
	*x = HealthCheckResponse{}
	mi := &file_v1_module_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HealthCheckResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HealthCheckResponse) ProtoMessage() {}

func (x *HealthCheckResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HealthCheckResponse) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{4}
}

func (x *HealthCheckResponse) GetHealthy() bool {
	if x != nil {
		return x.Healthy
	}
	return false
}

func (x *HealthCheckResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *HealthCheckResponse) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

var File_v1_module_proto protoreflect.FileDescriptor

const file_v1_module_proto_rawDesc = "" +
	"\n" +
	"\x0fv1/module.proto\x12\tgimpel.v1\"\xe8\x01\n" +
	"\x0eConnectionInfo\x12#\n" +
	"\rconnection_id\x18\x01 \x01(\tR\fconnectionId\x12\x1b\n" +
	"\tsource_ip\x18\x02 \x01(\tR\bsourceIp\x12\x1f\n" +
	"\vsource_port\x18\x03 \x01(\rR\n" +
	"sourcePort\x12\x17\n" +
	"\adest_ip\x18\x04 \x01(\tR\x06destIp\x12\x1b\n" +
	"\tdest_port\x18\x05 \x01(\rR\bdestPort\x12\x1a\n" +
	"\bprotocol\x18\x06 \x01(\tR\bprotocol\x12!\n" +
	"\ftimestamp_ns\x18\a \x01(\x03R\vtimestampNs\"T\n" +
	"\x17HandleConnectionRequest\x129\n" +
	"\n" +
	"connection\x18\x01 \x01(\v2\x19.gimpel.v1.ConnectionInfoR\n" +
	"connection\"S\n" +
	"\x18HandleConnectionResponse\x12\x1a\n" +
	"\baccepted\x18\x01 \x01(\bR\baccepted\x12\x1b\n" +
	"\tdata_port\x18\x02 \x01(\x05R\bdataPort\"\x14\n" +
	"\x12HealthCheckRequest\"\xce\x01\n" +
	"\x13HealthCheckResponse\x12\x18\n" +
	"\ahealthy\x18\x01 \x01(\bR\ahealthy\x12\x16\n" +
	"\x06status\x18\x02 \x01(\tR\x06status\x12H\n" +
	"\bmetadata\x18\x03 \x03(\v2,.gimpel.v1.HealthCheckResponse.MetadataEntryR\bmetadata\x1a;\n" +
	"\rMetadataEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x012\xba\x01\n" +
	"\rModuleService\x12[\n" +
	"\x10HandleConnection\x12\".gimpel.v1.HandleConnectionRequest\x1a#.gimpel.v1.HandleConnectionResponse\x12L\n" +
	"\vHealthCheck\x12\x1d.gimpel.v1.HealthCheckRequest\x1a\x1e.gimpel.v1.HealthCheckResponseB5Z3github.com/nohaxxjustlags/gimpel/api/go/v1;gimpelv1b\x06proto3"

var (
	file_v1_module_proto_rawDescOnce sync.Once
	file_v1_module_proto_rawDescData []byte
)

func file_v1_module_proto_rawDescGZIP() []byte {
	file_v1_module_proto_rawDescOnce.Do(func() {
		file_v1_module_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_v1_module_proto_rawDesc), len(file_v1_module_proto_rawDesc)))
	})
	return file_v1_module_proto_rawDescData
}

var file_v1_module_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_v1_module_proto_goTypes = []any{
	(*ConnectionInfo)(nil),
	(*HandleConnectionRequest)(nil),
	(*HandleConnectionResponse)(nil),
	(*HealthCheckRequest)(nil),
	(*HealthCheckResponse)(nil),
	nil,
}
var file_v1_module_proto_depIdxs = []int32{
	0,
	5,
	1,
	3,
	2,
	4,
	4,
	2,
	2,
	2,
	0,
}

func init() { file_v1_module_proto_init() }
func file_v1_module_proto_init() {
	if File_v1_module_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_v1_module_proto_rawDesc), len(file_v1_module_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_module_proto_goTypes,
		DependencyIndexes: file_v1_module_proto_depIdxs,
		MessageInfos:      file_v1_module_proto_msgTypes,
	}.Build()
	File_v1_module_proto = out.File
	file_v1_module_proto_goTypes = nil
	file_v1_module_proto_depIdxs = nil
}
