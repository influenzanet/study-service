// Code generated by protoc-gen-go. DO NOT EDIT.
// source: study-service-api.proto

package api

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
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

type SurveyAndContext struct {
	Survey               *Survey        `protobuf:"bytes,1,opt,name=survey,proto3" json:"survey,omitempty"`
	Context              *SurveyContext `protobuf:"bytes,2,opt,name=context,proto3" json:"context,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *SurveyAndContext) Reset()         { *m = SurveyAndContext{} }
func (m *SurveyAndContext) String() string { return proto.CompactTextString(m) }
func (*SurveyAndContext) ProtoMessage()    {}
func (*SurveyAndContext) Descriptor() ([]byte, []int) {
	return fileDescriptor_81f0d8f98f9be15c, []int{0}
}

func (m *SurveyAndContext) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SurveyAndContext.Unmarshal(m, b)
}
func (m *SurveyAndContext) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SurveyAndContext.Marshal(b, m, deterministic)
}
func (m *SurveyAndContext) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SurveyAndContext.Merge(m, src)
}
func (m *SurveyAndContext) XXX_Size() int {
	return xxx_messageInfo_SurveyAndContext.Size(m)
}
func (m *SurveyAndContext) XXX_DiscardUnknown() {
	xxx_messageInfo_SurveyAndContext.DiscardUnknown(m)
}

var xxx_messageInfo_SurveyAndContext proto.InternalMessageInfo

func (m *SurveyAndContext) GetSurvey() *Survey {
	if m != nil {
		return m.Survey
	}
	return nil
}

func (m *SurveyAndContext) GetContext() *SurveyContext {
	if m != nil {
		return m.Context
	}
	return nil
}

type CreateSurveyReq struct {
	Token                *TokenInfos `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	SurveyDef            *SurveyItem `protobuf:"bytes,2,opt,name=surveyDef,proto3" json:"surveyDef,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *CreateSurveyReq) Reset()         { *m = CreateSurveyReq{} }
func (m *CreateSurveyReq) String() string { return proto.CompactTextString(m) }
func (*CreateSurveyReq) ProtoMessage()    {}
func (*CreateSurveyReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_81f0d8f98f9be15c, []int{1}
}

func (m *CreateSurveyReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateSurveyReq.Unmarshal(m, b)
}
func (m *CreateSurveyReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateSurveyReq.Marshal(b, m, deterministic)
}
func (m *CreateSurveyReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateSurveyReq.Merge(m, src)
}
func (m *CreateSurveyReq) XXX_Size() int {
	return xxx_messageInfo_CreateSurveyReq.Size(m)
}
func (m *CreateSurveyReq) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateSurveyReq.DiscardUnknown(m)
}

var xxx_messageInfo_CreateSurveyReq proto.InternalMessageInfo

func (m *CreateSurveyReq) GetToken() *TokenInfos {
	if m != nil {
		return m.Token
	}
	return nil
}

func (m *CreateSurveyReq) GetSurveyDef() *SurveyItem {
	if m != nil {
		return m.SurveyDef
	}
	return nil
}

type SubmitResponseReq struct {
	Token                *TokenInfos     `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	StudyKey             string          `protobuf:"bytes,2,opt,name=study_key,json=studyKey,proto3" json:"study_key,omitempty"`
	Response             *SurveyResponse `protobuf:"bytes,3,opt,name=response,proto3" json:"response,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *SubmitResponseReq) Reset()         { *m = SubmitResponseReq{} }
func (m *SubmitResponseReq) String() string { return proto.CompactTextString(m) }
func (*SubmitResponseReq) ProtoMessage()    {}
func (*SubmitResponseReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_81f0d8f98f9be15c, []int{2}
}

func (m *SubmitResponseReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SubmitResponseReq.Unmarshal(m, b)
}
func (m *SubmitResponseReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SubmitResponseReq.Marshal(b, m, deterministic)
}
func (m *SubmitResponseReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubmitResponseReq.Merge(m, src)
}
func (m *SubmitResponseReq) XXX_Size() int {
	return xxx_messageInfo_SubmitResponseReq.Size(m)
}
func (m *SubmitResponseReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SubmitResponseReq.DiscardUnknown(m)
}

var xxx_messageInfo_SubmitResponseReq proto.InternalMessageInfo

func (m *SubmitResponseReq) GetToken() *TokenInfos {
	if m != nil {
		return m.Token
	}
	return nil
}

func (m *SubmitResponseReq) GetStudyKey() string {
	if m != nil {
		return m.StudyKey
	}
	return ""
}

func (m *SubmitResponseReq) GetResponse() *SurveyResponse {
	if m != nil {
		return m.Response
	}
	return nil
}

type EnterStudyRequest struct {
	Token                *TokenInfos `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	StudyKey             string      `protobuf:"bytes,2,opt,name=study_key,json=studyKey,proto3" json:"study_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *EnterStudyRequest) Reset()         { *m = EnterStudyRequest{} }
func (m *EnterStudyRequest) String() string { return proto.CompactTextString(m) }
func (*EnterStudyRequest) ProtoMessage()    {}
func (*EnterStudyRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_81f0d8f98f9be15c, []int{3}
}

func (m *EnterStudyRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EnterStudyRequest.Unmarshal(m, b)
}
func (m *EnterStudyRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EnterStudyRequest.Marshal(b, m, deterministic)
}
func (m *EnterStudyRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EnterStudyRequest.Merge(m, src)
}
func (m *EnterStudyRequest) XXX_Size() int {
	return xxx_messageInfo_EnterStudyRequest.Size(m)
}
func (m *EnterStudyRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_EnterStudyRequest.DiscardUnknown(m)
}

var xxx_messageInfo_EnterStudyRequest proto.InternalMessageInfo

func (m *EnterStudyRequest) GetToken() *TokenInfos {
	if m != nil {
		return m.Token
	}
	return nil
}

func (m *EnterStudyRequest) GetStudyKey() string {
	if m != nil {
		return m.StudyKey
	}
	return ""
}

type GetSurveyRequest struct {
	Token                *TokenInfos `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	StudyKey             string      `protobuf:"bytes,2,opt,name=study_key,json=studyKey,proto3" json:"study_key,omitempty"`
	SurveyKey            string      `protobuf:"bytes,3,opt,name=survey_key,json=surveyKey,proto3" json:"survey_key,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetSurveyRequest) Reset()         { *m = GetSurveyRequest{} }
func (m *GetSurveyRequest) String() string { return proto.CompactTextString(m) }
func (*GetSurveyRequest) ProtoMessage()    {}
func (*GetSurveyRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_81f0d8f98f9be15c, []int{4}
}

func (m *GetSurveyRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetSurveyRequest.Unmarshal(m, b)
}
func (m *GetSurveyRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetSurveyRequest.Marshal(b, m, deterministic)
}
func (m *GetSurveyRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetSurveyRequest.Merge(m, src)
}
func (m *GetSurveyRequest) XXX_Size() int {
	return xxx_messageInfo_GetSurveyRequest.Size(m)
}
func (m *GetSurveyRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetSurveyRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetSurveyRequest proto.InternalMessageInfo

func (m *GetSurveyRequest) GetToken() *TokenInfos {
	if m != nil {
		return m.Token
	}
	return nil
}

func (m *GetSurveyRequest) GetStudyKey() string {
	if m != nil {
		return m.StudyKey
	}
	return ""
}

func (m *GetSurveyRequest) GetSurveyKey() string {
	if m != nil {
		return m.SurveyKey
	}
	return ""
}

type StatusReportRequest struct {
	Token                *TokenInfos     `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	StudyKeys            []string        `protobuf:"bytes,2,rep,name=study_keys,json=studyKeys,proto3" json:"study_keys,omitempty"`
	StatusSurvey         *SurveyResponse `protobuf:"bytes,3,opt,name=status_survey,json=statusSurvey,proto3" json:"status_survey,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *StatusReportRequest) Reset()         { *m = StatusReportRequest{} }
func (m *StatusReportRequest) String() string { return proto.CompactTextString(m) }
func (*StatusReportRequest) ProtoMessage()    {}
func (*StatusReportRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_81f0d8f98f9be15c, []int{5}
}

func (m *StatusReportRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatusReportRequest.Unmarshal(m, b)
}
func (m *StatusReportRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatusReportRequest.Marshal(b, m, deterministic)
}
func (m *StatusReportRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatusReportRequest.Merge(m, src)
}
func (m *StatusReportRequest) XXX_Size() int {
	return xxx_messageInfo_StatusReportRequest.Size(m)
}
func (m *StatusReportRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StatusReportRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StatusReportRequest proto.InternalMessageInfo

func (m *StatusReportRequest) GetToken() *TokenInfos {
	if m != nil {
		return m.Token
	}
	return nil
}

func (m *StatusReportRequest) GetStudyKeys() []string {
	if m != nil {
		return m.StudyKeys
	}
	return nil
}

func (m *StatusReportRequest) GetStatusSurvey() *SurveyResponse {
	if m != nil {
		return m.StatusSurvey
	}
	return nil
}

func init() {
	proto.RegisterType((*SurveyAndContext)(nil), "inf.study_service_api.SurveyAndContext")
	proto.RegisterType((*CreateSurveyReq)(nil), "inf.study_service_api.CreateSurveyReq")
	proto.RegisterType((*SubmitResponseReq)(nil), "inf.study_service_api.SubmitResponseReq")
	proto.RegisterType((*EnterStudyRequest)(nil), "inf.study_service_api.EnterStudyRequest")
	proto.RegisterType((*GetSurveyRequest)(nil), "inf.study_service_api.GetSurveyRequest")
	proto.RegisterType((*StatusReportRequest)(nil), "inf.study_service_api.StatusReportRequest")
}

func init() {
	proto.RegisterFile("study-service-api.proto", fileDescriptor_81f0d8f98f9be15c)
}

var fileDescriptor_81f0d8f98f9be15c = []byte{
	// 536 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0xe1, 0x6a, 0x13, 0x41,
	0x10, 0x26, 0x8d, 0x8d, 0xcd, 0x24, 0x9a, 0x66, 0xa4, 0x35, 0x5e, 0x11, 0xca, 0x89, 0x5a, 0x0a,
	0x77, 0x85, 0xd6, 0xff, 0x12, 0x6b, 0x89, 0xc5, 0x5f, 0xde, 0x89, 0x82, 0x08, 0xe1, 0xd2, 0x4c,
	0xc2, 0xd1, 0x64, 0x77, 0x7b, 0xbb, 0x57, 0xcc, 0x9b, 0xf8, 0x02, 0xbe, 0x83, 0x8f, 0x57, 0x6e,
	0x77, 0x2f, 0x49, 0x73, 0xbd, 0x42, 0x20, 0x3f, 0x77, 0xe6, 0x9b, 0xef, 0x9b, 0x9d, 0xd9, 0x6f,
	0xe1, 0xa5, 0x54, 0xe9, 0x70, 0xe6, 0x49, 0x4a, 0x6e, 0xe3, 0x2b, 0xf2, 0x22, 0x11, 0xfb, 0x22,
	0xe1, 0x8a, 0xe3, 0x5e, 0xcc, 0x46, 0xbe, 0x4e, 0xf6, 0x6d, 0xb2, 0x1f, 0x89, 0xd8, 0xc1, 0xf1,
	0x84, 0x0f, 0xa2, 0x89, 0xa7, 0x66, 0x82, 0xa4, 0x81, 0x3a, 0x0d, 0x0d, 0xb3, 0x87, 0xa6, 0x4c,
	0x93, 0x5b, 0xca, 0x4f, 0x7b, 0xe6, 0xe4, 0x25, 0x24, 0x05, 0x67, 0x92, 0x6c, 0xf8, 0x60, 0xcc,
	0xf9, 0x78, 0x42, 0x27, 0xfa, 0x34, 0x48, 0x47, 0x27, 0x34, 0x15, 0xca, 0xd6, 0xb8, 0x12, 0x76,
	0x43, 0x5d, 0xd5, 0x65, 0xc3, 0x73, 0xce, 0x14, 0xfd, 0x51, 0x78, 0x0c, 0x35, 0xc3, 0xd4, 0xa9,
	0x1c, 0x56, 0x8e, 0x1a, 0xa7, 0xe8, 0xeb, 0xf6, 0x8c, 0x94, 0x41, 0x07, 0x16, 0x81, 0x67, 0xf0,
	0xf4, 0xca, 0x94, 0x75, 0xb6, 0x34, 0xf8, 0x55, 0x11, 0x6c, 0x79, 0x83, 0x1c, 0xe9, 0x32, 0x68,
	0x9d, 0x27, 0x14, 0x29, 0xb2, 0x64, 0x74, 0x83, 0x6f, 0x61, 0x5b, 0xf1, 0x6b, 0x62, 0x56, 0xb2,
	0xa5, 0x59, 0xbe, 0x67, 0x91, 0x4b, 0x36, 0xe2, 0x32, 0x30, 0x59, 0xfc, 0x00, 0x75, 0x43, 0xfd,
	0x99, 0x46, 0x56, 0x70, 0xbf, 0x28, 0x78, 0xa9, 0x68, 0x1a, 0x2c, 0x80, 0xee, 0xdf, 0x0a, 0xb4,
	0xc3, 0x74, 0x30, 0x8d, 0x55, 0x60, 0x47, 0xb3, 0x86, 0xe4, 0x01, 0xd4, 0xcd, 0x66, 0xae, 0x69,
	0xa6, 0x25, 0xeb, 0xc1, 0x8e, 0x0e, 0x7c, 0xa5, 0x19, 0x7e, 0x84, 0x9d, 0x7c, 0xda, 0x9d, 0xaa,
	0xa6, 0x79, 0xb3, 0xd4, 0x4e, 0x7f, 0xbe, 0x89, 0xfc, 0xa2, 0x56, 0x7d, 0x5e, 0xe4, 0xfe, 0x84,
	0xf6, 0x05, 0x53, 0x94, 0x84, 0x19, 0x63, 0x40, 0x37, 0x29, 0x49, 0xb5, 0x89, 0xce, 0xdc, 0x14,
	0x76, 0x7b, 0xa4, 0xe6, 0x03, 0xde, 0x14, 0x2f, 0xbe, 0x06, 0xb0, 0x97, 0xcb, 0xb2, 0x55, 0x9d,
	0xb5, 0xa3, 0xce, 0x64, 0xff, 0x55, 0xe0, 0x45, 0xa8, 0x22, 0x95, 0xca, 0x80, 0x04, 0x4f, 0xd4,
	0x9a, 0xd2, 0x19, 0x7b, 0x2e, 0x2d, 0x3b, 0x5b, 0x87, 0x55, 0xcd, 0x6e, 0xb5, 0x25, 0x7e, 0x81,
	0x67, 0x52, 0x93, 0xf7, 0xed, 0x03, 0x5d, 0x63, 0xe6, 0x4d, 0x53, 0x69, 0xa2, 0xa7, 0xff, 0x9f,
	0x40, 0x4b, 0xcf, 0x3c, 0x34, 0x7e, 0xeb, 0x8a, 0x18, 0x3d, 0xa8, 0x99, 0xd6, 0x71, 0xdf, 0x37,
	0x9e, 0xf1, 0x73, 0xcf, 0xf8, 0x17, 0x99, 0x67, 0x9c, 0x86, 0x16, 0xb2, 0xa0, 0x00, 0x60, 0xb1,
	0x3a, 0x3c, 0xf2, 0x1f, 0xf4, 0xb0, 0x5f, 0xd8, 0xae, 0xe3, 0x2c, 0x90, 0x7e, 0x57, 0xca, 0x78,
	0xcc, 0x68, 0x68, 0xba, 0x92, 0xd8, 0x05, 0xec, 0x91, 0x5a, 0x8d, 0xae, 0x4e, 0xeb, 0x51, 0x0a,
	0x82, 0x76, 0x81, 0x02, 0xdf, 0x97, 0x74, 0xb7, 0xfa, 0x44, 0x9c, 0x32, 0x60, 0xe1, 0x93, 0xf8,
	0x0d, 0x68, 0x2c, 0xb5, 0xbc, 0x6d, 0x3c, 0x2e, 0x2b, 0x2f, 0x3e, 0x89, 0x47, 0x2f, 0xd1, 0x83,
	0xe7, 0xf7, 0x0d, 0x5b, 0x3a, 0xdf, 0x82, 0xaf, 0xef, 0x2f, 0xe9, 0x1b, 0x34, 0x97, 0xbf, 0x1a,
	0x7c, 0x57, 0x42, 0xb3, 0xf2, 0x1f, 0x39, 0x0f, 0x7c, 0x63, 0x3f, 0x28, 0x91, 0x31, 0x67, 0x9f,
	0xb6, 0x7f, 0x55, 0x23, 0x11, 0x0f, 0x6a, 0xfa, 0x6d, 0x9c, 0xdd, 0x05, 0x00, 0x00, 0xff, 0xff,
	0x4d, 0xd3, 0x96, 0x5a, 0xd5, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// StudyServiceApiClient is the client API for StudyServiceApi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StudyServiceApiClient interface {
	Status(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Status, error)
	// Study flow
	EnterStudy(ctx context.Context, in *EnterStudyRequest, opts ...grpc.CallOption) (*AssignedSurveys, error)
	GetAssignedSurveys(ctx context.Context, in *TokenInfos, opts ...grpc.CallOption) (*AssignedSurveys, error)
	GetAssignedSurvey(ctx context.Context, in *GetSurveyRequest, opts ...grpc.CallOption) (*SurveyAndContext, error)
	SubmitStatusReport(ctx context.Context, in *StatusReportRequest, opts ...grpc.CallOption) (*AssignedSurveys, error)
	SubmitResponse(ctx context.Context, in *SubmitResponseReq, opts ...grpc.CallOption) (*Status, error)
	CreateSurvey(ctx context.Context, in *CreateSurveyReq, opts ...grpc.CallOption) (*SurveyVersion, error)
}

type studyServiceApiClient struct {
	cc grpc.ClientConnInterface
}

func NewStudyServiceApiClient(cc grpc.ClientConnInterface) StudyServiceApiClient {
	return &studyServiceApiClient{cc}
}

func (c *studyServiceApiClient) Status(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/inf.study_service_api.StudyServiceApi/Status", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studyServiceApiClient) EnterStudy(ctx context.Context, in *EnterStudyRequest, opts ...grpc.CallOption) (*AssignedSurveys, error) {
	out := new(AssignedSurveys)
	err := c.cc.Invoke(ctx, "/inf.study_service_api.StudyServiceApi/EnterStudy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studyServiceApiClient) GetAssignedSurveys(ctx context.Context, in *TokenInfos, opts ...grpc.CallOption) (*AssignedSurveys, error) {
	out := new(AssignedSurveys)
	err := c.cc.Invoke(ctx, "/inf.study_service_api.StudyServiceApi/GetAssignedSurveys", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studyServiceApiClient) GetAssignedSurvey(ctx context.Context, in *GetSurveyRequest, opts ...grpc.CallOption) (*SurveyAndContext, error) {
	out := new(SurveyAndContext)
	err := c.cc.Invoke(ctx, "/inf.study_service_api.StudyServiceApi/GetAssignedSurvey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studyServiceApiClient) SubmitStatusReport(ctx context.Context, in *StatusReportRequest, opts ...grpc.CallOption) (*AssignedSurveys, error) {
	out := new(AssignedSurveys)
	err := c.cc.Invoke(ctx, "/inf.study_service_api.StudyServiceApi/SubmitStatusReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studyServiceApiClient) SubmitResponse(ctx context.Context, in *SubmitResponseReq, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/inf.study_service_api.StudyServiceApi/SubmitResponse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studyServiceApiClient) CreateSurvey(ctx context.Context, in *CreateSurveyReq, opts ...grpc.CallOption) (*SurveyVersion, error) {
	out := new(SurveyVersion)
	err := c.cc.Invoke(ctx, "/inf.study_service_api.StudyServiceApi/CreateSurvey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StudyServiceApiServer is the server API for StudyServiceApi service.
type StudyServiceApiServer interface {
	Status(context.Context, *empty.Empty) (*Status, error)
	// Study flow
	EnterStudy(context.Context, *EnterStudyRequest) (*AssignedSurveys, error)
	GetAssignedSurveys(context.Context, *TokenInfos) (*AssignedSurveys, error)
	GetAssignedSurvey(context.Context, *GetSurveyRequest) (*SurveyAndContext, error)
	SubmitStatusReport(context.Context, *StatusReportRequest) (*AssignedSurveys, error)
	SubmitResponse(context.Context, *SubmitResponseReq) (*Status, error)
	CreateSurvey(context.Context, *CreateSurveyReq) (*SurveyVersion, error)
}

// UnimplementedStudyServiceApiServer can be embedded to have forward compatible implementations.
type UnimplementedStudyServiceApiServer struct {
}

func (*UnimplementedStudyServiceApiServer) Status(ctx context.Context, req *empty.Empty) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (*UnimplementedStudyServiceApiServer) EnterStudy(ctx context.Context, req *EnterStudyRequest) (*AssignedSurveys, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnterStudy not implemented")
}
func (*UnimplementedStudyServiceApiServer) GetAssignedSurveys(ctx context.Context, req *TokenInfos) (*AssignedSurveys, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAssignedSurveys not implemented")
}
func (*UnimplementedStudyServiceApiServer) GetAssignedSurvey(ctx context.Context, req *GetSurveyRequest) (*SurveyAndContext, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAssignedSurvey not implemented")
}
func (*UnimplementedStudyServiceApiServer) SubmitStatusReport(ctx context.Context, req *StatusReportRequest) (*AssignedSurveys, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitStatusReport not implemented")
}
func (*UnimplementedStudyServiceApiServer) SubmitResponse(ctx context.Context, req *SubmitResponseReq) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitResponse not implemented")
}
func (*UnimplementedStudyServiceApiServer) CreateSurvey(ctx context.Context, req *CreateSurveyReq) (*SurveyVersion, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSurvey not implemented")
}

func RegisterStudyServiceApiServer(s *grpc.Server, srv StudyServiceApiServer) {
	s.RegisterService(&_StudyServiceApi_serviceDesc, srv)
}

func _StudyServiceApi_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudyServiceApiServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inf.study_service_api.StudyServiceApi/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudyServiceApiServer).Status(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudyServiceApi_EnterStudy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EnterStudyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudyServiceApiServer).EnterStudy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inf.study_service_api.StudyServiceApi/EnterStudy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudyServiceApiServer).EnterStudy(ctx, req.(*EnterStudyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudyServiceApi_GetAssignedSurveys_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenInfos)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudyServiceApiServer).GetAssignedSurveys(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inf.study_service_api.StudyServiceApi/GetAssignedSurveys",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudyServiceApiServer).GetAssignedSurveys(ctx, req.(*TokenInfos))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudyServiceApi_GetAssignedSurvey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSurveyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudyServiceApiServer).GetAssignedSurvey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inf.study_service_api.StudyServiceApi/GetAssignedSurvey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudyServiceApiServer).GetAssignedSurvey(ctx, req.(*GetSurveyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudyServiceApi_SubmitStatusReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudyServiceApiServer).SubmitStatusReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inf.study_service_api.StudyServiceApi/SubmitStatusReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudyServiceApiServer).SubmitStatusReport(ctx, req.(*StatusReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudyServiceApi_SubmitResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitResponseReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudyServiceApiServer).SubmitResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inf.study_service_api.StudyServiceApi/SubmitResponse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudyServiceApiServer).SubmitResponse(ctx, req.(*SubmitResponseReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _StudyServiceApi_CreateSurvey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateSurveyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudyServiceApiServer).CreateSurvey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/inf.study_service_api.StudyServiceApi/CreateSurvey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudyServiceApiServer).CreateSurvey(ctx, req.(*CreateSurveyReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _StudyServiceApi_serviceDesc = grpc.ServiceDesc{
	ServiceName: "inf.study_service_api.StudyServiceApi",
	HandlerType: (*StudyServiceApiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Status",
			Handler:    _StudyServiceApi_Status_Handler,
		},
		{
			MethodName: "EnterStudy",
			Handler:    _StudyServiceApi_EnterStudy_Handler,
		},
		{
			MethodName: "GetAssignedSurveys",
			Handler:    _StudyServiceApi_GetAssignedSurveys_Handler,
		},
		{
			MethodName: "GetAssignedSurvey",
			Handler:    _StudyServiceApi_GetAssignedSurvey_Handler,
		},
		{
			MethodName: "SubmitStatusReport",
			Handler:    _StudyServiceApi_SubmitStatusReport_Handler,
		},
		{
			MethodName: "SubmitResponse",
			Handler:    _StudyServiceApi_SubmitResponse_Handler,
		},
		{
			MethodName: "CreateSurvey",
			Handler:    _StudyServiceApi_CreateSurvey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "study-service-api.proto",
}
