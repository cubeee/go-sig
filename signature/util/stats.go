package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Stat struct {
	Skill Skill
	Xp    int
}

func GetStats(username string) (map[int]Stat, error) {
	stats := map[int]Stat{}

	url := fmt.Sprintf("http://services.runescape.com/m=hiscore/index_lite.ws?player=%s", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return stats, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return stats, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return stats, errors.New(fmt.Sprintf("HTTP request failed, received status code %d", resp.StatusCode))
	}

	body, _ := ioutil.ReadAll(resp.Body)
	content := strings.Split(string(body), "\n")
	for i := 1; i <= len(Skills); i++ {
		parts := strings.Split(content[i], ",")

		id := i - 1
		skill, err := GetSkillById(id)
		if err != nil {
			continue
		}
		xp, err := strconv.Atoi(parts[2])
		if err != nil {
			xp = 0
		}
		if xp < 0 {
			xp = 0
		}

		stats[id] = Stat{
			Skill: skill,
			Xp:    xp,
		}
	}

	return stats, nil
}

func GetStatBySkill(stats map[int]Stat, skill Skill) Stat {
	var s Stat
	for _, stat := range stats {
		if stat.Skill == skill {
			return stat
		}
	}
	return s
}
