package binder

import (
	"net/http"
	"reflect"
	"strings"
)

type queryBinding struct{}

func (*queryBinding) Name() string {
	return "query"
}

func (b *queryBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)
	var err error

	for k, v := range r.URL.Query() {
		if err != nil {
			return err
		}

		v := strings.Join(v, ",")

		if strings.Contains(k, "[") {
			k, err = parseParamSquareBrackets(k)
		}

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
