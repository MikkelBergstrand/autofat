#!/bin/bash 

# Thsi script launches 3 elevator clients and servers (simulators), 
# ending up in a nice tmux-ed view.

SIMULATOR_PATH=SimElevatorServer
LAUNCH_DIRECTORY='/home/mikkel/dev/prosjekt/sanntidsprosjekt'
LAUNCH_PROGRAM='/snap/bin/go run main.go'

PORTS=(12345 12346 12347)
ELEVATOR_NAMES=('A' 'B' 'C')

tmux new-session -d -s autofat 
tmux rename-window 'ElevatorView'
tmux select-window -t autofat:1 
tmux send-keys "./${SIMULATOR_PATH} --port ${PORTS[0]}" 'C-m'

tmux split-window -h 
tmux send-keys "./${SIMULATOR_PATH} --port ${PORTS[1]}" 'C-m'

tmux split-window -v
tmux send-keys "./${SIMULATOR_PATH} --port ${PORTS[2]}" 'C-m'


for i in "${!PORTS[@]}"
do 
    cd $LAUNCH_DIRECTORY
    $LAUNCH_PROGRAM ":${PORTS[$i]}" ${ELEVATOR_NAMES[$i]} > /dev/null &
done

tmux -2 attach-session -t autofat
