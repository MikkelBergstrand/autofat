package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"
)

type ElevatorConfig struct {
	UserAddrPort       netip.AddrPort `json:"user_addr"`
	EvaulationAddrPort netip.AddrPort `json:"evaluation_addr"`
	NetworkNamespace   string         `json:"network_namespace"`
	NetworkInterface   string         `json:"iface"`
}

type LogConfig struct {
	Performance bool `json:"performance"`
	Events      bool `json:"events"`
	Network     bool `json:"network"`
	DSLOutput   bool `json:"dsl_output"`
}

type Config struct {
	StudentProgramDir          string            `json:"student_program"`
	TestFile                   string            `json:"test"`
	CompileParser              bool              `json:"compile_parser"`
	StudentProgramWaitTime     int               `json:"student_program_wait_time"`
	Elevators                  [3]ElevatorConfig `json:"elevators"`
	SimElevatorServerPath      string            `json:"sim_elevator_server_path"`
	LogStudentApplictionOutput bool              `json:"log_student_application_output"`
	Logging                    LogConfig         `json:"logs"`
}

func elevatorFlags(config *Config) {
	for i := 0; i < 3; i++ {
		//Namespaces
		def := fmt.Sprintf("container%d", i)
		flag.StringVar(&config.Elevators[i].NetworkNamespace, def, def,
			fmt.Sprintf("Name of network namespace %d", i))

		//IPs and ports
		config.Elevators[i].UserAddrPort = parseIpPortFlag(fmt.Sprintf("studaddr%d", i), config.Elevators[i].UserAddrPort)
		config.Elevators[i].EvaulationAddrPort = parseIpPortFlag(fmt.Sprintf("simaddr%d", i), config.Elevators[i].EvaulationAddrPort)
	}
}

func LoadFromFlags() Config {
	var config Config

	config_json, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(config_json, &config)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&config.TestFile, "test", config.TestFile, "Name of test file to be run. Lies in 'testfiles/$FILENAME'")
	flag.StringVar(&config.StudentProgramDir, "studentdir", config.StudentProgramDir, "sets directory of student program (relevant to the executing directory)")
	flag.StringVar(&config.SimElevatorServerPath, "simserverpath", config.SimElevatorServerPath, "path of the simulator executable.")
	flag.BoolVar(&config.CompileParser, "compile-parser", config.CompileParser, "Recreates the LR(1)-parser tables regardless of cache status.")
	flag.BoolVar(&config.LogStudentApplictionOutput, "log-stud-output", config.LogStudentApplictionOutput, "Logs the output of the student programs to logs/ directory")

	var wait_time_seconds int
	flag.IntVar(&wait_time_seconds, "studwaittime", config.StudentProgramWaitTime, "How many seconds to wait between launching student programs.")

	elevatorFlags(&config)

	flag.Parse()

	if config.StudentProgramDir == "" {
		panic("You must specify the student program directory with --studentdir!")
	}
	if config.TestFile == "" {
		panic("You must specify a test to be ran!")
	}

	if len(config.Elevators) != 3 {
		panic("Config must be specified for 3 elevators!")
	}
	return config
}

func (cfg Config) GetElevatorConfig(id int) ElevatorConfig {
	return cfg.Elevators[id]
}

func (cfg Config) GetAllElevatorConfigs() []ElevatorConfig {
	var out []ElevatorConfig
	for i := 0; i < 3; i++ {
		out = append(out, cfg.GetElevatorConfig(i))
	}
	return out
}

func parseIpPortFlag(flagName string, def netip.AddrPort) netip.AddrPort {
	var temp string
	flag.StringVar(&temp, flagName, def.String(),
		"Address and port of simulator "+flagName+". Format: addr:port, e.g. 127.0.0.1:9999")
	return netip.MustParseAddrPort(temp)
}
