package common

/*
Status Effects
==============
  - Invisible (not visible)
  - Stunned (no actions)
  - Blinded (no vision, high chance to miss even when attacking the right direction)
  - Slow / Hasted (less / more actions)
  - Scared (run away)
  - Silenced (no magic) / Pacified (no attacks)
  - Poisoned (damage over time)
  - Confused (random actions)
  - Charmed (change faction)
  - Berserker Rage (random attacks) / Haywire (random magic)
  - Flying (can move over some obstacles, chance to dodge melee attacks)
  - Increased/reduced stats/resources temporarily
*/
func IsValidStatusEffect(name string) bool {
	switch StatusEffect(name) {
	case StatusInvisible, StatusStunned, StatusPoisoned, StatusImmobile, StatusBlinded, StatusSlow, StatusHasted, StatusScared, StatusSilenced, StatusPacified, StatusDamageOverTime, StatusConfused, StatusCharmed, StatusBerserk, StatusHaywire, StatusFlying, StatusWinCondition:
		return true
	}
	return false
}

type StatusEffect string

const (
	// binary
	StatusInvisible StatusEffect = "invisible" // cannot be seen
	StatusSeeInvisible StatusEffect = "see_invisible" // can see invisible things
	StatusStunned   StatusEffect = "stunned"   // cannot act
	StatusScared    StatusEffect = "scared"    // runs away
	StatusSilenced  StatusEffect = "silenced"  // cannot cast spells
	StatusPacified  StatusEffect = "pacified"  // cannot attack
	StatusImmobile  StatusEffect = "immobile"  // cannot move
	StatusConfused  StatusEffect = "confused"  // random actions
	StatusPoisoned  StatusEffect = "poisoned"  // random actions
	StatusCharmed   StatusEffect = "charmed"   // change faction
	StatusBerserk   StatusEffect = "berserk"   // random attacks
	StatusExplodeOnDeath   StatusEffect = "explode_on_death"   // explodes on death
	StatusHaywire   StatusEffect = "haywire"   // random magic
	StatusFlying    StatusEffect = "flying"    // can move over some obstacles, chance to dodge melee attacks
	// parametrized
	StatusBlinded        StatusEffect = "blinded"          // are really timed modifiers
	StatusSlow           StatusEffect = "slow"             // are really timed modifiers
	StatusHasted         StatusEffect = "hasted"           // are really timed modifiers
	StatusDamageOverTime StatusEffect = "damage over time" // timed resource changes
	StatusWinCondition   StatusEffect = "win condition"    // killing this enemy is part of the win condition
)

type ParametrizedStatusEffect struct {
	name StatusEffect
	nextTurn func(a *Actor)
	turnsLeft int
}
func NewParametrizedStatusEffect(name StatusEffect, nextTurn func(a *Actor), turnsLeft int) *ParametrizedStatusEffect {
	return &ParametrizedStatusEffect{name: name, nextTurn: nextTurn, turnsLeft: turnsLeft}
}
func (e *ParametrizedStatusEffect) NextTurn(a *Actor) {
	if e.turnsLeft <= 0 {
		return
	}
	e.nextTurn(a)
	e.turnsLeft--
}

func (e *ParametrizedStatusEffect) TurnsLeft() int {
	return e.turnsLeft
}

func (e *ParametrizedStatusEffect) GetName() string {
	return string(e.name)
}


