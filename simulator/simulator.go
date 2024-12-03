package simulator

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
const NUM_CHANNELS = 10

var _simulators []Simulator
var _chan_Terminated chan bool = make(chan bool)

type InitializationParams struct {
	InitialFloor  int
	BetweenFloors bool //Above InitialFloor. So if InitialFloor=1, we will be between 1 and 2.
}

type Simulator struct {
	Params              InitializationParams
	io                  *elevio.ElevIO
	Config              config.ElevatorConfig
	Chan_FloorSensor    chan int
	Chan_ButtonPresser  chan elevio.ButtonEvent
	Chan_OrderLights    chan elevio.ButtonEvent
	Chan_FloorLight     chan int
	Chan_Door           chan bool
	Chan_Obstruction    chan bool
	Chan_SetObstruction chan bool
	Chan_Outofbounds    chan bool
	Chan_Direction      chan elevio.MotorDirection
	Chan_ToggleEngine   chan bool
	Chan_Kill           chan bool
}

func Init(config config.ElevatorConfig, params InitializationParams) {
	_simulators = append(_simulators, Simulator{})
	instance := &_simulators[len(_simulators)-1]

	instance.Chan_FloorSensor = make(chan int)
	instance.Chan_ButtonPresser = make(chan elevio.ButtonEvent)
	instance.Chan_OrderLights = make(chan elevio.ButtonEvent)
	instance.Chan_FloorLight = make(chan int)
	instance.Chan_Door = make(chan bool)
	instance.Chan_Obstruction = make(chan bool)
	instance.Chan_SetObstruction = make(chan bool)
	instance.Chan_Outofbounds = make(chan bool)
	instance.Chan_Kill = make(chan bool)
	instance.Chan_Direction = make(chan elevio.MotorDirection)
	instance.Chan_ToggleEngine = make(chan bool)
	instance.Config = config
	instance.Params = params

}

func Count() int {
	return len(_simulators)
}

func Get(id int) *Simulator {
	return &_simulators[id]
}

func Run(id int) {
	elevator := &_simulators[id]
	elevator.io = &elevio.ElevIO{}
	fmt.Printf("Launching simulator process, port=%d, externalPort=%d\n", elevator.Config.UserAddrPort.Port(), elevator.Config.ExternalAddrPort.Port())

	args := []string{
		"--port", strconv.Itoa(int(elevator.Config.UserAddrPort.Port())),
		"--externalPort", strconv.Itoa(int(elevator.Config.ExternalAddrPort.Port())),
		"--startFloor", strconv.Itoa(int(elevator.Params.InitialFloor)),
	}
	if elevator.Params.BetweenFloors {
		args = append(args, "--randomStart")
	}

	cmd := exec.Command(LAUNCH_SIMLATOR, args...)

	tmux.LaunchInPane(cmd, tmux.WINDOW_ELEVATORS, id)

	//Wait for process to start, then init the IO interface
	time.Sleep(1 * time.Second)

	elevator.io.Init(fmt.Sprintf(":%d", elevator.Config.ExternalAddrPort.Port()), elevio.N_FLOORS, elevator.Chan_Kill)

	go elevator.io.PollFloorSensor(elevator.Chan_FloorSensor)
	go elevator.io.PollFloorLight(elevator.Chan_FloorLight)
	go elevator.io.PollDoor(elevator.Chan_Door)
	go elevator.io.PollOrderLights(elevator.Chan_OrderLights)
	go elevator.io.PollObstructionSwitch(elevator.Chan_Obstruction)
	go elevator.io.PollOutofbounds(elevator.Chan_Outofbounds)
	go elevator.io.PollDirection(elevator.Chan_Direction)

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

	go func() {
		for {
			select {
			case engine_status := <-elevator.Chan_ToggleEngine:
				elevator.io.SetEngineState(engine_status)
			case <-elevator.Chan_Kill:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case obstruction := <-elevator.Chan_SetObstruction:
				elevator.io.SetObstruction(obstruction)
			case <-elevator.Chan_Kill:
				return
			}
		}
	}()
}

func TerminateAll() {
	for id := len(_simulators) - 1; id >= 0; id-- {
		go func() {
			//Kill all polling channels.
			for i := 0; i < NUM_CHANNELS; i++ {
				_simulators[id].Chan_Kill <- true
			}
			_simulators[id].io.Close()
			_chan_Terminated <- true
		}()
	}

	//Wait for all channels to close.
	terminate_count := 0
	for {
		<-_chan_Terminated
		terminate_count++
		if terminate_count == len(_simulators) {
			break
		}
	}
	//Clear the array
	_simulators = make([]Simulator, 0)
}

// Set whether or not the engine is working.
// Note that TRUE means the engine FAILS.
func SetEngineFailureState(id int, state bool) {
	_simulators[id].Chan_ToggleEngine <- state
}

func SetObstruction(id int, state bool) {
	_simulators[id].Chan_SetObstruction <- state
}

func MakeOrder(id int, btn elevio.ButtonType, floor int) {
	_simulators[id].Chan_ButtonPresser <- elevio.ButtonEvent{
		Button: btn,
		Floor:  floor,
	}
}
