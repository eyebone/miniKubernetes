// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.12.2
// source: google/cloud/ml/v1/prediction_service.proto

package ml

import (
	context "context"
	reflect "reflect"
	sync "sync"

	_ "google.golang.org/genproto/googleapis/api/annotations"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Request for predictions to be issued against a trained model.
//
// The body of the request is a single JSON object with a single top-level
// field:
//
// <dl>
//
//	<dt>instances</dt>
//	<dd>A JSON array containing values representing the instances to use for
//	    prediction.</dd>
//
// </dl>
//
// The structure of each element of the instances list is determined by your
// model's input definition. Instances can include named inputs or can contain
// only unlabeled values.
//
// Not all data includes named inputs. Some instances will be simple
// JSON values (boolean, number, or string). However, instances are often lists
// of simple values, or complex nested lists. Here are some examples of request
// bodies:
//
// CSV data with each row encoded as a string value:
// <pre>
// {"instances": ["1.0,true,\\"x\\"", "-2.0,false,\\"y\\""]}
// </pre>
// Plain text:
// <pre>
// {"instances": ["the quick brown fox", "la bruja le dio"]}
// </pre>
// Sentences encoded as lists of words (vectors of strings):
// <pre>
//
//	{
//	  "instances": [
//	    ["the","quick","brown"],
//	    ["la","bruja","le"],
//	    ...
//	  ]
//	}
//
// </pre>
// Floating point scalar values:
// <pre>
// {"instances": [0.0, 1.1, 2.2]}
// </pre>
// Vectors of integers:
// <pre>
//
//	{
//	  "instances": [
//	    [0, 1, 2],
//	    [3, 4, 5],
//	    ...
//	  ]
//	}
//
// </pre>
// Tensors (in this case, two-dimensional tensors):
// <pre>
//
//	{
//	  "instances": [
//	    [
//	      [0, 1, 2],
//	      [3, 4, 5]
//	    ],
//	    ...
//	  ]
//	}
//
// </pre>
// Images can be represented different ways. In this encoding scheme the first
// two dimensions represent the rows and columns of the image, and the third
// contains lists (vectors) of the R, G, and B values for each pixel.
// <pre>
//
//	{
//	  "instances": [
//	    [
//	      [
//	        [138, 30, 66],
//	        [130, 20, 56],
//	        ...
//	      ],
//	      [
//	        [126, 38, 61],
//	        [122, 24, 57],
//	        ...
//	      ],
//	      ...
//	    ],
//	    ...
//	  ]
//	}
//
// </pre>
// JSON strings must be encoded as UTF-8. To send binary data, you must
// base64-encode the data and mark it as binary. To mark a JSON string
// as binary, replace it with a JSON object with a single attribute named `b64`:
// <pre>{"b64": "..."} </pre>
// For example:
//
// Two Serialized tf.Examples (fake data, for illustrative purposes only):
// <pre>
// {"instances": [{"b64": "X5ad6u"}, {"b64": "IA9j4nx"}]}
// </pre>
// Two JPEG image byte strings (fake data, for illustrative purposes only):
// <pre>
// {"instances": [{"b64": "ASa8asdf"}, {"b64": "JLK7ljk3"}]}
// </pre>
// If your data includes named references, format each instance as a JSON object
// with the named references as the keys:
//
// JSON input data to be preprocessed:
// <pre>
//
//	{
//	  "instances": [
//	    {
//	      "a": 1.0,
//	      "b": true,
//	      "c": "x"
//	    },
//	    {
//	      "a": -2.0,
//	      "b": false,
//	      "c": "y"
//	    }
//	  ]
//	}
//
// </pre>
// Some models have an underlying TensorFlow graph that accepts multiple input
// tensors. In this case, you should use the names of JSON name/value pairs to
// identify the input tensors, as shown in the following exmaples:
//
// For a graph with input tensor aliases "tag" (string) and "image"
// (base64-encoded string):
// <pre>
//
//	{
//	  "instances": [
//	    {
//	      "tag": "beach",
//	      "image": {"b64": "ASa8asdf"}
//	    },
//	    {
//	      "tag": "car",
//	      "image": {"b64": "JLK7ljk3"}
//	    }
//	  ]
//	}
//
// </pre>
// For a graph with input tensor aliases "tag" (string) and "image"
// (3-dimensional array of 8-bit ints):
// <pre>
//
//	{
//	  "instances": [
//	    {
//	      "tag": "beach",
//	      "image": [
//	        [
//	          [138, 30, 66],
//	          [130, 20, 56],
//	          ...
//	        ],
//	        [
//	          [126, 38, 61],
//	          [122, 24, 57],
//	          ...
//	        ],
//	        ...
//	      ]
//	    },
//	    {
//	      "tag": "car",
//	      "image": [
//	        [
//	          [255, 0, 102],
//	          [255, 0, 97],
//	          ...
//	        ],
//	        [
//	          [254, 1, 101],
//	          [254, 2, 93],
//	          ...
//	        ],
//	        ...
//	      ]
//	    },
//	    ...
//	  ]
//	}
//
// </pre>
// If the call is successful, the response body will contain one prediction
// entry per instance in the request body. If prediction fails for any
// instance, the response body will contain no predictions and will contian
// a single error entry instead.
type PredictRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. The resource name of a model or a version.
	//
	// Authorization: requires `Viewer` role on the parent project.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	//
	// Required. The prediction request body.
	HttpBody *httpbody.HttpBody `protobuf:"bytes,2,opt,name=http_body,json=httpBody,proto3" json:"http_body,omitempty"`
}

