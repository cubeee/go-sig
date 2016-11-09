package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"

	"signature/generators"
	"signature/generators/rs3"
	"signature/generators/rs3/multi"
	"signature/util"
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) {
	return 0, nil
}

type SignatureRequest struct {
	hash      string
	generator generators.BaseGenerator
}

var (
	imageRoot      = "images"
	publicPath     = "public/"
	updateInterval = 10.0
	indexTemplate  = pongo2.Must(pongo2.FromFile("templates/index.tpl"))
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func createAndSaveSignature(writer http.ResponseWriter, req util.SignatureRequest, generator generators.BaseGenerator) error {
	// Create the signature image
	sig, err := generator.CreateSignature(req.Req)
	if err != nil {
		return err
	}

	// note: queue saving if it causes performance issues?
	// Save the image to disk with the given hash as the file name
	saveImage(req.Hash, sig.Image)
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

// Update the signature based on the image's last modification date
func updateSignature(writer http.ResponseWriter, req util.SignatureRequest, generator generators.BaseGenerator) {
	imagePath := fmt.Sprintf("%s/%s", imageRoot, req.Hash)
	if stat, err := os.Stat(imagePath); err == nil {
		modTime := stat.ModTime()
		now := time.Now()
		age := now.Sub(modTime)

		if age.Minutes() >= updateInterval {
			err = createAndSaveSignature(writer, req, generator)
			if err != nil {
				writeTextResponse(writer, err.Error())
				return
			}
		}
	}
}

// Write an image as a response to the client
func writeImageResponse(writer http.ResponseWriter, signature util.Signature) {
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
func serveSignature(c web.C, writer http.ResponseWriter, r *http.Request, req util.SignatureRequest, generator generators.BaseGenerator) {
	attemptUpdate := true

	// Check if an image already exists and create it if not
	if _, err := os.Stat(fmt.Sprintf("%s/%s", imageRoot, req.Hash)); os.IsNotExist(err) {
		err = createAndSaveSignature(writer, req, generator)
		if err != nil {
			writeTextResponse(writer, err.Error())
			return
		}
		attemptUpdate = false
	}

	if attemptUpdate {
		updateSignature(writer, req, generator)
	}

	http.ServeFile(writer, r, fmt.Sprintf("%s/%s", imageRoot, req.Hash))
}

// Front page
func index(c web.C, writer http.ResponseWriter, r *http.Request) {
	if err := indexTemplate.ExecuteWriter(pongo2.Context{
		"skills":  util.SkillNames,
		"has_aes": util.AES_KEY,
	}, writer); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func registerGenerator(generator generators.BaseGenerator) {
	goji.Get(generator.Url(), func(c web.C, writer http.ResponseWriter, request *http.Request) {
		parsedReq, err := generator.ParseSignatureRequest(c, request)
		if err != nil {
			writeTextResponse(writer, "Failed to parse the request: "+err.Error())
			return
		}
		hash := finalizeHash(generator.Name(), generator.CreateHash(parsedReq))
		req := util.SignatureRequest{parsedReq, hash}

		serveSignature(c, writer, request, req, generator)
	})

	formUrl := generator.FormUrl()
	if formUrl != "" {
		goji.Post(formUrl, func(c web.C, writer http.ResponseWriter, request *http.Request) {
			generator.HandleForm(c, writer, request)
		})
	}
}

func finalizeHash(name, hash string) string {
	return fmt.Sprintf("%s-%s", name, hash)
}

func main() {
	log.Println("Starting go-sig/web")

	if path := os.Getenv("IMG_PATH"); path != "" {
		imageRoot = path
	}
	log.Printf("Using image root: %s", imageRoot)
	if _, err := os.Stat(imageRoot); os.IsNotExist(err) {
		os.MkdirAll(imageRoot, 0750)
	}

	if procs := os.Getenv("PROCS"); procs != "" {
		if p, err := strconv.Atoi(procs); err != nil {
			runtime.GOMAXPROCS(p)
		}
	}

	if key := os.Getenv("AES_KEY"); key != "" {
		util.AES_KEY = []byte(key)
	}

	disableLogging := os.Getenv("DISABLE_LOGGING")
	if disableLogging == "1" || disableLogging == "true" {
		log.SetOutput(new(NullWriter))
	}

	// Routes
	log.Println("Mapping routes...")
	goji.Get("/", index)

	// Setup static files
	static := web.New()
	static.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(publicPath))))
	http.Handle("/assets/", static)

	profile := os.Getenv("ENABLE_DEBUG")
	if profile == "1" || profile == "true" {
		log.Println("Mapping debug routes...")
		goji.Handle("/debug/pprof/", pprof.Index)
		goji.Handle("/debug/pprof/cmdline", pprof.Cmdline)
		goji.Handle("/debug/pprof/profile", pprof.Profile)
		goji.Handle("/debug/pprof/symbol", pprof.Symbol)
		goji.Handle("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
		goji.Handle("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
		goji.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
		goji.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	}

	// Generators
	log.Println("Registering generators...")
	registerGenerator(new(rs3.BoxGoalGenerator))
	registerGenerator(new(multi.MultiGoalGenerator))
	//registerGenerator(new(rs3.ExampleGenerator))

	// Serve
	goji.Serve()
}
