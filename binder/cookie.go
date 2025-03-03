package binder

import (
	"net/http"
)

type cookieBinding struct {
	EnableSplitting bool
}

func (*cookieBinding) Name() string {
	return "cookie"
}

func (b *cookieBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)

	for _, cookie := range r.Cookies() {
		k := cookie.Name
		v := cookie.Value

		if err := formatBindData(out, data, k, v, b.EnableSplitting, true); err != nil {
			return err
		}
	}

	return parse(b.Name(), out, data)
}
