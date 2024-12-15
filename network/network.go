package network

import (
	"autofat/config"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const FILE_NAME = "ports.cfg"

type Port struct {
	Protocol string
	Port     uint16
}

var _ports []Port

// Run shell command, and panic if it fails.
func runOrPanic(cmd string) {
	parsed_cmd := strings.Split(cmd, " ")
	err := exec.Command(parsed_cmd[0], parsed_cmd[1:]...).Run()
	if err != nil {
		panic(fmt.Sprintf("Failed to execute command %s: %s", cmd, string(err.Error())))
	}
}

func Init(dir string, cfg config.Config) {
	clearIPTables()

	ports, err := loadPortsFromFile(dir)
	if err != nil {
		fmt.Println("Warning: could not load ports from ports.cfg: ", err.Error())
	}

	systemPorts := getSystemPorts(cfg)

	for _, porta := range systemPorts {
		for _, portb := range ports {
			if porta == portb.Port {
				panic(fmt.Sprintf("Student program attempts to reserve system port %d!", porta))
			}
		}
	}

	_ports = ports
}

func clearIPTables() {
	fmt.Println("Clearing iptables rules")
	runOrPanic("sudo iptables -P INPUT ACCEPT")
	runOrPanic("sudo iptables -P FORWARD ACCEPT")
	runOrPanic("sudo iptables -P OUTPUT ACCEPT")
	runOrPanic("sudo iptables -t nat -F")
	runOrPanic("sudo iptables -t mangle -F")
	runOrPanic("sudo iptables -F")
	runOrPanic("sudo iptables -X")
}

func SetPacketLoss(percentage int) {
	if percentage < 0 || percentage > 100 {
		panic("Packet loss percentage must be between 0 and 100.")
	}

	//If we want no packet loss, we might as well clear iptables rules.
	if percentage == 0 {
		clearIPTables()
		return
	}

	//Format string as 0.X if 0<=x<=99, or as 1.00 if x==100
	//like this to avoid potential rounding errors, might be overkill.
	percentage_str := "1.00"
	if percentage < 100 {
		percentage_str = fmt.Sprintf("0.%d", percentage)
	}

	for _, port := range _ports {
		fmt.Printf("Setting packet loss on port %s:%s to %s\n", port.Protocol, strconv.Itoa(int(port.Port)), percentage_str)
		runOrPanic(fmt.Sprintf("sudo iptables -A INPUT -p %s --dport %s -m statistic --mode random --probability %s -j DROP",
			port.Protocol, strconv.Itoa(int(port.Port)), percentage_str))
	}
}

func getSystemPorts(cfg config.Config) []uint16 {
	var ports_reserved []uint16
	for _, cfg_elev := range cfg.SimulatorAddresses {
		ports_reserved = append(ports_reserved, cfg_elev.Port())
		ports_reserved = append(ports_reserved, cfg_elev.Port())
	}
	return ports_reserved
}

func loadPortsFromFile(dir string) ([]Port, error) {
	file, err := os.Open(dir + "/" + FILE_NAME)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ret []Port

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile("^(tcp|udp) ([0-9]{1,6})$")
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindAllStringSubmatch(line, 2)
		if matches == nil {
			return nil, fmt.Errorf("could not parse port info '%s'", line)
		}

		porti, err := strconv.Atoi(matches[0][2])
		if err != nil {
			return nil, err
		}

		ret = append(ret, Port{
			Port:     uint16(porti),
			Protocol: matches[0][1],
		})
		fmt.Println("Port", ret[len(ret)-1], "loaded from ports.cfg")
	}
	return ret, nil
}
