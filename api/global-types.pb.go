// Code generated by protoc-gen-go. DO NOT EDIT.
// source: global-types.proto

package api

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type Status_StatusValue int32

const (
	Status_NORMAL  Status_StatusValue = 0
	Status_PROBLEM Status_StatusValue = 1
)

var Status_StatusValue_name = map[int32]string{
	0: "NORMAL",
	1: "PROBLEM",
}

var Status_StatusValue_value = map[string]int32{
	"NORMAL":  0,
	"PROBLEM": 1,
}

func (x Status_StatusValue) String() string {
	return proto.EnumName(Status_StatusValue_name, int32(x))
}

func (Status_StatusValue) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3fa09321e03748b2, []int{2, 0}
}

//
// Authentication types used in multiple services
type UserCredentials struct {
	Email                string   `protobuf:"bytes,1,opt,name=email,proto3" json:"email,omitempty"`
	Password             string   `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
	InstanceId           string   `protobuf:"bytes,3,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UserCredentials) Reset()         { *m = UserCredentials{} }
func (m *UserCredentials) String() string { return proto.CompactTextString(m) }
func (*UserCredentials) ProtoMessage()    {}
func (*UserCredentials) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fa09321e03748b2, []int{0}
}

func (m *UserCredentials) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserCredentials.Unmarshal(m, b)
}
func (m *UserCredentials) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserCredentials.Marshal(b, m, deterministic)
}
func (m *UserCredentials) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserCredentials.Merge(m, src)
}
func (m *UserCredentials) XXX_Size() int {
	return xxx_messageInfo_UserCredentials.Size(m)
}
func (m *UserCredentials) XXX_DiscardUnknown() {
	xxx_messageInfo_UserCredentials.DiscardUnknown(m)
}

var xxx_messageInfo_UserCredentials proto.InternalMessageInfo

func (m *UserCredentials) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *UserCredentials) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *UserCredentials) GetInstanceId() string {
	if m != nil {
		return m.InstanceId
	}
	return ""
}

type TokenInfos struct {
	Id                   string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	InstanceId           string            `protobuf:"bytes,2,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	IssuedAt             int64             `protobuf:"varint,3,opt,name=issued_at,json=issuedAt,proto3" json:"issued_at,omitempty"`
	Payload              map[string]string `protobuf:"bytes,4,rep,name=payload,proto3" json:"payload,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *TokenInfos) Reset()         { *m = TokenInfos{} }
func (m *TokenInfos) String() string { return proto.CompactTextString(m) }
func (*TokenInfos) ProtoMessage()    {}
func (*TokenInfos) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fa09321e03748b2, []int{1}
}

func (m *TokenInfos) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TokenInfos.Unmarshal(m, b)
}
func (m *TokenInfos) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TokenInfos.Marshal(b, m, deterministic)
}
func (m *TokenInfos) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TokenInfos.Merge(m, src)
}
func (m *TokenInfos) XXX_Size() int {
	return xxx_messageInfo_TokenInfos.Size(m)
}
func (m *TokenInfos) XXX_DiscardUnknown() {
	xxx_messageInfo_TokenInfos.DiscardUnknown(m)
}

var xxx_messageInfo_TokenInfos proto.InternalMessageInfo

func (m *TokenInfos) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *TokenInfos) GetInstanceId() string {
	if m != nil {
		return m.InstanceId
	}
	return ""
}

func (m *TokenInfos) GetIssuedAt() int64 {
	if m != nil {
		return m.IssuedAt
	}
	return 0
}

func (m *TokenInfos) GetPayload() map[string]string {
	if m != nil {
		return m.Payload
	}
	return nil
}

