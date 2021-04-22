Elevator Project - TTK4145 Sanntidsprogrammering
================

Summary
-------
The task for this project was to create software for controlling `n` elevators working in parallel across `m` floors. We used a floating master with peer to peer and UDP broadcast to solve this problem.

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

**Feeting master**

The elevator that recieves an external order calculates the cost of every elevator based on their states, and thereby delegates the order to the most suitable elevator. This decition as well as the order is broadcasted to the network. This solution will handle the event of network loss of a node, such that as long as there exists elevators, the orders will always be redelegated to the most suitible node.

When an order is distributed, it is implicity acknowledged by the other elevators. Network loss will cause the connected elevatoros to reassign the orders of the lost elevator, and the disconnected elevator will go on as a individual elevator.

**UDP broadcast**

The idea to broadcast everything all the time will support our fleeting master, as opposed to a TCP wich requires a hand shake protocol. This means that every elevator knows everyone's states and orders at all times. Backup and restore of orders in case of network loss and power loss is also easy to handle with UDP. Every massage is ID'ed to differentiate between the messages.

The system perfomed great. For example packet loss is not a problem due to the continous spam of packets. However, there are improvements that could be done to enhance the readability of the program and code. One module tured out to be a "fix-it-all", and could be modifid to enhance the neatness of the system.

All in all, we are satisfied with our superduperelevator(s) which is so smooth we (almost) cannot believe it .

Simulator
---------

We ran the elevator(s) locally with the [simulator](https://github.com/TTK4145/Simulator-v2). 

Additional resources
--------------------

Go to [the project resources repository](https://github.com/TTK4145/Project-resources) to find more resources for doing the project. 
