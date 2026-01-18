
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

type ModuleImage struct {
	state       protoimpl.MessageState `protogen:"open.v1"`
	Id          string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Version     string                 `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	Description string                 `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	ImageRef  string `protobuf:"bytes,5,opt,name=image_ref,json=imageRef,proto3" json:"image_ref,omitempty"`
	Digest    string `protobuf:"bytes,6,opt,name=digest,proto3" json:"digest,omitempty"`
	SizeBytes int64  `protobuf:"varint,7,opt,name=size_bytes,json=sizeBytes,proto3" json:"size_bytes,omitempty"`
	Signature []byte `protobuf:"bytes,8,opt,name=signature,proto3" json:"signature,omitempty"`
	SignedBy  string `protobuf:"bytes,9,opt,name=signed_by,json=signedBy,proto3" json:"signed_by,omitempty"`
	SignedAt  int64  `protobuf:"varint,10,opt,name=signed_at,json=signedAt,proto3" json:"signed_at,omitempty"`
	RequiredCapabilities []string `protobuf:"bytes,11,rep,name=required_capabilities,json=requiredCapabilities,proto3" json:"required_capabilities,omitempty"`
	RequiresPrivileged   bool     `protobuf:"varint,12,opt,name=requires_privileged,json=requiresPrivileged,proto3" json:"requires_privileged,omitempty"`
	MinAgentVersion      string   `protobuf:"bytes,13,opt,name=min_agent_version,json=minAgentVersion,proto3" json:"min_agent_version,omitempty"`
	Protocols []*ModuleProtocol `protobuf:"bytes,14,rep,name=protocols,proto3" json:"protocols,omitempty"`
	Resources *ResourceRequirements `protobuf:"bytes,15,opt,name=resources,proto3" json:"resources,omitempty"`
	Labels        map[string]string `protobuf:"bytes,16,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	CreatedAt     int64             `protobuf:"varint,17,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt     int64             `protobuf:"varint,18,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ModuleImage) Reset() {
	*x = ModuleImage{}
	mi := &file_v1_module_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleImage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleImage) ProtoMessage() {}

func (x *ModuleImage) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ModuleImage) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{5}
}

func (x *ModuleImage) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ModuleImage) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ModuleImage) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *ModuleImage) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *ModuleImage) GetImageRef() string {
	if x != nil {
		return x.ImageRef
	}
	return ""
}

func (x *ModuleImage) GetDigest() string {
	if x != nil {
		return x.Digest
	}
	return ""
}

func (x *ModuleImage) GetSizeBytes() int64 {
	if x != nil {
		return x.SizeBytes
	}
	return 0
}

func (x *ModuleImage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *ModuleImage) GetSignedBy() string {
	if x != nil {
		return x.SignedBy
	}
	return ""
}

func (x *ModuleImage) GetSignedAt() int64 {
	if x != nil {
		return x.SignedAt
	}
	return 0
}

func (x *ModuleImage) GetRequiredCapabilities() []string {
	if x != nil {
		return x.RequiredCapabilities
	}
	return nil
}

func (x *ModuleImage) GetRequiresPrivileged() bool {
	if x != nil {
		return x.RequiresPrivileged
	}
	return false
}

func (x *ModuleImage) GetMinAgentVersion() string {
	if x != nil {
		return x.MinAgentVersion
	}
	return ""
}

func (x *ModuleImage) GetProtocols() []*ModuleProtocol {
	if x != nil {
		return x.Protocols
	}
	return nil
}

func (x *ModuleImage) GetResources() *ResourceRequirements {
	if x != nil {
		return x.Resources
	}
	return nil
}

func (x *ModuleImage) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *ModuleImage) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *ModuleImage) GetUpdatedAt() int64 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

