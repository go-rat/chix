package chix_test

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-rat/chix"
)

func TestRender_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.ContentType("application/json")
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestRender_Status(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.Status(http.StatusNotFound)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestRender_Header(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.Header("X-Custom-Header", "value")
	require.Equal(t, "value", w.Header().Get("X-Custom-Header"))
}

func TestRender_Cookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	cookie := &http.Cookie{Name: "test", Value: "value"}
	r.Cookie(cookie)
	require.Equal(t, "test=value", w.Header().Get("Set-Cookie"))
}

func TestRender_WithoutCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.WithoutCookie("test")
	require.Equal(t, "test=; Max-Age=0", w.Header().Get("Set-Cookie"))
}

func TestRender_PlainText(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.PlainText("hello")
	require.Equal(t, "hello", w.Body.String())
}

func TestRender_Data(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.Data([]byte("data"))
	require.Equal(t, "data", w.Body.String())
}

func TestRender_HTML(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.HTML("<p>hello</p>")
	require.Equal(t, "<p>hello</p>", w.Body.String())
}

func TestRender_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.JSON(map[string]string{"key": "value"})
	require.Equal(t, `{"key":"value"}`+"\n", w.Body.String())
}

func TestRender_JSONP(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.JSONP("callback", map[string]string{"key": "value"})
	require.Equal(t, `callback({"key":"value"}`+"\n"+`);`, w.Body.String())
}

func TestRender_XML(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)

	type KeyValue struct {
		XMLName xml.Name `xml:"map"`
		Key     string   `xml:"key"`
		Value   string   `xml:"value"`
	}

	data := KeyValue{Key: "key", Value: "value"}
	r.XML(data)
	require.Equal(t, xml.Header+`<map><key>key</key><value>value</value></map>`, w.Body.String())
}

func TestRender_NoContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := chix.NewRender(w)
	r.NoContent()
	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestRender_File(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	f, err := os.CreateTemp("", "test.txt")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("test file content")
	require.NoError(t, err)
	r := chix.NewRender(w, req)
	r.File(f.Name())
	require.Equal(t, "test file content", w.Body.String())
}

func TestRender_Download(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	f, err := os.CreateTemp("", "test.txt")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("test file content")
	require.NoError(t, err)
	r := chix.NewRender(w, req)
	r.Download(f.Name(), "test.txt")
	require.Equal(t, `attachment; filename="test.txt"`, w.Header().Get("Content-Disposition"))
	require.Equal(t, "test file content", w.Body.String())
}
