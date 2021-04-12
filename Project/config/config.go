package config

type Direction int

const (
	Up   Direction = 1
	Down Direction = -1
	Stop Direction = 0
)

type RequestState int

const (
	None      RequestState = 0
	Order     RequestState = 1
	Comfirmed RequestState = 2
	Complete  RequestState = 3
)

type Behaviour int

const (
	Idle     Behaviour = 0
	DoorOpen Behaviour = 1
	Moving   Behaviour = 2
)

type ButttonType int

const (
	Cab      ButttonType = 0
	HallUp   ButttonType = 1
	HallDown ButttonType = 2
)

type Request struct {
	Floor  int
	Button ButttonType
}

type DistributorElevator struct {
	Id       string
	Floor    int
	Dir      Direction
	Requests [][]RequestState
	Behave   Behaviour
}
