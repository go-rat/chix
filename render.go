package chix

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

// Render struct
type Render struct {
	w http.ResponseWriter
	r *http.Request
}

// NewRender creates a new NewRender instance.
func NewRender(w http.ResponseWriter, r ...*http.Request) *Render {
	if len(r) == 0 {
		return &Render{w: w}
	}

	return &Render{w: w, r: r[0]}
}

// Status is a warpper for WriteHeader method.
func (r *Render) Status(status int) *Render {
	r.w.WriteHeader(status)
	return r
}

// Header sets the provided header key/value pair in the response.
func (r *Render) Header(key, value string) *Render {
	r.w.Header().Set(key, value)
	return r
}

// Cookie sets a cookie in the response.
func (r *Render) Cookie(cookie *http.Cookie) *Render {
	http.SetCookie(r.w, cookie)
	return r
}

// WithoutCookie deletes a cookie in the response.
func (r *Render) WithoutCookie(name string) *Render {
	http.SetCookie(r.w, &http.Cookie{
		Name:   name,
		MaxAge: -1,
	})
	return r
}

// Redirect replies to the request with a redirect to url, which may be a path
// relative to the request path.
func (r *Render) Redirect(url string) *Render {
	if r.r == nil {
		http.Error(r.w, "chix: Redirect requires passing *http.Request", http.StatusInternalServerError)
	}

	http.Redirect(r.w, r.r, url, http.StatusFound)
	return r
}

// PlainText writes a string to the response, setting the Content-Type as
// text/plain.
func (r *Render) PlainText(v string) *Render {
	r.w.Header().Set(HeaderContentType, MIMETextPlainCharsetUTF8)
	_, _ = r.w.Write([]byte(v))
	return r
}

// Data writes raw bytes to the response, setting the Content-Type as
// application/octet-stream.
func (r *Render) Data(v []byte) *Render {
	r.w.Header().Set(HeaderContentType, MIMEOctetStream)
	_, _ = r.w.Write(v)
	return r
}

// HTML writes a string to the response, setting the Content-Type as text/html.
func (r *Render) HTML(v string) *Render {
	r.w.Header().Set(HeaderContentType, MIMETextHTMLCharsetUTF8)
	_, _ = r.w.Write([]byte(v))
	return r
}

// JSON marshals 'v' to JSON, automatically escaping HTML and setting the
// Content-Type as application/json.
func (r *Render) JSON(v any) *Render {
	buf := new(bytes.Buffer)
	enc := JSONEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return r
	}

	r.w.Header().Set(HeaderContentType, MIMEApplicationJSONCharsetUTF8)
	_, _ = r.w.Write(buf.Bytes())
	return r
}

// XML marshals 'v' to XML, setting the Content-Type as application/xml. It
// will automatically prepend a generic XML header (see encoding/xml.Header) if
// one is not found in the first 100 bytes of 'v'.
func (r *Render) XML(v any) *Render {
	buf := new(bytes.Buffer)
	enc := XMLEncoder(buf)
	if err := enc.Encode(v); err != nil {
		http.Error(r.w, err.Error(), http.StatusInternalServerError)
		return r
	}

	r.w.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)

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
	return r
}

// NoContent returns a HTTP 204 "No Content" response.
func (r *Render) NoContent() *Render {
	return r.Status(http.StatusNoContent)
}

// EventStream writes a stream of JSON objects from a channel to the response and setting the
// Content-Type as text/event-stream.
func (r *Render) EventStream(v any) *Render {
	if r.r == nil {
		http.Error(r.w, "chix: Redirect requires passing *http.Request", http.StatusInternalServerError)
	}
	if reflect.TypeOf(v).Kind() != reflect.Chan {
		http.Error(r.w, fmt.Sprintf("render: event stream expects a channel, not %v", reflect.TypeOf(v).Kind()), http.StatusInternalServerError)
	}

	r.w.Header().Set(HeaderContentType, MIMEEventStreamCharsetUTF8)
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
			return r

		default: // equivalent to: case v, ok := <-stream
			if !ok {
				_, _ = r.w.Write([]byte("event: EOF\n\n"))
				return r
			}

			v := recv.Interface()
			if err := enc.Encode(v); err != nil {
				_, _ = r.w.Write([]byte(fmt.Sprintf("event: error\ndata: {\"error\":\"%v\"}\n\n", err)))
				r.Flush()
				continue
			}

			_, _ = r.w.Write([]byte(fmt.Sprintf("event: data\ndata: %s\n\n", buf.String())))
			r.Flush()
			buf.Reset()
		}
	}
}

// File sends a file to the client.
func (r *Render) File(filepath string) *Render {
	if r.r == nil {
		http.Error(r.w, "chix: File requires passing *http.Request", http.StatusInternalServerError)
	}

	http.ServeFile(r.w, r.r, filepath)
	return r
}

// Download sends a file to the client prompting it to be downloaded.
func (r *Render) Download(filepath, filename string) *Render {
	if r.r == nil {
		http.Error(r.w, "chix: Download requires passing *http.Request", http.StatusInternalServerError)
	}
	if isASCII(filename) {
		r.Header(HeaderContentDisposition, `attachment; filename="`+quoteEscape(filename)+`"`)
	} else {
		r.Header(HeaderContentDisposition, `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}

	http.ServeFile(r.w, r.r, filepath)
	return r
}

// Flush sends any buffered data to the client.
func (r *Render) Flush() *Render {
	if f, ok := r.w.(http.Flusher); ok {
		f.Flush()
	}
	return r
}
