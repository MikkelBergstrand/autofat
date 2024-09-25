package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const TMUX_SESSION_NAME string = "autofat"

func Cleanup() {
	fmt.Println("Deleting previous tmux-session " + TMUX_SESSION_NAME + ", if any..")
	exec.Command("tmux", "kill-session",  "-t",  TMUX_SESSION_NAME).Run()
}

func Launch() {
	exec.Command("tmux", "new-session",  "-d", "-s", "autofat").Run()
	exec.Command("tmux", "rename-window", "Elevators").Run()
	exec.Command("tmux", "split-window", "-h").Run()
	exec.Command("tmux", "split-window", "-v").Run()
}

func ExecuteCommand(paneIndex int, args ...string) {
	exec.Command("tmux", "select-pane", "-t", strconv.Itoa(paneIndex)).Run()
	exec.Command("tmux", "send-keys", strings.Join(args, " "), "C-m").Run()
}
