package config

import (
	"flag"
	"net/netip"
)

type ElevatorConfig struct {
	UserAddrPort     netip.AddrPort
	ExternalAddrPort netip.AddrPort
}

type Config struct {
	StudentProgramDir string
	NoTests           bool
}

func LoadFromFlags() Config {
	var config Config
	flag.StringVar(&config.StudentProgramDir, "studentdir", "", "sets directory of student program (relevant to the executing directory)")
	flag.BoolVar(&config.NoTests, "notests", false, "Only launches student programs / simulators. Do not run any test.")
	flag.Parse()

	return config
}
