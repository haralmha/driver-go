package main

import (
	"fmt"

	"./elevio"
)

var numFloors int = 4
var activeOrders [3][4]bool

type Elevator struct {
	floor int
	dir   elevio.MotorDirection
}

func shouldStop(floor int, dir elevio.MotorDirection) bool {
	if floor == 0 || floor == numFloors {
		return true
	}

	if activeOrders[elevio.BT_HallUp][floor] && (dir == elevio.MD_Up) {
		return true
	}
	if activeOrders[elevio.BT_HallDown][floor] && dir == elevio.MD_Down {
		return true
	}
	if activeOrders[elevio.BT_Cab][floor] {
		return true
	}

	return false
}

func anyOrders() bool {
	for i := elevio.BT_HallUp; i <= elevio.BT_Cab; i++ {
		for j := 0; j < numFloors; j++ {
			if activeOrders[i][j] {
				return true
			}
		}
	}
	return false
}

func setDir(floor int, prevDir int) elevio.MotorDirection {
	if !anyOrders() {
		return elevio.MD_Stop
	}
	// if floor == 0 && prevDir == int(elevio.MD_Down) {
	// 	return elevio.MD_Stop
	// }
	// if floor == numFloors-1 && prevDir == int(elevio.MD_Up) {
	// 	return elevio.MD_Stop
	// }
	// check for orders in direction of travel
	for i := floor + prevDir; 0 <= i && i < numFloors; i += prevDir {
		for j := elevio.BT_HallUp; j <= elevio.BT_Cab; j++ {
			if activeOrders[j][i] {
				return elevio.MotorDirection(prevDir)
			}
		}
	}
	//turn elevator around
	return elevio.MotorDirection(-prevDir)
}

func clearFloorOrders(floor int) {
	for i := elevio.BT_HallUp; i <= elevio.BT_Cab; i++ {
		activeOrders[i][floor] = false
		elevio.SetButtonLamp(i, floor, false)
	}
}

func main() {

	numFloors := 4
	var elev Elevator

	elevio.Init("localhost:15657", numFloors)

	for i := 0; i < numFloors; i++ {
		clearFloorOrders(i)
	}

	elev.dir = elevio.MD_Up
	elev.floor = 0
	var prevDir elevio.MotorDirection = elev.dir
	elevio.SetMotorDirection(elev.dir)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)
			activeOrders[a.Button][a.Floor] = true
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			elev.dir = setDir(a.Floor, int(prevDir))
			elevio.SetMotorDirection(elevio.MD_Stop)

		case floor := <-drv_floors:
			fmt.Printf("%+v\n", floor)
			elevio.SetFloorIndicator(floor)
			elev.floor = floor

			if shouldStop(floor, elev.dir) {
				clearFloorOrders(floor)
				prevDir = elev.dir
				elevio.SetMotorDirection(elevio.MD_Stop)
			}
			elev.dir = setDir(floor, int(prevDir))
			elevio.SetMotorDirection(elev.dir)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(elev.dir)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
