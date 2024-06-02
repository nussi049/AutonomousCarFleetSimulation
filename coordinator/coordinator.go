package coordinator

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"context"
	"image/color"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"google.golang.org/grpc"
)

var (
	carinfos     = make([]*api.CarInfo, 0)
	carinfoMutex sync.Mutex
	carInfoCh    = make(chan *api.CarInfo)
	routeCh      = make(chan *api.Route)
	gridData     = utils.CreateDataGrid()
)

type CoordinatorServiceServer struct {
	api.CoordinatorServiceServer
}

func (s *CoordinatorServiceServer) SendCarInfo(ctx context.Context, req *api.CarInfo) (*api.CarInfoResponse, error) {
	// Send CarInfo to the channel
	carInfoCh <- req
	log.Printf("Car info received successfully from: %v", req.Identifier)

	// Return success message
	return &api.CarInfoResponse{
		Message: "Car info received successfully",
	}, nil
}

func startServer() {
	// Create a gRPC server
	server := grpc.NewServer()

	// Register your server implementation
	coordinatorServer := &CoordinatorServiceServer{}
	api.RegisterCoordinatorServiceServer(server, coordinatorServer)

	// Start the server on a specific port
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Server started")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

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

func Run() {

	go startServer()

	go generateRandomRoute()

	go func() {
		window := new(app.Window)
		err := display(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()

}

func display(window *app.Window) error {
	window.Option(app.Size(2500, 2500))

	theme := material.NewTheme()

	theme.Face = "monospace"

	var ops op.Ops
	go func() {
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
	}()

	// Event loop
	for {
		e := window.Event()
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			drawGrid(gtx, theme)
			e.Frame(gtx.Ops)
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

func drawGrid(gtx layout.Context, th *material.Theme) layout.Dimensions {
	var rows []layout.FlexChild
	for _, rowData := range gridData {
		row := rowData
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return drawRow(gtx, th, row)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func drawRow(gtx layout.Context, th *material.Theme, data [][2]string) layout.Dimensions {
	var widgets []layout.FlexChild
	for _, cell := range data {
		cellContent := cell[0]
		cellColor := cell[1]

		widgets = append(widgets, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := material.Body1(th, cellContent)
			label.TextSize = unit.Sp(utils.Settings.FontSize)
			label.Font.Weight = font.Bold

			var col color.NRGBA
			switch cellColor {
			case "Rot":
				col = color.NRGBA{R: 255, G: 0, B: 0, A: 255} // Rot
			case "Grün":
				col = color.NRGBA{R: 0, G: 255, B: 0, A: 255} // Grün
			case "Blau":
				col = color.NRGBA{R: 0, G: 0, B: 255, A: 255} // Blau
			case "Cyan":
				col = color.NRGBA{R: 0, G: 255, B: 255, A: 255} // Cyan
			case "Magenta":
				col = color.NRGBA{R: 255, G: 0, B: 255, A: 255} // Magenta
			case "Orange":
				col = color.NRGBA{R: 255, G: 165, B: 0, A: 255} // Orange
			case "Pink":
				col = color.NRGBA{R: 255, G: 192, B: 203, A: 255} // Pink
			case "Lila":
				col = color.NRGBA{R: 128, G: 0, B: 128, A: 255} // Lila
			case "Braun":
				col = color.NRGBA{R: 165, G: 42, B: 42, A: 255} // Braun
			default:
				col = color.NRGBA{R: 0, G: 0, B: 0, A: 255} // Schwarz
			}
			label.Color = col

			return label.Layout(gtx)
		}))
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, widgets...)
}
