package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
	"image"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var (
	hasher         = md5.New()
	UsernameRegex  = regexp.MustCompile("^_?[a-zA-Z0-9-_+]+$")
	resultTemplate = pongo2.Must(pongo2.FromFile("templates/result.tpl"))
	AES_KEY        = []byte{}
)

type GoalType int

const (
	GoalLevel GoalType = iota
	GoalXP
)

type Signature struct {
	Username string
	Image    image.Image
}

type SignatureRequest struct {
	Req  ParsedSignatureRequest
	Hash string
}

type ParsedSignatureRequest struct {
	properties map[string]interface{}
}

func (s ParsedSignatureRequest) AddProperty(name string, value interface{}) {
	s.properties[name] = value
}

func (s ParsedSignatureRequest) GetProperty(name string) interface{} {
	return s.properties[name]
}

func NewSignatureRequest() ParsedSignatureRequest {
	return ParsedSignatureRequest{properties: make(map[string]interface{})}
}

func ServeResultPage(c web.C, writer http.ResponseWriter, r *http.Request, url string) {
	if err := resultTemplate.ExecuteWriter(pongo2.Context{
		"url": url,
	}, writer); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func GetMD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Format(n int) string {
	in := strconv.FormatInt(int64(n), 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func ParseUsername(username string) string {
	if AES_KEY != nil && strings.Index(username, "_") == 0 {
		hex := username[1:]
		name, err := decrypt(hex)
		if err != nil {
			return hex
		}
		return name
	}
	return username
}

func Encrypt(str string) (string, error) {
	block, err := aes.NewCipher(AES_KEY)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(str))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(str))

	return hex.EncodeToString(ciphertext), nil
}

func decrypt(str string) (plaintext string, err error) {
	ciphertext, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}

	var block cipher.Block
	if block, err = aes.NewCipher(AES_KEY); err != nil {
		return
	}

	if len(ciphertext) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return
	}

	iv := ciphertext[:aes.BlockSize]
	fmt.Println(iv)
	ciphertext = ciphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	plaintext = string(ciphertext)
	return
}
