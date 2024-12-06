#!/usr/bin/bash

# This genius script
# 1) creates 3 Linux virtual network namespaces (contexts)
# 2) a bridge (e.g. virtual network switch) in the main network context
# 3) peers (e.g. virtual cables) from the bridge to an endpoint in each namespace
# 4) changes the loopback interface in each virtual network namespace,
#    so that localhost works as expected.
# Run me as SUDO!

# Create namespaces
ip netns add container0 
ip netns add container1
ip netns add container2

# Create the bridge
ip link add v-net-0 type bridge
ip link set dev v-net-0 up

# Add peers (virtual cables)
ip link add veth-0 type veth peer name veth-0-br
ip link add veth-1 type veth peer name veth-1-br
ip link add veth-2 type veth peer name veth-2-br

# Move peer end-points to proper namespace, 
# and connect peer start-points to bridge.

ip link set veth-0 netns container0
ip link set veth-0-br master v-net-0
ip link set veth-1 netns container1
ip link set veth-1-br master v-net-0
ip link set veth-2 netns container2
ip link set veth-2-br master v-net-0

# Set IP adresses foreach container.
ip -n container0 addr add 10.10.0.1/24 dev veth-0
ip -n container1 addr add 10.10.0.2/24 dev veth-1
ip -n container2 addr add 10.10.0.3/24 dev veth-2

# Bring to life all interfaces.
ip -n container0 link set veth-0 up
ip -n container1 link set veth-1 up
ip -n container2 link set veth-2 up
ip link set veth-0-br up
ip link set veth-1-br up
ip link set veth-2-br up

# Configure gateway on bridge
ip addr add 10.0.0.10/24 dev v-net-0

# Configure endpoints routes to reach gateway
ip -n container0 route add default via 10.0.0.10
ip -n container1 route add default via 10.0.0.10
ip -n container2 route add default via 10.0.0.10

# Fix localhost in each namespace (route 127.0.0.0/8 to itself)
ip -n container0 addr add 127.0.0.1/8 dev lo
ip -n container1 addr add 127.0.0.1/8 dev lo
ip -n container2 addr add 127.0.0.1/8 dev lo

# NAT will translate 10.0.0.0, so that the network may reach
# the outside world (useful since some students poll 8.8.8.8 or such to check
# network connectivity)
iptables --table nat -A POSTROUTING -s 10.0.0.0/24 -j MASQUERADE