type ModuleProtocol struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	Name            string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	DefaultPort     uint32                 `protobuf:"varint,2,opt,name=default_port,json=defaultPort,proto3" json:"default_port,omitempty"`
	HighInteraction bool                   `protobuf:"varint,3,opt,name=high_interaction,json=highInteraction,proto3" json:"high_interaction,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *ModuleProtocol) Reset() {
	*x = ModuleProtocol{}
	mi := &file_v1_module_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleProtocol) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleProtocol) ProtoMessage() {}

func (x *ModuleProtocol) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ModuleProtocol) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{6}
}

func (x *ModuleProtocol) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ModuleProtocol) GetDefaultPort() uint32 {
	if x != nil {
		return x.DefaultPort
	}
	return 0
}

func (x *ModuleProtocol) GetHighInteraction() bool {
	if x != nil {
		return x.HighInteraction
	}
	return false
}

type ResourceRequirements struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	MemoryMb      int64                  `protobuf:"varint,1,opt,name=memory_mb,json=memoryMb,proto3" json:"memory_mb,omitempty"`
	CpuMillicores int32                  `protobuf:"varint,2,opt,name=cpu_millicores,json=cpuMillicores,proto3" json:"cpu_millicores,omitempty"`
	DiskMb        int64                  `protobuf:"varint,3,opt,name=disk_mb,json=diskMb,proto3" json:"disk_mb,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ResourceRequirements) Reset() {
	*x = ResourceRequirements{}
	mi := &file_v1_module_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ResourceRequirements) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceRequirements) ProtoMessage() {}

func (x *ResourceRequirements) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ResourceRequirements) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{7}
}

func (x *ResourceRequirements) GetMemoryMb() int64 {
	if x != nil {
		return x.MemoryMb
	}
	return 0
}

func (x *ResourceRequirements) GetCpuMillicores() int32 {
	if x != nil {
		return x.CpuMillicores
	}
	return 0
}

func (x *ResourceRequirements) GetDiskMb() int64 {
	if x != nil {
		return x.DiskMb
	}
	return 0
}

type ModuleCatalog struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Modules       []*ModuleImage         `protobuf:"bytes,1,rep,name=modules,proto3" json:"modules,omitempty"`
	Version       int64                  `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
	UpdatedAt     int64                  `protobuf:"varint,3,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	Signature     []byte                 `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
	SignedBy      string                 `protobuf:"bytes,5,opt,name=signed_by,json=signedBy,proto3" json:"signed_by,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ModuleCatalog) Reset() {
	*x = ModuleCatalog{}
	mi := &file_v1_module_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleCatalog) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleCatalog) ProtoMessage() {}

func (x *ModuleCatalog) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ModuleCatalog) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{8}
}

func (x *ModuleCatalog) GetModules() []*ModuleImage {
	if x != nil {
		return x.Modules
	}
	return nil
}

func (x *ModuleCatalog) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *ModuleCatalog) GetUpdatedAt() int64 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

func (x *ModuleCatalog) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *ModuleCatalog) GetSignedBy() string {
	if x != nil {
		return x.SignedBy
	}
	return ""
}

type ModuleAssignment struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	ModuleId string                 `protobuf:"bytes,1,opt,name=module_id,json=moduleId,proto3" json:"module_id,omitempty"`
	Version  string                 `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Listeners []*ListenerAssignment `protobuf:"bytes,3,rep,name=listeners,proto3" json:"listeners,omitempty"`
	Env map[string]string `protobuf:"bytes,4,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	ResourceOverrides *ResourceRequirements `protobuf:"bytes,5,opt,name=resource_overrides,json=resourceOverrides,proto3" json:"resource_overrides,omitempty"`
	ExecutionMode  string `protobuf:"bytes,6,opt,name=execution_mode,json=executionMode,proto3" json:"execution_mode,omitempty"`
	ConnectionMode string `protobuf:"bytes,7,opt,name=connection_mode,json=connectionMode,proto3" json:"connection_mode,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ModuleAssignment) Reset() {
	*x = ModuleAssignment{}
	mi := &file_v1_module_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleAssignment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleAssignment) ProtoMessage() {}

