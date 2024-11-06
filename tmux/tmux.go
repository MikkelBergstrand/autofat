package tmux

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

const TMUX_SESSION_NAME string = "autofat"

const WINDOW_ELEVATORS string = "Elevators"
const WINDOW_PROGRAMS string = "Programs"

func Cleanup() {
	fmt.Println("Deleting previous tmux-session " + TMUX_SESSION_NAME + ", if any..")
	exec.Command("tmux", "kill-session", "-t", TMUX_SESSION_NAME).Run()
}

func LaunchInPane(ex *exec.Cmd, window string, paneId int) {
	exec.Command("tmux", "select-window", "-t", window)

	err := exec.Command("tmux", "select-pane", "-t", strconv.Itoa(paneId)).Run()
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command("tmux", "send-keys", (ex.Path + " " + strings.Join(ex.Args, " ")), "C-m").Run()
	if err != nil {
		log.Fatal(err)
	}
}

func Launch() {
	exec.Command("tmux", "new-session", "-d", "-s", "autofat").Run()
	exec.Command("tmux", "rename-window", WINDOW_ELEVATORS).Run()
	exec.Command("tmux", "split-window", "-h").Run()
	exec.Command("tmux", "split-window", "-v").Run()
}
