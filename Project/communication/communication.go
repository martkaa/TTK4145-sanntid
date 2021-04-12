package communication

import (
	"Project/elevator"
	"Project/elevio"
	"Project/network/bcast"
	"Project/network/peers"
)

/*
	"fmt"
	"os"
	"time"

	"../network/localip"
*/

/*
Interface of communication module are channels; ch_transmit, ch_receive, ch_peer_updateCh
*/

/*
The elevator struct can be sent over the network by writing to channel ch_transmit.
Elevator struct on the network can be read from channel ch_receive.
*/

/*
Note that all members we want to transmit must be public. Any private members will be received as zero-values.
*/

type CommunicationElevator struct {
	Id       int
	Floor    int
	Dir      elevio.MotorDirection
	Requests [elevator.NumFloors][elevator.NumButtons]int
	Behave   elevator.Behaviour
}

func CommunicationInit(ch_receive chan<- CommunicationElevator, ch_transmit chan<- CommunicationElevator) {

	/* Start the transmitter/receiver pair on some port*/
	go bcast.Transmitter(16569, ch_receive)
	go bcast.Transmitter(16569, ch_transmit)

}

func PeerUpdateInit(id string, ch_peerUpdate chan<- peers.PeerUpdate, ch_peerTxEnable chan bool) {

	/* Start the transmitter/receiver pair on some port*/
	go peers.Transmitter(15647, id, ch_peerTxEnable)
	go peers.Receiver(15647, ch_peerUpdate)
}
