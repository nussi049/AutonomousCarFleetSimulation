package coordinator

import (
	"AutonomousCarFleetSimulation/api"
	"AutonomousCarFleetSimulation/utils"
	"net"
	"strings"
	"sync"

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
)

var gridData = make([][]string, 8)

type DisplaySettings struct {
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
	EmptyAscii: createEmptyString(),
	CarAscii:   "  ______\n /|_||_\\.__\n(   _    _ _\\\n=`-(_)--(_)-'",
	RouteAscii: createSquare(),
}

type CoordinatorServiceServer struct {
	api.UnimplementedCoordinatorServiceServer
}

func (s *CoordinatorServiceServer) SendCarInfo(req *api.CarInfoRequest, srv api.CoordinatorService_SendCarInfoServer) error {
	// Extract CarInfo from the request and send it to the channel
	carInfo := utils.CarInfo{
		Identifier:  req.Identifier,
		Position:    utils.Coordinate{X: req.Position.X, Y: req.Position.Y},
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
		converted = append(converted, utils.Coordinate{X: c.X, Y: c.Y})
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

func Run() {

	go startServer()

	// Create empty datagrid
	for i := range gridData {
		gridData[i] = make([]string, 8)
		for j := range gridData[i] {
			gridData[i][j] = settings.EmptyAscii
		}
	}

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
		for carInfo := range carInfoCh {
			var oldCarInfo utils.CarInfo = updateCarinfo(carInfo)
			updateGridData(oldCarInfo, carInfo)
			window.Invalidate()
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
