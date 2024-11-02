package tmux

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

const TMUX_SESSION_NAME string = "autofat"

var _pane_index = 1

func Cleanup() {
	fmt.Println("Deleting previous tmux-session " + TMUX_SESSION_NAME + ", if any..")
	exec.Command("tmux", "kill-session", "-t", TMUX_SESSION_NAME).Run()
}

func LaunchInPane(ex *exec.Cmd) {
	err := exec.Command("tmux", "select-pane", "-t", strconv.Itoa(_pane_index)).Run()
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command("tmux", "send-keys", (ex.Path + " " + strings.Join(ex.Args, " ")), "C-m").Run()
	if err != nil {
		log.Fatal(err)
	}

	_pane_index++
}

func Launch() {
	exec.Command("tmux", "new-session", "-d", "-s", "autofat").Run()
	exec.Command("tmux", "rename-window", "Elevators").Run()
	exec.Command("tmux", "split-window", "-h").Run()
	exec.Command("tmux", "split-window", "-v").Run()
}