func (x *ModuleAssignment) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ModuleAssignment) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{9}
}

func (x *ModuleAssignment) GetModuleId() string {
	if x != nil {
		return x.ModuleId
	}
	return ""
}

func (x *ModuleAssignment) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *ModuleAssignment) GetListeners() []*ListenerAssignment {
	if x != nil {
		return x.Listeners
	}
	return nil
}

func (x *ModuleAssignment) GetEnv() map[string]string {
	if x != nil {
		return x.Env
	}
	return nil
}

func (x *ModuleAssignment) GetResourceOverrides() *ResourceRequirements {
	if x != nil {
		return x.ResourceOverrides
	}
	return nil
}

func (x *ModuleAssignment) GetExecutionMode() string {
	if x != nil {
		return x.ExecutionMode
	}
	return ""
}

func (x *ModuleAssignment) GetConnectionMode() string {
	if x != nil {
		return x.ConnectionMode
	}
	return ""
}

type ListenerAssignment struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	Id              string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Protocol        string                 `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Port            uint32                 `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	HighInteraction bool                   `protobuf:"varint,4,opt,name=high_interaction,json=highInteraction,proto3" json:"high_interaction,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *ListenerAssignment) Reset() {
	*x = ListenerAssignment{}
	mi := &file_v1_module_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListenerAssignment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListenerAssignment) ProtoMessage() {}

func (x *ListenerAssignment) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ListenerAssignment) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{10}
}

func (x *ListenerAssignment) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ListenerAssignment) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *ListenerAssignment) GetPort() uint32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *ListenerAssignment) GetHighInteraction() bool {
	if x != nil {
		return x.HighInteraction
	}
	return false
}

type AgentModuleConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	AgentId       string                 `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
	Assignments   []*ModuleAssignment    `protobuf:"bytes,2,rep,name=assignments,proto3" json:"assignments,omitempty"`
	Version       int64                  `protobuf:"varint,3,opt,name=version,proto3" json:"version,omitempty"`
	Signature     []byte                 `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AgentModuleConfig) Reset() {
	*x = AgentModuleConfig{}
	mi := &file_v1_module_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AgentModuleConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AgentModuleConfig) ProtoMessage() {}

func (x *AgentModuleConfig) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*AgentModuleConfig) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{11}
}

func (x *AgentModuleConfig) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

func (x *AgentModuleConfig) GetAssignments() []*ModuleAssignment {
	if x != nil {
		return x.Assignments
	}
	return nil
}

func (x *AgentModuleConfig) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *AgentModuleConfig) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type GetCatalogRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	CurrentVersion int64                  `protobuf:"varint,1,opt,name=current_version,json=currentVersion,proto3" json:"current_version,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *GetCatalogRequest) Reset() {
	*x = GetCatalogRequest{}
	mi := &file_v1_module_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCatalogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCatalogRequest) ProtoMessage() {}

func (x *GetCatalogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetCatalogRequest) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{12}
}

func (x *GetCatalogRequest) GetCurrentVersion() int64 {
	if x != nil {
		return x.CurrentVersion
	}
	return 0
}

type GetCatalogResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Updated       bool                   `protobuf:"varint,1,opt,name=updated,proto3" json:"updated,omitempty"`
	Catalog       *ModuleCatalog         `protobuf:"bytes,2,opt,name=catalog,proto3" json:"catalog,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetCatalogResponse) Reset() {
	*x = GetCatalogResponse{}
	mi := &file_v1_module_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCatalogResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCatalogResponse) ProtoMessage() {}

func (x *GetCatalogResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetCatalogResponse) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{13}
}

func (x *GetCatalogResponse) GetUpdated() bool {
	if x != nil {
		return x.Updated
	}
	return false
}

func (x *GetCatalogResponse) GetCatalog() *ModuleCatalog {
	if x != nil {
		return x.Catalog
	}
	return nil
}

type GetModuleAssignmentsRequest struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	AgentId        string                 `protobuf:"bytes,1,opt,name=agent_id,json=agentId,proto3" json:"agent_id,omitempty"`
	CurrentVersion int64                  `protobuf:"varint,2,opt,name=current_version,json=currentVersion,proto3" json:"current_version,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *GetModuleAssignmentsRequest) Reset() {
	*x = GetModuleAssignmentsRequest{}
	mi := &file_v1_module_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetModuleAssignmentsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetModuleAssignmentsRequest) ProtoMessage() {}

