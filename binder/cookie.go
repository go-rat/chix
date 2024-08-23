package binder

import (
	"net/http"
	"reflect"
	"strings"
)

type cookieBinding struct{}

func (*cookieBinding) Name() string {
	return "cookie"
}

func (b *cookieBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)

	for _, cookie := range r.Cookies() {
		k := cookie.Name
		v := cookie.Value

		if strings.Contains(v, ",") && equalFieldType(out, reflect.Slice, k) {
			values := strings.Split(v, ",")
			for i := 0; i < len(values); i++ {
				data[k] = append(data[k], values[i])
			}
		} else {
			data[k] = append(data[k], v)
		}
	}

	return parse(b.Name(), out, data)
}
