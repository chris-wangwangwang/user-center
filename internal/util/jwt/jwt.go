package jwt

import (
	"errors"
	"sync"
	"time"

	"github.com/chris/usercenter/internal/config"
	"github.com/chris/usercenter/internal/exception"
	"github.com/golang-jwt/jwt/v5"
)

const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

type Claims struct {
	UserID   string `json:"sub"`
	Username string `json:"username"`
	Type     string `json:"typ"`
	jwt.RegisteredClaims
}

type Manager interface {
	IssueAccess(userID, username string) (string, time.Time, error)
	IssueRefresh(userID, username string) (string, time.Time, error)
	Parse(token string) (*Claims, *exception.Exception)
}

type managerImpl struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

var (
	mgrInstance Manager
	mgrOnce     sync.Once
)

func GetManager() Manager {
	mgrOnce.Do(func() {
		cfg := config.Get()
		mgrInstance = &managerImpl{
			secret:     []byte(cfg.JWTSecret),
			issuer:     cfg.JWTIssuer,
			accessTTL:  cfg.AccessTokenTTL(),
			refreshTTL: cfg.RefreshTokenTTL(),
		}
	})
	return mgrInstance
}

func (m *managerImpl) IssueAccess(userID, username string) (string, time.Time, error) {
	return m.issue(userID, username, TypeAccess, m.accessTTL)
}

func (m *managerImpl) IssueRefresh(userID, username string) (string, time.Time, error) {
	return m.issue(userID, username, TypeRefresh, m.refreshTTL)
}

func (m *managerImpl) issue(userID, username, kind string, ttl time.Duration) (string, time.Time, error) {
	exp := time.Now().Add(ttl)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Type:     kind,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return s, exp, nil
}

func (m *managerImpl) Parse(token string) (*Claims, *exception.Exception) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, exception.NewExpiredToken("token expired")
		}
		return nil, exception.NewInvalidToken("invalid token")
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, exception.NewInvalidToken("invalid token")
	}
	return claims, nil
}
