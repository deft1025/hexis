package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesWithoutDatabase(t *testing.T) {
	r := New(nil)
	for _, path := range []string{"/api/v1/champions", "/api/v1/champions/Aatrox", "/api/v1/augments", "/api/v1/combos"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("GET %s returned %d", path, w.Code)
		}
	}
}

func TestStaticContentType(t *testing.T) {
	r := New(nil)
	for _, test := range []struct {
		path    string
		content string
	}{
		{"/", "text/html"},
		{"/champions.html", "text/html"},
		{"/champion.html", "text/html"},
		{"/augments.html", "text/html"},
		{"/items.html", "text/html"},
		{"/combos.html", "text/html"},
		{"/app.css", "text/css"},
		{"/app.js", "javascript"},
	} {
		response := httptest.NewRecorder()
		r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Header().Get("Content-Type"), test.content) {
			t.Errorf("GET %s: status %d, content type %q", test.path, response.Code, response.Header().Get("Content-Type"))
		}
	}
}
