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

type InitializationParams struct {
	InitialFloor  int
	BetweenFloors bool //Above InitialFloor. So if InitialFloor=1, we will be between 1 and 2.
}

type SimulatedElevator struct {
	Params             InitializationParams
	io                 *elevio.ElevIO
	Config             config.ElevatorConfig
	Chan_FloorSensor   chan int
	Chan_ButtonPresser chan elevio.ButtonEvent
	Chan_OrderLights   chan elevio.ButtonEvent
	Chan_FloorLight    chan int
	Chan_Door          chan bool
	Chan_Obstruction   chan bool
	Chan_Outofbounds   chan bool
	Chan_Kill          chan bool
}

func (instance *SimulatedElevator) Init(config config.ElevatorConfig, params InitializationParams) {
	instance.Chan_FloorSensor = make(chan int)
	instance.Chan_ButtonPresser = make(chan elevio.ButtonEvent)
	instance.Chan_OrderLights = make(chan elevio.ButtonEvent)
	instance.Chan_FloorLight = make(chan int)
	instance.Chan_Door = make(chan bool)
	instance.Chan_Obstruction = make(chan bool)
	instance.Chan_Outofbounds = make(chan bool)
	instance.Chan_Kill = make(chan bool)
	instance.Config = config
	instance.Params = params
}

func (elevator *SimulatedElevator) Run(tmuxPane int) {
	elevator.io = &elevio.ElevIO{}


	fmt.Printf("Launching simulator process, port=%d, fatPort=%d\n", elevator.Config.UserAddrPort.Port(), elevator.Config.FatAddrPort.Port())
	cmd := exec.Command(LAUNCH_SIMLATOR,
		"--port", strconv.Itoa(int(elevator.Config.UserAddrPort.Port())),
		"--externalPort", strconv.Itoa(int(elevator.Config.FatAddrPort.Port())),
		"--startFloor", strconv.Itoa(int(elevator.Params.InitialFloor)),
		"--randomStart", strconv.FormatBool(elevator.Params.BetweenFloors),
	)
	tmux.LaunchInPane(cmd, tmux.WINDOW_ELEVATORS, tmuxPane)

	//Wait for process to start, then init the IO interface
	time.Sleep(1 * time.Second)

	elevator.io.Init(fmt.Sprintf(":%d", elevator.Config.FatAddrPort.Port()), elevio.N_FLOORS, elevator.Chan_Kill)

	go elevator.io.PollFloorSensor(elevator.Chan_FloorSensor)
	go elevator.io.PollFloorLight(elevator.Chan_FloorLight)
	go elevator.io.PollDoor(elevator.Chan_Door)
	go elevator.io.PollOrderLights(elevator.Chan_OrderLights)
	go elevator.io.PollObstructionSwitch(elevator.Chan_Obstruction)
	go elevator.io.PollOutofbounds(elevator.Chan_Outofbounds)

	go func() {
		for {
			select {
			case new_button_press := <-elevator.Chan_ButtonPresser:
				elevator.io.PressButton(new_button_press.Button, new_button_press.Floor)
			case <-elevator.Chan_Kill:
				return
			}
		} 
	}()
}

func (elevator *SimulatedElevator) Terminate() {
	//Kill all polling channels.
	for i := 0; i < 7; i++ {
		elevator.Chan_Kill <- true
	}
	elevator.io.Close()
}
