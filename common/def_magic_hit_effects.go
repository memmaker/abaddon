package common

import (
	"fmt"
	"math/rand"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/stats"
	"rpg/util"
)
// Spell Effects of Angband
// https://github.com/angband/angband/blob/master/lib/gamedata/activation.txt

// Monster Effects of Angband
// https://github.com/angband/angband/blob/master/lib/gamedata/blow_effects.txt
// https://github.com/angband/angband/blob/master/lib/gamedata/blow_methods.txt
// https://github.com/angband/angband/blob/master/lib/gamedata/monster_spell.txt
/*
Spell Effects of Moria
 - Sleep Monster
 - Detect Treasure
 - Detect Invisible
 - Detect Monsters
 - Light Area
 - Map Area
 - Aggravate Monster
 - Magic Bolt
 - Fireball
 - Recharge item
 - Drain life
 - Speed Monster
 - Confuse Monster
 - Stone to mud
 - Polymorph other
 - Build Wall
 - Clone Monster
 - Teleport Self / Other
 - (Mass) Genocide
 - Cure Status Effect
 - Earthquake
 - Destruction
 - Enchant Item
 - Remove curse
 - Warding glyph
 - Slow poison
 - Bless

From Brogue
 - Discord (Turn monsters against each other)


Angband Categories
 - Summon Monster
 - Slow/Haste
 - Blind
 - Scare
 - Short Beam of Elemental Damage - Ray with short/medium length
 - Bolt of Elemental Damage - One projectile
 - Orbs of Elemental Damage - Multiple projectiles?
 - Breath of Elemental Damage - Breath Cone


*/

func (g *Factory) GetRandomTargetedAction(randomSource *rand.Rand, dungeonLevel int) *TargetedAction {
	var allowed []defs.TargetedActionDefinition
	for _, targetedDef := range g.definitions.TargetedDefinitions {
		if targetedDef.DungeonLevel <= dungeonLevel {
			allowed = append(allowed, targetedDef)
		}
	}
	randomDefinition := allowed[randomSource.Intn(len(allowed))]
	tAction := g.NewTargetedActionFromDefinition(randomDefinition, nil)
	tAction.SetDungeonLevel(dungeonLevel)
	return tAction
}

func (g *Factory) confuseTarget(args util.Arguments) *HitEffect {
	turns := 5
	if !args.IsEmpty() {
		turns = util.ParseDice(args.Get(0)).Roll()
	}
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			if !currentMap.IsActorAt(target) {
				return
			}
			hitActor := currentMap.ActorAt(target)
			g.actionBroker.Execute(NewApplyStatusAction(NewActorCause("Confused", attacker), hitActor, StatusConfused, turns))
		},
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Confuse target for %d turns", turns)),
		},
	}
}

func (g *Factory) berserkTarget(args util.Arguments) *HitEffect {
	turns := 5
	if !args.IsEmpty() {
		turns = util.ParseDice(args.Get(0)).Roll()
	}
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			if !currentMap.IsActorAt(target) {
				return
			}
			hitActor := currentMap.ActorAt(target)
			g.actionBroker.Execute(NewApplyStatusAction(NewActorCause("Berserked", attacker), hitActor, StatusBerserk, turns))
		},
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Berserk target for %d turns", turns)),
		},
	}
}

func (g *Factory) poisonTarget(args util.Arguments) *HitEffect {
	turns := 3
	damagePerTurn := util.ParseDice("1d3")
	if !args.IsEmpty() {
		turns = util.ParseDice(args.Get(0)).Roll()
		damagePerTurn = util.ParseDice(args.Get(1))
	}
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			if !currentMap.IsActorAt(target) {
				return
			}
			hitActor := currentMap.ActorAt(target)
			poisonEffect := g.NewPoisonedEffect(turns, damagePerTurn.Roll())
			g.actionBroker.Execute(NewApplyStatusParametrizedAction(NewActorCause("Poisoned", attacker), hitActor, poisonEffect))
		},
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Poison target for %d turns, dealing %s damage per turn", turns, damagePerTurn.String())),
		},
	}
}
func (g *Factory) immobilizeTarget(args util.Arguments) *HitEffect {
	turns := args.GetInt(0)
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			if !currentMap.IsActorAt(target) {
				return
			}
			hitActor := currentMap.ActorAt(target)
			g.actionBroker.Execute(NewApplyStatusAction(NewActorCause("Poisoned", attacker), hitActor, StatusImmobile, turns))
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Immobilize target"),
		},
	}
}

func (g *Factory) throwFireball(args util.Arguments) *HitEffect {
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			action := NewDamageSpellAction(NewActorCause("Fireball", attacker), attacker, 10, target, hitPositions)
			g.actionBroker.Execute(action)
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Throw a fireball"),
		},
	}
}

func (g *Factory) projectFire(args util.Arguments) *HitEffect {
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			for _, hitPos := range hitPositions {
				g.actionBroker.Execute(NewSpawnAction(NewActorCause("Fire", attacker), "fire", func() {
					g.spawnFireAt(hitPos)
				}))
			}
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Create fire at target location"),
		},
	}
}

