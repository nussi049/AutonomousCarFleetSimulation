// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: coordinator.proto

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	CarClientService_SendRoute_FullMethodName = "/CarClientService/SendRoute"
)

// CarClientServiceClient is the client API for CarClientService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CarClientServiceClient interface {
	SendRoute(ctx context.Context, in *RouteRequest, opts ...grpc.CallOption) (*RouteResponse, error)
}

type carClientServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCarClientServiceClient(cc grpc.ClientConnInterface) CarClientServiceClient {
	return &carClientServiceClient{cc}
}

func (c *carClientServiceClient) SendRoute(ctx context.Context, in *RouteRequest, opts ...grpc.CallOption) (*RouteResponse, error) {
	out := new(RouteResponse)
	err := c.cc.Invoke(ctx, CarClientService_SendRoute_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CarClientServiceServer is the server API for CarClientService service.
// All implementations must embed UnimplementedCarClientServiceServer
// for forward compatibility
type CarClientServiceServer interface {
	SendRoute(context.Context, *RouteRequest) (*RouteResponse, error)
	mustEmbedUnimplementedCarClientServiceServer()
}

// UnimplementedCarClientServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCarClientServiceServer struct {
}

func (UnimplementedCarClientServiceServer) SendRoute(context.Context, *RouteRequest) (*RouteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendRoute not implemented")
}
func (UnimplementedCarClientServiceServer) mustEmbedUnimplementedCarClientServiceServer() {}

// UnsafeCarClientServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CarClientServiceServer will
// result in compilation errors.
type UnsafeCarClientServiceServer interface {
	mustEmbedUnimplementedCarClientServiceServer()
}

func RegisterCarClientServiceServer(s grpc.ServiceRegistrar, srv CarClientServiceServer) {
	s.RegisterService(&CarClientService_ServiceDesc, srv)
}

func _CarClientService_SendRoute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RouteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CarClientServiceServer).SendRoute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CarClientService_SendRoute_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CarClientServiceServer).SendRoute(ctx, req.(*RouteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CarClientService_ServiceDesc is the grpc.ServiceDesc for CarClientService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CarClientService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "CarClientService",
	HandlerType: (*CarClientServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendRoute",
			Handler:    _CarClientService_SendRoute_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "coordinator.proto",
}

const (
	CoordinatorService_SendCarInfo_FullMethodName = "/CoordinatorService/SendCarInfo"
)

// CoordinatorServiceClient is the client API for CoordinatorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CoordinatorServiceClient interface {
	SendCarInfo(ctx context.Context, in *CarInfoRequest, opts ...grpc.CallOption) (CoordinatorService_SendCarInfoClient, error)
}

type coordinatorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCoordinatorServiceClient(cc grpc.ClientConnInterface) CoordinatorServiceClient {
	return &coordinatorServiceClient{cc}
}

func (c *coordinatorServiceClient) SendCarInfo(ctx context.Context, in *CarInfoRequest, opts ...grpc.CallOption) (CoordinatorService_SendCarInfoClient, error) {
	stream, err := c.cc.NewStream(ctx, &CoordinatorService_ServiceDesc.Streams[0], CoordinatorService_SendCarInfo_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &coordinatorServiceSendCarInfoClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type CoordinatorService_SendCarInfoClient interface {
	Recv() (*CarInfoResponse, error)
	grpc.ClientStream
}

type coordinatorServiceSendCarInfoClient struct {
	grpc.ClientStream
}

func (x *coordinatorServiceSendCarInfoClient) Recv() (*CarInfoResponse, error) {
	m := new(CarInfoResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CoordinatorServiceServer is the server API for CoordinatorService service.
// All implementations must embed UnimplementedCoordinatorServiceServer
// for forward compatibility
type CoordinatorServiceServer interface {
	SendCarInfo(*CarInfoRequest, CoordinatorService_SendCarInfoServer) error
	mustEmbedUnimplementedCoordinatorServiceServer()
}

// UnimplementedCoordinatorServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCoordinatorServiceServer struct {
}

func (UnimplementedCoordinatorServiceServer) SendCarInfo(*CarInfoRequest, CoordinatorService_SendCarInfoServer) error {
	return status.Errorf(codes.Unimplemented, "method SendCarInfo not implemented")
}
func (UnimplementedCoordinatorServiceServer) mustEmbedUnimplementedCoordinatorServiceServer() {}

// UnsafeCoordinatorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CoordinatorServiceServer will
// result in compilation errors.
type UnsafeCoordinatorServiceServer interface {
	mustEmbedUnimplementedCoordinatorServiceServer()
}

func RegisterCoordinatorServiceServer(s grpc.ServiceRegistrar, srv CoordinatorServiceServer) {
	s.RegisterService(&CoordinatorService_ServiceDesc, srv)
}

func _CoordinatorService_SendCarInfo_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CarInfoRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CoordinatorServiceServer).SendCarInfo(m, &coordinatorServiceSendCarInfoServer{stream})
}

type CoordinatorService_SendCarInfoServer interface {
	Send(*CarInfoResponse) error
	grpc.ServerStream
}

type coordinatorServiceSendCarInfoServer struct {
	grpc.ServerStream
}

func (x *coordinatorServiceSendCarInfoServer) Send(m *CarInfoResponse) error {
	return x.ServerStream.SendMsg(m)
}

// CoordinatorService_ServiceDesc is the grpc.ServiceDesc for CoordinatorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CoordinatorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "CoordinatorService",
	HandlerType: (*CoordinatorServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendCarInfo",
			Handler:       _CoordinatorService_SendCarInfo_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "coordinator.proto",
}
