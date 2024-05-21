package carclient

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	rand.Seed(time.Now().UnixNano()) // Initialize the random number generator
}

type Client struct {
	Car         utils.CarInfo
	Conn        *grpc.ClientConn
	Client      api.CoordinatorServiceClient
	GridWidth   int
	GridHeight  int
	LastMoveDir int // 0: up, 1: down, 2: left, 3: right
}

func newCarClient(identifier string, startPos utils.Coordinate, gridWidth, gridHeight int) *Client {
	// Establish a connection to the coordinator service via gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect to coordinator:", err)
		return nil
	}

	client := api.NewCoordinatorServiceClient(conn)

	return &Client{
		Car: utils.CarInfo{
			Identifier:  identifier,
			Position:    startPos,
			Route:       []utils.Coordinate{}, // Empty route to start with
			ActiveRoute: true,
		},
		Conn:        conn,
		Client:      client,
		GridWidth:   gridWidth,
		GridHeight:  gridHeight,
		LastMoveDir: -1, // Initialize to an invalid direction
	}
}

func (c *Client) drive() {
	if len(c.Car.Route) > 0 {
		c.Car.Position = c.Car.Route[0] // Update the current position to the first element
		c.Car.Route = c.Car.Route[1:]   // Remove the first element from the route
	} else {
		// No route defined or route completed, perform a random drive
		c.randomDrive()
	}
	fmt.Printf("Driving to new position: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
	c.sendPosition() // Send updated position to the coordinator
}

func (c *Client) randomDrive() {
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
		c.Car.Position = newPosition
		break
	}
}

func (c *Client) oppositeDirection() int {
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

func (c *Client) sendPosition() {
	// Create and send a position update request
	req := &api.CarInfoRequest{
		Identifier:  c.Car.Identifier,
		Position:    &api.Coordinate{X: c.Car.Position.X, Y: c.Car.Position.Y},
		Route:       nil, // Not updating the route in this call
		ActiveRoute: c.Car.ActiveRoute,
	}

	stream, err := c.Client.SendCarInfo(context.Background(), req)
	if err != nil {
		fmt.Println("Error sending position data:", err)
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

func (c *Client) periodicPositionUpdate() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		c.sendPosition()
	}
}

func StartClient() {
	startPos := utils.Coordinate{X: 3, Y: 3}
	gridWidth := 8
	gridHeight := 8
	car := newCarClient("TestCar1", startPos, gridWidth, gridHeight)
	if car == nil {
		fmt.Println("Failed to create car client")
		return
	}
	fmt.Printf("Starting car: %+v\n", car.Car)
	go car.startDriving()           // Start driving in a separate goroutine
	go car.periodicPositionUpdate() // Start periodic updates in a separate goroutine
	select {}                       // Block forever
}

func (c *Client) startDriving() {
	for {
		c.drive()
		time.Sleep(1 * time.Second) // Simulate driving time
	}
}
