package studentprogram

import (
	"autofat/config"
	"autofat/procmanager"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

const CONFIG_FILENAME = "init.cfg"

func InitalizeFromConfig(ctx context.Context, programDir string, config []config.ElevatorConfig, nElevators int) {
	data, err := os.ReadFile(programDir + "/" + CONFIG_FILENAME)
	if err != nil {
		log.Panic(err)
	}

	re := regexp.MustCompile("{PORT}")            //Replace PORT with actual port
	re2 := regexp.MustCompile(`^([^\s]*?) (.*)$`) //Parse command as executable + parameters
	commands := strings.Split(string(data), "\n")
	for i := range nElevators {
		cmdStr := re.ReplaceAllString(commands[i], strconv.Itoa((int)(config[i].UserAddrPort.Port())))

		matches := re2.FindStringSubmatch(cmdStr)
		go studentProcess(ctx, programDir, matches[1], strings.Split(matches[2], " "))
	}
}

func studentProcess(ctx context.Context, programDir string, cmdExecutable string, cmdParams []string) {
	//Launching with context so that we abort when the program aborts.
	cmd := exec.CommandContext(ctx, cmdExecutable, cmdParams...)

	//This *should* according to some online guides make it so that child processes are killed
	//with the parent, but it does not seem like os/exec respects this....
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}

	cmd.Dir = programDir
	fmt.Println("Launching user process, as ", cmdExecutable, cmdParams)
	err := cmd.Start()
	if err != nil {
		log.Panic("Could not launch user process: ", err)
	}

	fmt.Println("User process ", cmdExecutable, cmdParams, "running with PID", cmd.Process.Pid)
	procmanager.AddProcess(cmd.Process.Pid)

	result := cmd.Wait()
	fmt.Println("Exit code: ", result)

	procmanager.DeleteProcess(cmd.Process.Pid)

}
