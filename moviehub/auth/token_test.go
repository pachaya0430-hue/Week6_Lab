package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret-key"
	memberID := int64(123)

	tokenString, err := GenerateToken(
		memberID,
		"member",
		secret,
		time.Hour,
	)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims.MemberID != memberID {
		t.Errorf("MemberID = %d, want %d", claims.MemberID, memberID)
	}

	if claims.Role != "member" {
		t.Errorf("Role = %q, want %q", claims.Role, "member")
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	tokenString, err := GenerateToken(
		123,
		"member",
		"correct-secret",
		time.Hour,
	)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	_, err = ParseToken(tokenString, "wrong-secret")
	if err == nil {
		t.Fatal("ParseToken() should reject a token signed with the wrong secret")
	}
}

func TestParseTokenRejectsInvalidToken(t *testing.T) {
	_, err := ParseToken("this-is-not-a-valid-token", "test-secret")
	if err == nil {
		t.Fatal("ParseToken() should reject an invalid token")
	}
}
