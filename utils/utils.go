package utils

import (
	"AutonomousCarFleetSimulation/api"
	"container/list"
	"fmt"
	"math"
	"strings"
)

type DisplaySettings struct {
	GridSize         int
	FontSize         int
	EmptyAscii       string
	CarAscii         string
	RouteAscii       string
	CarAndRouteAscii string
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
	GridSize:   16,
	FontSize:   8,
	EmptyAscii: createEmptyString(),
	CarAscii:   "  ______\n /|_||_\\.__\n(   _    _ _\\\n=`-(_)--(_)-'",
	RouteAscii: createSquare(),
	CarAndRouteAscii: `X X  XX
	X|_||X\.__
	(XXX_XX_X_\
	=` + "`" + `-(_)--(_)-'`,
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

// Struktur zur Speicherung eines Schritts im Pfad
type Step struct {
	Coord *api.Coordinate
	Path  []*api.Coordinate
}

// Funktion zur Berechnung des Pfads von start nach end unter Vermeidung der avoidRoute
func CalculatePath(start *api.Coordinate, end *api.Coordinate, avoidRoute *api.Route) []*api.Coordinate {
	// Set zur schnellen Überprüfung der avoidRoute-Koordinaten
	avoidSet := make(map[string]bool)
	if avoidRoute != nil {
		for _, coord := range avoidRoute.Coordinates {
			key := fmt.Sprintf("%d,%d", coord.X, coord.Y)
			avoidSet[key] = true
		}
	}

	// BFS-Initialisierung
	queue := list.New()
	startStep := Step{Coord: start, Path: []*api.Coordinate{start}}
	queue.PushBack(startStep)
	visited := make(map[string]bool)
	visited[fmt.Sprintf("%d,%d", start.X, start.Y)] = true

	// BFS-Schleife
	for queue.Len() > 0 {
		element := queue.Front()
		step := element.Value.(Step)
		queue.Remove(element)

		current := step.Coord
		path := step.Path

		// Ziel erreicht
		if current.X == end.X && current.Y == end.Y {
			return path
		}

		// Bewegung in vier Richtungen
		directions := []struct {
			dx, dy int32
		}{
			{dx: 1, dy: 0},
			{dx: -1, dy: 0},
			{dx: 0, dy: 1},
			{dx: 0, dy: -1},
		}

		for _, dir := range directions {
			newX, newY := current.X+dir.dx, current.Y+dir.dy
			newCoord := &api.Coordinate{X: newX, Y: newY}
			key := fmt.Sprintf("%d,%d", newX, newY)

			// Überprüfen, ob die neue Koordinate in der avoidRoute liegt oder bereits besucht wurde
			if (newX != end.X || newY != end.Y) && avoidSet[key] {
				continue
			}
			if visited[key] {
				continue
			}

			// Neue Koordinate zum Pfad hinzufügen und in die Warteschlange einfügen
			newPath := append([]*api.Coordinate{}, path...)
			newPath = append(newPath, newCoord)
			queue.PushBack(Step{Coord: newCoord, Path: newPath})
			visited[key] = true
		}
	}

	// Kein Pfad gefunden
	return nil
}
