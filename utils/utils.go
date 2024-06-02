package utils

import (
	"AutonomousCarFleetSimulation/api"
	"math"
	"math/rand"
	"strings"
)

type DisplaySettings struct {
	GridSize   int
	FontSize   int
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

var Settings = DisplaySettings{
	GridSize:   12,
	FontSize:   10,
	EmptyAscii: createEmptyString(),
	CarAscii:   "  ______\n /|_||_\\.__\n(   _    _ _\\\n=`-(_)--(_)-'",
	RouteAscii: createSquare(),
}

// CreateDataGrid erstellt ein zweidimensionales Array von Strings
func CreateDataGrid() [][][2]string {
	gridData := make([][][2]string, Settings.GridSize)

	// Create empty datagrid
	for i := range gridData {
		gridData[i] = make([][2]string, Settings.GridSize)
		for j := range gridData[i] {
			gridData[i][j] = [2]string{Settings.EmptyAscii, "E"} // Standardfarbe 'E'
		}
	}
	return gridData
}

func Distance(start *api.Coordinate, end *api.Coordinate) float64 {
	return math.Abs(float64(start.X)-float64(end.X)) + math.Abs(float64(start.Y)-float64(end.Y))
}

func CalculatePath(start *api.Coordinate, end *api.Coordinate) []*api.Coordinate {
	var path []*api.Coordinate
	current := &api.Coordinate{
		X: start.X,
		Y: start.Y,
	}

	// Entscheide zufällig, ob zuerst horizontal oder vertikal bewegt werden soll
	moveHorizontalFirst := true
	if start.X != end.X && start.Y != end.Y {
		moveHorizontalFirst = rand.Intn(2) == 0
	}

	// Bewege zuerst horizontal, dann vertikal
	if moveHorizontalFirst {
		for current.X != end.X {
			if current.X < end.X {
				current.X++
			} else {
				current.X--
			}
			path = append(path, &api.Coordinate{
				X: current.X,
				Y: current.Y,
			})
		}

		for current.Y != end.Y {
			if current.Y < end.Y {
				current.Y++
			} else {
				current.Y--
			}
			path = append(path, &api.Coordinate{
				X: current.X,
				Y: current.Y,
			})
		}
	} else {
		// Bewege zuerst vertikal, dann horizontal
		for current.Y != end.Y {
			if current.Y < end.Y {
				current.Y++
			} else {
				current.Y--
			}
			path = append(path, &api.Coordinate{
				X: current.X,
				Y: current.Y,
			})
		}

		for current.X != end.X {
			if current.X < end.X {
				current.X++
			} else {
				current.X--
			}
			path = append(path, &api.Coordinate{
				X: current.X,
				Y: current.Y,
			})
		}
	}

	// Füge die Endkoordinate hinzu
	path = append(path, &api.Coordinate{
		X: end.X,
		Y: end.Y,
	})

	return path
}
