#!/usr/bin/bash

# Undoes the changes made in network_context.sh

# Delete containers
ip netns del container0
ip netns del container1
ip netns del container2

# Delete bridge/peer interfaces
ip link delete veth-0-br
ip link delete veth-1-br
ip link delete veth-2-br
ip link delete v-net-0 

# Undo iptables stuff
iptables --table nat -D POSTROUTING -s 10.0.0.0/24 -j MASQUERADE

