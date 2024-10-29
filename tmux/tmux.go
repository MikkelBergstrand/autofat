package tmux

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var _tty_maps = make(map[int]string)

const TMUX_SESSION_NAME string = "autofat"

func Cleanup() {
	fmt.Println("Deleting previous tmux-session " + TMUX_SESSION_NAME + ", if any..")
	exec.Command("tmux", "kill-session", "-t", TMUX_SESSION_NAME).Run()
}

func Launch() {
	exec.Command("tmux", "new-session", "-d", "-s", "autofat").Run()
	exec.Command("tmux", "rename-window", "Elevators").Run()
	exec.Command("tmux", "split-window", "-h").Run()
	exec.Command("tmux", "split-window", "-v").Run()

	tmux_tty_info, err := exec.Command("tmux", "list-panes", "-F", "\"#{pane_index} #{pane_tty}\"").Output()
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile("^\"([0-9]) (.*)\"$")
	for _, line := range strings.Split(string(tmux_tty_info), "\n") {
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		pane_id, err := strconv.Atoi(matches[1])
		if err != nil {
			log.Fatal(err)
		}
		_tty_maps[pane_id] = matches[2]
	}
}

func GetTTYFromPane(pane_id int) string {
	return _tty_maps[pane_id]
}
