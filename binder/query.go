package binder

import (
	"net/http"
	"strings"
)

type queryBinding struct {
	EnableSplitting bool
}

func (*queryBinding) Name() string {
	return "query"
}

func (b *queryBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)

	for k, v := range r.URL.Query() {
		if err := formatBindData(out, data, k, strings.Join(v, ","), b.EnableSplitting, true); err != nil {
			return err
		}
	}

	return parse(b.Name(), out, data)
}
