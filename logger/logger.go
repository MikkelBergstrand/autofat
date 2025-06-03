package logger

import (
	"autofat/config"
	"fmt"
)

type LogType int

var _config config.Config

const (
	EVENT LogType = iota
	NETWORK
	PERF
	DSL_OUTPUT
)

func shoudLog(logType LogType) bool {
	switch logType {
	case EVENT:
		return _config.Logging.Events
	case NETWORK:
		return _config.Logging.Network
	case PERF:
		return _config.Logging.Performance
	case DSL_OUTPUT:
		return _config.Logging.DSLOutput
	}
	return true
}
func Init(config config.Config) {
	_config = config
}

func Log(logType LogType, msg ...any) {
	if !shoudLog(logType) {
		return
	}
	fmt.Println(msg...)
}

func Logf(logType LogType, msg string, args ...any) {
	if !shoudLog(logType) {
		return
	}
	fmt.Printf(msg, args...)
}
