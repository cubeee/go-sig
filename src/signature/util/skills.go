package util

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

type Skill struct {
	Name string
	Id   int
}

const (
	MAX_LEVEL int = 126
)

var (
	SkillNames = []string{
		"Attack",        //  0
		"Defence",       //  1
		"Strength",      //  2
		"Constitution",  //  3
		"Ranged",        //  4
		"Prayer",        //  5
		"Magic",         //  6
		"Cooking",       //  7
		"Woodcutting",   //  8
		"Fletching",     //  9
		"Fishing",       // 10
		"Firemaking",    // 11
		"Crafting",      // 12
		"Smithing",      // 13
		"Mining",        // 14
		"Herblore",      // 15
		"Agility",       // 16
		"Thieving",      // 17
		"Slayer",        // 18
		"Farming",       // 19
		"Runecrafting",  // 20
		"Hunter",        // 21
		"Construction",  // 22
		"Summoning",     // 23
		"Dungeoneering", // 24
		"Divination",    // 25
	}
	Skills        = map[int]Skill{}
	ExpThresholds = map[int]int{}
)

func init() {
	for idx := 0; idx < len(SkillNames); idx++ {
		Skills[idx] = Skill{SkillNames[idx], idx}
	}

	points, output := 0.0, 0.0
	for level := 1; level <= MAX_LEVEL+1; level++ {
		ExpThresholds[level] = int(output)

		points += math.Floor(float64(level) + 300.0*math.Pow(2.0, float64(level)/7.0))
		output = points / 4
	}
}

func GetSkillByName(name string) (Skill, error) {
	name = strings.ToLower(name)
	var s Skill
	for _, skill := range Skills {
		if strings.ToLower(skill.Name) == name {
			s = skill
			return s, nil
		}
	}
	return s, errors.New("No skill found with the given name")
}

func GetSkillById(id int) (Skill, error) {
	var s Skill
	if id < 0 || id >= len(Skills) {
		return s, errors.New(fmt.Sprintf("Id out of bounds, 0-%d expected", len(Skills)))
	}
	return Skills[id], nil
}

func XPToLevel(currentXp, targetLevel int) int {
	targetXp := XPForLevel(targetLevel)
	return targetXp - currentXp
}

func XPForLevel(level int) int {
	return ExpThresholds[level]
}

func LevelFromXP(xp int) int {
	for level := 1; level <= len(ExpThresholds); level++ {
		if xp < ExpThresholds[level] {
			return level - 1
		}
	}
	return 99
}
