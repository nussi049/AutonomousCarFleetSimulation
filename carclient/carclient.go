package carclient

import (
	"AutonomousCarFleetSimulation/utils"
	"fmt"
)

func Run() {
	car := utils.CarInfo{
		Identifier: "Test",
		PositionX:  0,
		PositionY:  0,
	}
	fmt.Printf("%+v\n", car)
}
