package common

import (
	"fmt"
	"image/color"
	"math/rand"
	"rpg/defs"
	"rpg/geometry"
	"rpg/stats"
	"rpg/util"
	"strings"
)

type VendorType string

const (
	VendorTypeNone    VendorType = ""
	VendorTypeGeneral VendorType = "store_general"
	VendorTypeWands   VendorType = "store_alchemy"
	VendorTypeArmor   VendorType = "store_armor"
	VendorTypeWeapons VendorType = "store_weapons"
)

type Actor struct {
	mapPosition  geometry.Point
	drawPosition geometry.PointF
	icon         int32
	unicodeIcon  int32
	asciiIcon    int32
	cp437Icon    int32
	corpseIcon   int32
	fov          *geometry.FOV
	name         string

	inventory                 *Inventory
	equipment                 *Equipment
	statusEffects             *Flags
	parametrizedStatusEffects []*ParametrizedStatusEffect
	isAnimatingMovement       bool

	stats      stats.AttributeHolder
	timeEnergy int

	hurtTicks int

	isVendorOfType     VendorType
	lastMoveWasWest    bool
	speechMap          map[defs.ActorSpeechTrigger][]string
	isUnique           bool
	internalName       string
	disableCorpse      bool
	deathStringFormats []string
	targetedSkills     []*TargetedAction
	untargetedSkills   []*UntargetedAction
	damageDice         util.Dice
	color              color.RGBA
	faction            defs.Faction
	leader             *Actor
	canSpeak           bool
}

type AttackType int

const (
	MeleeAttack AttackType = iota
	RangedAttack
)

func (a *Actor) Tick() {
	if a.IsHurt() {
		a.hurtTicks--
	}
}
func (a *Actor) SetHurt() {
	a.hurtTicks = int(a.hurtTicksNeeded())
}
func (a *Actor) hurtTicksNeeded() int {
	return int(util.SecondsToTicks(0.3))
}

func (a *Actor) GetHurtTicksAsPercentage() float64 {
	return float64(a.hurtTicks) / float64(a.hurtTicksNeeded())
}

func (a *Actor) IsHurt() bool {
	return a.hurtTicks > 0
}
func (a *Actor) AddEnergy(energy int) {
	//println(fmt.Sprintf("%s restored %d energy", a.name, energy))
	a.timeEnergy += energy
}

func NewActor(icon int32) *Actor {
	actorStats := stats.NewAttributeHolder()
	visionRange := actorStats.GetAttribute(stats.VisionRange)
	return &Actor{
		icon:          icon,
		stats:         actorStats,
		inventory:     NewInventory(32),
		equipment:     NewEquipment(),
		fov:           geometry.NewFOV(geometry.NewRect(-visionRange, -visionRange, visionRange, visionRange)),
		statusEffects: NewFlags(),
		corpseIcon:    defs.IconCorpse,
		speechMap:     make(map[defs.ActorSpeechTrigger][]string),
		damageDice:    util.ParseDice("1d3"),
	}
}

func (a *Actor) GetInventory() *Inventory {
	return a.inventory
}
func (a *Actor) GetEquipment() *Equipment {
	return a.equipment
}
func (a *Actor) VisionMap() []geometry.Point {
	return a.fov.Visibles
}

func (a *Actor) SetIcon(newIcon int32) {
	a.icon = newIcon
}

func (a *Actor) IsAlive() bool {
	return a.GetStats().GetResource(stats.Health).IsPositive()
}

func (a *Actor) Name() string {
	return a.name
}

func (a *Actor) GetStats() *stats.AttributeHolder {
	return &a.stats
}

func (a *Actor) MapPosition() geometry.Point {
	return a.mapPosition
}
func (a *Actor) SetMapPosition(pos geometry.Point) {
	if a.mapPosition != pos {
		a.SetLookDirFromPoI(pos)
	}
	a.mapPosition = pos
}

func (a *Actor) SetLookDirFromPoI(location geometry.Point) {
	if a.mapPosition.X > location.X {
		a.lastMoveWasWest = true
	} else if a.mapPosition.X < location.X {
		a.lastMoveWasWest = false
	}
}

