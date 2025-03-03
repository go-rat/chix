package binder

import (
	"net/http"
	"strings"
)

type headerBinding struct {
	EnableSplitting bool
}

func (*headerBinding) Name() string {
	return "header"
}

func (b *headerBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)

	for k, v := range r.Header {
		if err := formatBindData(out, data, k, strings.Join(v, ","), b.EnableSplitting, false); err != nil {
			return err
		}
	}

	return parse(b.Name(), out, data)
}
