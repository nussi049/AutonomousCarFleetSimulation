package coordinator

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

type DisplaySettings struct {
	EmptyAscii string
	CarAscii   string
	RouteAscii string
}

var gridData = make([][]string, 8)

func createEmptyString(carString string) string {
	lines := strings.Split(carString, "\n")
	height := len(lines)
	width := len(lines[0])

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
	EmptyAscii: createEmptyString("  ______\n /|_||_\\.__\n(   _    _ _\\\n=`-(_)--(_)-'"),
	CarAscii:   "  ______\n /|_||_\\.__\n(   _    _ _\\\n=`-(_)--(_)-'",
	RouteAscii: createSquare(),
}

func Run() {

	for i := range gridData {
		gridData[i] = make([]string, 8)
		for j := range gridData[i] {
			gridData[i][j] = settings.EmptyAscii
		}
	}

	fmt.Println("Car ASCII:")
	fmt.Println(settings.CarAscii)

	go func() {
		window := new(app.Window)
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()

}
func run(window *app.Window) error {
	window.Option(app.Size(2500, 2500))

	theme := material.NewTheme()

	theme.Face = "monospace"

	var ops op.Ops
	go func() {
		for {
			// Randomly choose a row and column to update
			row := rand.Intn(len(gridData))
			col := rand.Intn(len(gridData[row]))

			// Randomly switch between CarAscii and EmptyAscii
			if rand.Intn(2) == 0 {
				gridData[row][col] = settings.EmptyAscii
			} else {
				gridData[row][col] = settings.CarAscii
			}

			// Trigger a rerender of the app
			window.Invalidate()

			// Sleep for a while before the next update
			time.Sleep(1 * time.Second)
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
