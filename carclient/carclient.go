package carclient

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"context"
	"flag"
	"fmt"
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
	Car         *api.CarInfo
	Conn        *grpc.ClientConn
	Client      api.CoordinatorServiceClient
	GridWidth   int
	GridHeight  int
	LastMoveDir int // 0: up, 1: down, 2: left, 3: right
	mu          sync.Mutex
}

func newCar(identifier string, startPos *api.Coordinate, color string) *Car {
	// Establish a connection to the car client service via gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect to car client service:", err)
		return nil
	}

	client := api.NewCoordinatorServiceClient(conn)

	return &Car{
		Car: &api.CarInfo{
			Identifier:  identifier,
			Position:    startPos,
			Route:       &api.Route{Coordinates: []*api.Coordinate{}}, // Empty route to start with
			ActiveRoute: false,
			Color:       color,
		},
		Conn:        conn,
		Client:      client,
		GridWidth:   utils.Settings.GridSize, // Assuming the grid size is 8, adjust if needed
		GridHeight:  utils.Settings.GridSize, // Assuming the grid size is 8, adjust if needed
		LastMoveDir: -1,                      // Initialize to an invalid direction
	}
}

func (c *Car) drive() {
	c.mu.Lock()
	if c.Car.ActiveRoute && len(c.Car.Route.Coordinates) > 0 {
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

		newPosition := *c.Car.Position

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
		c.Car.Position = &newPosition
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
	resp, err := c.Client.SendCarInfo(context.Background(), c.Car)
	if err != nil {
		fmt.Println("Error sending car info:", err)
		return
	}
	c.mu.Unlock()

	fmt.Println("Response from server:", resp.Message)
}

func (c *Car) periodicCarInfoUpdate() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		c.sendCarInfo()
	}
}

func (c *Car) SendRoute(ctx context.Context, req *api.Route) (*api.RouteResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Debug-Ausgabe der neuen Route
	fmt.Println("Received new route:")
	for _, coord := range req.Coordinates {
		fmt.Printf("Coordinate: X=%d, Y=%d\n", coord.X, coord.Y)
	}

	c.Car.Route = req
	c.Car.ActiveRoute = true
	fmt.Println("Route updated successfully")

	return &api.RouteResponse{Message: "Route received successfully"}, nil
}

func (c *Car) driveRoute() {
	if len(c.Car.Route.Coordinates) == 0 {
		return
	}

	// Drive to the first position in the route
	toRouteStart := utils.CalculatePath(c.Car.Position, c.Car.Route.Coordinates[0], c.Car.Route)
	fmt.Println("Path to route start:", toRouteStart)

	for _, coord := range toRouteStart {
		c.mu.Lock()
		c.Car.Position = coord
		fmt.Printf("Driving to route start: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
		c.mu.Unlock()
		c.sendCarInfo()
		time.Sleep(1 * time.Second)
	}

	// Drive the remaining route
	//for i := 1; i < len(c.Car.Route.Coordinates); i++ {
	//	toNext := utils.CalculatePath(c.Car.Position, c.Car.Route.Coordinates[i])
	for _, coord := range c.Car.Route.Coordinates {
		c.mu.Lock()
		c.Car.Position = coord
		fmt.Printf("Driving to route position: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
		c.mu.Unlock()
		c.sendCarInfo()
		time.Sleep(1 * time.Second)
	}
	//}

	fmt.Println("Route completed. Checking for new route or switching to random drive after 1 seconds.")
	time.Sleep(1 * time.Second)
	c.mu.Lock()
	c.Car.ActiveRoute = false // Route is completed, switch to random drive if no new route
	fmt.Printf("Set Active Route to false")

	c.mu.Unlock()
}

type CarClientServiceServer struct {
	api.CarClientServiceServer
	car *Car
}

func (s *CarClientServiceServer) SendRoute(ctx context.Context, req *api.Route) (*api.RouteResponse, error) {
	return s.car.SendRoute(ctx, req)
}

func startCarClientServer(car *Car, port string) {
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

func StartClient() {
	port := flag.Int("port", 50000, "Port for the server to listen on")
	color := flag.String("color", "", "Color of car")
	x := flag.Int("x", 3, "X Coordinate to start")
	y := flag.Int("y", 3, "Y Coordinate to start")
	flag.Parse()

	startPos := &api.Coordinate{X: int32(*x), Y: int32(*y)}

	println(fmt.Sprintf("localhost:%d", *port))
	car := newCar(fmt.Sprintf("localhost:%d", *port), startPos, *color)
	if car == nil {
		fmt.Println("Failed to create car client")
		return
	}
	fmt.Printf("Starting car: %+v\n", car.Car)
	go car.startDriving()          // Start driving in a separate goroutine
	go car.periodicCarInfoUpdate() // Start periodic updates in a separate goroutine

	// Start the car client gRPC server
	go startCarClientServer(car, fmt.Sprintf(":%d", *port))

	select {} // Block forever
}

func (c *Car) startDriving() {
	for {
		c.drive()
		time.Sleep(1 * time.Second) // Simulate driving time
	}
}
