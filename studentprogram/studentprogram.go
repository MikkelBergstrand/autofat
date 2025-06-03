package studentprogram

import (
	"autofat/config"
	"autofat/network"
	"autofat/procmanager"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ProgramStatus int

const (
	RUNNING    = iota
	TERMINATED //Terminated on purpose by test.
	CRASHED    //Crashed unexpectedly, due to bad student.
)

type StudentProgram struct {
	Status     ProgramStatus
	Chan_Kill  chan bool
	Chan_Crash chan bool
	ProgramDir string
	Executable string
}

var _studentPrograms map[int]StudentProgram

var _config config.Config
var _launchTime time.Time

const CONFIG_FILENAME = "init.cfg"

func InitalizeFromConfig(cfg config.Config, waitTime time.Duration, programDir string, config []config.ElevatorConfig, nElevators int) {
	_config = cfg
	_launchTime = time.Now()
	_studentPrograms = make(map[int]StudentProgram)
	data, err := os.ReadFile(programDir + "/" + CONFIG_FILENAME)
	if err != nil {
		log.Panic(err)
	}

	re := regexp.MustCompile("{PORT}") //Replace PORT with actual port
	commands := strings.Split(string(data), "\n")
	for i := 0; i < nElevators; i++ {
		cmdStr := re.ReplaceAllString(commands[i], strconv.Itoa((int)(config[i].UserAddrPort.Port())))

		prog := StudentProgram{
			Status:     RUNNING,
			Executable: cmdStr,
			ProgramDir: programDir,
			Chan_Kill:  make(chan bool),
			Chan_Crash: make(chan bool),
		}
		_studentPrograms[i] = prog
		go runprocess(i)
		time.Sleep(waitTime)
	}
}

func runprocess(elevatorId int) {
	prog := _studentPrograms[elevatorId]
	//Launching with context so that we abort when the program aborts.
	cmd := network.CommandInNamespace(elevatorId, prog.Executable)

	//This *should* according to some online guides make it so that child processes are killed
	//with the parent, but it does not seem like os/exec respects this....
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}

	cmd.Dir = prog.ProgramDir

	if _config.LogStudentApplictionOutput {
		fileName := fmt.Sprintf("logs/%d%d%d%d%d%d_%d",
			_launchTime.Year(), _launchTime.Month(), _launchTime.Day(), _launchTime.Hour(), _launchTime.Minute(), _launchTime.Second(), elevatorId)
		f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		cmd.Stdout = f
	}

	err := cmd.Start()
	if err != nil {
		log.Panic("Could not launch user process: ", err)
	}

	wasInterrupted := false

	//Create thread to listen for kill signal.
	go func() {
		for {
			<-prog.Chan_Kill
			wasInterrupted = true
			procmanager.KillProcess(cmd.Process.Pid)
		}
	}()
	procmanager.AddProcess(cmd.Process.Pid)
	cmd.Wait()

	if !wasInterrupted {
		//Program died unexpectedly. Report this as an event to the state manager.
		procmanager.DeleteProcess(cmd.Process.Pid)
		_studentPrograms[elevatorId] = prog
		prog.Status = CRASHED
		prog.Chan_Crash <- true
	} else {
		prog.Status = TERMINATED
		_studentPrograms[elevatorId] = prog
	}
}

func Get(elevatorId int) StudentProgram {
	return _studentPrograms[elevatorId]
}

func KillProgram(elevatorId int) {
	_studentPrograms[elevatorId].Chan_Kill <- true
}

func KillAll() {
	for i := range _studentPrograms {
		KillProgram(i)
	}
}

func StartProgram(elevatorId int) {
	go runprocess(elevatorId)
}