func (a *Actor) DrawPosition() geometry.PointF {
	return a.drawPosition
}

func (a *Actor) SetDrawPosition(pos geometry.PointF) {
	a.drawPosition = pos
}

func (a *Actor) Icon(tick uint64) int32 {
	if !a.IsAlive() {
		return a.corpseIcon
	}
	return a.icon
}

func (a *Actor) IsHostileTo(other *Actor) bool {
	if a.IsAlly(other) {
		return false
	}
	if a.faction == defs.FactionMonsters || other.faction == defs.FactionMonsters {
		return true
	}
	return false
}

func (a *Actor) CanSee(pos geometry.Point) bool {
	return a.fov.Visible(pos)
}

func (a *Actor) GetFOV() *geometry.FOV {
	return a.fov
}

func (a *Actor) IsAnimatingMovement() bool {
	return a.isAnimatingMovement
}

func (a *Actor) SetMovingAnimation(isAnimatingMovement bool) {
	a.isAnimatingMovement = isAnimatingMovement
}

func (a *Actor) SetName(name string) {
	a.name = name
}

func (a *Actor) GetEnergy() int {
	return a.timeEnergy
}

func (a *Actor) RemoveEnergy(energy int) {
	a.timeEnergy -= energy
	//println(fmt.Sprintf("%s used %d energy", a.name, energy))
	a.BeforeTurn()
}

func (a *Actor) GetAttackSpeed(attackType AttackType) int {
	// TODO - missing:
	// - class based attack speed modifier per weapon type (eg. barb swinging axes faster, mages using staves faster)
	baseSpeed := a.GetBaseSpeed()

	attackFactor := a.stats.GetAttributeAsFloat(stats.AttackSpeedFactor)

	baseAttackSpeed := float64(baseSpeed) * attackFactor

	weapon := a.GetEquipment().GetWeapon(attackType)

	if weapon == nil {
		return int(baseAttackSpeed)
	}

	attackSpeed := int(baseAttackSpeed * weapon.GetWeaponStats().AttackSpeedFactor)
	return max(1, attackSpeed) // minimum attack speed is 1
}

