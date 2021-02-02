// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: sms.proto

package sms

import (
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SmsProvider int32

const (
	SmsProvider_ONFON SmsProvider = 0
)

// Enum value maps for SmsProvider.
var (
	SmsProvider_name = map[int32]string{
		0: "ONFON",
	}
	SmsProvider_value = map[string]int32{
		"ONFON": 0,
	}
)

func (x SmsProvider) Enum() *SmsProvider {
	p := new(SmsProvider)
	*p = x
	return p
}

func (x SmsProvider) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SmsProvider) Descriptor() protoreflect.EnumDescriptor {
	return file_sms_proto_enumTypes[0].Descriptor()
}

func (SmsProvider) Type() protoreflect.EnumType {
	return &file_sms_proto_enumTypes[0]
}

func (x SmsProvider) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SmsProvider.Descriptor instead.
func (SmsProvider) EnumDescriptor() ([]byte, []int) {
	return file_sms_proto_rawDescGZIP(), []int{0}
}

type SMS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DestinationPhones []string `protobuf:"bytes,2,rep,name=destination_phones,json=destinationPhones,proto3" json:"destination_phones,omitempty"`
	Keyword           string   `protobuf:"bytes,1,opt,name=keyword,proto3" json:"keyword,omitempty"`
	Message           string   `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *SMS) Reset() {
	*x = SMS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sms_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SMS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SMS) ProtoMessage() {}

func (x *SMS) ProtoReflect() protoreflect.Message {
	mi := &file_sms_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SMS.ProtoReflect.Descriptor instead.
func (*SMS) Descriptor() ([]byte, []int) {
	return file_sms_proto_rawDescGZIP(), []int{0}
}

func (x *SMS) GetDestinationPhones() []string {
	if x != nil {
		return x.DestinationPhones
	}
	return nil
}

func (x *SMS) GetKeyword() string {
	if x != nil {
		return x.Keyword
	}
	return ""
}

func (x *SMS) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type SendSMSRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sms      *SMS                    `protobuf:"bytes,1,opt,name=sms,proto3" json:"sms,omitempty"`
	Auth     *SendSMSRequest_SMSAuth `protobuf:"bytes,2,opt,name=auth,proto3" json:"auth,omitempty"`
	Provider SmsProvider             `protobuf:"varint,3,opt,name=provider,proto3,enum=gidyon.apis.SmsProvider" json:"provider,omitempty"`
}

func (x *SendSMSRequest) Reset() {
	*x = SendSMSRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sms_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendSMSRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendSMSRequest) ProtoMessage() {}

func (x *SendSMSRequest) ProtoReflect() protoreflect.Message {
	mi := &file_sms_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendSMSRequest.ProtoReflect.Descriptor instead.
func (*SendSMSRequest) Descriptor() ([]byte, []int) {
	return file_sms_proto_rawDescGZIP(), []int{1}
}

func (x *SendSMSRequest) GetSms() *SMS {
	if x != nil {
		return x.Sms
	}
	return nil
}

func (x *SendSMSRequest) GetAuth() *SendSMSRequest_SMSAuth {
	if x != nil {
		return x.Auth
	}
	return nil
}

func (x *SendSMSRequest) GetProvider() SmsProvider {
	if x != nil {
		return x.Provider
	}
	return SmsProvider_ONFON
}

type SendSMSRequest_Cookie struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *SendSMSRequest_Cookie) Reset() {
	*x = SendSMSRequest_Cookie{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sms_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendSMSRequest_Cookie) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendSMSRequest_Cookie) ProtoMessage() {}

func (x *SendSMSRequest_Cookie) ProtoReflect() protoreflect.Message {
	mi := &file_sms_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendSMSRequest_Cookie.ProtoReflect.Descriptor instead.
func (*SendSMSRequest_Cookie) Descriptor() ([]byte, []int) {
	return file_sms_proto_rawDescGZIP(), []int{1, 0}
}

func (x *SendSMSRequest_Cookie) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SendSMSRequest_Cookie) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type SendSMSRequest_SMSAuth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// [sms_url, sender_id, api_key, client_id, auth_token, access_key, cookies]
	ApiUrl    string                   `protobuf:"bytes,1,opt,name=api_url,json=apiUrl,proto3" json:"api_url,omitempty"`
	SenderId  string                   `protobuf:"bytes,2,opt,name=sender_id,json=senderId,proto3" json:"sender_id,omitempty"`
	ApiKey    string                   `protobuf:"bytes,3,opt,name=api_key,json=apiKey,proto3" json:"api_key,omitempty"`
	ClientId  string                   `protobuf:"bytes,4,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	AuthToken string                   `protobuf:"bytes,5,opt,name=auth_token,json=authToken,proto3" json:"auth_token,omitempty"`
	AccessKey string                   `protobuf:"bytes,6,opt,name=access_key,json=accessKey,proto3" json:"access_key,omitempty"`
	Cookies   []*SendSMSRequest_Cookie `protobuf:"bytes,7,rep,name=cookies,proto3" json:"cookies,omitempty"`
}

