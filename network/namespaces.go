package network

import (
	"os/exec"
	"strings"
)

var (
	_containerNames []string
)

func InitNamespaceConfig(containerNames []string) {
	_containerNames = containerNames
}

func CommandInNamespace(id int, command string, args []string) *exec.Cmd {
	commandStr := command + " " + strings.Join(args, " ")
	commandStr = "sudo ip netns exec " + _containerNames[id] + " " + commandStr

	new_args := strings.Split(commandStr, " ")
	return exec.Command(new_args[0], new_args[1:]...)
}
