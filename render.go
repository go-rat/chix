package chix

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/go-rat/chix/renderer"
)

var renderPool = sync.Pool{
	New: func() any {
		return new(Render)
	},
}

// Render struct
type Render struct {
	w              http.ResponseWriter
	r              *http.Request
	contentTypeSet bool
}

// NewRender creates a new NewRender instance.
func NewRender(w http.ResponseWriter, r ...*http.Request) *Render {
	render := renderPool.Get().(*Render)
	render.w = w
	if len(r) > 0 {
		render.r = r[0]
	}

	return render
}

// ContentType sets the Content-Type header for an HTTP response.
func (r *Render) ContentType(v string) {
	r.contentTypeSet = true
	r.w.Header().Set(HeaderContentType, v)
}

// Status is a warpper for WriteHeader method.
func (r *Render) Status(status int) {
	r.w.WriteHeader(status)
}

// Header sets the provided header key/value pair in the response.
func (r *Render) Header(key, value string) {
	r.w.Header().Set(key, value)
}

// Cookie sets a cookie in the response.
func (r *Render) Cookie(cookie *http.Cookie) {
	http.SetCookie(r.w, cookie)
}

// WithoutCookie deletes a cookie in the response.
func (r *Render) WithoutCookie(name string) {
	http.SetCookie(r.w, &http.Cookie{
		Name:   name,
		MaxAge: -1,
	})
}

// Redirect replies to the request with a redirect to url, which may be a path
// relative to the request path.
func (r *Render) Redirect(url string) {
	if r.r == nil {
		http.Error(r.w, "chix: Redirect requires passing *http.Request", http.StatusInternalServerError)
		return
	}

	http.Redirect(r.w, r.r, url, http.StatusFound)
}

// RedirectPermanent replies to the request with a redirect to url, which may be
// a path relative to the request path.
func (r *Render) RedirectPermanent(url string) {
	if r.r == nil {
		http.Error(r.w, "chix: RedirectPermanent requires passing *http.Request", http.StatusInternalServerError)
		return
	}

	http.Redirect(r.w, r.r, url, http.StatusMovedPermanently)
}

// PlainText writes a string to the response, setting the Content-Type as
// text/plain if not set.
func (r *Render) PlainText(v string) {
	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	}
	_, _ = r.w.Write([]byte(v))
}

// Data writes raw bytes to the response, setting the Content-Type as
// application/octet-stream if not set.
func (r *Render) Data(v []byte) {
	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMEOctetStream)
	}
	_, _ = r.w.Write(v)
}

// HTML writes a string to the response, setting the Content-Type as text/html
// if not set.
func (r *Render) HTML(v string) {
	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	}
	_, _ = r.w.Write([]byte(v))
}

// JSON marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/json if not set.
func (r *Render) JSON(v any) {
	buf := new(bytes.Buffer)
	enc := JSONEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	}
	_, _ = r.w.Write(buf.Bytes())
}

// JSONP marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/javascript if not set.
func (r *Render) JSONP(callback string, v any) {
	buf := new(bytes.Buffer)
	enc := JSONEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMEApplicationJavaScriptCharsetUTF8)
	}
	_, _ = r.w.Write([]byte(callback + "("))
	_, _ = r.w.Write(buf.Bytes())
	_, _ = r.w.Write([]byte(");"))
}

// XML marshals 'v' to XML, setting the Content-Type as application/xml if not set. It
// will automatically prepend a generic XML header (see encoding/xml.Header) if
// one is not found in the first 100 bytes of 'v'.
func (r *Render) XML(v any) {
	buf := new(bytes.Buffer)
	enc := XMLEncoder(buf)
	if err := enc.Encode(v); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	}

	// Try to find <?xml header in first 100 bytes (just in case there're some XML comments).
	findHeaderUntil := buf.Len()
	if findHeaderUntil > 100 {
		findHeaderUntil = 100
	}
	if !bytes.Contains(buf.Bytes()[:findHeaderUntil], []byte("<?xml")) {
		// No header found. Print it out first.
		_, _ = r.w.Write([]byte(xml.Header))
	}

	_, _ = r.w.Write(buf.Bytes())
}

// NoContent returns a HTTP 204 "No Content" response.
func (r *Render) NoContent() {
	r.Status(http.StatusNoContent)
}

