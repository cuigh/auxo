// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.21.2
// source: msg.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID      *uint64  `protobuf:"varint,1,opt,name=ID,proto3,oneof" json:"ID,omitempty"`
	Service *string  `protobuf:"bytes,2,opt,name=Service,proto3,oneof" json:"Service,omitempty"`
	Method  *string  `protobuf:"bytes,3,opt,name=Method,proto3,oneof" json:"Method,omitempty"`
	Labels  []*Label `protobuf:"bytes,4,rep,name=Labels,proto3" json:"Labels,omitempty"`
	Args    [][]byte `protobuf:"bytes,5,rep,name=Args,proto3" json:"Args,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{0}
}

func (x *Request) GetID() uint64 {
	if x != nil && x.ID != nil {
		return *x.ID
	}
	return 0
}

func (x *Request) GetService() string {
	if x != nil && x.Service != nil {
		return *x.Service
	}
	return ""
}

func (x *Request) GetMethod() string {
	if x != nil && x.Method != nil {
		return *x.Method
	}
	return ""
}

func (x *Request) GetLabels() []*Label {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *Request) GetArgs() [][]byte {
	if x != nil {
		return x.Args
	}
	return nil
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID     *uint64 `protobuf:"varint,1,opt,name=ID,proto3,oneof" json:"ID,omitempty"`
	Result []byte  `protobuf:"bytes,2,opt,name=Result,proto3,oneof" json:"Result,omitempty"`
	Error  *Error  `protobuf:"bytes,3,opt,name=Error,proto3,oneof" json:"Error,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{1}
}

func (x *Response) GetID() uint64 {
	if x != nil && x.ID != nil {
		return *x.ID
	}
	return 0
}

func (x *Response) GetResult() []byte {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *Response) GetError() *Error {
	if x != nil {
		return x.Error
	}
	return nil
}

type Label struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=Value,proto3" json:"Value,omitempty"`
}

func (x *Label) Reset() {
	*x = Label{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Label) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Label) ProtoMessage() {}

func (x *Label) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Label.ProtoReflect.Descriptor instead.
func (*Label) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{2}
}

func (x *Label) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Label) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code    *int32  `protobuf:"varint,1,opt,name=Code,proto3,oneof" json:"Code,omitempty"`
	Message *string `protobuf:"bytes,2,opt,name=Message,proto3,oneof" json:"Message,omitempty"`
	Detail  *string `protobuf:"bytes,3,opt,name=Detail,proto3,oneof" json:"Detail,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_msg_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_msg_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_msg_proto_rawDescGZIP(), []int{3}
}

func (x *Error) GetCode() int32 {
	if x != nil && x.Code != nil {
		return *x.Code
	}
	return 0
}

func (x *Error) GetMessage() string {
	if x != nil && x.Message != nil {
		return *x.Message
	}
	return ""
}

func (x *Error) GetDetail() string {
	if x != nil && x.Detail != nil {
		return *x.Detail
	}
	return ""
}

var File_msg_proto protoreflect.FileDescriptor

var file_msg_proto_rawDesc = []byte{
	0x0a, 0x09, 0x6d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xb2, 0x01, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x13,
	0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x48, 0x00, 0x52, 0x02, 0x49, 0x44,
	0x88, 0x01, 0x01, 0x12, 0x1d, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x88,
	0x01, 0x01, 0x12, 0x1b, 0x0a, 0x06, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x02, 0x52, 0x06, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x88, 0x01, 0x01, 0x12,
	0x24, 0x0a, 0x06, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4c, 0x61, 0x62, 0x65, 0x6c, 0x52, 0x06, 0x4c,
	0x61, 0x62, 0x65, 0x6c, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x41, 0x72, 0x67, 0x73, 0x18, 0x05, 0x20,
	0x03, 0x28, 0x0c, 0x52, 0x04, 0x41, 0x72, 0x67, 0x73, 0x42, 0x05, 0x0a, 0x03, 0x5f, 0x49, 0x44,
	0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x42, 0x09, 0x0a, 0x07,
	0x5f, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x22, 0x81, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x13, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x48, 0x00, 0x52, 0x02, 0x49, 0x44, 0x88, 0x01, 0x01, 0x12, 0x1b, 0x0a, 0x06, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x01, 0x52, 0x06, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x88, 0x01, 0x01, 0x12, 0x27, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x48, 0x02, 0x52, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x88, 0x01, 0x01, 0x42,
	0x05, 0x0a, 0x03, 0x5f, 0x49, 0x44, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x31, 0x0a, 0x05, 0x4c,
	0x61, 0x62, 0x65, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x7c,
	0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x17, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x88, 0x01, 0x01,
	0x12, 0x1d, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x01, 0x52, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x88, 0x01, 0x01, 0x12,
	0x1b, 0x0a, 0x06, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48,
	0x02, 0x52, 0x06, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x88, 0x01, 0x01, 0x42, 0x07, 0x0a, 0x05,
	0x5f, 0x43, 0x6f, 0x64, 0x65, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x42, 0x03, 0x5a, 0x01,
	0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_msg_proto_rawDescOnce sync.Once
	file_msg_proto_rawDescData = file_msg_proto_rawDesc
)

func file_msg_proto_rawDescGZIP() []byte {
	file_msg_proto_rawDescOnce.Do(func() {
		file_msg_proto_rawDescData = protoimpl.X.CompressGZIP(file_msg_proto_rawDescData)
	})
	return file_msg_proto_rawDescData
}

var file_msg_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_msg_proto_goTypes = []interface{}{
	(*Request)(nil),  // 0: proto.Request
	(*Response)(nil), // 1: proto.Response
	(*Label)(nil),    // 2: proto.Label
	(*Error)(nil),    // 3: proto.Error
}
var file_msg_proto_depIdxs = []int32{
	2, // 0: proto.Request.Labels:type_name -> proto.Label
	3, // 1: proto.Response.Error:type_name -> proto.Error
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_msg_proto_init() }
func file_msg_proto_init() {
	if File_msg_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_msg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_msg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_msg_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Label); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_msg_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_msg_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_msg_proto_msgTypes[1].OneofWrappers = []interface{}{}
	file_msg_proto_msgTypes[3].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_msg_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_msg_proto_goTypes,
		DependencyIndexes: file_msg_proto_depIdxs,
		MessageInfos:      file_msg_proto_msgTypes,
	}.Build()
	File_msg_proto = out.File
	file_msg_proto_rawDesc = nil
	file_msg_proto_goTypes = nil
	file_msg_proto_depIdxs = nil
}
