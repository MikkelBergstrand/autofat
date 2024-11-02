package fatelevator

import (
	"autofat/config"
	"autofat/elevio"
	"autofat/tmux"
	"fmt"
	"os/exec"
	"strconv"
	"time"
)

const LAUNCH_SIMLATOR string = "./SimElevatorServer"

type SimulatedElevator struct {
	Config 			   config.ElevatorConfig
	Chan_FloorSensor   chan int
	Chan_ButtonPresser chan elevio.ButtonEvent
	Chan_OrderLights   chan elevio.ButtonEvent
	Chan_ProcessKiller chan int
	Chan_ProcessDone   chan int
	Chan_FloorLight    chan int
	Chan_Door          chan bool
}

func (instance *SimulatedElevator) Init(config config.ElevatorConfig) {
	instance.Chan_ProcessKiller = make(chan int)
	instance.Chan_ProcessDone = make(chan int)
	instance.Chan_FloorSensor = make(chan int)
	instance.Chan_ButtonPresser = make(chan elevio.ButtonEvent)
	instance.Chan_OrderLights = make(chan elevio.ButtonEvent)
	instance.Chan_FloorLight = make(chan int)
	instance.Chan_Door = make(chan bool)
	instance.Config = config
}

func RunSimulator(io *elevio.ElevIO, elevator SimulatedElevator) {

	fmt.Printf("Launching simulator process, port=%d, fatPort=%d\n", elevator.Config.UserAddrPort.Port(), elevator.Config.FatAddrPort.Port())
	cmd := exec.Command(LAUNCH_SIMLATOR, 
		"--port", strconv.Itoa(int(elevator.Config.UserAddrPort.Port())), 
		"--externalPort", strconv.Itoa(int(elevator.Config.FatAddrPort.Port())))
	tmux.LaunchInPane(cmd)

	//Wait for process to start, then init the IO interface
	time.Sleep(1 * time.Second)

	io.Init(fmt.Sprintf(":%d", elevator.Config.FatAddrPort.Port()), elevio.N_FLOORS)
	go io.PollFloorSensor(elevator.Chan_FloorSensor)
	go io.PollFloorLight(elevator.Chan_FloorLight)
	go io.PollDoor(elevator.Chan_Door)
	go io.PollOrderLights(elevator.Chan_OrderLights)

	go func() {
		for {
			new_button_press := <-elevator.Chan_ButtonPresser
			io.PressButton(new_button_press.Button, new_button_press.Floor)
		}
	}()
}
