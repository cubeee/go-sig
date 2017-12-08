package main

import (
	"bufio"
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

	"github.com/cubeee/go-sig/signature/generators"
	"github.com/cubeee/go-sig/signature/generators/rs3"
	"github.com/cubeee/go-sig/signature/generators/rs3/multi"
	"github.com/cubeee/go-sig/signature/util"
	"github.com/cubeee/go-sig/signature"
)

type NullWriter int

func (NullWriter) Write([]byte) (int, error) {
	return 0, nil
}

var (
	indexTemplate  = pongo2.Must(pongo2.FromFile("resources/templates/index.tpl"))
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func createAndSaveSignature(req util.SignatureRequest, generator generators.BaseGenerator) error {
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
	out, err := os.Create(vars.ImageRoot + "/" + hash)
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
	imagePath := fmt.Sprintf("%s/%s", vars.ImageRoot, req.Hash)
	if stat, err := os.Stat(imagePath); err == nil {
		modTime := stat.ModTime()
		now := time.Now()
		age := now.Sub(modTime)

		if age.Minutes() >= vars.UpdateInterval {
			err = createAndSaveSignature(req, generator)
			if err != nil {
				writeTextResponse(writer, err.Error())
				return
			}
		}
	}
}

// Write text as a response to the client
func writeTextResponse(writer http.ResponseWriter, text string) {
	fmt.Fprintf(writer, text)
}

// Show an existing signature
func serveSignature(writer http.ResponseWriter, r *http.Request, req util.SignatureRequest, generator generators.BaseGenerator) {
	attemptUpdate := true

	// Check if an image already exists and create it if not
	if _, err := os.Stat(fmt.Sprintf("%s/%s", vars.ImageRoot, req.Hash)); os.IsNotExist(err) {
		err = createAndSaveSignature(req, generator)
		if err != nil {
			writeTextResponse(writer, err.Error())
			return
		}
		attemptUpdate = false
	}

	if attemptUpdate {
		updateSignature(writer, req, generator)
	}

	http.ServeFile(writer, r, fmt.Sprintf("%s/%s", vars.ImageRoot, req.Hash))
}

// Front page
func index(_ web.C, writer http.ResponseWriter, _ *http.Request) {
	if err := indexTemplate.ExecuteWriter(pongo2.Context{
		"skills":  util.SkillNames,
		"has_aes": util.AesKey,
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
		req := util.SignatureRequest{Req: parsedReq, Hash: hash}

		serveSignature(writer, request, req, generator)
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
	disableLogging := os.Getenv("DISABLE_LOGGING")
	if disableLogging == "1" || disableLogging == "true" {
		log.SetOutput(new(NullWriter))
	}

	log.Println("Starting go-sig")

	if secure := os.Getenv("SECURE"); secure != "" {
		if sec, err := strconv.ParseBool(secure);  err == nil {
			log.Println("secure:", sec)
			proto := "https"
			if !sec {
				proto = "http"
			}
			vars.Protocol = proto
		} else {
			log.Println(err.Error())
		}
	}

	if vHost := os.Getenv("VIRTUAL_HOST"); vHost != "" {
		vars.VirtualHost = vHost
	}
	log.Printf("Using virtual host: %s", vars.VirtualHost)

	if path := os.Getenv("IMG_PATH"); path != "" {
		vars.ImageRoot = path
	}
	log.Printf("Using image root: %s", vars.ImageRoot)
	if _, err := os.Stat(vars.ImageRoot); os.IsNotExist(err) {
		os.MkdirAll(vars.ImageRoot, 0740)
	}

	if key := os.Getenv("AES_KEY"); key != "" {
		util.AesKey = []byte(key)
	}

	if procs := os.Getenv("PROCS"); procs != "" {
		if p, err := strconv.Atoi(procs); err != nil {
			runtime.GOMAXPROCS(p)
		}
	}

	// Routes
	log.Println("Mapping routes...")
	goji.Get("/", index)

	// Setup static files
	static := web.New()
	static.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(vars.PublicPath))))
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
