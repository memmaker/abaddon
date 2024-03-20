package common

import (
	"fmt"
	"math/rand"
	"rpg/defs"
	"rpg/stats"
	"time"
)

type ChangeResourceAction struct {
	BaseAction
	changeFactor         float64
	amount               float64
	actor                *Actor
	changedByAmount      float64
	resourceName         stats.Name
	shouldStartAnimation func() bool
	animations           []Animation
}

func NewChangeResourceAction(cause Cause, actor *Actor, resourceName stats.Name, amount float64) *ChangeResourceAction {
	return &ChangeResourceAction{
		BaseAction:   BaseAction{cause: cause, name: ActionNameChangeResource},
		actor:        actor,
		amount:       amount,
		changeFactor: 1.0,
		resourceName: resourceName,
	}
}
func (d *ChangeResourceAction) Execute(context ActionContext) []Action {
	receiver := d.actor
	levelName := context.GameState.GetMap().GetName()
	isHellTown := levelName == string(defs.LevelNameTownHell)

	// 7drl hack
	if receiver.IsWinCondition() && !isHellTown {
		context.Msg(fmt.Sprintf("%s is not affected by any damage!", receiver.Name()))
		if rand.Intn(3) == 0 {
			context.Speak(receiver, defs.SpeechTriggerHurtInTown)
		}
		return nil
	}

	d.changedByAmount = float64(d.amount) * d.changeFactor

	receiver.GetStats().IncrementResourceBy(d.resourceName, d.changedByAmount)

	if d.changedByAmount < 0 { // this needs to be synched to the animation

		receiver.SetHurt()
	}

	d.animations = d.prepareAnimations(context)
	return nil
}

func (d *ChangeResourceAction) GetChangeFactor() float64 {
	return d.changeFactor
}

func (d *ChangeResourceAction) SetChangeFactor(factor float64) {
	d.changeFactor = factor
}
func (d *ChangeResourceAction) GetActor() *Actor {
	return d.actor
}

func (d *ChangeResourceAction) GetAnimations() []Animation { // TODO: Not allowed to create new animations here..
	return d.animations
}

func (d *ChangeResourceAction) prepareAnimations(context ActionContext) []Animation {
	randomSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	var animations []Animation
	amount := int(d.changedByAmount)
	if d.resourceName == stats.Health && amount != 0 {
		numberAnimation := NewDamageNumberAnimation(d.actor.GetInternalName(), d.actor.IsAlive(), d.actor.MapPosition(), amount)
		numberAnimation.SetAudio(amount < 0)
		numberAnimation.StartWhenTrue(d.shouldStartAnimation)
		animations = append(animations, numberAnimation)
		if amount < 0 {
			hitAnimation := NewHitAnimation(d.actor.MapPosition(),0.2, 43, 2)
			hitAnimation.StartWhenTrue(d.shouldStartAnimation)
			hitAnimation.SetOnFinished(func() {
				freeNeighbor := context.GameState.GetMap().GetRandomNeighbor(randomSource, d.actor.MapPosition())
				roll := randomSource.Intn(9)
				if roll == 0 {
					freeNeighbor = d.actor.MapPosition()
				}
				context.GameState.AddBloodDecalAt(freeNeighbor)
			})
			animations = append(animations, hitAnimation)
		}
	}
	return animations
}
func (d *ChangeResourceAction) GetResourceName() stats.Name {
	return d.resourceName
}

func (d *ChangeResourceAction) GetChangeAmount() float64 {
	return d.amount
}

func (d *ChangeResourceAction) SetAnimationStartPredicate(shouldStartAnimation func() bool) {
	d.shouldStartAnimation = shouldStartAnimation
}
