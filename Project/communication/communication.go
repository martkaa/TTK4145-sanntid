package communication

import (
	"../distributor"
	"../network/bcast"
	"../network/peers"

	"flag"
)

/*
	"fmt"
	"os"
	"time"

	"../network/localip"
*/

/*
The elevator struct can be sent over the network by writing to channel ch_transmitter. Elevator struct on the network can be read from channel ch_receive.
*/

/*
Note that all members we want to transmit must be public. Any private members will be received as zero-values.
*/

func communicationInit() {

	/* Set id from command line using 'go run main.go -id=our_id'*/
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	/* Channels for sending and receiving elevator struct*/
	ch_receive := make(chan distributor.DistributorElevator)
	ch_transmit := make(chan distributor.DistributorElevator)

	/* Start the transmitter/receiver pair on some port*/
	go bcast.Transmitter(16569, ch_receive)
	go bcast.Transmitter(16569, ch_transmit)

}

func peerUpdateInit(peerUpdateCh chan<- peers.PeerUpdate) {

	/* We can disable/enable the transmitter after it has been started.*/
	/* This could be used to signal that we are somehow "unavailable".*/
	peerTxEnable := make(chan bool)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)
}
