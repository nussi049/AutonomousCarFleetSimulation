package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	"AutonomousCarFleetSimulation/api"
)

func main() {
	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	// Create a client instance.
	client := api.NewCoordinatorServiceClient(conn)

	// Create a car info request.
	request := &api.CarInfoRequest{
		Identifier:  "car12",
		Position:    &api.Coordinate{X: 2, Y: 2},
		Route:       []*api.Coordinate{},
		ActiveRoute: true,
	}

	// Send the car info request to the server.
	stream, err := client.SendCarInfo(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to send car info: %v", err)
	}

	// Receive and handle responses from the server.
	for {
		response, err := stream.Recv()
		if err != nil {
			log.Fatalf("Failed to receive response: %v", err)
		}
		log.Printf("Response from server: %s", response.Message)
	}
}
