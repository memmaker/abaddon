package common

type TimedBeing interface {
	AddEnergy(energy int)
	RemoveEnergy(used int)
	IsAlive() bool
	GetBaseSpeed() int
	GetEnergy() int
}

type TimeKeeper struct {
	autCounter         uint64
	turnCounter        uint64
	timed              []TimedBeing
	hasEnergyForAction func(being TimedBeing) (int, func())
	onRemoveTimed      func(thing TimedBeing)
}

func NewTimeKeeper(hasEnergyForAction func(being TimedBeing) (int, func())) *TimeKeeper {
	return &TimeKeeper{
		autCounter:         0,
		turnCounter:        0,
		timed:              make([]TimedBeing, 0),
		hasEnergyForAction: hasEnergyForAction,
	}
}

func (t *TimeKeeper) SetTimed(timedBeings []TimedBeing) {
	t.timed = timedBeings
}
func (t *TimeKeeper) PlayerTakesTurn(playerAutUsed int) {
	t.autCounter += uint64(playerAutUsed)
	t.turnCounter++
	//println(fmt.Sprintf("Player takes turn %d using %d aut", t.turnCounter, playerAutUsed))
	//for _, actor := range t.timed {
	// PROBLEM: We are technically doing a recursive call of this function here.
	// Each time an after hook executes during our turn that then triggers another turn..

	// PROBLEM: Delayed teleport will trigger during this loop and a map change will occur,
	// which in turn will overwrite the list we are currently iterating over.

	// The solution, the map change should be queued and executed after the turn is over.
	// Hmmm... This should also occur when something dies, or when something is spawned.
	// Nope, death won't remove things from the list directly.. we do it here in the loop.
	// Any other events that could potentially change the list of active beings?

	for i := len(t.timed) - 1; i >= 0; i-- {
		timedBeing := t.timed[i]
		if !timedBeing.IsAlive() {
			t.removeAtIndex(i)
			continue
		}
		addedEnergy := int((float64(playerAutUsed) / 10.0) * float64(timedBeing.GetBaseSpeed()))
		timedBeing.AddEnergy(addedEnergy)
		for energyUsed, action := t.hasEnergyForAction(timedBeing); // init
		timedBeing.IsAlive() && action != nil; // condition
		energyUsed, action = t.hasEnergyForAction(timedBeing) { // executed after each iteration
			timedBeing.RemoveEnergy(energyUsed) // this does update status effects, such as poison, thus could kill the being..
			if timedBeing.IsAlive() {
				action()
			} else {
				t.removeAtIndex(i)
				break
			}
		}
	}
}

func (t *TimeKeeper) AddTimed(actor TimedBeing) {
	t.timed = append(t.timed, actor)
}

func (t *TimeKeeper) removeAtIndex(i int) {
	timedBeing := t.timed[i]
	t.timed = append(t.timed[:i], t.timed[i+1:]...)
	if t.onRemoveTimed != nil {
		t.onRemoveTimed(timedBeing)
	}
}
func (t *TimeKeeper) SetOnRemoveTimed(onRemoveTimed func(thing TimedBeing)) {
	t.onRemoveTimed = onRemoveTimed
}
func (t *TimeKeeper) ClearTrackedBeings(shouldRemoveBeing func(being TimedBeing) bool) []TimedBeing {
	var removed []TimedBeing
	for i := len(t.timed) - 1; i >= 0; i-- {
		if shouldRemoveBeing(t.timed[i]) {
			removed = append(removed, t.timed[i])
			t.removeAtIndex(i)
		}
	}
	return removed
}

func (t *TimeKeeper) GetBeingsByFilter(keep func(being TimedBeing) bool) []TimedBeing {
	var result []TimedBeing
	for i := len(t.timed) - 1; i >= 0; i-- {
		timedBeing := t.timed[i]
		if timedBeing.IsAlive() && keep(timedBeing) {
			result = append(result, timedBeing)
		}
	}
	return result
}

func (t *TimeKeeper) Remove(being TimedBeing) {
	for i := len(t.timed) - 1; i >= 0; i-- {
		if t.timed[i] == being {
			t.removeAtIndex(i)
			return
		}
	}
}
