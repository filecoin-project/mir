//
//Copyright IBM Corp. All Rights Reserved.
//
//SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: messagepb/messagepb.proto

package messagepb

import (
	isspb "github.com/filecoin-project/mir/pkg/pb/isspb"
	requestpb "github.com/filecoin-project/mir/pkg/pb/requestpb"
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

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module string `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	// Types that are assignable to Type:
	//	*Message_Iss
	//	*Message_DummyPreprepare
	Type isMessage_Type `protobuf_oneof:"type"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_messagepb_messagepb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_messagepb_messagepb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_messagepb_messagepb_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (m *Message) GetType() isMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Message) GetIss() *isspb.ISSMessage {
	if x, ok := x.GetType().(*Message_Iss); ok {
		return x.Iss
	}
	return nil
}

func (x *Message) GetDummyPreprepare() *DummyPreprepare {
	if x, ok := x.GetType().(*Message_DummyPreprepare); ok {
		return x.DummyPreprepare
	}
	return nil
}

type isMessage_Type interface {
	isMessage_Type()
}

type Message_Iss struct {
	Iss *isspb.ISSMessage `protobuf:"bytes,2,opt,name=iss,proto3,oneof"`
}

type Message_DummyPreprepare struct {
	DummyPreprepare *DummyPreprepare `protobuf:"bytes,100,opt,name=dummy_preprepare,json=dummyPreprepare,proto3,oneof"`
}

func (*Message_Iss) isMessage_Type() {}

func (*Message_DummyPreprepare) isMessage_Type() {}

type DummyPreprepare struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sn    uint64           `protobuf:"varint,1,opt,name=sn,proto3" json:"sn,omitempty"`
	Batch *requestpb.Batch `protobuf:"bytes,2,opt,name=batch,proto3" json:"batch,omitempty"`
}

func (x *DummyPreprepare) Reset() {
	*x = DummyPreprepare{}
	if protoimpl.UnsafeEnabled {
		mi := &file_messagepb_messagepb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DummyPreprepare) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DummyPreprepare) ProtoMessage() {}

func (x *DummyPreprepare) ProtoReflect() protoreflect.Message {
	mi := &file_messagepb_messagepb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DummyPreprepare.ProtoReflect.Descriptor instead.
func (*DummyPreprepare) Descriptor() ([]byte, []int) {
	return file_messagepb_messagepb_proto_rawDescGZIP(), []int{1}
}

func (x *DummyPreprepare) GetSn() uint64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *DummyPreprepare) GetBatch() *requestpb.Batch {
	if x != nil {
		return x.Batch
	}
	return nil
}

var File_messagepb_messagepb_proto protoreflect.FileDescriptor

var file_messagepb_messagepb_proto_rawDesc = []byte{
	0x0a, 0x19, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x1a, 0x19, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x70,
	0x62, 0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x11, 0x69, 0x73, 0x73, 0x70, 0x62, 0x2f, 0x69, 0x73, 0x73, 0x70, 0x62, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x99, 0x01, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x25, 0x0a, 0x03, 0x69, 0x73, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x69, 0x73, 0x73, 0x70, 0x62, 0x2e, 0x49, 0x53,
	0x53, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x00, 0x52, 0x03, 0x69, 0x73, 0x73, 0x12,
	0x47, 0x0a, 0x10, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x5f, 0x70, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70,
	0x61, 0x72, 0x65, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x70, 0x62, 0x2e, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x50, 0x72, 0x65, 0x70, 0x72,
	0x65, 0x70, 0x61, 0x72, 0x65, 0x48, 0x00, 0x52, 0x0f, 0x64, 0x75, 0x6d, 0x6d, 0x79, 0x50, 0x72,
	0x65, 0x70, 0x72, 0x65, 0x70, 0x61, 0x72, 0x65, 0x42, 0x06, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x22, 0x49, 0x0a, 0x0f, 0x44, 0x75, 0x6d, 0x6d, 0x79, 0x50, 0x72, 0x65, 0x70, 0x72, 0x65, 0x70,
	0x61, 0x72, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x02, 0x73, 0x6e, 0x12, 0x26, 0x0a, 0x05, 0x62, 0x61, 0x74, 0x63, 0x68, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x10, 0x2e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x42,
	0x61, 0x74, 0x63, 0x68, 0x52, 0x05, 0x62, 0x61, 0x74, 0x63, 0x68, 0x42, 0x32, 0x5a, 0x30, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f,
	0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_messagepb_messagepb_proto_rawDescOnce sync.Once
	file_messagepb_messagepb_proto_rawDescData = file_messagepb_messagepb_proto_rawDesc
)

func file_messagepb_messagepb_proto_rawDescGZIP() []byte {
	file_messagepb_messagepb_proto_rawDescOnce.Do(func() {
		file_messagepb_messagepb_proto_rawDescData = protoimpl.X.CompressGZIP(file_messagepb_messagepb_proto_rawDescData)
	})
	return file_messagepb_messagepb_proto_rawDescData
}

var file_messagepb_messagepb_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_messagepb_messagepb_proto_goTypes = []interface{}{
	(*Message)(nil),          // 0: messagepb.Message
	(*DummyPreprepare)(nil),  // 1: messagepb.DummyPreprepare
	(*isspb.ISSMessage)(nil), // 2: isspb.ISSMessage
	(*requestpb.Batch)(nil),  // 3: requestpb.Batch
}
var file_messagepb_messagepb_proto_depIdxs = []int32{
	2, // 0: messagepb.Message.iss:type_name -> isspb.ISSMessage
	1, // 1: messagepb.Message.dummy_preprepare:type_name -> messagepb.DummyPreprepare
	3, // 2: messagepb.DummyPreprepare.batch:type_name -> requestpb.Batch
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_messagepb_messagepb_proto_init() }
func file_messagepb_messagepb_proto_init() {
	if File_messagepb_messagepb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_messagepb_messagepb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
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
		file_messagepb_messagepb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DummyPreprepare); i {
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
	file_messagepb_messagepb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Message_Iss)(nil),
		(*Message_DummyPreprepare)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_messagepb_messagepb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_messagepb_messagepb_proto_goTypes,
		DependencyIndexes: file_messagepb_messagepb_proto_depIdxs,
		MessageInfos:      file_messagepb_messagepb_proto_msgTypes,
	}.Build()
	File_messagepb_messagepb_proto = out.File
	file_messagepb_messagepb_proto_rawDesc = nil
	file_messagepb_messagepb_proto_goTypes = nil
	file_messagepb_messagepb_proto_depIdxs = nil
}
