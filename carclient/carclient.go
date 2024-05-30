package carclient

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // Initialize the random number generator
}

type Car struct {
	Car         utils.CarInfo
	Conn        *grpc.ClientConn
	Client      api.CoordinatorServiceClient
	GridWidth   int
	GridHeight  int
	LastMoveDir int // 0: up, 1: down, 2: left, 3: right
	mu          sync.Mutex
}

func newCar(identifier string, startPos utils.Coordinate, gridWidth, gridHeight int) *Car {
	// Establish a connection to the car client service via gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect to car client service:", err)
		return nil
	}

	client := api.NewCoordinatorServiceClient(conn)

	return &Car{
		Car: utils.CarInfo{
			Identifier:  identifier,
			Position:    startPos,
			Route:       []utils.Coordinate{}, // Empty route to start with
			ActiveRoute: false,
		},
		Conn:        conn,
		Client:      client,
		GridWidth:   gridWidth,
		GridHeight:  gridHeight,
		LastMoveDir: -1, // Initialize to an invalid direction
	}
}

func (c *Car) drive() {
	c.mu.Lock()
	if c.Car.ActiveRoute && len(c.Car.Route) > 0 {
		c.mu.Unlock()
		fmt.Println("Switching to driveRoute mode")
		c.driveRoute()
	} else {
		c.mu.Unlock()
		c.randomDrive()
	}
	c.mu.Lock()
	fmt.Printf("Driving to new position: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
	c.mu.Unlock()
	c.sendCarInfo() // Send updated position to the coordinator
}

func (c *Car) randomDrive() {
	var moveDirection int

	for {
		moveDirection = rand.Intn(4) // Randomly choose between 0 (up), 1 (down), 2 (left), 3 (right)

		if moveDirection == c.oppositeDirection() {
			continue // Skip if it is the opposite of the last move
		}

		newPosition := c.Car.Position

		switch moveDirection {
		case 0:
			// Move vertically up
			if newPosition.Y > 0 {
				newPosition.Y -= 1
			} else {
				continue // Skip if the move is out of bounds
			}
		case 1:
			// Move vertically down
			if newPosition.Y < int32(c.GridHeight-1) {
				newPosition.Y += 1
			} else {
				continue // Skip if the move is out of bounds
			}
		case 2:
			// Move horizontally left
			if newPosition.X > 0 {
				newPosition.X -= 1
			} else {
				continue // Skip if the move is out of bounds
			}
		case 3:
			// Move horizontally right
			if newPosition.X < int32(c.GridWidth-1) {
				newPosition.X += 1
			} else {
				continue // Skip if the move is out of bounds
			}
		}

		// If the new position is valid and not reversing the last move, update the position and break the loop
		c.LastMoveDir = moveDirection
		c.mu.Lock()
		c.Car.Position = newPosition
		c.mu.Unlock()
		break
	}
}

func (c *Car) oppositeDirection() int {
	switch c.LastMoveDir {
	case 0:
		return 1 // Opposite of up is down
	case 1:
		return 0 // Opposite of down is up
	case 2:
		return 3 // Opposite of left is right
	case 3:
		return 2 // Opposite of right is left
	default:
		return -1 // No valid last direction
	}
}

func (c *Car) sendCarInfo() {
	// Create and send a CarInfo request
	c.mu.Lock()
	req := &api.CarInfoRequest{
		Identifier:  c.Car.Identifier,
		Position:    &api.Coordinate{X: c.Car.Position.X, Y: c.Car.Position.Y},
		Route:       convertToProtoCoordinates(c.Car.Route),
		ActiveRoute: c.Car.ActiveRoute,
	}
	c.mu.Unlock()

	stream, err := c.Client.SendCarInfo(context.Background(), req)
	if err != nil {
		fmt.Println("Error sending car info:", err)
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break // No more messages
		}
		if err != nil {
			fmt.Println("Error receiving response from server:", err)
			return
		}
		fmt.Println("Response from server:", resp.Message)
	}
}

func (c *Car) periodicCarInfoUpdate() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		c.sendCarInfo()
	}
}

func (c *Car) SendRoute(ctx context.Context, req *api.RouteRequest) (*api.RouteResponse, error) {
	route := convertFromProtoCoordinates(req.Route)
	c.mu.Lock()
	c.Car.Route = route
	c.Car.ActiveRoute = true
	c.mu.Unlock()
	fmt.Println("Route received successfully")
	return &api.RouteResponse{Message: "Route received successfully"}, nil
}

func (c *Car) driveRoute() {
	for len(c.Car.Route) > 0 {
		c.mu.Lock()
		c.Car.Position = c.Car.Route[0] // Update the current position to the first element
		c.Car.Route = c.Car.Route[1:]   // Remove the first element from the route
		fmt.Printf("Driving to new position: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
		c.mu.Unlock()
		c.sendCarInfo()             // Send updated position to the coordinator
		time.Sleep(1 * time.Second) // Simulate driving time
	}
	fmt.Println("Route completed. Switching to random drive after 5 seconds.")
	time.Sleep(5 * time.Second) // Stay at the final position for 5 seconds
	c.mu.Lock()
	c.Car.ActiveRoute = false // Route is completed, switch to random drive
	c.mu.Unlock()
}

func convertToProtoCoordinates(coords []utils.Coordinate) []*api.Coordinate {
	var protoCoords []*api.Coordinate
	for _, c := range coords {
		protoCoords = append(protoCoords, &api.Coordinate{X: c.X, Y: c.Y})
	}
	return protoCoords
}

func convertFromProtoCoordinates(coords []*api.Coordinate) []utils.Coordinate {
	var converted []utils.Coordinate
	for _, c := range coords {
		converted = append(converted, utils.Coordinate{X: c.X, Y: c.Y})
	}
	return converted
}

type CarClientServiceServer struct {
	api.UnimplementedCarClientServiceServer
	car *Car
}

func (s *CarClientServiceServer) SendRoute(ctx context.Context, req *api.RouteRequest) (*api.RouteResponse, error) {
	return s.car.SendRoute(ctx, req)
}

func startCarClientServer(car *Car) {
	server := grpc.NewServer()
	api.RegisterCarClientServiceServer(server, &CarClientServiceServer{car: car})

	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		fmt.Println("Failed to listen:", err)
		return
	}
	fmt.Println("Car client server started")
	if err := server.Serve(listener); err != nil {
		fmt.Println("Failed to serve:", err)
	}
}

func StartClient() {
	startPos := utils.Coordinate{X: 3, Y: 3}
	gridWidth := 8
	gridHeight := 8
	car := newCar("localhost:50052", startPos, gridWidth, gridHeight)
	if car == nil {
		fmt.Println("Failed to create car client")
		return
	}
	fmt.Printf("Starting car: %+v\n", car.Car)
	go car.startDriving()          // Start driving in a separate goroutine
	go car.periodicCarInfoUpdate() // Start periodic updates in a separate goroutine

	// Start the car client gRPC server
	go startCarClientServer(car)

	select {} // Block forever
}

func (c *Car) startDriving() {
	for {
		c.drive()
		time.Sleep(1 * time.Second) // Simulate driving time
	}
}