func (a *Actor) GetMoveSpeed() int {
	baseSpeed := a.GetBaseSpeed()
	moveFactor := a.stats.GetAttributeAsFloat(stats.MoveSpeedFactor)
	moveSpeed := int(float64(baseSpeed) * moveFactor)
	return max(1, moveSpeed) // minimum move speed is 1
}
func (a *Actor) GetDamageDice(attackType AttackType, enemyName string) util.Dice {
	weapons := a.GetEquipment().GetWeapon(attackType)
	hasWeapons := weapons != nil

	var damageDice util.Dice

	if hasWeapons {
		weaponStats := weapons.GetWeaponStats()
		if weaponStats.HasSpecialDamage() {
			damageDice = weaponStats.GetSpecialDamageDice(a, enemyName)
		} else {
			weaponPlus := weaponStats.Plus
			damageDice = weaponStats.Damage.WithBonus(weaponPlus)
		}
	} else if attackType == RangedAttack { // no ranged weapon, no ranged damage
		return util.NewDice(0, 0, 0)
	} else { // no weapon, built in damage
		damageDice = a.damageDice
	}

	return damageDice
}
func (a *Actor) SetBaseSpeed(speed int) {
	a.stats.SetAttribute(stats.BaseSpeed, speed)
}
func (a *Actor) GetBaseSpeed() int {
	return a.stats.GetAttribute(stats.BaseSpeed)
}
func (a *Actor) GetInfoText() []string {
	tableLines := []string{
		fmt.Sprintf("HP: %s", a.GetStats().GetResource(stats.Health).ToString()),
		fmt.Sprintf("AC: %d", a.GetArmorClass()),
		fmt.Sprintf("ADR: %d%%", int(a.GetArmorDamageReduction()*100)),
		fmt.Sprintf("Movement speed: %d", a.GetMoveSpeed()),
		fmt.Sprintf("Melee speed: %d", a.GetAttackSpeed(MeleeAttack)),
		fmt.Sprintf("Ranged speed: %d", a.GetAttackSpeed(RangedAttack)),
	}

	armor := a.GetEquipment().GetArmor()
	if len(armor) > 0 {
		var armorInfo []string
		armorInfo = append(armorInfo, "", "Armor")
		for _, armorPiece := range armor {
			armorInfo = append(armorInfo, " "+armorPiece.Name())
		}
		tableLines = append(tableLines, armorInfo...)
	}

	meleeWeapon := a.GetEquipment().GetWeapon(MeleeAttack)
	var meleeInfo []string
	meleeDamageString := a.GetDamageDice(MeleeAttack, "").String()
	if meleeWeapon != nil {
		meleeInfo = append(meleeInfo, "", "Melee weapon")
		meleeInfo = append(meleeInfo, fmt.Sprintf(" %s - %s", meleeWeapon.Name(), meleeDamageString))
	} else {
		meleeInfo = append(meleeInfo, "", "Melee weapon")
		meleeInfo = append(meleeInfo, fmt.Sprintf(" unarmed - %s", meleeDamageString))
	}
	tableLines = append(tableLines, meleeInfo...)

	hasRanged := a.GetEquipment().HasRangedWeaponEquipped()
	if hasRanged {
		var rangedInfo []string
		rangedWeapon := a.GetEquipment().GetWeapon(RangedAttack)
		if rangedWeapon != nil {
			rangedDamageString := a.GetDamageDice(RangedAttack, "").String()
			rangedInfo = append(rangedInfo, "", "Ranged weapon")
			rangedInfo = append(rangedInfo, fmt.Sprintf(" %s - %s", rangedWeapon.Name(), rangedDamageString))
		}
		tableLines = append(tableLines, rangedInfo...)
	}

	utSkills := a.GetUntargetedSkills()
	taSkills := a.GetTargetedSkills()
	if len(utSkills) > 0 || len(taSkills) > 0 {
		tableLines = append(tableLines, "", "Skills")
	}
	for _, utSkill := range utSkills {
		tableLines = append(tableLines, " "+utSkill.Name())
	}
	for _, taSkill := range taSkills {
		tableLines = append(tableLines, " "+taSkill.Name())
	}

	baseInfoTable := tableLines

	nameInfo := []string{a.name}
	if a.IsVendor() {
		nameInfo = append(nameInfo, "Vendor", "")
	} else {
		nameInfo = append(nameInfo, "")
	}

	statusEffects := a.statusEffects.AsStrings()

	if len(statusEffects) > 0 {
		baseInfoTable = append(baseInfoTable, "")
		baseInfoTable = append(baseInfoTable, statusEffects...)
	}

	return append(nameInfo, baseInfoTable...)
}

func (a *Actor) IsVendor() bool {
	return a.isVendorOfType != VendorTypeNone
}

func (a *Actor) GetArmorClass() int {
	if a.HasStatusEffect(StatusBerserk) {
		return 0
	}
	baseArmorClass := a.stats.GetAttribute(stats.ArmorClass)
	armor := baseArmorClass
	for _, item := range a.GetEquipment().GetArmor() {
		armor += item.GetArmorStats().GetTotalArmorClass()
	}
	armor = int(float64(armor))
	return armor
}

func (a *Actor) GetArmorDamageReduction() float64 {
	armorClass := a.GetArmorClass()
	return ADRFromAC(armorClass)
}

func (a *Actor) HealCompletely() {
	a.GetStats().GetResource(stats.Health).SetToMax()
}

func (a *Actor) SetVendorType(vendorType VendorType) {
	a.isVendorOfType = vendorType
	if vendorType != VendorTypeNone {
		a.SetNeutralFaction()
	}
}

func (a *Actor) SetNeutralFaction() {
	a.faction = defs.FactionNeutral
}

func (a *Actor) SetMonsterFaction() {
	a.faction = defs.FactionMonsters
}

