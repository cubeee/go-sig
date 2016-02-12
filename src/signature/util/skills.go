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
	MAX_LEVEL           int = 126
	INVENTION_MAX_LEVEL int = 150
	INVENTION_ID        int = 26
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
		"Invention",     // 26
	}
	Skills                 = map[int]Skill{}
	ExpThresholds          = []int{}
	InventionExpThresholds = []int{
		0, 830, 1861, 2902, 3980, 5126, 6380, 7787, 9400, 11275,
		13605, 16372, 19656, 23546, 28134, 33520, 39809, 47109, 55535,
		65209, 77190, 90811, 106221, 123573, 143025, 164742, 188893, 215651,
		245196, 277713, 316311, 358547, 404634, 454796, 509259, 568254, 632019,
		700797, 774834, 854383, 946227, 1044569, 1149696, 1261903, 1381488, 1508756,
		1644015, 1787581, 1939773, 2100917, 2283490, 2476369, 2679917, 2894505, 3120508,
		3358307, 3608290, 3870846, 4146374, 4435275, 4758122, 5096111, 5449685, 5819299,
		6205407, 6608473, 7028964, 7467354, 7924122, 8399751, 8925664, 9472665, 10041285,
		10632061, 11245538, 11882262, 12542789, 13227679, 13937496, 14672812, 15478994,
		16313404, 17176661, 18069395, 18992239, 19945833, 20930821, 21947856, 22997593,
		24080695, 25259906, 26475754, 27728955, 29020233, 30350318, 31719944, 33129852,
		34580790, 36073511, 37608773, 39270442, 40978509, 42733789, 44537107, 46389292,
		48291180, 50243611, 52247435, 54303504, 56412678, 58575824, 60793812, 63067521,
		65397835, 67785643, 70231841, 72737330, 75303019, 77929820, 80618654, 83370445,
		86186124, 89066630, 92012904, 95025896, 98106559, 101255855, 104474750, 107764216,
		111125230, 114558777, 118065845, 121647430, 125304532, 129038159, 132849323, 136739041,
		140708338, 144758242, 148889790, 153104021, 157401983, 161784728, 166253312, 170808801,
		175452262, 180184770, 185007406, 189921255, 194927409,
	}
)

func init() {
	for idx := 0; idx < len(SkillNames); idx++ {
		Skills[idx] = Skill{SkillNames[idx], idx}
	}

	ExpThresholds = make([]int, MAX_LEVEL+1)

	points, output := 0.0, 0.0
	for level := 1; level <= MAX_LEVEL+1; level++ {
		ExpThresholds[level-1] = int(output)

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

func XPToLevel(skill Skill, currentXp, targetLevel int) int {
	targetXp := XPForLevel(skill, targetLevel)
	return targetXp - currentXp
}

func XPForLevel(skill Skill, level int) int {
	if skill.Id == INVENTION_ID {
		return InventionExpThresholds[level-1]
	} else {
		return ExpThresholds[level-1]
	}
}

func LevelFromXP(skill Skill, xp int) int {
	var maxLevel int
	var xpTable []int
	if skill.Id == INVENTION_ID {
		maxLevel = INVENTION_MAX_LEVEL
		xpTable = InventionExpThresholds
	} else {
		maxLevel = MAX_LEVEL
		xpTable = ExpThresholds
	}

	for level := 1; level <= maxLevel; level++ {
		if xp < xpTable[level-1] {
			return level - 1
		}
	}
	return maxLevel
}