func (x *PredictRequest) Reset() {
	*x = PredictRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_cloud_ml_v1_prediction_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PredictRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PredictRequest) ProtoMessage() {}

func (x *PredictRequest) ProtoReflect() protoreflect.Message {
	mi := &file_google_cloud_ml_v1_prediction_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PredictRequest.ProtoReflect.Descriptor instead.
func (*PredictRequest) Descriptor() ([]byte, []int) {
	return file_google_cloud_ml_v1_prediction_service_proto_rawDescGZIP(), []int{0}
}

func (x *PredictRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PredictRequest) GetHttpBody() *httpbody.HttpBody {
	if x != nil {
		return x.HttpBody
	}
	return nil
}

var File_google_cloud_ml_v1_prediction_service_proto protoreflect.FileDescriptor

var file_google_cloud_ml_v1_prediction_service_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x6d,
	0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6d, 0x6c, 0x2e, 0x76,
	0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x68, 0x74, 0x74, 0x70,
	0x62, 0x6f, 0x64, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x57, 0x0a, 0x0e, 0x50, 0x72,
	0x65, 0x64, 0x69, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x31, 0x0a, 0x09, 0x68, 0x74, 0x74, 0x70, 0x5f, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x48, 0x74, 0x74, 0x70, 0x42, 0x6f, 0x64, 0x79, 0x52, 0x08, 0x68, 0x74, 0x74, 0x70, 0x42,
	0x6f, 0x64, 0x79, 0x32, 0x89, 0x01, 0x0a, 0x17, 0x4f, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x50, 0x72,
	0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x6e, 0x0a, 0x07, 0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x12, 0x22, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6d, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x50, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x14,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x48, 0x74, 0x74, 0x70,
	0x42, 0x6f, 0x64, 0x79, 0x22, 0x29, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x23, 0x22, 0x1e, 0x2f, 0x76,
	0x31, 0x2f, 0x7b, 0x6e, 0x61, 0x6d, 0x65, 0x3d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73,
	0x2f, 0x2a, 0x2a, 0x7d, 0x3a, 0x70, 0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x3a, 0x01, 0x2a, 0x42,
	0x6c, 0x0a, 0x1a, 0x63, 0x6f, 0x6d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x6d, 0x6c, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x42, 0x16, 0x50,
	0x72, 0x65, 0x64, 0x69, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x34, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x67, 0x65, 0x6e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2f, 0x6d, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x6d, 0x6c, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_cloud_ml_v1_prediction_service_proto_rawDescOnce sync.Once
	file_google_cloud_ml_v1_prediction_service_proto_rawDescData = file_google_cloud_ml_v1_prediction_service_proto_rawDesc
)

