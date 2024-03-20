package common

import (
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/util"
)

type MoveAction struct {
	BaseAction
	actor                     *Actor
	targetPos                 geometry.Point
	originPos                 geometry.Point
	isPlayer                  bool
	animationDelay            float64
	isMovementVisibleToPlayer bool
	animationDisabled         bool
	moveCostFactor            float64
}

func (d *MoveAction) IsMovementVisibleToPlayer() bool {
	return d.isMovementVisibleToPlayer
}

func NewMoveAction(source Cause, actor *Actor, targetPos geometry.Point, isPlayer bool) *MoveAction {
	return &MoveAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameMove,
		},
		actor:          actor,
		targetPos:      targetPos,
		isPlayer:       isPlayer,
		animationDelay: 0.2,
		moveCostFactor: 1.0,
	}
}

func (d *MoveAction) SetAnimationDelay(delay float64) {
	d.animationDelay = delay
}

func (d *MoveAction) GetDelayInAut() int {
	autUsed := int(float64(100 / d.actor.GetMoveSpeed()) * d.moveCostFactor)
	return autUsed
}

// PROBLEM: Since all enemy actions are triggered in one tick,
// but the re-positioning of the actor takes place over multiple ticks,
// all the following actors will have a false idea of the world state.
// SOLUTION: The actor should be moved immediately, but the animation should be delayed.

func (d *MoveAction) Execute(context ActionContext) []Action {
	state := context.GameState
	gameMap := state.GetMap() // read state
	actor := d.actor
	d.originPos = actor.MapPosition()

	if gameMap.IsSpecialAt(d.originPos, gridmap.SpecialTileWater) || gameMap.IsSpecialAt(d.originPos, gridmap.SpecialTileLava) {
		d.moveCostFactor = 1.5
	}

	gameMap.MoveActor(actor, d.targetPos) // write state

	d.isMovementVisibleToPlayer = false
	if d.isPlayer {
		d.isMovementVisibleToPlayer = true
	} else {
		player := state.GetPlayer()
		d.isMovementVisibleToPlayer = player.CanSee(d.GetOldPos()) || player.CanSee(d.GetNewPos())
	}
	if d.isNotAnimated() {
		actor.SetDrawPosition(d.targetPos.ToPointF())
	}
	//println(fmt.Sprintf("Actor %s moved to %s", actor.Action(), d.targetPos))
	return nil
}

func (d *MoveAction) GetAnimations() []Animation {
	if d.isNotAnimated() {
		return nil
	}
	ticksNeeded := util.SecondsToTicks(d.animationDelay)
	return OneAnimation(&MovementAnimation{
		actor:       d.actor,
		path:        []geometry.Point{d.originPos, d.targetPos},
		ticksNeeded: ticksNeeded,
		isForPlayer: d.isPlayer,
	})
	//return &DummyAnimation{time: 5, cancelRunningAnimations: true}
}

func (d *MoveAction) isNotAnimated() bool {
	return !d.isMovementVisibleToPlayer || d.animationDisabled
}

func (d *MoveAction) GetActor() *Actor {
	return d.actor
}

func (d *MoveAction) GetOldPos() geometry.Point {
	return d.originPos
}

func (d *MoveAction) GetNewPos() geometry.Point {
	return d.targetPos
}

func (d *MoveAction) DisableAnimation() {
	d.animationDisabled = true
}

func PlayerMoveToAttack(actor *Actor, relativeDirection geometry.Point, forceAttack bool, context ActionContext) Action {
	state := context.GameState
	gameMap := state.GetMap()
	targetPos := actor.MapPosition().Add(relativeDirection)
	if !gameMap.Contains(targetPos) {
		return nil
	}
	isObjectAt := gameMap.IsObjectAt(targetPos)

	if isObjectAt {
		objectAtPos := gameMap.ObjectAt(targetPos)
		if state.IsClosedDoorAt(targetPos) {
			return NewOpenDoorAction(NewActorCause("Open door", actor), actor, targetPos)
		} else if !objectAtPos.IsWalkable(actor) {
			objectAtPos.ActorBumpedInto(actor, AnimateNow())
			return nil
		}
	}

	isActorAt := gameMap.IsActorAt(targetPos)
	isTargetFree := !isActorAt && gameMap.IsWalkable(targetPos)
	if isActorAt {
		actorAtPos := gameMap.ActorAt(targetPos)
		isAlly := actor.IsAlly(actorAtPos)
		if isAlly {
			return NewSwapPositionsAction(NewActorCause("Swap positions", actor), actor, actorAtPos, state.GetPlayer() == actor)
		} else if actorAtPos.IsVendor() && !forceAttack {
			return NewTalkToAction(NewActorCause("Talk to", actor), actorAtPos)
		} else {
			return NewMeleeAttackAction(NewActorCause("Melee attack", actor), actor, actorAtPos)
		}
	} else if isTargetFree {
		return NewMoveAction(NewActorCause("Move", actor), actor, targetPos, state.GetPlayer() == actor)
	}

	state.GetMap().SetExplored(targetPos)
	return nil
}

func AnimateNow() func() bool {
	return func() bool {
		return true
	}
}
