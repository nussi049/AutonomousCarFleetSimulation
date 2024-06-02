package coordinator

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"log"

	"gioui.org/app"
	"google.golang.org/grpc"
)

var (
	carinfos     = make([]*api.CarInfo, 0)
	carinfoMutex sync.Mutex
	carInfoCh    = make(chan *api.CarInfo)
	routeCh      = make(chan *api.Route)
	gridData     = utils.CreateDataGrid()
)

func generateRandomRoute() {
	for {
		time.Sleep(10 * time.Second)
		start := &api.Coordinate{X: int32(rand.Intn(int(utils.Settings.GridSize))), Y: int32(rand.Intn(int(utils.Settings.GridSize)))}
		end := &api.Coordinate{X: int32(rand.Intn(int(utils.Settings.GridSize))), Y: int32(rand.Intn(int(utils.Settings.GridSize)))}
		route := utils.CalculatePath(start, end, nil)
		routeCh <- &api.Route{Coordinates: route}
		log.Printf("Generated Random Route from %v to %v and path %v", start, end, route)
	}
}

func sendRoute(carinfo *api.CarInfo, route *api.Route) {
	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial(carinfo.Identifier, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	// Create a client instance.
	client := api.NewCarClientServiceClient(conn)

	// Send the car info request to the server.
	response, err := client.SendRoute(context.Background(), route)
	if err != nil {
		log.Fatalf("Failed to send route to car: %v", err)
	}
	log.Printf("Response from server: %s", response.Message)
}

func waitForUpdates(window *app.Window) bool {
	for {
		select {
		case carInfo := <-carInfoCh:
			var oldCarInfo = updateCarinfo(carInfo)
			updateGridData(oldCarInfo, carInfo)
			updateCarinfo(carInfo)
			window.Invalidate()
		case route := <-routeCh:
			updateGridDataRoute(route, "")
			go sendRouteWhenFree(carinfos, route)
			window.Invalidate()
		}
	}
}

func sendRouteWhenFree(carInfos []*api.CarInfo, route *api.Route) {
	freeCar := findCarWithShortestPath(carInfos, route)
	sendRoute(freeCar, route)
}

func findCarWithShortestPath(carInfos []*api.CarInfo, route *api.Route) *api.CarInfo {
	if len(route.Coordinates) == 0 {
		return nil
	}
	startPoint := route.Coordinates[0]

	for {
		carinfoMutex.Lock()
		var shortestCar *api.CarInfo
		shortestLength := math.MaxFloat64

		for _, carInfo := range carInfos {
			if !carInfo.ActiveRoute {
				dist := utils.Distance(carInfo.Position, startPoint)
				if dist < shortestLength {
					shortestLength = dist
					shortestCar = carInfo
				}
			}
		}

		if shortestCar != nil {
			shortestCar.ActiveRoute = true
			shortestCar.Route = route
			updateGridDataRoute(route, shortestCar.Color)
			carinfoMutex.Unlock()
			log.Printf("Shortest Path to route: %v", shortestCar.Identifier)
			return shortestCar
		}
		carinfoMutex.Unlock()
		log.Println("No free car found, waiting for 1 second")
		time.Sleep(1 * time.Second)
	}
}

func updateCarinfo(newCarInfo *api.CarInfo) *api.CarInfo {
	carinfoMutex.Lock()
	defer carinfoMutex.Unlock()

	var oldCarInfo *api.CarInfo
	for i, car := range carinfos {
		if car.Identifier == newCarInfo.Identifier {
			// Save the old carinfo
			oldCarInfo = car

			// Update the carinfo slice with the new carinfo
			carinfos[i] = newCarInfo

			return oldCarInfo
		}
	}

	// Append new CarInfo if not found
	carinfos = append(carinfos, newCarInfo)
	return nil
}

func updateGridData(oldCarInfo *api.CarInfo, newCarInfo *api.CarInfo) {
	carinfoMutex.Lock()
	defer carinfoMutex.Unlock()

	if oldCarInfo != nil && gridData[oldCarInfo.Position.X][oldCarInfo.Position.Y][0] == utils.Settings.CarAscii {
		// Delete old position of car
		gridData[oldCarInfo.Position.X][oldCarInfo.Position.Y] = [2]string{utils.Settings.EmptyAscii, ""}
	}

	// If new field empty: Set CarAscii
	if gridData[newCarInfo.Position.X][newCarInfo.Position.Y][0] == utils.Settings.EmptyAscii {
		gridData[newCarInfo.Position.X][newCarInfo.Position.Y] = [2]string{utils.Settings.CarAscii, newCarInfo.Color}
	}
	// If new field route
	if gridData[newCarInfo.Position.X][newCarInfo.Position.Y][0] == utils.Settings.RouteAscii {
		var isRoute bool = false
		for _, coord := range newCarInfo.Route.Coordinates {
			if coord.X == newCarInfo.Position.X && coord.Y == newCarInfo.Position.Y {
				isRoute = true
				break
			}
		}
		// if field is coord of own route: Set CarAscii
		if isRoute {
			gridData[newCarInfo.Position.X][newCarInfo.Position.Y] = [2]string{utils.Settings.CarAscii, newCarInfo.Color}
			if oldCarInfo != nil {
				gridData[oldCarInfo.Position.X][oldCarInfo.Position.Y] = [2]string{utils.Settings.EmptyAscii, ""}
			}
		} else {
			// if field is not coord in own route: Set CarAndRouteAscii with color of old value
			gridData[newCarInfo.Position.X][newCarInfo.Position.Y] = [2]string{utils.Settings.CarAndRouteAscii, gridData[newCarInfo.Position.X][newCarInfo.Position.Y][1]}
		}
	}
	// If CarAndRouteAscii is set: Set old position to Route Ascii with old color and new position to new CarAscii
	if oldCarInfo != nil && gridData[oldCarInfo.Position.X][oldCarInfo.Position.Y][0] == utils.Settings.CarAndRouteAscii {
		gridData[oldCarInfo.Position.X][oldCarInfo.Position.Y] = [2]string{utils.Settings.RouteAscii, gridData[oldCarInfo.Position.X][oldCarInfo.Position.Y][1]}
	}
}

func updateGridDataRoute(route *api.Route, color string) {
	for _, coord := range route.Coordinates {
		gridData[coord.X][coord.Y] = [2]string{utils.Settings.RouteAscii, color}
	}
}

func Run() {

	go startServer()

	go generateRandomRoute()

	window := new(app.Window)

	go display(window)

	go waitForUpdates(window)

	app.Main()
}
