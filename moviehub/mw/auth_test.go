package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"moviehub/auth"
)

func TestLoginAndRoleMiddleware(t *testing.T) {
	issuer := &auth.Issuer{
		Secret: []byte("0123456789abcdef0123456789abcdef"),
		TTL:    time.Hour,
	}
	endpoint := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	adminOnly := RequireLogin(issuer, RequireRole("admin", endpoint))

	adminToken, err := issuer.Issue(1, "admin")
	if err != nil {
		t.Fatal(err)
	}
	memberToken, err := issuer.Issue(2, "member")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "missing token", wantStatus: http.StatusUnauthorized},
		{name: "member token", token: memberToken, wantStatus: http.StatusForbidden},
		{name: "admin token", token: adminToken, wantStatus: http.StatusNoContent},
		{name: "invalid token", token: "not-a-token", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/admin/stats", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			adminOnly(rec, req)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
