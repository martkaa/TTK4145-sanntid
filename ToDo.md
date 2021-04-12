Todo:
-------
-  Fix import cycle between communication and distributor by making a new Communication struct for sending and receiving. This struct may be similar or equal to the one in distributor, but the import cycle makes it not possible to use DistributorElevator in the communication-module. 