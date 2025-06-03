package network

import (
	"autofat/config"
	"os/exec"
	"strings"
)

var (
	_containerNames []string
	_ifaceNames     []string
)

func InitNamespaceConfig(cfg config.Config) {
	clear(_containerNames)
	clear(_ifaceNames)

	for _, elev_cfg := range cfg.Elevators {
		_containerNames = append(_containerNames, elev_cfg.NetworkNamespace)
		_ifaceNames = append(_ifaceNames, elev_cfg.NetworkInterface)
	}
}

func CommandInNamespace(id int, command string) *exec.Cmd {
	command = "sudo ip netns exec " + _containerNames[id] + " " + command

	new_args := strings.Split(command, " ")
	return exec.Command(new_args[0], new_args[1:]...)
}

func RunOrPanicInNamespace(id int, command string) {
	command = "sudo ip netns exec " + _containerNames[id] + " " + command
	runOrPanic(command)
}
