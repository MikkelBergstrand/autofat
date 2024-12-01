package elevio

type MotorDirection int

const (
	MD_Up   MotorDirection = 2
	MD_Down MotorDirection = 0
	MD_Stop MotorDirection = 1
)

func (md MotorDirection) String() string {
	switch md {
	case MD_Up:
		return "UP"
	case MD_Down:
		return "DOWN"
	case MD_Stop:
		return "STOP"
	default:
		return "ERRMD"
	}
}

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown ButtonType = 1
	BT_Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
	Value  bool
}
