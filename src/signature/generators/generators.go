package generators

import (
	"github.com/zenazn/goji/web"
	"signature/util"
)

type BaseGenerator interface {
	Name() string
	Url() string
	CreateSignature(req util.ParsedSignatureRequest) (util.Signature, error)
	CreateHash(req util.ParsedSignatureRequest) string
	ParseSignatureRequest(c web.C) (util.ParsedSignatureRequest, error)
}

type Generator struct {
	BaseGenerator
}
