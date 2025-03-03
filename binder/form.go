package binder

import (
	"mime/multipart"
	"net/http"
	"strings"
)

type formBinding struct {
	EnableSplitting bool
}

func (*formBinding) Name() string {
	return "form"
}

func (b *formBinding) Bind(r *http.Request, out any) error {
	data := make(map[string][]string)

	if err := r.ParseForm(); err != nil {
		return err
	}

	for k, v := range r.PostForm {
		if err := formatBindData(out, data, k, strings.Join(v, ","), b.EnableSplitting, true); err != nil {
			return err
		}
	}

	return parse(b.Name(), out, data)
}

func (b *formBinding) BindMultipart(r *http.Request, out any, size int64) error {
	if err := r.ParseMultipartForm(size); err != nil {
		return err
	}

	data := make(map[string][]string)
	for key, values := range r.MultipartForm.Value {
		if err := formatBindData(out, data, key, values, b.EnableSplitting, true); err != nil {
			return err
		}
	}

	files := make(map[string][]*multipart.FileHeader)
	for key, values := range r.MultipartForm.File {
		if err := formatBindData(out, files, key, values, b.EnableSplitting, true); err != nil {
			return err
		}
	}

	return parse(b.Name(), out, data, files)
}
