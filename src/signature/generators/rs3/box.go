package rs3

import (
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
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"signature/generators"
	"signature/util"
	"strconv"
)

var (
	baseWidth    = 161
	baseHeight   = 80
	baseImage    *image.RGBA
	dpi          = 72.0
	fontFile     = "./assets/fonts/MuseoSans_500.ttf"
	baseFont     *truetype.Font
	fontColor    = image.NewUniform(color.RGBA{245, 178, 65, 255})
	size         = 12.0
	salt         = "ded3b63a6f11a9efd8e0f6a9b84fbeb1"
	staticLabels = []StaticLabel{
		{"Current XP:", 7, 15},
		{"Target lvl:", 7, 30},
		{"Remainder:", 7, 45},
	}
)

func init() {
	baseImage = loadbaseImage()
	loadFonts()
}

type StaticLabel struct {
	str string
	x   int
	y   int
}

type BoxGoalGenerator struct {
	generators.Generator
}

func (b BoxGoalGenerator) CreateSignature(req util.ParsedSignatureRequest) (util.Signature, error) {
	baseImage := cloneImage(baseImage)

	username := req.GetProperty("username").(string)
	skill := req.GetProperty("skill").(util.Skill)
	goal := req.GetProperty("goal").(int)
	goaltype := req.GetProperty("goaltype").(util.GoalType)

	stats, err := util.GetStats(username) // introduce field
	if err != nil {
		var s util.Signature
		return s, errors.New(fmt.Sprintf("Failed to fetch stats for %s", username))
	}
	stat := util.GetStatBySkill(stats, skill)

	currentLevel := util.LevelFromXP(stat.Xp)
	currentXP := stat.Xp
	goalXP := util.XPForLevel(goal)
	remainder := util.XPToLevel(currentXP, goal)
	if goaltype == util.GoalXP {
		goalXP = goal
		remainder = goalXP - currentXP
	}
	goalLevel := util.LevelFromXP(goalXP)
	if remainder < 0 {
		remainder = 0
	}
	percent := int(float64(currentXP) / float64(goalXP) * 100.0)
	if percent > 100 {
		percent = 100
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

	drawer := createDrawer(baseImage, fontColor, baseFont, size, dpi,
		font.HintingFull)

	drawString := func(str string, x fixed.Int26_6, y int) {
		drawer.Dot = fixed.Point26_6{
			X: x,
			//Y: fixed.I(y + int((c.PointToFixed(size) >> 6))),
			Y: fixed.I(y + int(c.PointToFixed(size)>>6)),
		}
		drawer.DrawString(str)
	}

	drawRightAlignedString := func(str string, x, y int) {
		width := drawer.MeasureString(str)
		drawString(str, fixed.I(x)-width, y)
	}

	// Skill name and current level
	drawString(
		fmt.Sprintf("%s: %d/%d", skill.Name, currentLevel, goalLevel),
		fixed.I(7), 1)

	for _, label := range staticLabels {
		if label.str == "Target lvl:" && goaltype == util.GoalXP {
			label.str = "Target XP:"
		}

		drawString(label.str, fixed.I(label.x), label.y)
	}

	x, y := 150, 15

	// current xp
	drawRightAlignedString(util.Format(currentXP), x, y)
	y += 15

	// goal
	drawRightAlignedString(util.Format(goal), x, y)
	y += 15

	// remainder
	drawRightAlignedString(util.Format(remainder), x, y)
	y += 15

	// bar
	drawBar(baseImage, percent)

	// bar percentage
	x = 71
	y = 61
	color := image.White
	if percent >= 50 {
		color = image.Black
	}

	drawer = createDrawer(baseImage, color, baseFont, 11, dpi,
		font.HintingFull)
	drawString(fmt.Sprintf("%d%%", percent), fixed.I(x), y)

	return util.Signature{username, baseImage}, nil
}

func (b BoxGoalGenerator) Name() string {
	return "box"
}

func (b BoxGoalGenerator) Url() string {
	return "/:username/:skill/:goal"
}

func (b BoxGoalGenerator) FormUrl() string {
	return "/tooltip/create"
}

func (b BoxGoalGenerator) CreateHash(req util.ParsedSignatureRequest) string {
	skill := req.GetProperty("skill").(util.Skill)
	return fmt.Sprintf("%s-%d-%d", req.GetProperty("username"), skill.Id, req.GetProperty("goal"))
}

func (b BoxGoalGenerator) HandleForm(c web.C, writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	form := request.Form
	username := form.Get("username")
	skill := form.Get("skill")
	goal := form.Get("goal")

	// todo: validate input lol

	url := fmt.Sprintf("/%s/%s/%s", username, skill, goal)
	fmt.Println(url)
	//http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
	util.ServeResultPage(c, writer, request, url)
}

// Parse the request into a signature request
func (b BoxGoalGenerator) ParseSignatureRequest(c web.C) (util.ParsedSignatureRequest, error) {
	req := util.NewSignatureRequest()

	username := c.URLParams["username"] // todo: clean username
	usernameLength := len(username)
	if !util.UsernameRegex.MatchString(username) {
		return req, errors.New("Invalid username entered, allowed characters: alphabets, numbers, _ and +")
	}
	if usernameLength < 1 || usernameLength > 12 {
		return req, errors.New("Username has to be between 1 and 12 characters long")
	}

	// Read the skill id and make sure it is numeric
	id, err := strconv.Atoi(c.URLParams["skill"])
	var skill util.Skill
	if err == nil {
		// Get the skill by id
		skill, err = util.GetSkillById(id)
		if err != nil {
			return req, errors.New(fmt.Sprintf("No skill found for the given id, make sure it is between 0 and %d", len(util.Skills)))
		}
	} else {
		// Get the skill by name
		skill, err = util.GetSkillByName(c.URLParams["skill"])
		if err != nil {
			return req, errors.New("No skill found for the given skill name")
		}
	}

	// Read the level and make sure it is numeric
	goal, err := strconv.Atoi(c.URLParams["goal"])
	if err != nil {
		return req, errors.New("Invalid goal entered, make sure it is numeric")
	}

	// Make sure the level is within valid bounds
	if goal < 0 || goal > 200000000 {
		return req, errors.New("Invalid level/xp goal entered, make sure it 0-200,000,000")
	}

	// Switch the goal type if the goal exceeds the maximum skill level
	goaltype := util.GoalLevel
	if goal > util.MAX_LEVEL {
		goaltype = util.GoalXP
	}

	req.AddProperty("username", username)
	req.AddProperty("id", id)
	req.AddProperty("goal", goal)
	req.AddProperty("skill", skill)
	req.AddProperty("goaltype", goaltype)
	return req, nil
}

func drawBar(img draw.Image, percent int) {
	x := 15
	y := 62
	width := int(135.0 * (float64(percent) / 100.0))
	height := 14

	green := color.RGBA{0, 255, 0, 255}
	// todo: red
	bar := image.Rect(x, y, x+width, y+height)
	draw.Draw(img, bar, &image.Uniform{green}, image.ZP, draw.Src)
}

// Load base image to memory
func loadbaseImage() *image.RGBA {
	baseImageHandle, _ := os.Open("assets/img/base.png")
	defer baseImageHandle.Close()
	baseImage := image.NewRGBA(image.Rect(0, 0, baseWidth, baseHeight))
	img, _ := png.Decode(baseImageHandle)
	draw.Draw(baseImage, baseImage.Bounds(), img, image.ZP, draw.Src)
	return baseImage
}

// Load font(s) to memory
func loadFonts() {
	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		panic(err)
	}
	baseFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
}

// Clones an image
func cloneImage(src image.Image) draw.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)
	return dst
}
