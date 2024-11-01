package studentprogram

import (
	"autofat/config"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const CONFIG_FILENAME = "init.cfg"

func InitalizeFromConfig(programDir string, config []config.ElevatorConfig, nElevators int) {
	data, err := os.ReadFile(programDir + "/" + CONFIG_FILENAME); if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile("{PORT}") //Replace PORT with actual port
	re2 := regexp.MustCompile(`^([^\s]*?) (.*)$`) //Parse command as executable + parameters
	commands := strings.Split(string(data), "\n")
	for i := range nElevators {
		cmdStr := re.ReplaceAllString(commands[i], strconv.Itoa((int)(config[i].UserAddrPort.Port())))

		matches := re2.FindStringSubmatch(cmdStr)
		go userProcess(programDir, matches[1], strings.Split(matches[2], " "))
	}
}

func userProcess(programDir string, cmdExecutable string, cmdParams []string) {
	cmd := exec.Command(cmdExecutable, cmdParams...)
	cmd.Dir = programDir

	fmt.Println("Launching user process, as ", cmdExecutable, cmdParams)
	if err := cmd.Run(); err != nil {
		fmt.Println("User process failed!")
		fmt.Println(err)
	}
}
