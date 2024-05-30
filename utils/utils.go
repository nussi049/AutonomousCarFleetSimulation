package utils

import "math/rand"

type Coordinate struct {
	X, Y int32
}

type CarInfo struct {
	Identifier  string
	Position    Coordinate
	Route       []Coordinate
	ActiveRoute bool
}

func CalculatePath(start, end Coordinate) []Coordinate {
	var path []Coordinate
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
