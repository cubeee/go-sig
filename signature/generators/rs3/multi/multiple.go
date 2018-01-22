package multi

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/zenazn/goji/web"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"net/http"
	"net/url"
	"github.com/cubeee/go-sig/signature/generators"
	"github.com/cubeee/go-sig/signature/util"
	"strconv"
	"github.com/cubeee/go-sig/signature"
)

var (
	baseWidth     = 400
	baseHeight    = 25
	bottomPadding = 20
	paddingSides  = 5
	dpi           = 72.0
	baseFont      = util.LoadFont("./resources/assets/fonts/MuseoSans_500.ttf")
	fontColor     = image.NewUniform(color.RGBA{245, 178, 65, 255})
	size          = 15.0
)

type MultiGoalGenerator struct {
	generators.Generator
}

type MultiGoal struct {
	skill    util.Skill
	goal     int
	goalType util.GoalType
}

func (m MultiGoalGenerator) CreateSignature(req util.ParsedSignatureRequest) (util.Signature, error) {
	username := req.GetProperty("username").(string)
	goals := req.GetProperty("goals").([]MultiGoal)

	stats, err := util.GetStats(username)
	if err != nil {
		var s util.Signature
		return s, errors.New(fmt.Sprintf("Failed to fetch stats for %s", username))
	}

	c := freetype.NewContext()

	createDrawer := func(img draw.Image, color *image.Uniform, f *truetype.Font,
		size float64, dpi float64, hinting font.Hinting) *font.Drawer {
		return &font.Drawer{
			Dst: img,
			Src: color,
			Face: truetype.NewFace(f, &truetype.Options{
				Size:    size,
				DPI:     dpi,
				Hinting: hinting,
			})}
	}

	baseImage := loadBaseImage(len(goals))

	drawer := createDrawer(baseImage, fontColor, baseFont, size, dpi, font.HintingFull)

	drawString := func(str string, x fixed.Int26_6, y int) {
		drawer.Dot = fixed.Point26_6{
			X: x,
			Y: fixed.I(y + int(c.PointToFixed(size)>>6)),
		}
		drawer.DrawString(str)
	}

	drawRightAlignedString := func(str string, x, y int) {
		width := drawer.MeasureString(str)
		drawString(str, fixed.I(x)-width, y)
	}

	nameX, goalX := paddingSides, baseWidth-paddingSides
	y := paddingSides

	for _, goal := range goals {
		stat := util.GetStatBySkill(stats, goal.skill)

		currentLevel := util.LevelFromXP(stat.Skill, stat.Xp)
		currentXP := stat.Xp
		var goalXP int
		var remainder int
		if goal.goalType == util.GoalXP {
			goalXP = goal.goal
			remainder = goalXP - currentXP
		} else {
			goalXP = util.XPForLevel(stat.Skill, goal.goal)
			remainder = util.XPToLevel(stat.Skill, currentXP, goal.goal)
		}
		goalLevel := util.LevelFromXP(stat.Skill, goalXP)
		if remainder < 0 {
			remainder = 0
		}
		percent := int(float64(currentXP) / float64(goalXP) * 100.0)
		if percent > 100 {
			percent = 100
		}

		if currentLevel > goalLevel {
			currentLevel = goalLevel
		}
		if currentXP > goalXP {
			currentXP = goalXP
		}

		// Skill name and current level
		drawString(
			fmt.Sprintf("%s: %d/%d", goal.skill.Name, currentLevel, goalLevel),
			fixed.I(nameX), y)

		// Current and goal xp
		drawRightAlignedString(util.Format(currentXP)+"/"+util.Format(goalXP), goalX, y)

		// Bar
		drawBar(baseImage, percent, nameX, y+20, baseWidth-5-paddingSides, 1)

		y += baseHeight
	}

	// Watermark
	y -= 5
	drawer = createDrawer(baseImage, fontColor, baseFont, 11, dpi, font.HintingFull)
	drawRightAlignedString(vars.VirtualHost, goalX, y)

	return util.Signature{Username: username, Image: baseImage}, nil
}

func (m MultiGoalGenerator) Name() string {
	return "multi"
}

func (m MultiGoalGenerator) Url() string {
	return "/multi/:username"
}