func (x *SendSMSRequest_SMSAuth) Reset() {
	*x = SendSMSRequest_SMSAuth{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sms_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendSMSRequest_SMSAuth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendSMSRequest_SMSAuth) ProtoMessage() {}

func (x *SendSMSRequest_SMSAuth) ProtoReflect() protoreflect.Message {
	mi := &file_sms_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendSMSRequest_SMSAuth.ProtoReflect.Descriptor instead.
func (*SendSMSRequest_SMSAuth) Descriptor() ([]byte, []int) {
	return file_sms_proto_rawDescGZIP(), []int{1, 1}
}

func (x *SendSMSRequest_SMSAuth) GetApiUrl() string {
	if x != nil {
		return x.ApiUrl
	}
	return ""
}

func (x *SendSMSRequest_SMSAuth) GetSenderId() string {
	if x != nil {
		return x.SenderId
	}
	return ""
}

func (x *SendSMSRequest_SMSAuth) GetApiKey() string {
	if x != nil {
		return x.ApiKey
	}
	return ""
}

func (x *SendSMSRequest_SMSAuth) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

func (x *SendSMSRequest_SMSAuth) GetAuthToken() string {
	if x != nil {
		return x.AuthToken
	}
	return ""
}

func (x *SendSMSRequest_SMSAuth) GetAccessKey() string {
	if x != nil {
		return x.AccessKey
	}
	return ""
}

func (x *SendSMSRequest_SMSAuth) GetCookies() []*SendSMSRequest_Cookie {
	if x != nil {
		return x.Cookies
	}
	return nil
}

var File_sms_proto protoreflect.FileDescriptor

var file_sms_proto_rawDesc = []byte{
	0x0a, 0x09, 0x73, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x67, 0x69, 0x64,
	0x79, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x75,
	0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2c, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65,
	0x6e, 0x2d, 0x73, 0x77, 0x61, 0x67, 0x67, 0x65, 0x72, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xd4, 0x01, 0x0a, 0x03, 0x53, 0x4d, 0x53, 0x12, 0x32, 0x0a, 0x12, 0x64,
	0x65, 0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x68, 0x6f, 0x6e, 0x65,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52, 0x11, 0x64, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x68, 0x6f, 0x6e, 0x65, 0x73, 0x12,
	0x18, 0x0a, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x6b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x1d, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x03, 0xe0, 0x41, 0x02, 0x52,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x3a, 0x60, 0x92, 0x41, 0x5d, 0x0a, 0x5b, 0x2a,
	0x03, 0x53, 0x4d, 0x53, 0x32, 0x35, 0x53, 0x4d, 0x53, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x74,
	0x65, 0x78, 0x74, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x20, 0x74, 0x68, 0x61, 0x74,
	0x20, 0x69, 0x73, 0x20, 0x74, 0x6f, 0x20, 0x62, 0x65, 0x20, 0x73, 0x65, 0x6e, 0x74, 0x20, 0x74,
	0x6f, 0x20, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x28, 0x73, 0x29, 0xd2, 0x01, 0x12, 0x64, 0x65,
	0x73, 0x74, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x73,
	0xd2, 0x01, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x88, 0x04, 0x0a, 0x0e, 0x53,
	0x65, 0x6e, 0x64, 0x53, 0x4d, 0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a,
	0x03, 0x73, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x67, 0x69, 0x64,
	0x79, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x53, 0x4d, 0x53, 0x52, 0x03, 0x73, 0x6d,
	0x73, 0x12, 0x37, 0x0a, 0x04, 0x61, 0x75, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x23, 0x2e, 0x67, 0x69, 0x64, 0x79, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x53, 0x65,
	0x6e, 0x64, 0x53, 0x4d, 0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x4d, 0x53,
	0x41, 0x75, 0x74, 0x68, 0x52, 0x04, 0x61, 0x75, 0x74, 0x68, 0x12, 0x34, 0x0a, 0x08, 0x70, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x18, 0x2e, 0x67,
	0x69, 0x64, 0x79, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x53, 0x6d, 0x73, 0x50, 0x72,
	0x6f, 0x76, 0x69, 0x64, 0x65, 0x72, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x64, 0x65, 0x72,
	0x1a, 0x32, 0x0a, 0x06, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x1a, 0xf1, 0x01, 0x0a, 0x07, 0x53, 0x4d, 0x53, 0x41, 0x75, 0x74, 0x68,
	0x12, 0x17, 0x0a, 0x07, 0x61, 0x70, 0x69, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x61, 0x70, 0x69, 0x55, 0x72, 0x6c, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65, 0x6e,
	0x64, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65,
	0x6e, 0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x61, 0x70, 0x69, 0x5f, 0x6b, 0x65,
	0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x70, 0x69, 0x4b, 0x65, 0x79, 0x12,
	0x1b, 0x0a, 0x09, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a,
	0x61, 0x75, 0x74, 0x68, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x61, 0x75, 0x74, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x61,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x12, 0x3c, 0x0a, 0x07, 0x63, 0x6f,
	0x6f, 0x6b, 0x69, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x67, 0x69,
	0x64, 0x79, 0x6f, 0x6e, 0x2e, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x53, 0x4d,
	0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x52,
	0x07, 0x63, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x73, 0x3a, 0x3b, 0x92, 0x41, 0x38, 0x0a, 0x36, 0x2a,
	0x0e, 0x53, 0x65, 0x6e, 0x64, 0x53, 0x4d, 0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x32,
	0x1e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x20, 0x74, 0x6f, 0x20, 0x73, 0x65, 0x6e, 0x64,
	0x20, 0x73, 0x6d, 0x73, 0x20, 0x74, 0x6f, 0x20, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x73, 0xd2,
	0x01, 0x03, 0x73, 0x6d, 0x73, 0x2a, 0x18, 0x0a, 0x0b, 0x53, 0x6d, 0x73, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x64, 0x65, 0x72, 0x12, 0x09, 0x0a, 0x05, 0x4f, 0x4e, 0x46, 0x4f, 0x4e, 0x10, 0x00, 0x32,
	0x62, 0x0a, 0x06, 0x53, 0x4d, 0x53, 0x41, 0x50, 0x49, 0x12, 0x58, 0x0a, 0x07, 0x53, 0x65, 0x6e,
	0x64, 0x53, 0x4d, 0x53, 0x12, 0x1b, 0x2e, 0x67, 0x69, 0x64, 0x79, 0x6f, 0x6e, 0x2e, 0x61, 0x70,
	0x69, 0x73, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x53, 0x4d, 0x53, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x18, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x12, 0x12, 0x10, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x6d, 0x73, 0x3a, 0x73, 0x65, 0x6e, 0x64,
	0x53, 0x4d, 0x53, 0x42, 0xbd, 0x03, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x67, 0x69, 0x64, 0x79, 0x6f, 0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x69, 0x6e, 0x67, 0x2f, 0x73, 0x6d, 0x73, 0x92, 0x41, 0x87, 0x03, 0x12, 0xf3, 0x01, 0x0a,
	0x07, 0x53, 0x4d, 0x53, 0x20, 0x41, 0x50, 0x49, 0x12, 0x1f, 0x53, 0x65, 0x6e, 0x64, 0x20, 0x74,
	0x65, 0x78, 0x74, 0x20, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x20, 0x74, 0x6f, 0x20,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x28, 0x73, 0x29, 0x22, 0x7c, 0x0a, 0x15, 0x47, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x20, 0x3c, 0x47, 0x69, 0x64, 0x65, 0x6f, 0x6e, 0x20, 0x4b, 0x61, 0x6d, 0x61,
	0x75, 0x3e, 0x12, 0x4c, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x69, 0x64, 0x79, 0x6f, 0x6e, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x62, 0x6c, 0x6f, 0x62, 0x2f, 0x6d, 0x61, 0x73, 0x74,
	0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x69, 0x6e, 0x67, 0x2f, 0x73, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x15, 0x67, 0x69, 0x64, 0x65, 0x6f, 0x6e, 0x68, 0x61, 0x63, 0x65, 0x72, 0x40, 0x67, 0x6d,
	0x61, 0x69, 0x6c, 0x2e, 0x63, 0x6f, 0x6d, 0x2a, 0x45, 0x0a, 0x0b, 0x4d, 0x49, 0x54, 0x20, 0x4c,
	0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x12, 0x36, 0x68, 0x74, 0x74, 0x70, 0x73, 0x3a, 0x2f, 0x2f,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x69, 0x64, 0x79, 0x6f,
	0x6e, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x62, 0x6c, 0x6f, 0x62, 0x2f,
	0x6d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x2f, 0x4c, 0x49, 0x43, 0x45, 0x4e, 0x53, 0x45, 0x32, 0x02,
	0x76, 0x31, 0x2a, 0x02, 0x01, 0x02, 0x32, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x3a, 0x10, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x5a, 0x59, 0x0a, 0x57, 0x0a, 0x06,
	0x62, 0x65, 0x61, 0x72, 0x65, 0x72, 0x12, 0x4d, 0x08, 0x02, 0x12, 0x38, 0x41, 0x75, 0x74, 0x68,
	0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x74, 0x6f, 0x6b, 0x65, 0x6e,
	0x2c, 0x20, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x64, 0x20, 0x62, 0x79, 0x20, 0x42, 0x65,
	0x61, 0x72, 0x65, 0x72, 0x3a, 0x20, 0x42, 0x65, 0x61, 0x72, 0x65, 0x72, 0x20, 0x3c, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x3e, 0x1a, 0x0d, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x20, 0x02, 0x62, 0x0c, 0x0a, 0x0a, 0x0a, 0x06, 0x62, 0x65, 0x61, 0x72, 0x65,
	0x72, 0x12, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_sms_proto_rawDescOnce sync.Once
	file_sms_proto_rawDescData = file_sms_proto_rawDesc
)

func file_sms_proto_rawDescGZIP() []byte {
	file_sms_proto_rawDescOnce.Do(func() {
		file_sms_proto_rawDescData = protoimpl.X.CompressGZIP(file_sms_proto_rawDescData)
	})
	return file_sms_proto_rawDescData
}

var file_sms_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_sms_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_sms_proto_goTypes = []interface{}{
	(SmsProvider)(0),               // 0: gidyon.apis.SmsProvider
	(*SMS)(nil),                    // 1: gidyon.apis.SMS
	(*SendSMSRequest)(nil),         // 2: gidyon.apis.SendSMSRequest
	(*SendSMSRequest_Cookie)(nil),  // 3: gidyon.apis.SendSMSRequest.Cookie
	(*SendSMSRequest_SMSAuth)(nil), // 4: gidyon.apis.SendSMSRequest.SMSAuth
	(*empty.Empty)(nil),            // 5: google.protobuf.Empty
}
var file_sms_proto_depIdxs = []int32{
	1, // 0: gidyon.apis.SendSMSRequest.sms:type_name -> gidyon.apis.SMS
	4, // 1: gidyon.apis.SendSMSRequest.auth:type_name -> gidyon.apis.SendSMSRequest.SMSAuth
	0, // 2: gidyon.apis.SendSMSRequest.provider:type_name -> gidyon.apis.SmsProvider
	3, // 3: gidyon.apis.SendSMSRequest.SMSAuth.cookies:type_name -> gidyon.apis.SendSMSRequest.Cookie
	2, // 4: gidyon.apis.SMSAPI.SendSMS:input_type -> gidyon.apis.SendSMSRequest
	5, // 5: gidyon.apis.SMSAPI.SendSMS:output_type -> google.protobuf.Empty
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_sms_proto_init() }
func file_sms_proto_init() {
	if File_sms_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sms_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SMS); i {
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
		file_sms_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendSMSRequest); i {
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
		file_sms_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendSMSRequest_Cookie); i {
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
		file_sms_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendSMSRequest_SMSAuth); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_sms_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_sms_proto_goTypes,
		DependencyIndexes: file_sms_proto_depIdxs,
		EnumInfos:         file_sms_proto_enumTypes,
		MessageInfos:      file_sms_proto_msgTypes,
	}.Build()
	File_sms_proto = out.File
	file_sms_proto_rawDesc = nil
	file_sms_proto_goTypes = nil
	file_sms_proto_depIdxs = nil
}
