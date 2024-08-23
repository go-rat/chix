package binder

import (
	"net/http"
	"reflect"
	"strings"
)

type formBinding struct{}

func (*formBinding) Name() string {
	return "form"
}

func (b *formBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)
	var err error

	if err = r.ParseForm(); err != nil {
		return err
	}

	for k, v := range r.PostForm {
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

	if err != nil {
		return err
	}

	return parse(b.Name(), out, data)
}

func (b *formBinding) BindMultipart(r *http.Request, out any) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	return parse(b.Name(), out, r.MultipartForm.Value)
}
