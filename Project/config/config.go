package config

const NumFloors = 4
const NumButtons = 3

const NoOrder = -1

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

type ButtonType int

const (
	HallUp   ButtonType = 0
	HallDown ButtonType = 1
	Cab      ButtonType = 2
)



type Request struct {
	Floor  int
	Button ButtonType
}

type DistributorElevator struct {
	ID       string
	Floor    int
	Dir      Direction
	Requests [][]RequestState
	Behave   Behaviour
}

type CostRequest struct {
	Id         string
	Cost       int
	AssignedId string
	Req        Request
}

