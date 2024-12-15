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
	StudentProgramDir       string
	NoTests                 bool
	StudentProgramWaitTime  time.Duration
	NetworkNamespaces       [3]string
	SimulatorAddresses      [3]netip.AddrPort
	StudentProgramAddresses [3]netip.AddrPort
}

func LoadFromFlags() Config {
	var config Config

	config.SimulatorAddresses = [3]netip.AddrPort{
		netip.MustParseAddrPort("10.0.0.1:12345"),
		netip.MustParseAddrPort("10.0.0.2:12345"),
		netip.MustParseAddrPort("10.0.0.3:12345"),
	}

	config.StudentProgramAddresses = [3]netip.AddrPort{
		netip.MustParseAddrPort("10.0.0.1:12346"),
		netip.MustParseAddrPort("10.0.0.2:12346"),
		netip.MustParseAddrPort("10.0.0.3:12346"),
	}

	flag.StringVar(&config.StudentProgramDir, "studentdir", "", "sets directory of student program (relevant to the executing directory)")
	flag.BoolVar(&config.NoTests, "notests", false, "Only launches student programs / simulators. Do not run any test.")

	var wait_time_seconds int
	flag.IntVar(&wait_time_seconds, "studwaittime", 1, "How many seconds to wait between launching student programs (default=1 sec.).")

	for i := 0; i < 3; i++ {
		//Namespaces
		def := fmt.Sprintf("container%d", i)
		flag.StringVar(&config.NetworkNamespaces[i], def, def,
			fmt.Sprintf("Name of network namespace %d", i))

		//IPs and ports
		config.StudentProgramAddresses[i] = parseIpPortFlag(fmt.Sprintf("studaddr%d", i), config.StudentProgramAddresses[i])
		config.SimulatorAddresses[i] = parseIpPortFlag(fmt.Sprintf("simaddr%d", i), config.SimulatorAddresses[i])
	}

	flag.Parse()

	config.StudentProgramWaitTime = time.Second * time.Duration(wait_time_seconds)
	fmt.Println(config.StudentProgramWaitTime)
	return config
}

func (cfg Config) GetElevatorConfig(id int) ElevatorConfig {
	return ElevatorConfig{
		UserAddrPort:     cfg.StudentProgramAddresses[id],
		ExternalAddrPort: cfg.SimulatorAddresses[id],
	}
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