func (g *Factory) summonRandomMonster(args util.Arguments) *HitEffect {
	dlvl := util.ParseDice(args.Get(0)).Roll()
	return &HitEffect{
		Activate: func(caster *Actor, target geometry.Point, hitPositions []geometry.Point) {
			for _, hitPos := range hitPositions {
				g.actionBroker.Execute(NewSpawnAction(NewActorCause("Summon monster", caster), "monster", func() {
					newEnemy, _ := g.NewRandomEnemy(dlvl)
					currentMap := g.state.GetMap()
					if !currentMap.CanPlaceActorHere(hitPos) {
						freeCells := currentMap.GetFreeCellsForDistribution(hitPos,1, func(pos geometry.Point) bool {
							return currentMap.CanPlaceActorHere(pos)
						})
						if len(freeCells) == 0 {
							return
						}
						hitPos = freeCells[0]
					}
					currentMap.AddActor(newEnemy, hitPos)
					newEnemy.SetFaction(caster.GetFaction())
					newEnemy.SetDrawPosition(hitPos.ToPointF())

					timeKeeper := g.getTimeKeeper()
					timeKeeper.AddTimed(newEnemy)
					g.audioPlayer.PlayCue("king_spell")
					g.state.msg(fmt.Sprintf("%s summons %s", caster.Name(), newEnemy.Name()))
				}))
			}
		},
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Summon a random monster from dungeon level %d", dlvl)),
		},
	}
}

func (g *Factory) jumpStart(args util.Arguments) *HitEffect {
	maxSprint := args.GetIntOrDefault(0, 5)
	maxJump := args.GetIntOrDefault(1, 5)
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			jumpInfo := currentMap.JumpOverPositions(attacker.MapPosition(), target, maxSprint, maxJump)
			if !jumpInfo.ActorFound {
				return
			}
			isPlayer := attacker == g.state.GetPlayer()
			//attackedActor := currentMap.ActorAt(jumpInfo.JumpedPos)
			jumpExecution := g.actionBroker.Execute
			landingPos := jumpInfo.Jump[len(jumpInfo.Jump)-1]

			if len(jumpInfo.Sprint) > 0 {
				firstMoveTarget := jumpInfo.Sprint[len(jumpInfo.Sprint)-1]
				g.actionBroker.Execute(NewMoveAction(NewActorCause("JumpStart", attacker), attacker, firstMoveTarget, isPlayer))
				jumpExecution = g.actionBroker.QueueForExecution
			}

			jump := NewMoveAction(NewActorCause("JumpStart", attacker), attacker, landingPos, isPlayer)
			jump.DisableAnimation()
			jumpExecution(jump)

			if attacker.GetEquipment().HasRangedWeaponEquipped() {
				g.actionBroker.QueueForExecution(NewRangedAttackAction(NewActorCause("JumpStart", attacker), attacker, jumpInfo.JumpedPos))
			} else if currentMap.IsActorAt(jumpInfo.JumpedPos) {
				targetActor := currentMap.ActorAt(jumpInfo.JumpedPos)
				g.actionBroker.QueueForExecution(NewMeleeAttackAction(NewActorCause("JumpStart", attacker), attacker, targetActor))
			}

			actionCost := 100 / attacker.GetMoveSpeed()
			g.actionBroker.QueueForExecution(NewEndPlayerTurnAction(NewActorCause("JumpStart", attacker), actionCost))
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Athletic combat maneuver:"),
			util.LabelTextFromString("Jump over an enemy and attack from behind"),
		},
	}
}

func (g *Factory) spawnFireAt(location geometry.Point) {
	fire := NewFire(3)
	procFireAtLocation := func(location geometry.Point, startAnimating func() bool) {
		if g.state.GetMap().IsActorAt(location) {
			actor := g.state.GetMap().ActorAt(location)
			action := NewChangeResourceAction(AbstractCause("Fire"), actor, stats.Health, -3)
			action.SetAnimationStartPredicate(startAnimating)
			g.actionBroker.Execute(action)
		}
	}
	onTurn := func() {
		if !fire.IsAlive() {
			g.state.RemoveObject(fire)
			return
		}
		fire.DealDamage(nil)
	}
	fire.SetOnTurn(onTurn)
	fire.SetDealDamage(procFireAtLocation)
	g.getTimeKeeper().AddTimed(fire)

	g.state.SpawnObject(location, fire)
	procFireAtLocation(location, AnimateNow())
}

func (g *Factory) fireZone(args util.Arguments) *HitEffect {
	damageInterval := util.ParseDice(args.Get(0))
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			damagePerTile := damageInterval.Roll()
			damageSpellAction := NewDamageSpellAction(NewActorCause("Flask of Oil", attacker), attacker, damagePerTile, target, hitPositions)
			g.actionBroker.Execute(damageSpellAction)
		},
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Create a small fiery explosion,")),
			util.LabelTextFromString(fmt.Sprintf("dealing %s damage to all targets", damageInterval.String())),
		},
	}
}

