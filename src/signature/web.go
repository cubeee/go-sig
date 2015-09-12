package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) {
	return 0, nil
}

type GoalType int

type SignatureRequest struct {
	username string
	id       int
	goal     int
	skill    Skill
	hash     string
	goaltype GoalType
}

const (
	GoalLevel GoalType = iota
	GoalXP
)

var (
	imageRoot      = "/tmp/images"
	updateInterval = 4.0
	generator      = new(Generator)
	usernameRegex  = regexp.MustCompile("^[a-zA-Z0-9-_+]+$")
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

// Parse the request into a signature request
func GetSignatureRequest(c web.C) (SignatureRequest, error) {
	var req SignatureRequest

	username := c.URLParams["username"] // todo: clean username
	usernameLength := len(username)
	if !usernameRegex.MatchString(username) {
		return req, errors.New("Invalid username entered, allowed characters: alphabets, numbers, _ and +")
	}
	if usernameLength < 1 || usernameLength > 12 {
		return req, errors.New("Username has to be between 1 and 12 characters long")
	}

	// Read the skill id and make sure it is numeric
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		log.Println(err.Error())
		return req, errors.New("Invalid id entered, make sure it is numeric")
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

	// Get the skill by given id
	skill, err := GetSkillById(id)
	if err != nil {
		return req, err
	}

	// Switch the goal type if the goal exceeds the maximum skill level
	goaltype := GoalLevel
	if goal > MAX_LEVEL {
		goaltype = GoalXP
	}

	// Create the hash for the request
	hash := generator.CreateHash(username, id, goal)

	return SignatureRequest{username, id, goal, skill, hash, goaltype}, nil
}

func createAndSaveSignature(writer http.ResponseWriter, req SignatureRequest) error {
	// Create the signature image
	sig, err := generator.CreateSignature(req)
	if err != nil {
		return err
	}

	// note: queue saving if it causes performance issues?
	// Save the image to disk with the given hash as the file name
	saveImage(req.hash, sig.Image)
	return nil
}

// Save the image to disk with the given hash as the file name
func saveImage(hash string, img image.Image) {
	out, err := os.Create(imageRoot + "/" + hash)
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()
	writer := bufio.NewWriter(out)
	err = png.Encode(writer, img)
	if err != nil {
		log.Println(err)
		return
	}
	err = writer.Flush()
	if err != nil {
		log.Println(err)
		return
	}
}

func updateSignature(writer http.ResponseWriter, req SignatureRequest) {
	imagePath := fmt.Sprintf("%s/%s", imageRoot, req.hash)
	if stat, err := os.Stat(imagePath); err == nil {
		modTime := stat.ModTime()
		now := time.Now()
		age := now.Sub(modTime)

		if age.Minutes() >= updateInterval {
			log.Println("Updating signature....")
			err = createAndSaveSignature(writer, req)
			if err != nil {
				writeTextResponse(writer, err.Error())
				return
			}
		}
	}
}

// Write an image as a response to the client
func writeImageResponse(writer http.ResponseWriter, signature Signature) {
	buffer := new(bytes.Buffer)
	if err := png.Encode(buffer, signature.Image); err != nil {
		writeTextResponse(writer, "unable to encode image")
		return
	}
	writer.Header().Set("Content-Type", "image/png")
	writer.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := writer.Write(buffer.Bytes()); err != nil {
		writeTextResponse(writer, "unable to write image")
		return
	}
}

// Write text as a response to the client
func writeTextResponse(writer http.ResponseWriter, text string) {
	fmt.Fprintf(writer, text)
}

// Show an existing signature
func signature(c web.C, writer http.ResponseWriter, r *http.Request) {
	// Parse the request into a struct
	req, err := GetSignatureRequest(c)
	if err != nil {
		writeTextResponse(writer, err.Error())
		return
	}

	// Check if an image already exists and make it if not
	if _, err := os.Stat(fmt.Sprintf("%s/%s", imageRoot, req.hash)); os.IsNotExist(err) {
		err = createAndSaveSignature(writer, req)
		if err != nil {
			writeTextResponse(writer, err.Error())
			return
		}
		return
	}

	updateSignature(writer, req)

	http.ServeFile(writer, r, fmt.Sprintf("%s/%s", imageRoot, req.hash))
}

// Create a new signature with the given id and goal level
func create(c web.C, writer http.ResponseWriter, r *http.Request) {
	req, err := GetSignatureRequest(c)
	if err != nil {
		writeTextResponse(writer, err.Error())
		return
	}

	// Construct the url for signature display page
	url := fmt.Sprintf("/s/%s/%d/%d", req.username, req.id, req.goal)

	// Check if an image already exists
	if _, err := os.Stat(fmt.Sprintf("%s/%s", imageRoot, req.hash)); os.IsExist(err) {
		http.Redirect(writer, r, url, http.StatusMovedPermanently)
		return
	}

	err = createAndSaveSignature(writer, req)
	if err != nil {
		writeTextResponse(writer, err.Error())
		return
	}

	// Redirect to the image
	http.Redirect(writer, r, url, http.StatusMovedPermanently)
}

func main() {
	log.Println("Starting go-sig/web")

	if path := os.Getenv("IMG_PATH"); path != "" {
		imageRoot = path
	}
	log.Printf("Using image root: %s", imageRoot)

	if procs := os.Getenv("PROCS"); procs != "" {
		if p, err := strconv.Atoi(procs); err != nil {
			runtime.GOMAXPROCS(p)
		}
	}

	disableLogging := os.Getenv("DISABLE_LOGGING")
	if disableLogging == "1" || disableLogging == "true" {
		// Disable logger
		log.SetOutput(new(NullWriter))
	}

	// Routes
	log.Println("Setting up routes...")
	goji.Get("/s/:username/:id/:goal", signature)
	goji.Get("/c/:username/:id/:goal", create) // todo: change to post
	// todo: handler for front page

	// Serve
	goji.Serve()
}
