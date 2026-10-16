package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutesWithoutDatabase(t *testing.T) {
	r := New(nil)
	for _, path := range []string{"/api/v1/champions", "/api/v1/champions/Aatrox", "/api/v1/champions/Aatrox/skills", "/api/v1/augments", "/api/v1/items", "/api/v1/combos", "/api/v1/guides"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("GET %s returned %d", path, w.Code)
		}
	}
}

func TestGuidePreviewRequiresExplicitServerOptIn(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		r := NewWithGuidePreview(nil, enabled)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/guides?preview=1", nil))
		want := http.StatusNotFound
		if enabled {
			want = http.StatusServiceUnavailable
		}
		if w.Code != want {
			t.Fatalf("enabled=%v: got %d, want %d", enabled, w.Code, want)
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
		{"/mini/", "text/html"},
		{"/mini/core.js", "javascript"},
		{"/mini/assets/book-open.png", "image/png"},
	} {
		response := httptest.NewRecorder()
		r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Header().Get("Content-Type"), test.content) {
			t.Errorf("GET %s: status %d, content type %q", test.path, response.Code, response.Header().Get("Content-Type"))
		}
	}
}