func (x *GetModuleAssignmentsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetModuleAssignmentsRequest) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{14}
}

func (x *GetModuleAssignmentsRequest) GetAgentId() string {
	if x != nil {
		return x.AgentId
	}
	return ""
}

func (x *GetModuleAssignmentsRequest) GetCurrentVersion() int64 {
	if x != nil {
		return x.CurrentVersion
	}
	return 0
}

type GetModuleAssignmentsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Updated       bool                   `protobuf:"varint,1,opt,name=updated,proto3" json:"updated,omitempty"`
	Config        *AgentModuleConfig     `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetModuleAssignmentsResponse) Reset() {
	*x = GetModuleAssignmentsResponse{}
	mi := &file_v1_module_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetModuleAssignmentsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetModuleAssignmentsResponse) ProtoMessage() {}

func (x *GetModuleAssignmentsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetModuleAssignmentsResponse) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{15}
}

func (x *GetModuleAssignmentsResponse) GetUpdated() bool {
	if x != nil {
		return x.Updated
	}
	return false
}

func (x *GetModuleAssignmentsResponse) GetConfig() *AgentModuleConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type DownloadModuleRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ModuleId      string                 `protobuf:"bytes,1,opt,name=module_id,json=moduleId,proto3" json:"module_id,omitempty"`
	Version       string                 `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DownloadModuleRequest) Reset() {
	*x = DownloadModuleRequest{}
	mi := &file_v1_module_proto_msgTypes[16]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DownloadModuleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadModuleRequest) ProtoMessage() {}

func (x *DownloadModuleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[16]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*DownloadModuleRequest) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{16}
}

func (x *DownloadModuleRequest) GetModuleId() string {
	if x != nil {
		return x.ModuleId
	}
	return ""
}

func (x *DownloadModuleRequest) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

