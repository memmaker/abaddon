package common

type Visibility struct {
	visibleActors map[*Actor]bool
}

func (v *Visibility) IsVisible(actor *Actor) bool {
	_, exists := v.visibleActors[actor]
	return exists
}

func (v *Visibility) GetVisibleActors() map[*Actor]bool {
	return v.visibleActors
}

func (v *Visibility) Set(visible map[*Actor]bool) {
	v.visibleActors = visible
}

func (v *Visibility) GetVisibleEnemyCount(isEnemy func(*Actor) bool) int {
	count := 0
	for actor, _ := range v.visibleActors {
		if isEnemy(actor) {
			count++
		}
	}
	return count
}

func (v *Visibility) RemoveActor(victim *Actor) {
	delete(v.visibleActors, victim)
}

func (v *Visibility) GetVisibleEnemies(player *Actor) []*Actor {
	enemies := make([]*Actor, 0)
	for actor, _ := range v.visibleActors {
		if actor.IsHostileTo(player) {
			enemies = append(enemies, actor)
		}
	}
	return enemies

}

func NewVisibility() *Visibility {
	return &Visibility{
		visibleActors: make(map[*Actor]bool),
	}
}
