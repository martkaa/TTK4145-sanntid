package main

import (
	"flag"

	"Project/config"
	"Project/distributor"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/fsm"
	"Project/network/bcast"
	"Project/network/peers"
	"Project/watchdog"
)

func main() {

	// Set id and port from command line using 'go run main.go -port=our_id -port=our_port'
	var port string
	flag.StringVar(&port, "port", "", "port of this peer")
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	elevio.Init("localhost:"+port, 4)

	// Channels
	ch_newLocalOrder := make(chan elevio.ButtonEvent, 100)
	ch_newLocalState := make(chan elevator.Elevator, 100)
	ch_msgFromNetwork := make(chan []config.DistributorElevator, 100)
	ch_msgToNetwork := make(chan []config.DistributorElevator, 100)
	ch_orderToLocal := make(chan elevio.ButtonEvent, 100)
	ch_peerUpdate := make(chan peers.PeerUpdate)
	ch_peerTxEnable := make(chan bool)
	ch_watchdogElevatorStuck := make(chan bool)
	ch_elevStuck := make(chan bool)
	ch_clearLocalHallOrders := make(chan bool)

	go fsm.Fsm(ch_orderToLocal, ch_newLocalState, ch_clearLocalHallOrders)
	go elevio.PollButtons(ch_newLocalOrder)

	// Goroutines for communication
	go bcast.Transmitter(16568, ch_msgToNetwork)
	go bcast.Receiver(16568, ch_msgFromNetwork)
	go peers.Transmitter(15647, id, ch_peerTxEnable)
	go peers.Receiver(15647, ch_peerUpdate)

	go watchdog.WatchdogElevatorStuck(5, ch_elevStuck, ch_watchdogElevatorStuck)

	go distributor.DistributorFsm(
		id,
		ch_newLocalOrder,
		ch_newLocalState,
		ch_msgFromNetwork,
		ch_msgToNetwork,
		ch_orderToLocal,
		ch_peerUpdate,
		ch_watchdogElevatorStuck,
		ch_elevStuck,
		ch_clearLocalHallOrders)

	select {}
}
