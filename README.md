Elevator Project - TTK4145 Sanntidsprogrammering
================

Summary
-------
The task for this project was to create software for controlling `n` elevators working in parallel across `m` floors. We used a fleeting master with peer to peer and UDP broadcast to solve this problem.

Requirements
-----------------
There were some requirements for the elevator's behaviour. These are summarized in points below.

**- No orders are lost**
  * Should handle errors like packet loss, losing network connection entirely, software that crashes, and losing power for both hall and cab orders

**- Multiple elevators should be more efficient than one**
  * All about communication and distributing to the most suitable elevator

**- An individual elevator should behave sensibly and efficiently**
  * As we implemented in TTK4235

**- The lights and buttons should function as expected** 

There were some permitted assumptions:

- At least one elevator is always working normally
  - No multiple simultaneous errors: Only one error happens at a time, but the system must still return to a fully operational state after this error
- Recall that network packet loss is *not* an error in this context, and must be considered regardless of any other (single) error that can occur
- No network partitioning: There will never be a situation where there are multiple sets of two or more elevators with no connection between them
- Cab call redundancy with a single elevator is not required
  - Given assumptions **1** and **2**, a system containing only one elevator is assumed to be unable to fail

Our solution
-------------

We wrote our solution program in `GoLang`. We found that the channel feature together with the handeling of concurrency with `go routines` was very useful to solve the task. The main features to our solution is commented below.

**Fleeting master**

Every peer on the network stores the states of the other peers of the network in an array, donated elevators. The states containes the floor of an given elevator, the behaviour, the direction, the requests and the ID. The requests of an elevator is a `m x n` matrix containing a number corresponding to the states of the request. An 0 donates no order, 1 denotes an order, 2 denotes an confirmed order and 3 denotes completed order.

The elevator that recieves an local order calculates the cost of every elevator based on their states, and thereby delegates the order to the most suitable elevator. This decision is broadcasted to the network. This solution will handle the event of network loss of a node, such that as long as there exists elevators, the orders will always be redelegated to the most suitible elevator.

When an order is distributed, it is implicity acknowledged by the other elevators. Network loss will cause the connected elevatoros to reassign the orders of the lost elevator, and the disconnected elevator will go on as a individual elevator.

**UDP broadcast**

The idea to broadcast everything using UDP all the time. This will support our fleeting master. This means that every elevator knows everyone's states and orders at all times. Backup and restore of orders in case of network loss and power loss is also easy to handle with UDP, as all elevators knows the last state of every other elevator.

The system perfomed great. For example packet loss is not a problem due to the continous spam of packets. However, there are improvements that could be done to enhance the readability of the program and code. One module tured out to almost be a "fix-it-all", and could be modifed to enhance the neatness of the system.

All in all, we are satisfied with our superduperelevator(s) which is so smooth we (almost) cannot believe it .

Simulator
---------

We ran the elevator(s) locally with the [simulator](https://github.com/TTK4145/Simulator-v2). 

Additional resources
--------------------

Go to [the project resources repository](https://github.com/TTK4145/Project-resources) to find more resources for doing the project. 
