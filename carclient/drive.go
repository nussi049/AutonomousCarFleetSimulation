package carclient

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"fmt"
	"math"
	"math/rand"
	"time"
)

func (c *Car) drive() {
	for {
		c.mu.Lock()
		if c.Car.ActiveRoute && len(c.Car.Route.Coordinates) > 0 {
			c.mu.Unlock()
			fmt.Println("Switching to driveRoute mode")
			c.driveRoute()
		} else {
			if c.advancedD {
				// Advanced Drive
				c.mu.Unlock()
				c.advancedDrive()
			} else {
				// Random Drive
				c.mu.Unlock()
				c.randomDrive()
			}
		}
		c.mu.Lock()
		fmt.Printf("Driving to new position: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
		c.mu.Unlock()
		c.updateCoordinator()       // Send updated position to the coordinator
		time.Sleep(1 * time.Second) // Simulate driving time
	}
}

func (c *Car) randomDrive() {
	var moveDirection int

	for {
		moveDirection = rand.Intn(4) // Randomly choose between 0 (up), 1 (down), 2 (left), 3 (right)

		if moveDirection == c.oppositeDirection() {
			continue // Skip if it is the opposite of the last move
		}

		// Create a new instance of api.Coordinate
		newPosition := &api.Coordinate{
			X: c.Car.Position.X,
			Y: c.Car.Position.Y,
		}

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

func (c *Car) manhattanDistance(p1, p2 *api.Coordinate) int {
	return int(math.Abs(float64(p1.X-p2.X)) + math.Abs(float64(p1.Y-p2.Y)))
}

func (c *Car) calculateCost(pos *api.Coordinate) float64 {
	c.peerMutex.Lock()
	defer c.peerMutex.Unlock()

	cost := 0.0
	for _, peer := range c.peers {
		distance := c.manhattanDistance(pos, peer.Position)
		if distance != 0 {
			cost += 1 / float64(distance)
		}
	}
	return cost
}

func (c *Car) advancedDrive() {
	potentialPositions := []*api.Coordinate{
		{X: c.Car.Position.X, Y: c.Car.Position.Y + 1}, // up
		{X: c.Car.Position.X, Y: c.Car.Position.Y - 1}, // down
		{X: c.Car.Position.X - 1, Y: c.Car.Position.Y}, // left
		{X: c.Car.Position.X + 1, Y: c.Car.Position.Y}, // right
		{X: c.Car.Position.X, Y: c.Car.Position.Y},     // hold
	}

	var bestPosition *api.Coordinate
	minCost := math.MaxFloat64

	for _, pos := range potentialPositions {
		// Check if the position is within bounds
		if pos.X < 0 || pos.X >= int32(c.GridWidth) || pos.Y < 0 || pos.Y >= int32(c.GridHeight) {
			continue
		}

		cost := c.calculateCost(pos)
		if cost < minCost {
			minCost = cost
			bestPosition = pos
		}
	}

	// Update the car's position
	if bestPosition != nil {
		c.mu.Lock()
		c.Car.Position = bestPosition
		c.mu.Unlock()
	}
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
		c.updateCoordinator()
		time.Sleep(1 * time.Second)
	}

	for _, coord := range c.Car.Route.Coordinates {
		c.mu.Lock()
		c.Car.Position = coord
		fmt.Printf("Driving to route position: X: %d, Y: %d\n", c.Car.Position.X, c.Car.Position.Y)
		c.mu.Unlock()
		c.updateCoordinator()
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Route completed. Checking for new route or switching to random drive after 1 seconds.")
	time.Sleep(1 * time.Second)
	c.mu.Lock()
	c.Car.ActiveRoute = false // Route is completed, switch to random drive if no new route
	fmt.Printf("Set Active Route to false")

	c.mu.Unlock()
}
