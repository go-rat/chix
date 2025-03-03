package chix

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/go-rat/chix/binder"
)

var bindPool = sync.Pool{
	New: func() any {
		return new(Bind)
	},
}

// Bind struct
type Bind struct {
	r               *http.Request
	enableSplitting bool
}

// NewBind creates a new Bind instance.
func NewBind(r *http.Request, enableSplitting ...bool) *Bind {
	b := bindPool.Get().(*Bind)
	b.r = r
	if len(enableSplitting) > 0 {
		b.enableSplitting = enableSplitting[0]
	}

	return b
}

// Header binds the request header strings into the struct, map[string]string and map[string][]string.
func (b *Bind) Header(out any) error {
	return binder.HeaderBinder.Bind(b.r, out)
}

// Cookie binds the request cookie strings into the struct, map[string]string and map[string][]string.
// NOTE: If your cookie is like key=val1,val2; they'll be binded as an slice if your map is map[string][]string. Else, it'll use last element of cookie.
func (b *Bind) Cookie(out any) error {
	return binder.CookieBinder.Bind(b.r, out)
}

// Query binds the query string into the struct, map[string]string and map[string][]string.
func (b *Bind) Query(out any) error {
	return binder.QueryBinder.Bind(b.r, out)
}

// JSON binds the body string into the struct.
func (b *Bind) JSON(out any) error {
	return binder.JSONBinder.Bind(JSONDecoder(b.r.Body), out)
}

// XML binds the body string into the struct.
func (b *Bind) XML(out any) error {
	return binder.XMLBinder.Bind(XMLDecoder(b.r.Body), out)
}

// Form binds the form into the struct, map[string]string and map[string][]string.
func (b *Bind) Form(out any) error {
	return binder.FormBinder.Bind(b.r, out)
}

// URI binds the route parameters into the struct, map[string]string and map[string][]string.
func (b *Bind) URI(out any) error {
	ctx := chi.RouteContext(b.r.Context())
	return binder.URIBinder.Bind(ctx.URLParams.Keys, ctx.URLParam, out)
}

// MultipartForm binds the multipart form into the struct, map[string]string and map[string][]string.
// Parameter size is the maximum memory in bytes used to parse the form, default is 32MB.
func (b *Bind) MultipartForm(out any, size ...int64) error {
	if len(size) == 0 {
		size = append(size, 32768<<10) // 32MB
	}

	return binder.FormBinder.BindMultipart(b.r, out, size[0])
}

// Body binds the request body into the struct, map[string]string and map[string][]string.
// It supports decoding the following content types based on the Content-Type header:
// application/json, application/xml, application/x-www-form-urlencoded, multipart/form-data
// If none of the content types above are matched, it'll take a look custom binders by checking the MIMETypes() method of custom binder.
// If there're no custom binder for mşme type of body, it will return a ErrUnprocessableEntity error.
func (b *Bind) Body(out any) error {
	// Get content-type
	ctype := strings.ToLower(b.r.Header.Get("Content-Type"))
	ctype = binder.FilterFlags(parseVendorSpecificContentType(ctype))

	// Parse body accordingly
	switch ctype {
	case MIMEApplicationJSON:
		return b.JSON(out)
	case MIMETextXML, MIMEApplicationXML:
		return b.XML(out)
	case MIMEApplicationForm:
		return b.Form(out)
	case MIMEMultipartForm:
		return b.MultipartForm(out)
	}

	// No suitable content type found
	return errors.New(http.StatusText(http.StatusUnprocessableEntity))
}

// Release releases the Bind instance back into the pool.
func (b *Bind) Release() {
	b.r = nil
	b.enableSplitting = false
	bindPool.Put(b)
}