// Stream sends a streaming response and returns a boolean
// indicates "Is client disconnected in middle of stream"
func (r *Render) Stream(step func(w io.Writer) bool) bool {
	if r.r == nil {
		http.Error(r.w, "chix: Stream requires passing *http.Request", http.StatusInternalServerError)
		return false
	}

	for {
		select {
		case <-r.r.Context().Done():
			return true
		default:
			keepOpen := step(r.w)
			r.Flush()
			if !keepOpen {
				return false
			}
		}
	}
}

// EventStream writes a stream of JSON objects from a channel to the response and setting the
// Content-Type as text/event-stream if not set.
func (r *Render) EventStream(v any) {
	if r.r == nil {
		http.Error(r.w, "chix: EventStream requires passing *http.Request", http.StatusInternalServerError)
		return
	}
	if reflect.TypeOf(v).Kind() != reflect.Chan {
		http.Error(r.w, fmt.Sprintf("chix: EventStream expects a channel, not %v", reflect.TypeOf(v).Kind()), http.StatusInternalServerError)
		return
	}

	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMEEventStreamCharsetUTF8)
	}
	r.w.Header().Set(HeaderCacheControl, "no-cache")

	if r.r.ProtoMajor == 1 {
		// An endpoint MUST NOT generate an HTTP/2 message containing connection-specific header fields.
		// Source: RFC7540
		r.w.Header().Set(HeaderConnection, "keep-alive")
	}

	r.w.WriteHeader(http.StatusOK)

	ctx := r.r.Context()
	buf := new(strings.Builder)
	enc := JSONEncoder(buf)
	enc.SetEscapeHTML(true)
	for {
		switch chosen, recv, ok := reflect.Select([]reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())},
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(v)},
		}); chosen {
		case 0: // equivalent to: case <-ctx.Done()
			_, _ = r.w.Write([]byte("event: error\ndata: {\"error\":\"Server Timeout\"}\n\n"))
			return

		default: // equivalent to: case v, ok := <-stream
			if !ok {
				_, _ = r.w.Write([]byte("event: EOF\n\n"))
				return
			}

			v := recv.Interface()
			if err := enc.Encode(v); err != nil {
				_, _ = fmt.Fprintf(r.w, "event: error\ndata: {\"error\":\"%v\"}\n\n", err)
				r.Flush()
				continue
			}

			_, _ = fmt.Fprintf(r.w, "event: data\ndata: %s\n\n", buf.String())
			r.Flush()
			buf.Reset()
		}
	}
}

// SSEvent writes a Server-Sent Event to the response and setting the
// Content-Type as text/event-stream if not set.
func (r *Render) SSEvent(event renderer.SSEvent) {
	if r.r == nil {
		http.Error(r.w, "chix: SSEvent requires passing *http.Request", http.StatusInternalServerError)
		return
	}

	if !r.contentTypeSet {
		r.w.Header().Set(HeaderContentType, MIMEEventStreamCharsetUTF8)
	}
	r.w.Header().Set(HeaderCacheControl, "no-cache")

	if r.r.ProtoMajor == 1 {
		// An endpoint MUST NOT generate an HTTP/2 message containing connection-specific header fields.
		// Source: RFC7540
		r.w.Header().Set(HeaderConnection, "keep-alive")
	}

	r.w.WriteHeader(http.StatusOK)
	_ = renderer.SSEventEncode(r.w, event)
}

// File sends a file to the response.
func (r *Render) File(filepath string) {
	if r.r == nil {
		http.Error(r.w, "chix: File requires passing *http.Request", http.StatusInternalServerError)
		return
	}

	http.ServeFile(r.w, r.r, filepath)
}

// Download sends a file to the response and prompting it to be downloaded
// by setting the Content-Disposition header.
func (r *Render) Download(filepath, filename string) {
	if r.r == nil {
		http.Error(r.w, "chix: Download requires passing *http.Request", http.StatusInternalServerError)
		return
	}
	if isASCII(filename) {
		r.Header(HeaderContentDisposition, `attachment; filename="`+quoteEscape(filename)+`"`)
	} else {
		r.Header(HeaderContentDisposition, `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}

	http.ServeFile(r.w, r.r, filepath)
}

// Flush sends any buffered data to the response.
func (r *Render) Flush() {
	if f, ok := r.w.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack returns the underlying Hijacker interface.
func (r *Render) Hijack() (http.Hijacker, bool) {
	h, ok := r.w.(http.Hijacker)
	return h, ok
}

// Release puts the Render instance back into the pool.
func (r *Render) Release() {
	r.w = nil
	r.r = nil
	r.contentTypeSet = false
	renderPool.Put(r)
}
