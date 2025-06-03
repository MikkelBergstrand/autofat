package network

import (
	"autofat/config"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const FILE_NAME = "ports.cfg"

var _cfg config.Config

// Run shell command, and panic if it fails.
func runOrPanic(cmd string) {
	parsed_cmd := strings.Split(cmd, " ")
	err := exec.Command(parsed_cmd[0], parsed_cmd[1:]...).Run()
	if err != nil {
		panic(fmt.Sprintf("Failed to execute command %s: %s", cmd, string(err.Error())))
	}
}

func Init(dir string, cfg config.Config) {
	_cfg = cfg
	for i := range _cfg.Elevators {
		clearIPTables(i)
	}
}

func clearIPTables(elevator int) {
	fmt.Println("Clearing iptables rules")
	RunOrPanicInNamespace(elevator, "sudo iptables -P INPUT ACCEPT")
	RunOrPanicInNamespace(elevator, "sudo iptables -P FORWARD ACCEPT")
	RunOrPanicInNamespace(elevator, "sudo iptables -P OUTPUT ACCEPT")
	RunOrPanicInNamespace(elevator, "sudo iptables -t nat -F")
	RunOrPanicInNamespace(elevator, "sudo iptables -t mangle -F")
	RunOrPanicInNamespace(elevator, "sudo iptables -F") 
	RunOrPanicInNamespace(elevator, "sudo iptables -X") 
}

func SetPacketLoss(elevator int, percentage int) {
	if percentage < 0 || percentage > 100 {
		panic("Packet loss percentage must be between 0 and 100.")
	}

	//Format string as 0.X if 0<=x<=99, or as 1.00 if x==100
	//like this to avoid potential rounding errors, might be overkill.
	percentage_str := "1.00"
	if percentage < 100 {
		percentage_str = fmt.Sprintf("0.%d", percentage)
	}

	fmt.Printf("Setting packet loss on elevator %d to %d%%\n", elevator, percentage)
	RunOrPanicInNamespace(elevator, fmt.Sprintf("sudo iptables -I INPUT -p tcp --dport %s -j ACCEPT",
		strconv.Itoa(int(_cfg.Elevators[elevator].EvaulationAddrPort.Port()))))
	RunOrPanicInNamespace(elevator, "sudo iptables -A INPUT -i lo -j ACCEPT")
	RunOrPanicInNamespace(elevator,

		fmt.Sprintf("sudo iptables -A INPUT -m statistic --mode random --probability %s -j DROP", percentage_str))
}

func SetGlobalPacketLoss(elevators []int, percentage int) {
	for _, elev := range elevators {
		SetPacketLoss(elev, percentage)
	}
}
