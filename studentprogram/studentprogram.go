package studentprogram

import (
	"autofat/config"
	"autofat/procmanager"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
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
	Params     []string
}

var _studentPrograms map[int]StudentProgram

const CONFIG_FILENAME = "init.cfg"

func InitalizeFromConfig(programDir string, config []config.ElevatorConfig, nElevators int) {
	_studentPrograms = make(map[int]StudentProgram)
	data, err := os.ReadFile(programDir + "/" + CONFIG_FILENAME)
	if err != nil {
		log.Panic(err)
	}

	re := regexp.MustCompile("{PORT}")            //Replace PORT with actual port
	re2 := regexp.MustCompile(`^([^\s]*?) (.*)$`) //Parse command as executable + parameters
	commands := strings.Split(string(data), "\n")
	for i := 0; i < nElevators; i++ {
		cmdStr := re.ReplaceAllString(commands[i], strconv.Itoa((int)(config[i].UserAddrPort.Port())))
		matches := re2.FindStringSubmatch(cmdStr)

		prog := StudentProgram{
			Status:     RUNNING,
			Executable: matches[1],
			ProgramDir: programDir,
			Params:     strings.Split(matches[2], " "),
			Chan_Kill:  make(chan bool),
			Chan_Crash: make(chan bool),
		}
		_studentPrograms[i] = prog

		go runprocess(i)
	}
}

func runprocess(elevatorId int) {
	prog := _studentPrograms[elevatorId]
	//Launching with context so that we abort when the program aborts.
	cmd := exec.Command(prog.Executable, prog.Params...)

	//This *should* according to some online guides make it so that child processes are killed
	//with the parent, but it does not seem like os/exec respects this....
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:   true,
		Pdeathsig: syscall.SIGKILL,
	}

	cmd.Dir = prog.ProgramDir
	fmt.Println("Launching user process, as ", prog.Executable, prog.Params)
	err := cmd.Start()
	if err != nil {
		log.Panic("Could not launch user process: ", err)
	}

	fmt.Println("User process ", prog.Executable, prog.Params, "running with PID", cmd.Process.Pid)
	wasInterrupted := false

	//Create thread to listen for kill signal.
	go func() {
		for {
			<-prog.Chan_Kill
			fmt.Println("Ending student program", elevatorId)
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
		fmt.Println("Killing student", i)
		KillProgram(i)
		fmt.Println("Killed")
	}
}

func StartProgram(elevatorId int) {
	fmt.Println("Starting student program ", elevatorId)
	go runprocess(elevatorId)
}
