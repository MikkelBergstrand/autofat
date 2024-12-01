package elevio

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const _pollRate = 20 * time.Millisecond

type ElevIO struct {
	initialized bool
	numFloors   int
	mtx         sync.Mutex
	conn        net.Conn
	killChan    <-chan bool
}

func (io *ElevIO) Init(addr string, numFloors int, killChan <-chan bool) {
	if io.initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	io.numFloors = numFloors
	io.mtx = sync.Mutex{}
	var err error
	io.conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	io.initialized = true
	io.killChan = killChan
	fmt.Println("io initialized for addr", addr)
}

func (io *ElevIO) Close() {
	io.initialized = false
	io.conn.Close()
}

func (io *ElevIO) SetEngineState(state bool) {
	io.write([4]byte{15, toByte(state), 0, 0})
}

func (io *ElevIO) SetMotorDirection(dir MotorDirection) {
	io.write([4]byte{1, byte(dir), 0, 0})
}

func (io *ElevIO) SetButtonLamp(button ButtonType, floor int, value bool) {
	io.write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

func (io *ElevIO) PressButton(button ButtonType, floor int) {
	io.write([4]byte{10, byte(button), byte(floor), 0})
}

func (io *ElevIO) SetFloorIndicator(floor int) {
	io.write([4]byte{3, byte(floor), 0, 0})
}

func (io *ElevIO) SetDoorOpenLamp(value bool) {
	io.write([4]byte{4, toByte(value), 0, 0})
}

func (io *ElevIO) SetStopLamp(value bool) {
	io.write([4]byte{5, toByte(value), 0, 0})
}

func (io *ElevIO) PollOrderLights(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, io.numFloors)
	ticker := time.NewTicker(_pollRate)
	for {
		select {
		case <-ticker.C:
			for f := 0; f < io.numFloors; f++ {
				for b := ButtonType(0); b < 3; b++ {
					v := io.GetOrderLight(b, f)
					if v != prev[f][b] {
						receiver <- ButtonEvent{Floor: f, Button: ButtonType(b), Value: v}
					}
					prev[f][b] = v
				}
			}
		case <-io.killChan:
			return
		}
	}
}

func (io *ElevIO) PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, io.numFloors)
	ticker := time.NewTicker(_pollRate)

	for {
		select {
		case <-ticker.C:
			for f := 0; f < io.numFloors; f++ {
				for b := ButtonType(0); b < 3; b++ {
					v := io.GetButton(b, f)
					if v != prev[f][b] && v {
						receiver <- ButtonEvent{Floor: f, Button: ButtonType(b)}
					}
					prev[f][b] = v
				}
			}
		case <-io.killChan:
			return
		}
	}
}

func (io *ElevIO) PollFloorSensor(receiver chan<- int) {
	pollInt(receiver, io.killChan, io.GetFloor)
}

func (io *ElevIO) PollFloorLight(receiver chan<- int) {
	pollInt(receiver, io.killChan, io.GetFloorLight)
}

func (io *ElevIO) PollObstructionSwitch(receiver chan<- bool) {
	pollBool(receiver, io.killChan, io.GetObstruction)
}

func (io *ElevIO) PollDoor(receiver chan<- bool) {
	pollBool(receiver, io.killChan, io.GetDoor)
}

func (io *ElevIO) PollOutofbounds(receiver chan<- bool) {
	pollBool(receiver, io.killChan, io.GetOutofbounds)
}

func (io *ElevIO) PollDirection(receiver chan<- MotorDirection) {
	ticker := time.NewTicker(_pollRate)
	prev := MD_Stop
	for {
		select {
		case <-ticker.C:
			v := (MotorDirection)(io.GetDirection())
			if v != prev {
				receiver <- v
			}
			prev = v
		case <-io.killChan:
			return
		}
	}
}

func (io *ElevIO) GetButton(button ButtonType, floor int) bool {
	a := io.read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func (io *ElevIO) GetOrderLight(button ButtonType, floor int) bool {
	a := io.read([4]byte{13, byte(button), byte(floor), 0})
	return toBool(a[1])
}

func (io *ElevIO) GetFloor() int {
	a := io.read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func (io *ElevIO) GetFloorLight() int {
	a := io.read([4]byte{11, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

func (io *ElevIO) GetStop() bool {
	a := io.read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

func (io *ElevIO) GetObstruction() bool {
	a := io.read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

func (io *ElevIO) GetDoor() bool {
	a := io.read([4]byte{12, 0, 0, 0})
	return toBool(a[1])
}

func (io *ElevIO) GetOutofbounds() bool {
	a := io.read([4]byte{14, 0, 0, 0})
	return toBool(a[1])
}

func (io *ElevIO) GetDirection() byte {
	a := io.read([4]byte{16, 0, 0, 0})
	return a[1]
}

func (io *ElevIO) read(in [4]byte) [4]byte {
	io.mtx.Lock()
	defer io.mtx.Unlock()

	_, err := io.conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = io.conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

func (io *ElevIO) write(in [4]byte) {
	io.mtx.Lock()
	defer io.mtx.Unlock()

	_, err := io.conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

func pollInt(receiver chan<- int, done <-chan bool, caller func() int) {
	ticker := time.NewTicker(_pollRate)
	prev := -1
	for {
		select {
		case <-ticker.C:
			v := caller()
			if v != prev && v != -1 {
				receiver <- v
			}
			prev = v
		case <-done:
			return
		}
	}
}

func pollBool(receiver chan<- bool, done <-chan bool, caller func() bool) {
	prev := false
	ticker := time.NewTicker(_pollRate)
	for {
		select {
		case <-ticker.C:
			v := caller()
			if v != prev {
				receiver <- v
			}
			prev = v
		case <-done:
			return
		}
	}
}
