package service

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"user-center/internal/database"
	"user-center/internal/exception"
	"user-center/internal/model"
	"user-center/internal/repository"
	"user-center/internal/util/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OIDCService interface {
	Authorize(userID string, req *AuthorizeRequest) (code, redirect string, state string, ex *exception.Exception)
	TokenExchange(req *TokenExchangeRequest) (*TokenInfo, *exception.Exception)
	UserInfo(accessToken string) (*UserInfo, *exception.Exception)
	Discovery(issuer string) map[string]interface{}
	JWKS() map[string]interface{}
}

type oidcServiceImpl struct {
	db          *gorm.DB
	jwtMgr      jwt.Manager
	clientRepo  repository.OAuthClientRepository
	userRepo    repository.UserRepository
	codeRepo    repository.AuthorizationCodeRepository
	issuer      string
}

var (
	oidcSvcInstance OIDCService
	oidcSvcOnce     sync.Once
)

func GetOIDCService() OIDCService {
	oidcSvcOnce.Do(func() {
		oidcSvcInstance = &oidcServiceImpl{
			db:         database.GetDB(),
			jwtMgr:     jwt.GetManager(),
			clientRepo: repository.GetOAuthClientRepository(),
			userRepo:   repository.GetUserRepository(),
			codeRepo:   repository.GetAuthorizationCodeRepository(),
		}
	})
	return oidcSvcInstance
}

func (s *oidcServiceImpl) Authorize(userID string, req *AuthorizeRequest) (string, string, string, *exception.Exception) {
	client, ex := s.clientRepo.FindByClientID(s.db, req.ClientID)
	if ex != nil {
		return "", "", req.State, exception.NewBadRequest("invalid client_id")
	}
	if !containsRedirect(client.RedirectURIs, req.RedirectURI) {
		return "", "", req.State, exception.NewBadRequest("redirect_uri not registered")
	}
	if req.ResponseType != "code" {
		return "", req.RedirectURI, req.State, exception.NewBadRequest("unsupported response_type")
	}
	code := randomString(48)
	ac := &model.AuthorizationCode{
		Code:        code,
		ClientID:    req.ClientID,
		UserID:      userID,
		RedirectURI: req.RedirectURI,
		Scope:       req.Scope,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}
	if ex := s.codeRepo.Create(s.db, ac); ex != nil {
		return "", req.RedirectURI, req.State, ex
	}
	return code, req.RedirectURI, req.State, nil
}

func (s *oidcServiceImpl) TokenExchange(req *TokenExchangeRequest) (*TokenInfo, *exception.Exception) {
	client, ex := s.clientRepo.FindByClientID(s.db, req.ClientID)
	if ex != nil {
		return nil, exception.NewBadRequest("invalid client")
	}
	if client.ClientSecret != req.ClientSecret {
		return nil, exception.NewUnauthorized("invalid client credentials")
	}

	switch req.GrantType {
	case "authorization_code":
		ac, ex := s.codeRepo.FindAndConsume(s.db, req.Code)
		if ex != nil {
			return nil, ex
		}
		if ac.ClientID != req.ClientID || ac.RedirectURI != req.RedirectURI {
			return nil, exception.NewBadRequest("code/client/redirect mismatch")
		}
		u, ex := s.userRepo.FindByID(s.db, ac.UserID)
		if ex != nil {
			return nil, exception.NewUnauthorized("user no longer exists")
		}
		return s.issueForClient(u, req.ClientID)
	case "refresh_token":
		claims, ex := s.jwtMgr.Parse(req.RefreshToken)
		if ex != nil {
			return nil, ex
		}
		if claims.Type != jwt.TypeRefresh {
			return nil, exception.NewInvalidToken("not a refresh token")
		}
		hash := hashToken(req.RefreshToken)
		rt, ex := s.tokenRepo().FindByHash(s.db, hash)
		if ex != nil {
			return nil, exception.NewInvalidToken("refresh token not recognized")
		}
		if rt.Revoked || time.Now().After(rt.ExpiresAt) {
			return nil, exception.NewInvalidToken("refresh token invalid")
		}
		_ = s.tokenRepo().Revoke(s.db, rt.ID)
		u, ex := s.userRepo.FindByID(s.db, claims.UserID)
		if ex != nil {
			return nil, exception.NewUnauthorized("user no longer exists")
		}
		return s.issueForClient(u, req.ClientID)
	default:
		return nil, exception.NewBadRequest("unsupported grant_type")
	}
}

func (s *oidcServiceImpl) UserInfo(accessToken string) (*UserInfo, *exception.Exception) {
	claims, ex := s.jwtMgr.Parse(accessToken)
	if ex != nil {
		return nil, ex
	}
	if claims.Type != jwt.TypeAccess {
		return nil, exception.NewInvalidToken("not an access token")
	}
	u, ex := s.userRepo.FindByID(s.db, claims.UserID)
	if ex != nil {
		return nil, ex
	}
	return toUserInfo(u), nil
}

func (s *oidcServiceImpl) Discovery(issuer string) map[string]interface{} {
	return map[string]interface{}{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oidc/authorize",
		"token_endpoint":                        issuer + "/oidc/token",
		"userinfo_endpoint":                     issuer + "/oidc/userinfo",
		"jwks_uri":                              issuer + "/oidc/jwks",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"HS256"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"scopes_supported":                      []string{"openid", "profile", "email"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"code_challenge_methods_supported":      []string{"S256", "plain"},
	}
}

func (s *oidcServiceImpl) JWKS() map[string]interface{} {
	return map[string]interface{}{"keys": []interface{}{}}
}

func (s *oidcServiceImpl) issueForClient(u *model.User, clientID string) (*TokenInfo, *exception.Exception) {
	access, accessExp, err := s.jwtMgr.IssueAccess(u.ID, u.Username)
	if err != nil {
		return nil, exception.NewInternal(err)
	}
	refresh, refreshExp, err := s.jwtMgr.IssueRefresh(u.ID, u.Username)
	if err != nil {
		return nil, exception.NewInternal(err)
	}
	rt := &model.RefreshToken{
		UserID:    u.ID,
		TokenHash: hashToken(refresh),
		ExpiresAt: refreshExp,
	}
	if ex := s.tokenRepo().Create(s.db, rt); ex != nil {
		return nil, ex
	}
	return &TokenInfo{
		AccessToken:      access,
		RefreshToken:     refresh,
		TokenType:        "Bearer",
		ExpiresIn:        int64(time.Until(accessExp).Seconds()),
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (s *oidcServiceImpl) tokenRepo() repository.RefreshTokenRepository {
	return repository.GetRefreshTokenRepository()
}

func randomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n]
}

func containsRedirect(list, target string) bool {
	for _, uri := range strings.Split(list, ",") {
		if strings.TrimSpace(uri) == target {
			return true
		}
	}
	return false
}
