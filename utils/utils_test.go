package utils

import (
	"AutonomousCarFleetSimulation/api"
	"container/list"
	"fmt"
	"testing"
)

func TestBFSInitializationAndPathCalculation(t *testing.T) {
	tests := []struct {
		name         string
		start        *api.Coordinate
		end          *api.Coordinate
		avoidRoute   *api.Route
		expectedLen  int
		expectedPath []*api.Coordinate
	}{
		{
			name:  "Path with obstacle",
			start: &api.Coordinate{X: 0, Y: 0},
			end:   &api.Coordinate{X: 3, Y: 0},
			avoidRoute: &api.Route{Coordinates: []*api.Coordinate{
				{X: 1, Y: 0},
			}},
			expectedLen: 6, // Expected path length adjusted to 6
			expectedPath: []*api.Coordinate{
				{X: 0, Y: 0},
				{X: 0, Y: 1},
				{X: 1, Y: 1},
				{X: 2, Y: 1},
				{X: 3, Y: 1},
				{X: 3, Y: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test BFS-Initialisierung
			t.Run("BFS Initialization", func(t *testing.T) {
				start := tt.start
				queue := list.New()
				startStep := Step{Coord: start, Path: []*api.Coordinate{start}}
				queue.PushBack(startStep)
				visited := make(map[string]bool)
				visited[fmt.Sprintf("%d,%d", start.X, start.Y)] = true

				if queue.Len() != 1 {
					t.Errorf("expected queue length 1, got %d", queue.Len())
				}

				step := queue.Front().Value.(Step)
				if step.Coord.X != start.X || step.Coord.Y != start.Y {
					t.Errorf("expected start coordinate (%d, %d), got (%d, %d)", start.X, start.Y, step.Coord.X, step.Coord.Y)
				}

				if len(visited) != 1 {
					t.Errorf("expected visited length 1, got %d", len(visited))
				}

				if !visited["0,0"] {
					t.Errorf("expected start coordinate (0,0) to be marked as visited")
				}
			})

			// Test CalculatePath
			path := CalculatePath(tt.start, tt.end, tt.avoidRoute)
			if len(path) != tt.expectedLen {
				t.Errorf("expected path length %d, got %d", tt.expectedLen, len(path))
			}
			fmt.Printf("Actual path: ")
			for _, coord := range path {
				fmt.Printf("(%d, %d) ", coord.X, coord.Y)
			}
			fmt.Println()

			// Compare the actual path with the expected path
			for i, coord := range tt.expectedPath {
				if i >= len(path) || path[i].X != coord.X || path[i].Y != coord.Y {
					t.Errorf("expected path[%d] = (%d, %d), got (%d, %d)", i, coord.X, coord.Y, path[i].X, path[i].Y)
				}
			}
		})
	}
}
