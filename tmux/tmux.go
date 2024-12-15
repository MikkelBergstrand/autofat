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

func LaunchInPane(ex *exec.Cmd, window string, paneId int) {
	exec.Command("tmux", "select-window", "-t", window)

	err := exec.Command("tmux", "select-pane", "-t", strconv.Itoa(paneId)).Run()
	if err != nil {
		log.Fatal(err)
	}

	err = exec.Command("tmux", "send-keys", (strings.Join(ex.Args, " ")), "C-m").Run()
	if err != nil {
		log.Fatal(err)
	}
}

func Launch() {
	err := exec.Command("tmux", "a", "-t", "autofat").Run()
	if err != nil {
		fmt.Println("tmux session not found, creating new...")
		exec.Command("tmux", "new-session", "-d", "-s", "autofat").Run()
	}

	err = exec.Command("tmux", "select-window", "-t", WINDOW_ELEVATORS).Run()
	if err != nil {
		fmt.Println("tmux window not found, creating new...")
		exec.Command("tmux", "rename-window", WINDOW_ELEVATORS).Run()
	} else {
		//Kill leftover panes, if any.
		exec.Command("tmux", "kill-pane", "-t", "2").Run()
		exec.Command("tmux", "kill-pane", "-t", "1").Run()

		//Kill whatever is running in the 0th pane, if anything.
		exec.Command("tmux", "select-pane", "-t", "0").Run()
		exec.Command("tmux", "send-keys", "C-c").Run()
	}

	exec.Command("tmux", "split-window", "-h").Run()
	exec.Command("tmux", "split-window", "-v").Run()

}
