package defs

import "fmt"

func NiceNameForStatusEffect(effect string) string {
    switch effect {
    case "explode_on_death":
        return "Explode on death"
    case "buff_speed": return "Haste"
    }
    return effect
}

func NiceNameForStatusEffectWithTurns(effect string, turnsLeft int) string {
    turnString := ""
    if turnsLeft > 0 {
        turnString = fmt.Sprintf(" (%d turns left)", turnsLeft)
    }
    return NiceNameForStatusEffect(effect) + turnString
}
func DescriptionForStatusEffect(effect string, turnsLeft int) string {
    description := NiceNameForStatusEffect(effect)
    turnString := ""
    if turnsLeft > 0 {
        turnString = fmt.Sprintf(" (%d turns left)", turnsLeft)
    }

    switch effect {
    case "explode_on_death":
        description = "When this creature dies, it explodes."
    case "buff_speed":
        description = "This creature is moving faster than usual."
    }
    return description + turnString
}
