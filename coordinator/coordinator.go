package coordinator

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"context"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
	"google.golang.org/grpc"
)

var (
	carinfo      = make([]utils.CarInfo, 0)
	carinfoMutex sync.Mutex
	carInfoCh    = make(chan utils.CarInfo)
	routeCh      = make(chan []utils.Coordinate)
)

var gridData = make([][]string, 8)

type DisplaySettings struct {
	GridSize   int
	EmptyAscii string
	CarAscii   string
	RouteAscii string
}

func createEmptyString() string {
	height := 4
	width := 13

	var emptyString strings.Builder
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			emptyString.WriteByte(' ')

		}
		emptyString.WriteByte('\n')
	}

	return emptyString.String()
}

func createSquare() string {
	height := 4
	width := 13
	var square strings.Builder
	square.WriteByte('\n')

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			square.WriteByte('X')
		}
		if i != height-1 {
			square.WriteByte('\n')
		}

	}

	return square.String()
}

var settings = DisplaySettings{
	GridSize:   8,
	EmptyAscii: createEmptyString(),
	CarAscii:   "  ______\n /|_||_\\.__\n(   _    _ _\\\n=`-(_)--(_)-'",
	RouteAscii: createSquare(),
}

type CoordinatorServiceServer struct {
	api.CoordinatorServiceServer
}

func (s *CoordinatorServiceServer) SendCarInfo(req *api.CarInfoRequest, srv api.CoordinatorService_SendCarInfoServer) error {
	// Extract CarInfo from the request and send it to the channel
	carInfo := utils.CarInfo{
		Identifier:  req.Identifier,
		Position:    utils.Coordinate{X: req.Position.GetX(), Y: req.Position.GetY()},
		Route:       convertCoordinates(req.Route),
		ActiveRoute: req.ActiveRoute,
	}

	// Send CarInfo to the channel
	carInfoCh <- carInfo
	log.Println("Successfully sent CarInfo to channel")

	// Return success message
	return srv.Send(&api.CarInfoResponse{
		Message: "Car info received successfully",
	})
}

func convertCoordinates(coords []*api.Coordinate) []utils.Coordinate {
	var converted []utils.Coordinate
	for _, c := range coords {
		converted = append(converted, utils.Coordinate{X: c.GetX(), Y: c.GetY()})
	}
	return converted
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

func calculatePath(start, end utils.Coordinate) []utils.Coordinate {
	var path []utils.Coordinate
	current := start

	// Solange wir nicht am Ziel sind
	for current != end {
		path = append(path, current)

		// Zufällige Entscheidung, ob in X- oder Y-Richtung bewegt wird, sofern beide Richtungen möglich sind
		if current.X != end.X && current.Y != end.Y {
			if rand.Intn(2) == 0 {
				// Bewegung in X-Richtung
				if current.X < end.X {
					current.X++
				} else {
					current.X--
				}
			} else {
				// Bewegung in Y-Richtung
				if current.Y < end.Y {
					current.Y++
				} else {
					current.Y--
				}
			}
		} else if current.X != end.X {
			// Nur Bewegung in X-Richtung möglich
			if current.X < end.X {
				current.X++
			} else {
				current.X--
			}
		} else if current.Y != end.Y {
			// Nur Bewegung in Y-Richtung möglich
			if current.Y < end.Y {
				current.Y++
			} else {
				current.Y--
			}
		}
	}

	// Füge die Endkoordinate hinzu
	path = append(path, end)

	return path
}

func generateRandomRoute() {
	for {
		time.Sleep(20 * time.Second)
		start := utils.Coordinate{X: int32(rand.Intn(int(settings.GridSize))), Y: int32(rand.Intn(int(settings.GridSize)))}
		end := utils.Coordinate{X: int32(rand.Intn(int(settings.GridSize))), Y: int32(rand.Intn(int(settings.GridSize)))}
		route := calculatePath(start, end)
		routeCh <- route
		log.Printf("Generated Random Route from %v to %v and path %v", start, end, route)
	}
}

func convertRoute(coords []utils.Coordinate) []*api.Coordinate {
	apiCoords := make([]*api.Coordinate, len(coords))
	for i, coord := range coords {
		apiCoords[i] = &api.Coordinate{
			X: coord.X,
			Y: coord.Y,
		}
	}
	return apiCoords
}

func sendRoute(carinfo utils.CarInfo, route []utils.Coordinate) {
	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial(carinfo.Identifier, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	// Create a client instance.
	client := api.NewCarClientServiceClient(conn)

	// Create a car info request.
	request := &api.RouteRequest{
		Route: convertRoute(route),
	}

	// Send the car info request to the server.
	response, err := client.SendRoute(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to send route to car: %v", err)
	}
	log.Printf("Response from server: %s", response.Message)
}

func Run() {

	go startServer()

	// Create empty datagrid
	for i := range gridData {
		gridData[i] = make([]string, settings.GridSize)
		for j := range gridData[i] {
			gridData[i][j] = settings.EmptyAscii
		}
	}

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
				var oldCarInfo utils.CarInfo = updateCarinfo(carInfo)
				updateGridData(oldCarInfo, carInfo)
				window.Invalidate()
			case route := <-routeCh:
				updateGridDataRoute(route)
				sendRoute(carinfo[0], route) //TODO calculate which car has shortestPath to starting point of route
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

func updateCarinfo(newCarInfo utils.CarInfo) utils.CarInfo {
	carinfoMutex.Lock()
	defer carinfoMutex.Unlock()

	var oldCarInfo utils.CarInfo
	for i, car := range carinfo {
		if car.Identifier == newCarInfo.Identifier {
			// Save the old carinfo
			oldCarInfo = car

			// Update the carinfo slice with the new carinfo
			carinfo[i] = newCarInfo

			return oldCarInfo
		}
	}

	carinfo = append(carinfo, newCarInfo)

	return utils.CarInfo{}
}

func updateGridData(oldCarInfo utils.CarInfo, newCarInfo utils.CarInfo) {
	carinfoMutex.Lock()
	defer carinfoMutex.Unlock()

	// Delete old position of car
	gridData[oldCarInfo.Position.X][oldCarInfo.Position.Y] = settings.EmptyAscii
	// set new position of car
	gridData[newCarInfo.Position.X][newCarInfo.Position.Y] = settings.CarAscii
}

func updateGridDataRoute(route []utils.Coordinate) {
	for _, coord := range route {
		gridData[coord.Y][coord.X] = settings.RouteAscii
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

func drawRow(gtx layout.Context, th *material.Theme, data []string) layout.Dimensions {
	var widgets []layout.FlexChild
	for _, cell := range data {
		cell := cell
		widgets = append(widgets, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Body1(th, cell).Layout(gtx)
		}))
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, widgets...)
}
