package defs

import (
	"image/color"
	"math/rand"
	"rpg/recfile"
	"rpg/util"
	"strings"
)

type Faction int

const (
	FactionNeutral Faction = iota
	FactionPlayer
	FactionMonsters
	FactionLoneWolf
)

type ActorSpeechTrigger int

const (
	SpeechTriggerSight ActorSpeechTrigger = iota
	SpeechTriggerHurtInTown
)

type EnemyDefinition struct {
	Icon               int32
	DLvl               int
	MaxLevel           int
	OutOfDepthChance   float64
	MinLevelForOoD     int
	Name               string
	InternalName       string
	BaseSpeed          int
	ResourceHealth     util.Dice
	DamageDice         util.Dice
	Stats              map[string]int
	Equipment          []string
	DeathStringFormats []string

	TargetedSkills   []string
	UntargetedSkills []string
	GroupSpawn       []string

	Tags EnemyTags

	AggroRange int

	Speech      map[ActorSpeechTrigger][]string
	AsciiIcon   rune
	UnicodeIcon rune
	CP437Icon   rune // As Corresponding Unicode Code Point
	Color       color.RGBA
}

func NewEnemyDefinitionFromRecord(record recfile.Record) EnemyDefinition {
	enemyDefinition := EnemyDefinition{
		Stats:  make(map[string]int),
		Speech: make(map[ActorSpeechTrigger][]string),
		AsciiIcon: -1,
		CP437Icon: -1,
		UnicodeIcon: -1,
	}

	for _, field := range record {
		// special case for stats
		if strings.HasPrefix(field.Name, "stats_") {
			statName := strings.TrimPrefix(field.Name, "stats_")
			enemyDefinition.Stats[statName] = field.AsInt()
			continue
		}
		switch field.Name {
		case "internal_name":
			enemyDefinition.InternalName = field.Value
		case "name":
			enemyDefinition.Name = field.Value
		case "icon":
			enemyDefinition.Icon = field.AsInt32()
		case "ascii":
			enemyDefinition.AsciiIcon = rune(field.Value[0])
		case "unicode":
			enemyDefinition.UnicodeIcon = rune(field.Value[0])
		case "cp437":
			cp437IconAsUnicode := rune(field.Value[0])
			cp437Byte := util.UnicodeToCP437Byte(cp437IconAsUnicode)
			enemyDefinition.CP437Icon = int32(cp437Byte)
		case "color":
			enemyDefinition.Color = util.GetColorFromRGBString(field.Value)
		case "dlvl":
			enemyDefinition.DLvl = field.AsInt()
		case "out_of_depth_chance":
			enemyDefinition.OutOfDepthChance = field.AsFloat()
		case "min_level_for_ood":
			enemyDefinition.MinLevelForOoD = field.AsInt()
		case "max_level":
			enemyDefinition.MaxLevel = field.AsInt()
		case "base_speed":
			enemyDefinition.BaseSpeed = field.AsInt()
		case "resource_health":
			enemyDefinition.ResourceHealth = util.ParseDice(field.Value)
		case "damage_dice":
			enemyDefinition.DamageDice = util.ParseDice(field.Value)
		case "equipment":
			enemyDefinition.Equipment = append(enemyDefinition.Equipment, field.Value)
		case "death_string":
			enemyDefinition.DeathStringFormats = append(enemyDefinition.DeathStringFormats, field.Value)
		case "targeted_skill":
			enemyDefinition.TargetedSkills = append(enemyDefinition.TargetedSkills, field.Value)
		case "untargeted_skill":
			enemyDefinition.UntargetedSkills = append(enemyDefinition.UntargetedSkills, field.Value)
		case "group_spawn":
			enemyDefinition.GroupSpawn = append(enemyDefinition.GroupSpawn, field.Value)
		case "aggro_range":
			enemyDefinition.AggroRange = field.AsInt()
		case "tags":
			enemyDefinition.Tags = EnemyTagsFromString(field.Value)
		case "speech_on_sight":
			enemyDefinition.Speech[SpeechTriggerSight] = append(enemyDefinition.Speech[SpeechTriggerSight], field.Value)
		}
	}
	if enemyDefinition.AsciiIcon == -1 {
		enemyDefinition.AsciiIcon = rune(enemyDefinition.Name[0])
	}
	if enemyDefinition.UnicodeIcon == -1 {
		enemyDefinition.UnicodeIcon = enemyDefinition.AsciiIcon
	}
	if enemyDefinition.CP437Icon == -1 {
		enemyDefinition.CP437Icon = enemyDefinition.AsciiIcon
	}
	return enemyDefinition
}

func (d EnemyDefinition) CanAppearAt(randomSource *rand.Rand, dungeonLevel int) bool {
	// reject early
	if d.MaxLevel > 0 && dungeonLevel > d.MaxLevel {
		return false
	}
	regularAppearance := dungeonLevel >= d.DLvl

	// apply out of depth chance
	oodAppearance := dungeonLevel >= d.MinLevelForOoD && randomSource.Float64() < d.OutOfDepthChance

	return regularAppearance || oodAppearance
}





type EnemyTags int

func (t EnemyTags) Has(tag EnemyTags) bool {
	return t&tag != 0
}

const (
	TagFlying EnemyTags = 1 << iota
	TagUndead
	TagUnique
	TagNoCorpse
	TagWinCondition
)
//IsUnique    bool
//	HasNoCorpse bool

func EnemyTagsFromString(tags string) EnemyTags {
	var result EnemyTags
	for _, tag := range strings.Split(tags, "|") {
		tag = strings.ToLower(strings.TrimSpace(tag))
		switch tag {
		case "flying":
			result |= TagFlying
		case "undead":
			result |= TagUndead
		case "unique":
			result |= TagUnique
		case "no_corpse":
			result |= TagNoCorpse
		case "win_condition":
			result |= TagWinCondition
		}
	}
	return result
}