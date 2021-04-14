Todo:
-------
- Fix errors in communication module
- Handling of new internal request does not account for cab orders.
- Should Cost and updateDistributorElevators happen in the same case because they happen in order and concern the same thing.
- Remove prefix on functions in modules. E.g. "distributorUpdateInternalState" to "UpdateInternalState"


Suggestion
- Triggers for Distributor:
	I.	Receive local order.
	II.	Receive something on Network.
	III.	Receive state update on local elev.
	IV.	Timer on elevator runs out. E.i. error handling(This part is hard:().

Less important
- change "Behave" to "behaviour" 
- UpdateInternalState to localElevUpdate?


Questions - Fred
- If we only send single elevator on network, why is updatedElevators in distributorUpdate an vector and not a single elevator?
- What should happen when we start a new elevator? Should communication or Distributor handle this?
- What should happen when an elevator loses connection?Should communication or Distributor handle this? Both two last problem should be handled by the same module as some of the same work may have to be done.
- What does distributorUpdateInternalState do?