//
// Status is typically used as a return value indicating if the method was performed normally, or the system has any internal error
// e.g. checking system status of a service
type Status struct {
	Status               Status_StatusValue `protobuf:"varint,1,opt,name=status,proto3,enum=inf.Status_StatusValue" json:"status,omitempty"`
	Msg                  string             `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Status) Reset()         { *m = Status{} }
func (m *Status) String() string { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()    {}
func (*Status) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fa09321e03748b2, []int{2}
}

func (m *Status) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Status.Unmarshal(m, b)
}
func (m *Status) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Status.Marshal(b, m, deterministic)
}
func (m *Status) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Status.Merge(m, src)
}
func (m *Status) XXX_Size() int {
	return xxx_messageInfo_Status.Size(m)
}
func (m *Status) XXX_DiscardUnknown() {
	xxx_messageInfo_Status.DiscardUnknown(m)
}

var xxx_messageInfo_Status proto.InternalMessageInfo

func (m *Status) GetStatus() Status_StatusValue {
	if m != nil {
		return m.Status
	}
	return Status_NORMAL
}

func (m *Status) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func init() {
	proto.RegisterEnum("inf.Status_StatusValue", Status_StatusValue_name, Status_StatusValue_value)
	proto.RegisterType((*UserCredentials)(nil), "inf.UserCredentials")
	proto.RegisterType((*TokenInfos)(nil), "inf.TokenInfos")
	proto.RegisterMapType((map[string]string)(nil), "inf.TokenInfos.PayloadEntry")
	proto.RegisterType((*Status)(nil), "inf.Status")
}

func init() { proto.RegisterFile("global-types.proto", fileDescriptor_3fa09321e03748b2) }

var fileDescriptor_3fa09321e03748b2 = []byte{
	// 320 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x51, 0x4d, 0x6b, 0xf2, 0x40,
	0x10, 0x7e, 0x93, 0xbc, 0x46, 0x9d, 0x14, 0x2b, 0x4b, 0xa1, 0xc1, 0x16, 0x2a, 0x39, 0x14, 0x2f,
	0x4d, 0xc1, 0x42, 0x29, 0xde, 0xb4, 0x78, 0x10, 0xb4, 0x4a, 0xfa, 0x71, 0xe8, 0x45, 0xd6, 0xee,
	0x2a, 0x8b, 0x71, 0x37, 0x64, 0x36, 0x2d, 0xf9, 0x93, 0xfd, 0x4d, 0x25, 0x9b, 0xd8, 0x4a, 0x7b,
	0xca, 0x3c, 0xf3, 0x7c, 0x0d, 0x59, 0x20, 0x9b, 0x58, 0xad, 0x68, 0x7c, 0xa5, 0xf3, 0x84, 0x63,
	0x98, 0xa4, 0x4a, 0x2b, 0xe2, 0x08, 0xb9, 0x0e, 0x18, 0x1c, 0x3f, 0x23, 0x4f, 0xef, 0x53, 0xce,
	0xb8, 0xd4, 0x82, 0xc6, 0x48, 0x4e, 0xa0, 0xc6, 0x77, 0x54, 0xc4, 0xbe, 0xd5, 0xb5, 0x7a, 0xcd,
	0xa8, 0x04, 0xa4, 0x03, 0x8d, 0x84, 0x22, 0x7e, 0xa8, 0x94, 0xf9, 0xb6, 0x21, 0xbe, 0x31, 0xb9,
	0x00, 0x4f, 0x48, 0xd4, 0x54, 0xbe, 0xf1, 0xa5, 0x60, 0xbe, 0x63, 0x68, 0xd8, 0xaf, 0x26, 0x2c,
	0xf8, 0xb4, 0x00, 0x9e, 0xd4, 0x96, 0xcb, 0x89, 0x5c, 0x2b, 0x24, 0x2d, 0xb0, 0x05, 0xab, 0xe2,
	0x6d, 0xf1, 0xc7, 0x6f, 0xff, 0xf6, 0x93, 0x33, 0x68, 0x0a, 0xc4, 0x8c, 0xb3, 0x25, 0xd5, 0x26,
	0xde, 0x89, 0x1a, 0xe5, 0x62, 0xa8, 0xc9, 0x2d, 0xd4, 0x13, 0x9a, 0xc7, 0x8a, 0x32, 0xff, 0x7f,
	0xd7, 0xe9, 0x79, 0xfd, 0xf3, 0x50, 0xc8, 0x75, 0xf8, 0xd3, 0x17, 0x2e, 0x4a, 0x7a, 0x2c, 0x75,
	0x9a, 0x47, 0x7b, 0x71, 0x67, 0x00, 0x47, 0x87, 0x04, 0x69, 0x83, 0xb3, 0xe5, 0x79, 0x75, 0x56,
	0x31, 0x16, 0x7f, 0xe2, 0x9d, 0xc6, 0x19, 0xaf, 0x2e, 0x2a, 0xc1, 0xc0, 0xbe, 0xb3, 0x02, 0x04,
	0xf7, 0x51, 0x53, 0x9d, 0x21, 0xb9, 0x06, 0x17, 0xcd, 0x64, 0x8c, 0xad, 0xfe, 0xa9, 0x29, 0x2f,
	0xc9, 0xea, 0xf3, 0x52, 0xd8, 0xa2, 0x4a, 0x56, 0xd4, 0xec, 0x70, 0x53, 0x45, 0x16, 0x63, 0x70,
	0x09, 0xde, 0x81, 0x90, 0x00, 0xb8, 0x0f, 0xf3, 0x68, 0x36, 0x9c, 0xb6, 0xff, 0x11, 0x0f, 0xea,
	0x8b, 0x68, 0x3e, 0x9a, 0x8e, 0x67, 0x6d, 0x6b, 0x54, 0x7b, 0x75, 0x68, 0x22, 0x56, 0xae, 0x79,
	0xbe, 0x9b, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x08, 0xf5, 0x4d, 0x59, 0xd4, 0x01, 0x00, 0x00,
}