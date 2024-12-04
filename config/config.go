package config

import (
	"flag"
	"fmt"
	"net/netip"
	"time"
)

type ElevatorConfig struct {
	UserAddrPort     netip.AddrPort
	ExternalAddrPort netip.AddrPort
}

type Config struct {
	StudentProgramDir      string
	NoTests                bool
	StudentProgramWaitTime time.Duration
}

func LoadFromFlags() Config {
	var config Config
	flag.StringVar(&config.StudentProgramDir, "studentdir", "", "sets directory of student program (relevant to the executing directory)")
	flag.BoolVar(&config.NoTests, "notests", false, "Only launches student programs / simulators. Do not run any test.")

	var wait_time_seconds int
	flag.IntVar(&wait_time_seconds, "studwaittime", 1, "How many seconds to wait between launching student programs (default=1 sec.).")
	flag.Parse()

	config.StudentProgramWaitTime = time.Second * time.Duration(wait_time_seconds)
	fmt.Println(config.StudentProgramWaitTime)
	return config
}