func (a *Actor) GetVendorType() VendorType {
	return a.isVendorOfType
}


func (a *Actor) GetXPForLevel(level int) int {
	if level == 1 {
		return 0
	}
	if level == 2 {
		return 100
	} else {
		return a.GetXPForLevel(level-1) + 250
	}
}

func (a *Actor) FlipX() bool {
	// spritesheet default is looking right
	return a.lastMoveWasWest
}

func (a *Actor) SetFlipLook(flip bool) {
	a.lastMoveWasWest = flip
}

func (a *Actor) CanPickup() bool {
	return !a.inventory.IsFull()
}

func (a *Actor) BeforeTurn() {
	a.canSpeak = true
	a.regenerateBasicResources()
	a.handleStatusEffects()
}

func (a *Actor) regenerateBasicResources() {
	//statsForActor := a.GetStats()
}
func (a *Actor) handleStatusEffects() {
	for flagName, value := range a.statusEffects.AllEntries() {
		//effectName := StatusEffect(flagName)
		if value == 0 {
			a.statusEffects.SetFalse(flagName)
		} else if value > 0 {
			newValue := value - 1
			if newValue == 0 {
				a.statusEffects.SetFalse(flagName)
			} else {
				a.statusEffects.SetToValue(flagName, newValue)
			}
		}
	}

	//for _, effect := range a.parametrizedStatusEffects {
	for i := len(a.parametrizedStatusEffects) - 1; i >= 0; i-- {
		effect := a.parametrizedStatusEffects[i]
		effect.NextTurn(a)
		if effect.TurnsLeft() == 0 {
			a.parametrizedStatusEffects = append(a.parametrizedStatusEffects[:i], a.parametrizedStatusEffects[i+1:]...)
		}
	}
}

func (a *Actor) AddParametrizedStatusEffect(effect *ParametrizedStatusEffect) {
	a.parametrizedStatusEffects = append(a.parametrizedStatusEffects, effect)
}

func (a *Actor) HasSpeech(trigger defs.ActorSpeechTrigger) bool {
	return len(a.speechMap[trigger]) > 0
}

func (a *Actor) GetSpeech(random *rand.Rand, trigger defs.ActorSpeechTrigger) []string {
	allSpeech := a.speechMap[trigger]
	if len(allSpeech) == 0 {
		return []string{}
	}
	chosenIndex := random.Intn(len(allSpeech))
	chosenLines := allSpeech[chosenIndex]
	//remove
	a.speechMap[trigger] = append(allSpeech[:chosenIndex], allSpeech[chosenIndex+1:]...)
	return strings.Split(chosenLines, "\n")
}

func (a *Actor) SetSpeech(speechMap map[defs.ActorSpeechTrigger][]string) {
	a.speechMap = speechMap
}

func (a *Actor) SetSpeechForTrigger(trigger defs.ActorSpeechTrigger, speech []string) {
	a.speechMap[trigger] = speech
}

func (a *Actor) SetUnique() {
	a.isUnique = true
}

func (a *Actor) IsUnique() bool {
	return a.isUnique
}

func (a *Actor) GetInternalName() string {
	return a.internalName
}

func (a *Actor) SetInternalName(internalName string) {
	a.internalName = internalName
}

func (a *Actor) SetNoCorpse() {
	a.disableCorpse = true
}

func (a *Actor) HasCorpse() bool {
	return !a.disableCorpse
}

func (a *Actor) GetDeathStringFormat() string {
	randomSource := rand.New(rand.NewSource(int64(a.GetStats().GetAttribute(stats.Health))))
	formats := a.deathStringFormats
	if len(formats) == 0 {
		formats = []string{
			"%s died.",
			"%s was killed.",
			"%s was slain.",
		}
	}
	return formats[randomSource.Intn(len(formats))]
}

func (a *Actor) SetDeathStringFormats(formats []string) {
	a.deathStringFormats = formats
}

func (a *Actor) GetStatusEffects() *Flags {
	return a.statusEffects
}

