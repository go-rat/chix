package chix_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/libtnb/chix"
)

func TestBind_HeaderBindsCorrectly(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Test-Header", "test-value")
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.Header(&out)
	require.NoError(t, err)
	require.Equal(t, "test-value", out["X-Test-Header"])
}

func TestBind_CookieBindsCorrectly(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "test-cookie", Value: "test-value"})
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.Cookie(&out)
	require.NoError(t, err)
	require.Equal(t, "test-value", out["test-cookie"])
}

func TestBind_QueryBindsCorrectly(t *testing.T) {
	req := httptest.NewRequest("GET", "/?key=value", nil)
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.Query(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out["key"])
}

func TestBind_JSONBindsCorrectly(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.JSON(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out["key"])
}

func TestBind_XMLBindsCorrectly(t *testing.T) {
	type XMLData struct {
		Key string `xml:"key"`
	}

	req := httptest.NewRequest("POST", "/", strings.NewReader(`<XMLData><key>value</key></XMLData>`))
	req.Header.Set("Content-Type", "application/xml")
	b := chix.NewBind(req)
	out := XMLData{}
	err := b.XML(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out.Key)
}

func TestBind_FormBindsCorrectly(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("key=value"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.Form(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out["key"])
}

func TestBind_URIBindsCorrectly(t *testing.T) {
	req := httptest.NewRequest("GET", "/test/value", nil)
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("key", "value")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.URI(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out["key"])
}

func TestBind_MultipartFormBindsCorrectly(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	require.NoError(t, writer.WriteField("key", "value"))
	require.NoError(t, writer.Close())
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.MultipartForm(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out["key"])
}

func TestBind_BodyBindsJSONCorrectly(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"key":"value"}`))
	req.Header.Set("Content-Type", "application/json")
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.Body(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out["key"])
}

func TestBind_BodyBindsXMLCorrectly(t *testing.T) {
	type XMLData struct {
		Key string `xml:"key"`
	}

	req := httptest.NewRequest("POST", "/", strings.NewReader(`<XMLData><key>value</key></XMLData>`))
	req.Header.Set("Content-Type", "application/xml")
	b := chix.NewBind(req)
	out := XMLData{}
	err := b.Body(&out)
	require.NoError(t, err)
	require.Equal(t, "value", out.Key)
}

func TestBind_BodyReturnsErrorForUnsupportedContentType(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader("unsupported content"))
	req.Header.Set("Content-Type", "text/plain")
	b := chix.NewBind(req)
	out := make(map[string]string)
	err := b.Body(&out)
	require.Error(t, err)
}
