package carclient

import (
	"AutonomousCarFleetSimulation/utils"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type CarClient struct {
	Car  utils.CarInfo
	Conn net.Conn
}

func NewCarClient(identifier string, posX, posY int) *CarClient {
	// Establish a connection to the coordinator server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Failed to connect to coordinator:", err)
		return nil
	}
	return &CarClient{
		Car: utils.CarInfo{
			Identifier:  identifier,
			PositionX:   posX,
			PositionY:   posY,
			ActiveRoute: false,
		},
		Conn: conn,
	}
}

func (c *CarClient) Drive() {
	c.Car.PositionX += 1
	c.Car.PositionY += 1
	fmt.Printf("Driving to new position: X: %d, Y: %d\n", c.Car.PositionX, c.Car.PositionY)
}

func (c *CarClient) SendPosition() {
	// Send the current position to the central coordinator using JSON format
	positionData, _ := json.Marshal(c.Car)
	_, err := c.Conn.Write(positionData)
	if err != nil {
		fmt.Println("Error sending position data:", err)
	}
}

func (c *CarClient) PeriodicPositionUpdate() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		c.SendPosition()
	}
}

func (c *CarClient) Run() {
	go c.StartDriving()
	go c.PeriodicPositionUpdate()
}

func (c *CarClient) StartDriving() {
	for {
		c.Drive()
		time.Sleep(1 * time.Second)
	}
}

func Run() {
	car := NewCarClient("TestCar1", 0, 0)
	if car == nil {
		fmt.Println("Failed to create car client")
		return
	}
	fmt.Printf("Starting car: %+v\n", car.Car)
	car.Run()
}
