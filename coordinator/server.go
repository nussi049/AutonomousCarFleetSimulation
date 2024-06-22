package coordinator

import (
	"AutonomousCarFleetSimulation/api"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
)

type CoordinatorServiceServer struct {
	api.CoordinatorServiceServer
}

func (s *CoordinatorServiceServer) SendCarInfo(ctx context.Context, req *api.CarInfo) (*api.CarInfoResponse, error) {
	// Send CarInfo to the channel
	carInfoCh <- req
	log.Printf("Car info received successfully from: %v", req.Identifier)

	// Return success message
	return &api.CarInfoResponse{
		Message: "Car info received successfully",
	}, nil
}

func startServer() {
	// Create a gRPC server
	server := grpc.NewServer()

	// Register your server implementation
	coordinatorServer := &CoordinatorServiceServer{}
	api.RegisterCoordinatorServiceServer(server, coordinatorServer)

	// Start the server on a specific port
	listener, err := net.Listen("tcp", ":50000")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Server started")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