func file_google_cloud_ml_v1_prediction_service_proto_rawDescGZIP() []byte {
	file_google_cloud_ml_v1_prediction_service_proto_rawDescOnce.Do(func() {
		file_google_cloud_ml_v1_prediction_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_cloud_ml_v1_prediction_service_proto_rawDescData)
	})
	return file_google_cloud_ml_v1_prediction_service_proto_rawDescData
}

var file_google_cloud_ml_v1_prediction_service_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_google_cloud_ml_v1_prediction_service_proto_goTypes = []interface{}{
	(*PredictRequest)(nil),    // 0: google.cloud.ml.v1.PredictRequest
	(*httpbody.HttpBody)(nil), // 1: google.api.HttpBody
}
var file_google_cloud_ml_v1_prediction_service_proto_depIdxs = []int32{
	1, // 0: google.cloud.ml.v1.PredictRequest.http_body:type_name -> google.api.HttpBody
	0, // 1: google.cloud.ml.v1.OnlinePredictionService.Predict:input_type -> google.cloud.ml.v1.PredictRequest
	1, // 2: google.cloud.ml.v1.OnlinePredictionService.Predict:output_type -> google.api.HttpBody
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_google_cloud_ml_v1_prediction_service_proto_init() }
func file_google_cloud_ml_v1_prediction_service_proto_init() {
	if File_google_cloud_ml_v1_prediction_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_cloud_ml_v1_prediction_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PredictRequest); i {
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
			RawDescriptor: file_google_cloud_ml_v1_prediction_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_google_cloud_ml_v1_prediction_service_proto_goTypes,
		DependencyIndexes: file_google_cloud_ml_v1_prediction_service_proto_depIdxs,
		MessageInfos:      file_google_cloud_ml_v1_prediction_service_proto_msgTypes,
	}.Build()
	File_google_cloud_ml_v1_prediction_service_proto = out.File
	file_google_cloud_ml_v1_prediction_service_proto_rawDesc = nil
	file_google_cloud_ml_v1_prediction_service_proto_goTypes = nil
	file_google_cloud_ml_v1_prediction_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// OnlinePredictionServiceClient is the client API for OnlinePredictionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type OnlinePredictionServiceClient interface {
	// Performs prediction on the data in the request.
	//
	// **** REMOVE FROM GENERATED DOCUMENTATION
	Predict(ctx context.Context, in *PredictRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error)
}

type onlinePredictionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOnlinePredictionServiceClient(cc grpc.ClientConnInterface) OnlinePredictionServiceClient {
	return &onlinePredictionServiceClient{cc}
}

func (c *onlinePredictionServiceClient) Predict(ctx context.Context, in *PredictRequest, opts ...grpc.CallOption) (*httpbody.HttpBody, error) {
	out := new(httpbody.HttpBody)
	err := c.cc.Invoke(ctx, "/google.cloud.ml.v1.OnlinePredictionService/Predict", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OnlinePredictionServiceServer is the server API for OnlinePredictionService service.
type OnlinePredictionServiceServer interface {
	// Performs prediction on the data in the request.
	//
	// **** REMOVE FROM GENERATED DOCUMENTATION
	Predict(context.Context, *PredictRequest) (*httpbody.HttpBody, error)
}

// UnimplementedOnlinePredictionServiceServer can be embedded to have forward compatible implementations.
type UnimplementedOnlinePredictionServiceServer struct {
}

func (*UnimplementedOnlinePredictionServiceServer) Predict(context.Context, *PredictRequest) (*httpbody.HttpBody, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Predict not implemented")
}

func RegisterOnlinePredictionServiceServer(s *grpc.Server, srv OnlinePredictionServiceServer) {
	s.RegisterService(&_OnlinePredictionService_serviceDesc, srv)
}

func _OnlinePredictionService_Predict_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PredictRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OnlinePredictionServiceServer).Predict(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.cloud.ml.v1.OnlinePredictionService/Predict",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OnlinePredictionServiceServer).Predict(ctx, req.(*PredictRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _OnlinePredictionService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.cloud.ml.v1.OnlinePredictionService",
	HandlerType: (*OnlinePredictionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Predict",
			Handler:    _OnlinePredictionService_Predict_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/cloud/ml/v1/prediction_service.proto",
}