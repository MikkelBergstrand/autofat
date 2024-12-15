# AutoFAT

This program is intended to automate evaluation of the TTK4145 FAT.

## Requirements

- Go (>v1.23)
- tmux (only tested compliant with default configuration)
- [dmd](https://dlang.org/download.html#dmd) (for building the simulator)

## Installation

First, clone the repository and change directory into it.

    git clone https://github.com/MikkelBergstrand/autofat && cd autofat

One must also install the modified `SimElevatorServer` program.

    git clone https://github.com/MikkelBergstrand/Simulator-v2 && cd Simulator-v2

Then, build it from source:

    dmd -w -g src/sim_server.d src/timer_event.d -ofSimElevatorServer

Give the executable execute permissions:

    chmod u+x SimElevatorServer

It is recommended to have the Simulator executable in the same directory as the
autofat respository. For instance, you can `cd` to your autofat directory and 
then make a symlink to the SimElevatorServer file:

    ln -s /path/to/SimElevatorServer SimElevatorServer

In order for the program to function as expected, some OS-level configuration is
necessary. First, we must establish the __networking context__, which is used
to essentially run each elevator program in its own "container":

    sudo ./network_context.sh

These changes can be undone at any time by rebooting or by running

    sudo ./destroy_network_context.sh

This program is not responsible for installing software necessary for launching
student programs. Users of this program should consult the official documentation
of the respective languages in question.

### Security bypass

To utilize the network context, one must bypass the sudo password prompts where 
it is necessary. Open the file using for instance vim:

    sudo vim /etc/sudoers

Here, add the following lines to the bottom of the file:

    yourusername  ALL=(ALL:ALL) NOPASSWD: /usr/sbin/ip
    yourusername  ALL=(ALL:ALL) NOPASSWD: /usr/sbin/iptables

This will allow your user to use ```iptables``` and ```ip``` without password,
so use with caution.

### Adding programs to sudo's $PATH

You probably want to be able to, for instance, run ```sudo go```, but this is
not possible (if you have a custom installation of Go, i.e. one that is not
installed with apt).
You can achieve this by adding directory paths to the ```secure_path``` variable
in ```/etc/sudoers```. For instance, here I have added the path ``/usr/local/go/bin``
to be able to launch go:

    Defaults        secure_path="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin:/usr/local/go/bin"

To add more directories, simply add a colon and then another path to the end of the string.

## Usage

To run the program, you obviously need a student application to test with. The program can
then be run inside the repository using

    go run main.go --studentdir="/path/to/student-application"

This will run all tests in sequential order. 

Extra command line parameters:
- `--container*` Name of network namespace * (* is 0, 1 or 2)
- `--simaddr*` IP address + port of simulator for the evaluation application * (* is 0, 1 or 2)
- `--studaddr*` IP address + port of simulator for the student application * (* is 0, 1 or 2)
- `--studwaittime` How many seconds to wait in between launching each student application instance.
- `--notests` Run no tests, just spawn three application + simulator instances. Useful for testing. 

The format of IP address + ports are of the format `IP:PORT`, e.g. `127.0.0.1:9999`. If the IP address
is omitted, it will default to `localhost`.
See these arguments again by running

    go run main.go --help