func (a *Actor) AddStatusEffect(effect StatusEffect, turnCount int) {
	a.statusEffects.SetToValue(string(effect), turnCount)
}
func (a *Actor) HasStatusEffect(effect StatusEffect) bool {
	return a.statusEffects.IsSet(string(effect))
}

func (a *Actor) SetTargetedSkills(skills []*TargetedAction) {
	a.targetedSkills = skills
}

func (a *Actor) GetTargetedSkills() []*TargetedAction {
	return a.targetedSkills
}

func (a *Actor) SetUntargetedSkills(skills []*UntargetedAction) {
	a.untargetedSkills = skills
}

func (a *Actor) GetUntargetedSkills() []*UntargetedAction {
	return a.untargetedSkills
}

func (a *Actor) VisionRange() int {
	return a.stats.GetAttribute(stats.VisionRange)
}

func (a *Actor) SetCorpseIcon(corpse int32) {
	a.corpseIcon = corpse
}

func (a *Actor) GetStealableItems() []*Item {
	var stealableItems []*Item
	for _, item := range a.GetInventory().Items() {
		if !a.GetEquipment().IsEquipped(item) {
			stealableItems = append(stealableItems, item)
		}
	}
	return stealableItems
}

func (a *Actor) CanAutoMove() bool {
	if !a.IsAlive() {
		return false
	}
	if a.HasStatusEffect(StatusStunned) {
		return false
	}
	if a.HasStatusEffect(StatusBerserk) {
		return false
	}
	if a.HasStatusEffect(StatusConfused) {
		return false
	}
	if a.HasStatusEffect(StatusImmobile) {
		return false
	}
	return true
}

func (a *Actor) CanUseItems() bool {
	if !a.IsAlive() {
		return false
	}
	if a.HasStatusEffect(StatusStunned) {
		return false
	}
	if a.HasStatusEffect(StatusBerserk) {
		return false
	}
	return true
}

func (a *Actor) GetFullStats() []util.LabelText {
	ourStats := a.GetStats()
	var hpColor color.Color
	defaultColor := defs.DarkTextColor
	hpColor = defaultColor

	hpColor = ourStats.GetResource(stats.Health).AsColor()

	fullStats := []util.LabelText{
		{
			Text:      fmt.Sprintf("Life: %d", ourStats.GetResource(stats.Health).Value()),
			TextColor: hpColor,
		},
		{
			Text:      fmt.Sprintf("Blood: %d", a.GetStats().GetAttribute(stats.Blood)),
			TextColor: defaultColor,
		},
		{
			Text:      fmt.Sprintf("AC: %d", a.GetArmorClass()),
			TextColor: defaultColor,
		},
		{
			Text:      fmt.Sprintf("ADR: %d%%", int(a.GetArmorDamageReduction()*100)),
			TextColor: defaultColor,
		},
		{
			Text:      fmt.Sprintf("Melee: %s", a.GetDamageDice(MeleeAttack, "").ShortString()),
			TextColor: defaultColor,
		},
		{
			Text:      fmt.Sprintf("Ranged: %s", a.GetDamageDice(RangedAttack, "").ShortString()),
			TextColor: defaultColor,
		},

		{
			Text:      fmt.Sprintf("Movement speed: %d", a.GetMoveSpeed()),
			TextColor: defaultColor,
		},
		{
			Text:      fmt.Sprintf("Melee speed: %d", a.GetAttackSpeed(MeleeAttack)),
			TextColor: defaultColor,
		},
		{
			Text:      fmt.Sprintf("Ranged speed: %d", a.GetAttackSpeed(RangedAttack)),
			TextColor: defaultColor,
		},
	}
	return fullStats
}

