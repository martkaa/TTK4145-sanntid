// Returnerer heis med lavest kost basert på et straffesystem/poeng
// Må deretter sorteres og delegeres
func costCalculator(order Request, elevList [NumElevators]Elevator, id int, onlineList [NumElevators]bool) int{
    if order.btn == BT_cab {
        return id
    }
    minCost := (NumButtons * NumFloors) * NumElevators
    bestElevator := id
    for elevator := 0; elevator < NumElevators; elevator++{
        if !onlineList[elevator] {
            // Neglect offline elevators
            continue
        }
        cost := order.Floor - elevList[elevator].Floor

        if cost = 0 && elevList[elevator].State != Moving {
            bestElevator = elevator
            return bestElevator
        }
        if cost < 0 {
            cost = -cost
            if elevList[elevator].Dir == MD_Up{
                cost += 3
            }
            
        } else if cost > 0 {
            if elevList[elevator].Dir == MD_Down{
                cost += 3
            }
        }
        if cost == 0 && elevList[elevator].State == Moving {
            cost += 4
        }
        if elevList[elevator].State == EB_DoorOpen{
            cost++
        }
        if cost < minCost {
            minCost = cost
            bestElevator = elevator
        }
    }
    return bestElevator
}

func distrubute(){
    var (
        elevList        [NumElevators]elev
        onlineList      [NumElevators]bool
        completedOrder  Keypress
    )
    completedOrder.DesignatedElevator = id
    elevList[id] = <-elevatorCh
    updateSyncCh <- elevList
}