type ModuleImageChunk struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Data          []byte                 `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Offset        int64                  `protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	TotalSize     int64                  `protobuf:"varint,3,opt,name=total_size,json=totalSize,proto3" json:"total_size,omitempty"`
	IsLast        bool                   `protobuf:"varint,4,opt,name=is_last,json=isLast,proto3" json:"is_last,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ModuleImageChunk) Reset() {
	*x = ModuleImageChunk{}
	mi := &file_v1_module_proto_msgTypes[17]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleImageChunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleImageChunk) ProtoMessage() {}

func (x *ModuleImageChunk) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[17]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ModuleImageChunk) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{17}
}

func (x *ModuleImageChunk) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *ModuleImageChunk) GetOffset() int64 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *ModuleImageChunk) GetTotalSize() int64 {
	if x != nil {
		return x.TotalSize
	}
	return 0
}

func (x *ModuleImageChunk) GetIsLast() bool {
	if x != nil {
		return x.IsLast
	}
	return false
}

type VerifyModuleRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ModuleId      string                 `protobuf:"bytes,1,opt,name=module_id,json=moduleId,proto3" json:"module_id,omitempty"`
	Version       string                 `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	Digest        string                 `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VerifyModuleRequest) Reset() {
	*x = VerifyModuleRequest{}
	mi := &file_v1_module_proto_msgTypes[18]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VerifyModuleRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifyModuleRequest) ProtoMessage() {}

func (x *VerifyModuleRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[18]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*VerifyModuleRequest) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{18}
}

func (x *VerifyModuleRequest) GetModuleId() string {
	if x != nil {
		return x.ModuleId
	}
	return ""
}

func (x *VerifyModuleRequest) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *VerifyModuleRequest) GetDigest() string {
	if x != nil {
		return x.Digest
	}
	return ""
}

type VerifyModuleResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Valid         bool                   `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
	Signature     []byte                 `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *VerifyModuleResponse) Reset() {
	*x = VerifyModuleResponse{}
	mi := &file_v1_module_proto_msgTypes[19]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *VerifyModuleResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerifyModuleResponse) ProtoMessage() {}

func (x *VerifyModuleResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_module_proto_msgTypes[19]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*VerifyModuleResponse) Descriptor() ([]byte, []int) {
	return file_v1_module_proto_rawDescGZIP(), []int{19}
}

func (x *VerifyModuleResponse) GetValid() bool {
	if x != nil {
		return x.Valid
	}
	return false
}

func (x *VerifyModuleResponse) GetSignature() []byte {
	if x != nil {
		return x.Signature
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
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\xd8\x05\n" +
	"\vModuleImage\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x12\x18\n" +
	"\aversion\x18\x03 \x01(\tR\aversion\x12 \n" +
	"\vdescription\x18\x04 \x01(\tR\vdescription\x12\x1b\n" +
	"\timage_ref\x18\x05 \x01(\tR\bimageRef\x12\x16\n" +
	"\x06digest\x18\x06 \x01(\tR\x06digest\x12\x1d\n" +
	"\n" +
	"size_bytes\x18\a \x01(\x03R\tsizeBytes\x12\x1c\n" +
	"\tsignature\x18\b \x01(\fR\tsignature\x12\x1b\n" +
	"\tsigned_by\x18\t \x01(\tR\bsignedBy\x12\x1b\n" +
	"\tsigned_at\x18\n" +
	" \x01(\x03R\bsignedAt\x123\n" +
	"\x15required_capabilities\x18\v \x03(\tR\x14requiredCapabilities\x12/\n" +
	"\x13requires_privileged\x18\f \x01(\bR\x12requiresPrivileged\x12*\n" +
	"\x11min_agent_version\x18\r \x01(\tR\x0fminAgentVersion\x127\n" +
	"\tprotocols\x18\x0e \x03(\v2\x19.gimpel.v1.ModuleProtocolR\tprotocols\x12=\n" +
	"\tresources\x18\x0f \x01(\v2\x1f.gimpel.v1.ResourceRequirementsR\tresources\x12:\n" +
	"\x06labels\x18\x10 \x03(\v2\".gimpel.v1.ModuleImage.LabelsEntryR\x06labels\x12\x1d\n" +
	"\n" +
	"created_at\x18\x11 \x01(\x03R\tcreatedAt\x12\x1d\n" +
	"\n" +
	"updated_at\x18\x12 \x01(\x03R\tupdatedAt\x1a9\n" +
	"\vLabelsEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"r\n" +
	"\x0eModuleProtocol\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12!\n" +
	"\fdefault_port\x18\x02 \x01(\rR\vdefaultPort\x12)\n" +
	"\x10high_interaction\x18\x03 \x01(\bR\x0fhighInteraction\"s\n" +
	"\x14ResourceRequirements\x12\x1b\n" +
	"\tmemory_mb\x18\x01 \x01(\x03R\bmemoryMb\x12%\n" +
	"\x0ecpu_millicores\x18\x02 \x01(\x05R\rcpuMillicores\x12\x17\n" +
	"\adisk_mb\x18\x03 \x01(\x03R\x06diskMb\"\xb5\x01\n" +
	"\rModuleCatalog\x120\n" +
	"\amodules\x18\x01 \x03(\v2\x16.gimpel.v1.ModuleImageR\amodules\x12\x18\n" +
	"\aversion\x18\x02 \x01(\x03R\aversion\x12\x1d\n" +
	"\n" +
	"updated_at\x18\x03 \x01(\x03R\tupdatedAt\x12\x1c\n" +
	"\tsignature\x18\x04 \x01(\fR\tsignature\x12\x1b\n" +
	"\tsigned_by\x18\x05 \x01(\tR\bsignedBy\"\x96\x03\n" +
	"\x10ModuleAssignment\x12\x1b\n" +
	"\tmodule_id\x18\x01 \x01(\tR\bmoduleId\x12\x18\n" +
	"\aversion\x18\x02 \x01(\tR\aversion\x12;\n" +
	"\tlisteners\x18\x03 \x03(\v2\x1d.gimpel.v1.ListenerAssignmentR\tlisteners\x126\n" +
	"\x03env\x18\x04 \x03(\v2$.gimpel.v1.ModuleAssignment.EnvEntryR\x03env\x12N\n" +
	"\x12resource_overrides\x18\x05 \x01(\v2\x1f.gimpel.v1.ResourceRequirementsR\x11resourceOverrides\x12%\n" +
	"\x0eexecution_mode\x18\x06 \x01(\tR\rexecutionMode\x12'\n" +
	"\x0fconnection_mode\x18\a \x01(\tR\x0econnectionMode\x1a6\n" +
	"\bEnvEntry\x12\x10\n" +
	"\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value:\x028\x01\"\x7f\n" +
	"\x12ListenerAssignment\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1a\n" +
	"\bprotocol\x18\x02 \x01(\tR\bprotocol\x12\x12\n" +
	"\x04port\x18\x03 \x01(\rR\x04port\x12)\n" +
	"\x10high_interaction\x18\x04 \x01(\bR\x0fhighInteraction\"\xa5\x01\n" +
	"\x11AgentModuleConfig\x12\x19\n" +
	"\bagent_id\x18\x01 \x01(\tR\aagentId\x12=\n" +
	"\vassignments\x18\x02 \x03(\v2\x1b.gimpel.v1.ModuleAssignmentR\vassignments\x12\x18\n" +
	"\aversion\x18\x03 \x01(\x03R\aversion\x12\x1c\n" +
	"\tsignature\x18\x04 \x01(\fR\tsignature\"<\n" +
	"\x11GetCatalogRequest\x12'\n" +
	"\x0fcurrent_version\x18\x01 \x01(\x03R\x0ecurrentVersion\"b\n" +
	"\x12GetCatalogResponse\x12\x18\n" +
	"\aupdated\x18\x01 \x01(\bR\aupdated\x122\n" +
	"\acatalog\x18\x02 \x01(\v2\x18.gimpel.v1.ModuleCatalogR\acatalog\"a\n" +
	"\x1bGetModuleAssignmentsRequest\x12\x19\n" +
	"\bagent_id\x18\x01 \x01(\tR\aagentId\x12'\n" +
	"\x0fcurrent_version\x18\x02 \x01(\x03R\x0ecurrentVersion\"n\n" +
	"\x1cGetModuleAssignmentsResponse\x12\x18\n" +
	"\aupdated\x18\x01 \x01(\bR\aupdated\x124\n" +
	"\x06config\x18\x02 \x01(\v2\x1c.gimpel.v1.AgentModuleConfigR\x06config\"N\n" +
	"\x15DownloadModuleRequest\x12\x1b\n" +
	"\tmodule_id\x18\x01 \x01(\tR\bmoduleId\x12\x18\n" +
	"\aversion\x18\x02 \x01(\tR\aversion\"v\n" +
	"\x10ModuleImageChunk\x12\x12\n" +
	"\x04data\x18\x01 \x01(\fR\x04data\x12\x16\n" +
	"\x06offset\x18\x02 \x01(\x03R\x06offset\x12\x1d\n" +
	"\n" +
	"total_size\x18\x03 \x01(\x03R\ttotalSize\x12\x17\n" +
	"\ais_last\x18\x04 \x01(\bR\x06isLast\"d\n" +
	"\x13VerifyModuleRequest\x12\x1b\n" +
	"\tmodule_id\x18\x01 \x01(\tR\bmoduleId\x12\x18\n" +
	"\aversion\x18\x02 \x01(\tR\aversion\x12\x16\n" +
	"\x06digest\x18\x03 \x01(\tR\x06digest\"J\n" +
	"\x14VerifyModuleResponse\x12\x14\n" +
	"\x05valid\x18\x01 \x01(\bR\x05valid\x12\x1c\n" +
	"\tsignature\x18\x02 \x01(\fR\tsignature2\xba\x01\n" +
	"\rModuleService\x12[\n" +
	"\x10HandleConnection\x12\".gimpel.v1.HandleConnectionRequest\x1a#.gimpel.v1.HandleConnectionResponse\x12L\n" +
	"\vHealthCheck\x12\x1d.gimpel.v1.HealthCheckRequest\x1a\x1e.gimpel.v1.HealthCheckResponse2\xee\x02\n" +
	"\x14ModuleCatalogService\x12I\n" +
	"\n" +
	"GetCatalog\x12\x1c.gimpel.v1.GetCatalogRequest\x1a\x1d.gimpel.v1.GetCatalogResponse\x12g\n" +
	"\x14GetModuleAssignments\x12&.gimpel.v1.GetModuleAssignmentsRequest\x1a'.gimpel.v1.GetModuleAssignmentsResponse\x12Q\n" +
	"\x0eDownloadModule\x12 .gimpel.v1.DownloadModuleRequest\x1a\x1b.gimpel.v1.ModuleImageChunk0\x01\x12O\n" +
	"\fVerifyModule\x12\x1e.gimpel.v1.VerifyModuleRequest\x1a\x1f.gimpel.v1.VerifyModuleResponseB5Z3github.com/nohaxxjustlags/gimpel/api/go/v1;gimpelv1b\x06proto3"

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

var file_v1_module_proto_msgTypes = make([]protoimpl.MessageInfo, 23)
var file_v1_module_proto_goTypes = []any{
	(*ConnectionInfo)(nil),
	(*HandleConnectionRequest)(nil),
	(*HandleConnectionResponse)(nil),
	(*HealthCheckRequest)(nil),
	(*HealthCheckResponse)(nil),
	(*ModuleImage)(nil),
	(*ModuleProtocol)(nil),
	(*ResourceRequirements)(nil),
	(*ModuleCatalog)(nil),
	(*ModuleAssignment)(nil),
	(*ListenerAssignment)(nil),
	(*AgentModuleConfig)(nil),
	(*GetCatalogRequest)(nil),
	(*GetCatalogResponse)(nil),
	(*GetModuleAssignmentsRequest)(nil),
	(*GetModuleAssignmentsResponse)(nil),
	(*DownloadModuleRequest)(nil),
	(*ModuleImageChunk)(nil),
	(*VerifyModuleRequest)(nil),
	(*VerifyModuleResponse)(nil),
	nil,
	nil,
	nil,
}
var file_v1_module_proto_depIdxs = []int32{
	0,
	20,
	6,
	7,
	21,
	5,
	10,
	22,
	7,
	9,
	8,
	11,
	1,
	3,
	12,
	14,
	16,
	18,
	2,
	4,
	13,
	15,
	17,
	19,
	18,
	12,
	12,
	12,
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
			NumMessages:   23,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_v1_module_proto_goTypes,
		DependencyIndexes: file_v1_module_proto_depIdxs,
		MessageInfos:      file_v1_module_proto_msgTypes,
	}.Build()
	File_v1_module_proto = out.File
	file_v1_module_proto_goTypes = nil
	file_v1_module_proto_depIdxs = nil
}
