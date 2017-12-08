package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/flosch/pongo2"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	md5Hash        = md5.New()
	UsernameRegex  = regexp.MustCompile("^_?[a-zA-Z0-9-_+]+$")
	resultTemplate = pongo2.Must(pongo2.FromFile("resources/templates/result.tpl"))
	AesKey []byte
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

func ServeResultPage(writer http.ResponseWriter, url string) {
	if err := resultTemplate.ExecuteWriter(pongo2.Context{
		"url": url,
	}, writer); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func GetMD5(text string) string {
	md5Hash = md5.New()
	md5Hash.Write([]byte(text))
	return hex.EncodeToString(md5Hash.Sum(nil))
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
	if AesKey != nil && strings.Index(username, "_") == 0 {
		nameHex := username[1:]
		name, err := decrypt(nameHex)
		if err != nil {
			return nameHex
		}
		return name
	}
	return username
}

func Encrypt(str string) (string, error) {
	block, err := aes.NewCipher(AesKey)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(str))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(str))

	return hex.EncodeToString(cipherText), nil
}

func decrypt(str string) (plaintext string, err error) {
	cipherText, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}

	var block cipher.Block
	if block, err = aes.NewCipher(AesKey); err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("ciphertext too short")
		return
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)

	plaintext = string(cipherText)
	return
}

func FromSuffixed(value string) (int, error) {
	lastCharacter := string(value[len(value)-1])
	if lastCharacter != "k" && lastCharacter != "m" {
		if val, err := strconv.Atoi(value); err == nil {
			return val, nil
		} else {
			return 0, errors.New("invalid suffix")
		}
	}
	noSuffix := value[:len(value)-1]
	number, err := strconv.Atoi(noSuffix)
	if lastCharacter == "k" && err == nil {
		return number * 1000, nil
	} else if lastCharacter == "m" && err == nil {
		return number * 1000000, nil
	}
	val, err := strconv.Atoi(value)
	if err != nil {
		val = 0
	}
	return val, errors.New("failed to parse suffixed number")
}

// Parse query parameters and return them in the right order
func ParseQueryParameters(query string) (params []QueryParameter, err error) {
	for query != "" {
		key := query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value := ""
		if i := strings.Index(key, "="); i >= 0 {
			key, value = key[:i], key[i+1:]
		}
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		params = append(params, QueryParameter{key, value})
	}
	return params, err
}

type QueryParameter struct {
	Key   string
	Value string
}

// Load font(s) to memory
func LoadFont(fontFile string) *truetype.Font {
	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		panic(err)
	}
	baseFont, err := freetype.ParseFont(fontBytes)
	if err != nil {
		panic(err)
	}
	return baseFont
}

func GetGoalType(skill Skill, goal int) GoalType {
	goalType := GoalLevel
	if (skill.Id == InventionId && goal > InventionLevelMax) || (skill.Id != InventionId && goal > LevelMax) {
		goalType = GoalXP
	}
	return goalType
}
