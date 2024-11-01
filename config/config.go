package config

import "net/netip"

type ElevatorConfig struct {
	UserAddrPort netip.AddrPort
	FatAddrPort  netip.AddrPort
}