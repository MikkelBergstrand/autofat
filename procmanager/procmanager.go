package procmanager

//This package is used to manage processes, in our case, user programs.
//Specifically it handles dangling processes that are the result of
//a bad program termination. It also helps with cleaning up spawned
//child processes

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var _map map[int]bool

// Add a process that is to be managed
func AddProcess(i int) {
	fmt.Println("Process ", i, " is now managed")
	_map[i] = true
	writeToFile()
}

// Make process no longer managed (because it died on its own, hopefully.)
func DeleteProcess(i int) {
	fmt.Println("Process ", i, " is no longer managed")
	delete(_map, i)
	writeToFile()
}

// Helper function: cleanup by killing all children of the process as well.
func cleanupProcess(pid int) {
	fmt.Println("Cleaning up process ", pid)
	// Little bit of evil Linux hacking
	// This targets all processes with a Parent Process ID (PPID) = pid
	// This is not a catch-all thing, but for common student cases
	// like running their program in shell script loop, it works.
	err := exec.Command("kill", "--", "-"+strconv.Itoa(pid)).Run()
	if err != nil {
		fmt.Println("Could not cleanup process with ID ", pid)
	}
}

// Write the map to a file, binary format
func writeToFile() {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	err := e.Encode(_map)
	if err != nil {
		log.Panic("Could not encode the map for saving!")
	}

	err = os.WriteFile("looseProcesses", b.Bytes(), 0644)
	if err != nil {
		log.Panic("Could not encode map for saving!")
	}
}

// Delete and cleanup all managed processes.
func KillAll() {
	for i := range _map {
		cleanupProcess(i)
	}
	_map = make(map[int]bool)
	writeToFile()
}

func Init() {
	_map = make(map[int]bool)
	fileBytes, err := os.ReadFile("looseProcesses")
	if err != nil || len(fileBytes) == 0 {
		return
	}

	b := new(bytes.Buffer)
	b.Write(fileBytes)
	d := gob.NewDecoder(b)
	err = d.Decode(&_map)

	if err != nil {
		panic(err)
	}
	KillAll()
}
