package binder

import (
	"net/http"
	"reflect"
	"strings"
)

type headerBinding struct{}

func (*headerBinding) Name() string {
	return "header"
}

func (b *headerBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)

	for k, v := range r.Header {
		v := strings.Join(v, ",")

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
