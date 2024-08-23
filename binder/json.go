package binder

import "encoding/json"

type jsonBinding struct{}

func (*jsonBinding) Name() string {
	return "json"
}

func (*jsonBinding) Bind(jsonDecoder *json.Decoder, out any) error {
	return jsonDecoder.Decode(out)
}
