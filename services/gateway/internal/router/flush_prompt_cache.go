package router

// FlushPromptCacheRequest and FlushPromptCacheResponse are manually defined
// until `make gen-proto` regenerates the protobuf code from router.proto.
// These match the proto definitions added in proto/router.proto.

import "google.golang.org/protobuf/runtime/protoimpl"

type FlushPromptCacheRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	OrgId         string                 `protobuf:"bytes,1,opt,name=org_id,json=orgId,proto3" json:"org_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FlushPromptCacheRequest) Reset()         { *x = FlushPromptCacheRequest{} }
func (x *FlushPromptCacheRequest) String() string  { return x.OrgId }
func (x *FlushPromptCacheRequest) ProtoMessage()   {}

func (x *FlushPromptCacheRequest) GetOrgId() string {
	if x != nil {
		return x.OrgId
	}
	return ""
}

type FlushPromptCacheResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	DeletedCount  int64                  `protobuf:"varint,2,opt,name=deleted_count,json=deletedCount,proto3" json:"deleted_count,omitempty"`
	Error         string                 `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *FlushPromptCacheResponse) Reset()         { *x = FlushPromptCacheResponse{} }
func (x *FlushPromptCacheResponse) String() string  { return x.Error }
func (x *FlushPromptCacheResponse) ProtoMessage()   {}

func (x *FlushPromptCacheResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *FlushPromptCacheResponse) GetDeletedCount() int64 {
	if x != nil {
		return x.DeletedCount
	}
	return 0
}

func (x *FlushPromptCacheResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}
