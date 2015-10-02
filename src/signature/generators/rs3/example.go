package rs3

import (
	"errors"
	"github.com/zenazn/goji/web"
	"image"
	"image/color"
	"image/draw"
	"signature/generators"
	"signature/util"
)

type ExampleGenerator struct {
	generators.Generator
}

func (g ExampleGenerator) Name() string {
	return "example"
}

func (g ExampleGenerator) Url() string {
	return "/hello/:username"
}

func (g ExampleGenerator) CreateSignature(req util.ParsedSignatureRequest) (util.Signature, error) {
	username := req.GetProperty("username").(string)

	baseImage := image.NewRGBA(image.Rect(0, 0, 500, 100))
	blue := color.RGBA{0, 0, 255, 255}

	draw.Draw(baseImage, baseImage.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	return util.Signature{username, baseImage}, nil
}

func (g ExampleGenerator) CreateHash(req util.ParsedSignatureRequest) string {
	username := req.GetProperty("username").(string)
	return username
}

func (g ExampleGenerator) ParseSignatureRequest(c web.C) (util.ParsedSignatureRequest, error) {
	req := util.NewSignatureRequest()

	username := c.URLParams["username"] // todo: clean username
	usernameLength := len(username)
	if !util.UsernameRegex.MatchString(username) {
		return req, errors.New("Invalid username entered, allowed characters: alphabets, numbers, _ and +")
	}
	if usernameLength < 1 || usernameLength > 12 {
		return req, errors.New("Username has to be between 1 and 12 characters long")
	}

	req.AddProperty("username", username)
	return req, nil
}