func (g *Factory) stealItem(args util.Arguments) *HitEffect {
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			currentMap.ApplyToActorsAt(hitPositions, func(hitActor *Actor) {
				g.actionBroker.Execute(NewStealItemAction(NewActorCause("Steal", attacker), attacker, hitActor))
			})
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Steal item"),
		},
	}
}

func (g *Factory) stealItemAndTeleport(args util.Arguments) *HitEffect {
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			currentMap.ApplyToActorsAt(hitPositions, func(hitActor *Actor) {
				g.actionBroker.Execute(NewStealItemAction(NewActorCause("Steal", attacker), attacker, hitActor))
			})
			teleportAction := g.NewUntargetedActionFromName("random_near_teleport")
			teleportAction.Invoke(attacker, "StealAndTeleport")
			attacker.SetNeutralFaction()
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Steal item and teleport"),
		},
	}
}

func (g *Factory) applyStatusEffectToSelection(args util.Arguments) *HitEffect {
	//effect common.StatusEffect, turnCount int
	effectName := args.Get(0)
	turnCount := args.GetInt(1)
	if !IsValidStatusEffect(effectName) {
		g.logError(fmt.Sprintf("Unknown status effect: %s", effectName))
		return nil
	}
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			currentMap.ApplyToActorsAt(hitPositions, func(actor *Actor) {
				g.actionBroker.Execute(NewApplyStatusAction(NewActorCause("ApplyStatus", attacker), actor, StatusEffect(effectName), turnCount))
			})
		},
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Apply %s to target(s)", effectName)),
		},
	}
}

func (g *Factory) blink(args util.Arguments) *HitEffect {
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			isPlayer := g.state.GetPlayer() == attacker
			action := NewMoveAction(NewActorCause("Blink", attacker), attacker, target, isPlayer)
			action.DisableAnimation()
			g.actionBroker.Execute(action)
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Teleport to target location"),
		},
	}
}

func (g *Factory) mindPush(args util.Arguments) *HitEffect {
	pushDistance := args.GetIntOrDefault(0, 3)
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			g.queuePush(currentMap, hitPositions, attacker, pushDistance)
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Push target away"),
		},
	}
}
func (g *Factory) earthShaker(args util.Arguments) *HitEffect {
	maxSprint := args.GetIntOrDefault(0, 5)
	pushDist := args.GetIntOrDefault(1, 1)
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()

			if !currentMap.IsValidSprintTarget(attacker.MapPosition(), target, maxSprint) {
				return
			}

			isPlayer := attacker == g.state.player
			sprintAction := NewMoveAction(NewActorCause("EarthShaker", attacker), attacker, target, isPlayer)
			g.actionBroker.Execute(sprintAction)
			var attackActions []Action
			for _, pos := range hitPositions {
				if !currentMap.IsActorAt(pos) {
					continue
				}
				actorAt := currentMap.ActorAt(pos)
				meleeAttackAction := NewMeleeAttackAction(NewActorCause("EarthShaker", attacker), attacker, actorAt)
				meleeAttackAction.DisableAnimation()
				attackActions = append(attackActions, meleeAttackAction)
			}
			g.actionBroker.QueueForExecutionAtTheSameTime(attackActions)
			g.queuePush(currentMap, hitPositions, attacker, pushDist)
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Rush forward to a specified location,"),
			util.LabelTextFromString("dealing damage to all adjacent locations"),
		},
	}
}

func (g *Factory) queuePush(currentMap *gridmap.GridMap[*Actor, *Item, Object], hitPositions []geometry.Point, attacker *Actor, pushDistance int) {
	var knockbackActions []Action
	currentMap.ApplyToActorsAt(hitPositions, func(hitActor *Actor) {
		hitPos := hitActor.MapPosition()
		pushDirection := hitPos.Sub(attacker.MapPosition())
		knockbackAction := NewKnockbackAction(NewActorCause("Push", attacker), hitActor, pushDirection, pushDistance)
		knockbackActions = append(knockbackActions, knockbackAction)
	})
	g.actionBroker.QueueForExecutionAtTheSameTime(knockbackActions)
}

func (g *Factory) bridge(args util.Arguments) *HitEffect {
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			for _, pos := range hitPositions {
				if currentMap.IsObviousHazardAt(pos) {
					currentMap.SetTile(pos, defs.FloorTile)
				}
			}
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Bridge a hazard"),
		},
	}
}

func (g *Factory) breakFloor(args util.Arguments) *HitEffect {
	return &HitEffect{
		Activate: func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) {
			currentMap := g.state.GetMap()
			for _, pos := range hitPositions {
				if currentMap.IsTileWalkable(pos) {
					currentMap.SetTile(pos, defs.ChasmTile)
					if currentMap.IsActorAt(pos) {
						actor := currentMap.ActorAt(pos)
						g.actionBroker.Execute(NewFallIntoChasmAction(NewActorCause("Fall", attacker), actor))
					}
				}
			}
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Create a chasm"),
		},
	}
}