package binder

import (
	"encoding/xml"
)

type xmlBinding struct{}

func (*xmlBinding) Name() string {
	return "xml"
}

func (*xmlBinding) Bind(xmlDecoder *xml.Decoder, out any) error {
	return xmlDecoder.Decode(out)
}
