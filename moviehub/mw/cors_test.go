package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	handlerCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})
	handler := CORS([]string{" http://localhost:3000 ", ""}, next)

	t.Run("allowed preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/movies", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Errorf("Access-Control-Allow-Origin = %q", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
			t.Errorf("Access-Control-Allow-Methods = %q", got)
		}
		if got := rec.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
			t.Errorf("Access-Control-Allow-Headers = %q", got)
		}
		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Errorf("Vary = %q, want Origin", got)
		}
		if handlerCalled {
			t.Error("preflight request reached the next handler")
		}
	})

	t.Run("disallowed preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/movies", nil)
		req.Header.Set("Origin", "http://localhost:3001")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		for _, name := range []string{
			"Access-Control-Allow-Origin",
			"Access-Control-Allow-Methods",
			"Access-Control-Allow-Headers",
			"Access-Control-Max-Age",
		} {
			if rec.Header().Get(name) != "" {
				t.Errorf("disallowed origin received %s", name)
			}
		}
		if got := rec.Header().Get("Vary"); got != "Origin" {
			t.Errorf("Vary = %q, want Origin", got)
		}
	})

	t.Run("allowed actual request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if !handlerCalled {
			t.Error("actual request did not reach the next handler")
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Errorf("Access-Control-Allow-Origin = %q", got)
		}
	})
}
