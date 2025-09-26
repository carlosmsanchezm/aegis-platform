package jwtutil

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Sub     string `json:"sub"`
	Wid     string `json:"wid"`
	Dest    string `json:"dest"`
	DNS     string `json:"dns,omitempty"`
	Cluster string `json:"cluster,omitempty"`
	jwt.RegisteredClaims
}

type Verifier struct {
	secret   []byte
	audience string
	clock    func() time.Time
}

func NewVerifier(secret []byte) *Verifier {
	return &Verifier{
		secret: secret,
		clock:  time.Now,
	}
}

func (v *Verifier) WithAudience(aud string) *Verifier {
	v.audience = aud
	return v
}

func (v *Verifier) Verify(token string) (*Claims, error) {
	if len(v.secret) == 0 {
		return nil, errors.New("jwt secret not configured")
	}
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %s", t.Method.Alg())
		}
		return v.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Before(v.clock()) {
		return nil, errors.New("token expired")
	}
	if len(claims.Audience) == 0 {
		return nil, errors.New("audience missing")
	}
	if err := claimStringNotEmpty(claims.Wid, "wid"); err != nil {
		return nil, err
	}
	if err := claimStringNotEmpty(claims.Dest, "dest"); err != nil {
		return nil, err
	}
	if claims.DNS != "" {
		if strings.Contains(claims.DNS, ":") {
			return nil, errors.New("dns must not include port")
		}
	}
	if claims.ID == "" {
		return nil, errors.New("jti missing")
	}
	if v.audience != "" {
		if !containsAudience(claims.Audience, v.audience) {
			return nil, errors.New("audience mismatch")
		}
	}
	return claims, nil
}

func claimStringNotEmpty(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s missing", field)
	}
	return nil
}

func containsAudience(auds jwt.ClaimStrings, expected string) bool {
	for _, aud := range auds {
		if aud == expected {
			return true
		}
	}
	return false
}
