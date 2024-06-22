package carclient

import (
	"AutonomousCarFleetSimulation/api"
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type CarClientServiceServer struct {
	api.CarClientServiceServer
	car *Car
}

func (s *CarClientServiceServer) GetCarInfo(ctx context.Context, in *api.Empty) (*api.CarInfo, error) {
	// Implementiere deine Logik hier
	return &api.CarInfo{
		Identifier:  s.car.Car.Identifier,
		Position:    s.car.Car.Position,
		Route:       s.car.Car.Route,
		ActiveRoute: s.car.Car.ActiveRoute,
		Color:       s.car.Car.Color,
	}, nil
}

func (s *CarClientServiceServer) SendRoute(ctx context.Context, req *api.Route) (*api.RouteResponse, error) {
	s.car.mu.Lock()
	defer s.car.mu.Unlock()

	// Debug-Ausgabe der neuen Route
	fmt.Println("Received new route:")
	for _, coord := range req.Coordinates {
		fmt.Printf("Coordinate: X=%d, Y=%d\n", coord.X, coord.Y)
	}

	s.car.Car.Route = req
	s.car.Car.ActiveRoute = true
	fmt.Println("Route updated successfully")

	return &api.RouteResponse{Message: "Route received successfully"}, nil
}

func (car *Car) startCarClientServer(port string) {
	server := grpc.NewServer()
	api.RegisterCarClientServiceServer(server, &CarClientServiceServer{car: car})

	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
	fmt.Println("Car client server started")
	if err := server.Serve(listener); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}
