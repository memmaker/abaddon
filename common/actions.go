package common

import (
	"rpg/audio"
	"rpg/defs"
	"rpg/geometry"
	"rpg/util"
)

type ActionName string

const (
	ActionNameMeleeAttack     ActionName = "ActionNameMeleeAttack"
	ActionNameRangedAttack    ActionName = "ActionNameRangedAttack"
	ActionNameThrow           ActionName = "ActionNameThrow"
	ActionNameMove            ActionName = "ActionNameMove"
	ActionNameMoveItem        ActionName = "ActionNameMoveItem"
	ActionNameSwap            ActionName = "ActionNameSwap"
	ActionNameChangeEquipment ActionName = "ActionNameChangeEquipment"
	ActionNameChangeResource  ActionName = "ActionNameChangeHealth"
	ActionNamePickup          ActionName = "ActionNamePickup"
	ActionNameStealItem       ActionName = "ActionNameStealItem"
	ActionNameDrop            ActionName = "ActionNameDrop"
	ActionNameOpenDoor        ActionName = "ActionNameOpenDoor"
	ActionNameSpawnObject     ActionName = "ActionNameSpawnObject"
	ActionNameUseItem         ActionName = "ActionNameUseItem"
	ActionNameFireball        ActionName = "ActionNameFireball"
	ActionNameLightning   ActionName = "ActionNameLightning"
	ActionNameExplosion   ActionName = "ActionNameExplosion"

	ActionNameDie               ActionName = "ActionNameDie"
	ActionNameApplyStatusEffect ActionName = "ActionNameApplyStatusEffect"
	ActionNameApplyStatusEffectParametrized ActionName = "ActionNameApplyStatusEffectParametrized"
	ActionNameEndTurn           ActionName = "ActionNameEndTurn"
	ActionNameTransition        ActionName = "ActionNameTransition"
	ActionNameTalkTo            ActionName = "ActionNameTalkToNpc"
)

type Animator interface {
	CancelAllAnimations()
	Tick()
	Draw(renderer TileRenderer)
}
type DebugActions struct {
	KillLevel      func()
	GotoHellTown   func()
	OpenDebugMenu  func()
	GoOneLevelDown func()
	GoOneLevelUp   func()
}

type UserInterface interface {
	// Queries
	GetVisibleMap() geometry.Rect
	IsPlayerForcingAggression() bool
	HasOverlays() bool

	// Features ParseColorCodedText(colorCodedText string) []util.ColoredTextPart

	// Logging
	ShowLogMessage(message string)

	// More generic
	Cancel()
	PopAll()
	ShowGameOverMenu()
	OpenMenu(options []util.MenuItem)
	OpenTextWindow(msg []util.LabelText)

	CancelTargeting()
	ShowFullStats()
	UpdateStatsLabel()

	// Specialized
	OpenStash(inv *Inventory, town *Inventory, drop func(item *Item))
	OnMapChanged()
	OpenClassSelectionMenu()

	// Action specific
	BeginTargeting(rules TargetRules, onHit *HitEffect)

	// Vendor interactions
	OpenShopMenu(npc *Actor)
	BeginPlayerSellingTo(vendor *Actor)

	OpenTransferEnchantmentMenu()
	OpenListInventoryForSelection(filter func(element *Item) bool, selected func(itemKey rune))
	UpdateMiniMap()
	OpenPlayerStashInTown()
	OnPlayerTurnTaken()
	SetPlayerMessage(message func(message string))
	SetDebugActions(actions func() DebugActions)
	FadeToBlack()
	FadeFromBlack()
	OpenCenteredTextWindowWithCallback(texts []util.LabelText, onClose func())

}
type ActionBroker interface {
	Execute(action Action)
	RegisterHook(hook ActionHook)
	QueueForExecution(action Action)
	QueueForExecutionAtTheSameTime(actions []Action)
	RemoveBeforeHook(hook ActionHook)
	RemoveAfterHook(hook ActionHook)
	ClearQueuedActions(filter func(action Action) bool)
}
type ActionContext struct {
	GameState     *State
	UserInterface func() UserInterface
	AudioPlayer   *audio.Player
	Data          defs.DataContainer
	ActionBroker  ActionBroker
	Msg           func(message string)
	Speak         func(actor *Actor, trigger defs.ActorSpeechTrigger)
}

type ActionHook interface {
	ActionName() ActionName
	Before(action Action, context ActionContext) []Action
	After(action Action, context ActionContext) []Action
	IsBefore() bool
	IsAfter() bool
}

type ItemActionHook interface {
	ActionHook
	SetItemContext(context *ItemContext)
	ToString() string
}

type Cause interface {
	Action() ActionName
	GetActor() *Actor
	ToString() string
}

type Action interface {
	Cause() Cause
	Name() ActionName
	Execute(context ActionContext) []Action
	GetAnimations() []Animation
	GetDelayInAut() int
}

func OneAction(action Action) []Action {
	return []Action{action}
}
