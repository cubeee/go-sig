package generators

import (
	"github.com/zenazn/goji/web"
	"net/http"
	"github.com/cubeee/go-sig/signature/util"
)

type BaseGenerator interface {
	Name() string
	Url() string
	FormUrl() string
	CreateSignature(req util.ParsedSignatureRequest) (util.Signature, error)
	CreateHash(req util.ParsedSignatureRequest) string
	ParseSignatureRequest(c web.C, r *http.Request) (util.ParsedSignatureRequest, error)
	HandleForm(c web.C, writer http.ResponseWriter, request *http.Request)
}

type Generator struct {
	BaseGenerator
}
