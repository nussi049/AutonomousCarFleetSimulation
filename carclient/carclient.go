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
	Car    utils.CarInfo
	Conn   *grpc.ClientConn
	Client api.CoordinatorServiceClient
}

func newCarClient(identifier string, route []utils.Coordinate) *Client {
	// Establish a connection to the coordinator service via gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect to coordinator:", err)
		return nil
	}

	client := api.NewCoordinatorServiceClient(conn)
	var startPos utils.Coordinate
	if len(route) > 0 {
		startPos = route[0]
	}

	return &Client{
		Car: utils.CarInfo{
			Identifier:  identifier,
			Position:    startPos,
			Route:       route,
			ActiveRoute: false,
		},
		Conn:   conn,
		Client: client,
	}
}

func (c *Client) drive() {
	if len(c.Car.Route) > 1 {
		c.Car.Route = c.Car.Route[1:]   // Move to the next position in the route
		c.Car.Position = c.Car.Route[0] // Update the current position
	} else {
		// No route defined, perform a random drive
		c.randomDrive()
	}
	fmt.Printf("Driving to new position: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
	c.sendPosition() // Send updated position to the coordinator
}

// randomDrive performs a random movement forward or sideways
func (c *Client) randomDrive() {
	// Choose randomly between moving horizontally or vertically
	moveVertically := rand.Intn(2) == 0 // Randomly choose between true (vertical) or false (horizontal)

	if moveVertically {
		// Move vertically up or down
		deltaY := rand.Intn(2) // Randomly 0 or 1
		if deltaY == 0 {
			deltaY = -1 // Move down
		}
		c.Car.Position.Y += int32(deltaY)
	} else {
		// Move horizontally left or right
		deltaX := rand.Intn(2) // Randomly 0 or 1
		if deltaX == 0 {
			deltaX = -1 // Move left
		}
		c.Car.Position.X += int32(deltaX)
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
	route := []utils.Coordinate{{X: 0, Y: 0}, {X: 1, Y: 1}, {X: 2, Y: 2}}
	car := newCarClient("TestCar1", route)
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
