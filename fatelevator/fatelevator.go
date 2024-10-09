package fatelevator

import (
	"autofat/elevio"
	"autofat/watchdog"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)


const LAUNCH_SIMLATOR string = "/home/mikkel/dev/prosjekt/autofat/SimElevatorServer"


type SimulatedElevator struct {
  Tty string 
  UserPort uint16
  FatPort uint16
  CurrentFloor uint8

  Chan_FloorSensor chan int
  Chan_ButtonPresser chan elevio.ButtonEvent
  Chan_ProcessKiller chan int
  Chan_ProcessDone chan int
}


func MakeElevatorInstance(userPort uint16, fatPort uint16, tty string, initialFloor uint8)  SimulatedElevator {
	var instance SimulatedElevator
	instance.Chan_ProcessKiller = make(chan int)
	instance.Chan_ProcessDone = make(chan int)
	instance.Chan_FloorSensor = make(chan int)
	instance.CurrentFloor = initialFloor
	instance.Tty = tty
  	instance.UserPort = userPort
  	instance.FatPort = fatPort
	return instance
}


func RunSimulator(elevator SimulatedElevator){
	fmt.Printf("Launching simulator process, port=%d, fatPort=%d\n", elevator.UserPort, elevator.FatPort)
	cmd := exec.Command(LAUNCH_SIMLATOR, "--port", strconv.Itoa(int(elevator.UserPort)), "--externalPort", strconv.Itoa(int(elevator.FatPort)))

	tty, err := os.OpenFile(elevator.Tty, os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
    defer tty.Close()

	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.Stdin = tty

	// Run the process
	go watchdog.KillerWatchdog(cmd.Process, elevator.Chan_ProcessKiller, elevator.Chan_ProcessDone)

  	go watchdog.AwaitProcessReturn(cmd, elevator.Chan_ProcessDone)

	//Wait for process to start, then init the IO interface
	time.Sleep(1 * time.Second)

	elevio.Init(fmt.Sprintf(":%d", elevator.FatPort), elevio.N_FLOORS)
	go elevio.PollFloorSensor(elevator.Chan_FloorSensor)
}