func (m MultiGoalGenerator) FormUrl() string {
	return "/multi/create"
}

func (m MultiGoalGenerator) CreateHash(req util.ParsedSignatureRequest) string {
	username := req.GetProperty("username").(string)
	goals := req.GetProperty("goals").([]MultiGoal)
	goalStr := username
	for _, goal := range goals {
		goalStr = fmt.Sprintf("%s-%v-%v", goalStr, goal.skill.Id, goal.goal)
	}
	return util.GetMD5(goalStr)
}

func (m MultiGoalGenerator) HandleForm(c web.C, writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	form := request.Form
	username := form.Get("username")

	// Workaround to preserve order
	var buf bytes.Buffer
	for id := 0; id < len(util.Skills); id++ {
		skill := form.Get("skill_" + strconv.Itoa(id))
		goal := form.Get("goal_" + strconv.Itoa(id))

		if skill == "" || goal == "" {
			continue
		}

		if _, err := util.GetSkillByName(skill); err != nil {
			continue
		}

		queryPrefix := url.QueryEscape(skill) + "="
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(queryPrefix)
		buf.WriteString(url.QueryEscape(goal))
	}

	hideUsername := form.Get("hide")
	if hideUsername == "on" && util.AesKey != nil && len(util.AesKey) > 0 {
		name, err := util.Encrypt(username)
		if err == nil {
			name = "_" + name
		}
		username = name
	}

	hash := buf.String()
	imageUrl := fmt.Sprintf("/multi/%s?%s", username, hash)
	util.ServeResultPage(writer, imageUrl)
}

// Parse the request into a signature request
func (m MultiGoalGenerator) ParseSignatureRequest(c web.C, r *http.Request) (util.ParsedSignatureRequest, error) {
	req := util.NewSignatureRequest()

	username := util.ParseUsername(c.URLParams["username"])
	usernameLength := len(username)
	if !util.UsernameRegex.MatchString(username) {
		return req, errors.New("invalid username entered, allowed characters: alphabets, numbers, _ and +")
	}
	if usernameLength < 1 || usernameLength > 12 {
		return req, errors.New("username has to be between 1 and 12 characters long")
	}

	var goals []MultiGoal
	params, _ := util.ParseQueryParameters(r.URL.RawQuery)
	for _, param := range params {
		skillName, skillGoal := param.Key, param.Value

		// Make sure the skill is valid
		skill, err := util.GetSkillByName(skillName)
		if err != nil {
			return req, errors.New("no skill found for the given skill name '" + skillName + "'")
		}

		// Check if goal has 'k' or 'm' suffix
		goal, err := util.FromSuffixed(skillGoal)
		if err != nil {
			// Make sure the goal is numeric
			goal, err = strconv.Atoi(skillGoal)
		}
		if err != nil {
			return req, errors.New("invalid goal entered for " + skillName + ", make sure it is numeric or has 'k'/'m' suffix")
		}

		// Make sure the goal is within valid bounds
		if goal < 0 || goal > 200000000 {
			return req, errors.New("invalid level/xp goal entered, make sure it 0-200,000,000")
		}

		// Switch the goal type if the goal exceeds the maximum skill level
		goalType := util.GetGoalType(skill, goal)

		goals = append(goals, MultiGoal{skill, goal, goalType})
	}

	req.AddProperty("username", username)
	req.AddProperty("goals", goals)
	return req, nil
}

func drawBar(img draw.Image, percent, x, y, width, height int) {
	greenWidth := int(float64(width) * (float64(percent) / 100.0))

	red := color.RGBA{R: 160, G: 0, B: 0, A: 255}
	redBar := image.Rect(x, y, width, y+height)
	green := color.RGBA{R: 0, G: 160, B: 0, A: 255}
	greenBar := image.Rect(x, y, x+greenWidth, y+height)

	draw.Draw(img, redBar, &image.Uniform{red}, image.ZP, draw.Src)
	draw.Draw(img, greenBar, &image.Uniform{green}, image.ZP, draw.Src)
}

// Load base image to memory
func loadBaseImage(goals int) *image.RGBA {
	baseImage := image.NewRGBA(image.Rect(0, 0, baseWidth, baseHeight*goals+bottomPadding))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	draw.Draw(baseImage, baseImage.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	return baseImage
}
