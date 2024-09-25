package main

import (
	"autofat/elevio"
	"autofat/tmux"
	"fmt"
	"log"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const LAUNCH_SIMLATOR string = "/home/mikkel/dev/prosjekt/autofat/SimElevatorServer"

var LAUNCH_PROGRAM_CMD = "go"
var LAUNCH_PROGRAM_DIR = "/home/mikkel/dev/prosjekt/sanntidsprosjekt"

type ElevatorConfig struct {
	UserAddrPort netip.AddrPort
	FatAddrPort    netip.AddrPort
}

func userProc(userPort uint16, id int) {
    cmd := exec.Command(LAUNCH_PROGRAM_CMD, "run", "main.go", ":" + strconv.Itoa(int(userPort)), strconv.Itoa(int(id)))
	cmd.Dir = LAUNCH_PROGRAM_DIR
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("Launching user process, port=%d\n", userPort)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	cmd.Wait()
	fmt.Println("done")
}

func commandProc(fatPort uint16) {
	// Demo command procedure: press cycling cab button every 5th second.

	ticker := time.NewTicker(5 * time.Second)
	elevio.Init(":"+strconv.Itoa(int(fatPort)), elevio.N_FLOORS)

	ch_ButtonEvent := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(ch_ButtonEvent)
	floor := 0

	for {
		select {
		case <-ticker.C:
			fmt.Println("Tick!", floor)
			elevio.PressButton(elevio.BT_Cab, floor)
			floor++
			if floor >= elevio.N_FLOORS {
				floor = 0
			}
		}
	}
}

func main() {
	N_ELEVATORS := 3
	USERPROGRAM_PORTS := [3]uint16{12345, 12346, 12347}
	FATPROGRAM_PORTS := [3]uint16{12348, 12349, 12350}

	LOCALHOST := [4]byte{127, 0, 0, 1}
	var elevators []ElevatorConfig

	// Create tmux environment, for display
	tmux.Cleanup()
	tmux.Launch()

	// Launch simulator servers, one in each tmux pane
	tmux.ExecuteCommand(1, LAUNCH_SIMLATOR, "--port", strconv.Itoa(int(USERPROGRAM_PORTS[0])), "--externalPort", strconv.Itoa(int(FATPROGRAM_PORTS[0])))
	time.Sleep(100 * time.Millisecond)
	tmux.ExecuteCommand(2, LAUNCH_SIMLATOR, "--port", strconv.Itoa(int(USERPROGRAM_PORTS[1])), "--externalPort", strconv.Itoa(int(FATPROGRAM_PORTS[1])))
	time.Sleep(100 * time.Millisecond)
	tmux.ExecuteCommand(3, LAUNCH_SIMLATOR, "--port", strconv.Itoa(int(USERPROGRAM_PORTS[2])), "--externalPort", strconv.Itoa(int(FATPROGRAM_PORTS[2])))

	for i := 0; i < N_ELEVATORS; i++ {
		elevators = append(elevators, ElevatorConfig{
			UserAddrPort: netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), USERPROGRAM_PORTS[i]),
			FatAddrPort:    netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), FATPROGRAM_PORTS[i]),
		})

		go userProc(elevators[i].UserAddrPort.Port(), i)

        if(i == 1) {
            go commandProc(elevators[i].FatAddrPort.Port())
        }

	}

    for {} //Wait :|
}
