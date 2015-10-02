package util

import (
	"crypto/md5"
	"encoding/hex"
	"image"
	"regexp"
	"strconv"
)

var (
	Salt          = "ded3b63a6f11a9efd8e0f6a9b84fbeb1"
	hasher        = md5.New()
	UsernameRegex = regexp.MustCompile("^[a-zA-Z0-9-_+]+$")
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
