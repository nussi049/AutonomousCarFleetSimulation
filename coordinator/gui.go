package coordinator

import (
	"AutonomousCarFleetSimulation/utils"
	"image/color"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func display(window *app.Window) error {
	window.Option(app.Size(2500, 2500))

	theme := material.NewTheme()

	theme.Face = "monospace"

	var ops op.Ops

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