func (a *Actor) Clone() *Actor {
	actorStats := stats.NewAttributeHolder()
	visionRange := actorStats.GetAttribute(stats.VisionRange)
	clonedActor := &Actor{
		mapPosition:    a.mapPosition,
		drawPosition:   a.drawPosition,
		icon:           a.icon,
		fov:            geometry.NewFOV(geometry.NewRect(-visionRange, -visionRange, visionRange, visionRange)),
		name:           a.name,
		inventory:      a.inventory.Clone(),
		equipment:      NewEquipment(),
		statusEffects:  NewFlags(),
		stats:          a.stats.Clone(),
		timeEnergy:     a.timeEnergy,
		isVendorOfType: a.isVendorOfType,
		faction:        a.faction,
		//speechMap:           a.sp, TODO?
		isUnique:           a.isUnique,
		internalName:       a.internalName,
		disableCorpse:      a.disableCorpse,
		deathStringFormats: a.deathStringFormats,
		targetedSkills:     a.targetedSkills,
		untargetedSkills:   a.untargetedSkills,
		corpseIcon:         a.corpseIcon,
	}
	// TODO equip weapons..
	return clonedActor
}

func (a *Actor) SetAggroRange(aggroRange int) {
	a.stats.SetAttribute(stats.AggroRange, aggroRange)
}

func (a *Actor) GetAggroRange() int {
	return a.stats.GetAttribute(stats.AggroRange)
}

func (a *Actor) SetDamageDice(dice util.Dice) {
	a.damageDice = dice
}

func (a *Actor) GetParametrizedStatusEffects() []*ParametrizedStatusEffect {
	return a.parametrizedStatusEffects
}

func (a *Actor) IsVisible() bool {
	return !a.HasStatusEffect(StatusInvisible)
}

func (a *Actor) GetRangedHitEffects() []*HitEffect {
	// TODO
	return []*HitEffect{}
}

func (a *Actor) AsciiIcon(ticksForAnimation uint64) int32 {
	return a.asciiIcon
}

func (a *Actor) CodePage437Icon(ticksForAnimation uint64) int32 {
	return a.cp437Icon
}

func (a *Actor) UnicodeIcon(ticksForAnimation uint64) int32 {
	return a.unicodeIcon
}

func (a *Actor) SetAsciiIcon(icon int32) {
	a.asciiIcon = icon
}

func (a *Actor) SetCP437Icon(icon int32) {
	a.cp437Icon = icon
}

func (a *Actor) SetUnicodeIcon(icon int32) {
	a.unicodeIcon = icon
}

func (a *Actor) SetColor(color color.RGBA) {
	a.color = color
}

func (a *Actor) GetColor() color.RGBA {
	if a.IsVendor() {
		switch a.isVendorOfType {
		case VendorTypeGeneral:
			return defs.Brown
		case VendorTypeArmor:
			return defs.LightGray
		case VendorTypeWeapons:
			return defs.DarkGray
		case VendorTypeWands:
			return defs.Blue
		}
	}
	return a.color
}

func (a *Actor) IsWinCondition() bool {
	return a.HasStatusEffect(StatusWinCondition)
}

func (a *Actor) SetFaction(faction defs.Faction) {
	a.faction = faction
}

func (a *Actor) SetFollowLeader(leader *Actor) {
	a.leader = leader
}

func (a *Actor) HasLeader() bool {
	return a.leader != nil
}

func (a *Actor) GetFaction() defs.Faction {
	return a.faction
}

func (a *Actor) GetLeader() *Actor {
	return a.leader
}

func (a *Actor) IsAlly(other *Actor) bool {
	return a.faction == other.GetFaction()
}

func (a *Actor) CanAct() bool {
	return !a.HasStatusEffect(StatusStunned)
}

func (a *Actor) CanMove() bool {
	return a.CanAct() && !a.HasStatusEffect(StatusImmobile)
}

func (a *Actor) CanAttack() bool {
	return a.CanAct() && !a.HasStatusEffect(StatusPacified)
}

func (a *Actor) SetInventory(inventory *Inventory) {
	a.inventory = inventory
}

func (a *Actor) SetEquipment(equipment *Equipment) {
	a.equipment = equipment
}

func (a *Actor) CanSpeak() bool {
	return a.canSpeak
}

func (a *Actor) SetHasSpoken() {
	a.canSpeak = false
}


func ADRFromAC(armorClass int) float64 {
	return util.Clamp(0.0, 0.9, (float64(armorClass)/float64(armorClass+20))*1.5)
}
