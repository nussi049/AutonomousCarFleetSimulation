package carclient

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Car struct {
	CarInfo     *api.CarInfo
	Conn        *grpc.ClientConn
	Client      api.CoordinatorServiceClient
	GridWidth   int
	GridHeight  int
	LastMoveDir int // 0: up, 1: down, 2: left, 3: right
	mu          sync.Mutex
	peerMutex   sync.Mutex
	peers       map[string]*api.CarInfo
	advancedD   bool
}

func newCar(identifier string, startPos *api.Coordinate, color string, advancedD bool) *Car {
	// Establish a connection to the car client service via gRPC
	conn, err := grpc.Dial("localhost:50000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Failed to connect to car client service:", err)
		return nil
	}

	client := api.NewCoordinatorServiceClient(conn)

	return &Car{
		CarInfo: &api.CarInfo{
			Identifier:  identifier,
			Position:    startPos,
			Route:       &api.Route{Coordinates: []*api.Coordinate{}}, // Empty route to start with
			ActiveRoute: false,
			Color:       color,
		},
		Conn:        conn,
		Client:      client,
		GridWidth:   utils.Settings.GridSize,       // Assuming the grid size is 8, adjust if needed
		GridHeight:  utils.Settings.GridSize,       // Assuming the grid size is 8, adjust if needed
		LastMoveDir: -1,                            // Initialize to an invalid direction
		peers:       make(map[string]*api.CarInfo), // Initialize peers map
		advancedD:   advancedD,
	}
}

func (c *Car) updateCoordinator() {
	// Create and send a CarInfo request
	c.mu.Lock()
	resp, err := c.Client.SendCarInfo(context.Background(), c.CarInfo)
	if err != nil {
		fmt.Println("Error sending car info:", err)
		return
	}
	c.mu.Unlock()

	fmt.Println("Response from server:", resp.Message)
}

func (c *Car) discoverPeers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for port := 50001; port <= 50100; port++ {
			address := fmt.Sprintf("localhost:%d", port)

			// Skip own port
			if address == c.CarInfo.Identifier {
				continue
			}

			// Skip peers which are already in peer group
			exists := false
			for _, peer := range c.peers {
				if peer.Identifier == address {
					exists = true
					break
				}
			}
			if exists {
				continue
			}

			conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				continue
			}

			client := api.NewCarClientServiceClient(conn)
			resp, err := client.GetCarInfo(context.Background(), &api.Empty{})
			if err != nil {
				conn.Close()
				continue
			}
			c.peerMutex.Lock()
			c.peers[resp.Identifier] = resp
			c.peerMutex.Unlock()
			conn.Close()
			fmt.Println("Found Peer at: %s", address)
			fmt.Println("%s", c.peers[resp.Identifier].Position)
		}
	}
	fmt.Print(c.peers)
}

func (c *Car) updatePeers() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.peerMutex.Lock()
		for id, peer := range c.peers {
			conn, err := grpc.Dial(peer.Identifier, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Printf("Failed to connect to peer %s: %v\n", peer.Identifier, err)
				delete(c.peers, id)
				continue
			}

			client := api.NewCarClientServiceClient(conn)
			resp, err := client.GetCarInfo(context.Background(), &api.Empty{})
			if err != nil || resp == nil {
				fmt.Printf("Failed to update peer info for %s: %v\n", peer.Identifier, err)
				conn.Close()
				delete(c.peers, id)
				continue
			}

			c.peers[resp.Identifier] = resp
			conn.Close()
		}
		c.peerMutex.Unlock()
	}
}

func Run() {
	// Parse console args
	port := flag.Int("port", 50001, "Port for the server to listen on")
	color := flag.String("color", "", "Color of car")
	x := flag.Int("x", 3, "X Coordinate to start")
	y := flag.Int("y", 3, "Y Coordinate to start")
	advancedD := flag.Bool("advancedDrive", false, "AdvancedDrive Function")
	flag.Parse()

	startPos := &api.Coordinate{X: int32(*x), Y: int32(*y)}

	println(fmt.Sprintf("localhost:%d", *port))
	car := newCar(fmt.Sprintf("localhost:%d", *port), startPos, *color, *advancedD)
	if car == nil {
		fmt.Println("Failed to create car client")
		return
	}
	fmt.Printf("Starting car: %+v\n", car.CarInfo)

	// Start the car client gRPC server
	go car.startCarClientServer(fmt.Sprintf(":%d", *port))

	go car.drive() // Start driving in a separate goroutine

	go car.discoverPeers() // Start peer discovery in a separate goroutine

	go car.updatePeers() // Ask periodically for positon updates of all peers

	select {} // Block forever
}
