package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Signature struct {
	hash     string
	username string
	Image    image.Image
}

type StaticLabel struct {
	str string
	x   int
	y   int
}

type Generator int

var (
	baseWidth    = 161
	baseHeight   = 80
	baseImage    *image.RGBA
	dpi          = 72.0
	fontFile     = "./MuseoSans_500.ttf"
	baseFont     *truetype.Font
	fontColor    = image.NewUniform(color.RGBA{245, 178, 65, 255})
	size         = 12.0
	salt         = "ded3b63a6f11a9efd8e0f6a9b84fbeb1"
	hasher       = md5.New()
	staticLabels = []StaticLabel{
		{"Current XP:", 7, 15},
		{"Target lvl:", 7, 30},
		{"Remainder:", 7, 45},
	}
)

func init() {
	// Load base image to memory
	log.Println("Loading assets to memory...")
	baseImage = loadbaseImage()
	loadFonts()
}

// Create a hash for the given parameters
func (g Generator) CreateHash(username string, id int, level int) string {
	return getMD5(fmt.Sprintf("%s-%d-%s-%d", username, id, salt, level))
}

// Generate a signature for the given request
func (g Generator) CreateSignature(req SignatureRequest) (Signature, error) {
	baseImage := cloneImage(baseImage)

	stats, err := GetStats(req.username) // introduce field
	if err != nil {
		var s Signature
		return s, errors.New(fmt.Sprintf("Failed to fetch stats for %s", req.username))
	}
	stat := GetStatBySkill(stats, req.skill)

	currentLevel := LevelFromXP(stat.xp)
	currentXP := stat.xp
	goalXP := XPForLevel(req.goal)
	remainder := XPToLevel(currentXP, req.goal)
	if req.goaltype == GoalXP {
		goalXP = req.goal
		remainder = goalXP - currentXP
	}
	goalLevel := LevelFromXP(goalXP)
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
		fmt.Sprintf("%s: %d/%d", req.skill.name, currentLevel, goalLevel),
		fixed.I(7), 1)

	for _, label := range staticLabels {
		if label.str == "Target lvl:" && req.goaltype == GoalXP {
			label.str = "Target XP:"
		}

		drawString(label.str, fixed.I(label.x), label.y)
	}

	x, y := 150, 15

	// current xp
	drawRightAlignedString(Format(currentXP), x, y)
	y += 15

	// goal
	drawRightAlignedString(Format(req.goal), x, y)
	y += 15

	// remainder
	drawRightAlignedString(Format(remainder), x, y)
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

	return Signature{req.hash, req.username, baseImage}, nil
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
	baseImageHandle, _ := os.Open("base.png")
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

// Creates a MD5 hash
func getMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
