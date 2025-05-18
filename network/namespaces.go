package network

import (
	"autofat/config"
	"os/exec"
	"strings"
)

var (
	_containerNames []string
)

func InitNamespaceConfig(cfg config.Config) {
	clear(_containerNames)
	for _, elev_cfg := range cfg.Elevators {
		_containerNames = append(_containerNames, elev_cfg.NetworkNamespace)
	}
}

func CommandInNamespace(id int, command string, args []string) *exec.Cmd {
	commandStr := command + " " + strings.Join(args, " ")
	commandStr = "sudo ip netns exec " + _containerNames[id] + " " + commandStr

	new_args := strings.Split(commandStr, " ")
	return exec.Command(new_args[0], new_args[1:]...)
}
