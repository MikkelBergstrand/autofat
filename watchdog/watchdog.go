package watchdog

import (
	"fmt"
	"os"
	"os/exec"
)

func KillerWatchdog(process *os.Process, chan_KillProcess chan int, chan_ProcessDone chan int) {
	for {
		select {
			case <-chan_ProcessDone:
				fmt.Println("Process is done, terminating killer watchdog")
				return
			case v := <-chan_KillProcess:
				fmt.Println("Watchdog killing process", v)
				process.Kill()
		}
	}
}

func AwaitProcessReturn(cmd *exec.Cmd, chan_ProcessDone chan int) {
	err := cmd.Run()

	//TODO: properly store command response code
	return_code := 0
	if err != nil {
		fmt.Println(err)
		return_code = 1
	}

	chan_ProcessDone <- return_code
}
