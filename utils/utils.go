package utils

type Coordinate struct {
	X, Y int32
}

type CarInfo struct {
	Identifier  string
	Position    Coordinate
	Route       []Coordinate
	ActiveRoute bool
}
