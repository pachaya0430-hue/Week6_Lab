package mw

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"moviehub/auth"
)

type contextKey int

const (
	claimsKey contextKey = iota
)

func claimsFrom(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*auth.Claims)
	return claims, ok
}

func WhoFrom(ctx context.Context) (*auth.Claims, bool) {
	return claimsFrom(ctx)
}

func MemberID(ctx context.Context) (int64, bool) {
	claims, ok := claimsFrom(ctx)
	if !ok {
		return 0, false
	}
	return claims.MemberID, true
}

func Role(ctx context.Context) string {
	claims, ok := claimsFrom(ctx)
	if !ok {
		return ""
	}
	return claims.Role
}

func fail(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{"code": code, "message": message},
	})
}

func RequireLogin(issuer *auth.Issuer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			fail(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
			return
		}

		claims, err := auth.ParseToken(parts[1], string(issuer.Secret))
		if err != nil {
			fail(w, http.StatusUnauthorized, "BAD_TOKEN", "ตั๋วไม่ถูกต้องหรือหมดอายุ กรุณาเข้าสู่ระบบใหม่")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}
