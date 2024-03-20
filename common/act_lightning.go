package common

import (
	"fmt"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/stats"
)

type LightningAction struct {
	BaseAction
	damageFactor         float64
	damageAmount         int
	hitPositions         []geometry.Point
	caster               *Actor
	actionDescription    ActionName
}

func NewLightningAction(cause Cause, caster *Actor, hitPositions []geometry.Point, damage int) *LightningAction {
	return &LightningAction{
		BaseAction:           BaseAction{cause: cause, name: ActionNameLightning},
		hitPositions:         hitPositions,
		damageFactor:         1.0,
		caster:               caster,
		actionDescription:    "Lightning",
		damageAmount:         damage,
	}
}

func (d *LightningAction) Execute(context ActionContext) []Action {
	d.damageAmount = int(float64(d.damageAmount) * d.damageFactor)
	state := context.GameState
	currentMap := state.GetMap()

	openList := make(map[geometry.Point]bool)
	closedList := make(map[geometry.Point]bool)
	for _, pos := range d.hitPositions {
		if !d.isConductorAt(currentMap, pos) {
			continue
		}
		openList[pos] = true
	}
	popFromOpenList := func() geometry.Point {
		for pos := range openList {
			delete(openList, pos)
			return pos
		}
		return geometry.Point{}
	}
	for len(openList) > 0 {
		nextPos := popFromOpenList()
		moreAffectedActorPositions := currentMap.NeighborsAll(nextPos, func(p geometry.Point) bool {
			if !d.isConductorAt(currentMap, p) {
				return false
			}
			if p == d.caster.MapPosition() {
				return false
			}
			if _, inClosed := closedList[p]; inClosed {
				return false
			}
			if _, inOpen := openList[p]; inOpen {
				return false
			}
			return true
		})
		for _, pos := range moreAffectedActorPositions {
			openList[pos] = true
		}
		closedList[nextPos] = true
	}

	var hitPositions []geometry.Point
	var hitActions []Action
	for hitPos := range closedList {
		if currentMap.IsActorAt(hitPos) {
			actor := currentMap.ActorAt(hitPos)
			hitActions = append(hitActions, NewChangeResourceAction(NewActorCause(d.actionDescription, d.caster), actor, stats.Health, float64(-d.damageAmount)))
		}
		hitPositions = append(hitPositions, hitPos)
	}
	allAdjacentBarrelLocations := currentMap.GetFilteredAdjacentPositions(hitPositions, func (pos geometry.Point) bool {
		if !currentMap.IsObjectAt(pos) {
			return false
		}

		object := currentMap.ObjectAt(pos)
		if typedObject, isObject := object.(Object); isObject {
			if _, exploder := typedObject.(*ExplodingObject); exploder {
				return true
			}
		}
		return false
	})
	for _, pos := range allAdjacentBarrelLocations {
		object := currentMap.ObjectAt(pos)
		typedObject := object.(*ExplodingObject)
		typedObject.DamageReceived(5)
	}
	d.hitPositions = hitPositions
	casterPart := fmt.Sprintf("%s casts", d.caster.Name())
	if d.caster == state.GetPlayer() {
		casterPart = "You cast"
	}
	context.Msg(fmt.Sprintf("%s %s", casterPart, d.actionDescription))
	return hitActions
}

func (d *LightningAction) GetAnimations() []Animation {
	animation := NewLightningAnimation(d.caster, d.hitPositions)
	return OneAnimation(animation)
}

func (d *LightningAction) GetDamageFactor() float64 {
	return d.damageFactor
}

func (d *LightningAction) SetDamageFactor(factor float64) {
	d.damageFactor = factor
}

func (d *LightningAction) GetDamageDealt() int {
	return d.damageAmount
}

func (d *LightningAction) isConductorAt(currentMap *gridmap.GridMap[*Actor, *Item, Object], pos geometry.Point) bool {
	if !currentMap.Contains(pos) {
		return false
	}
	if currentMap.IsActorAt(pos) || currentMap.IsDownedActorAt(pos) {
		return true
	}
	if currentMap.IsItemAt(pos) {
		itemAt := currentMap.ItemAt(pos)
		if itemAt.IsConductor() {
			return true
		}
	}
	if currentMap.IsObjectAt(pos) {
		objectAt := currentMap.ObjectAt(pos)
		if objectAt.IsConductor() {
			return true
		}
	}
	return false
}
