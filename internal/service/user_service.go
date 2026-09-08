package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/chris/usercenter/internal/database"
	"github.com/chris/usercenter/internal/exception"
	"github.com/chris/usercenter/internal/model"
	"github.com/chris/usercenter/internal/repository"
	"github.com/chris/usercenter/internal/util/jwt"
	"github.com/chris/usercenter/internal/util/password"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService interface {
	Register(req *RegisterRequest) (*UserInfo, *exception.Exception)
	Login(req *LoginRequest) (*TokenInfo, *exception.Exception)
	Refresh(refreshToken string) (*TokenInfo, *exception.Exception)
	Logout(userID, refreshToken string) *exception.Exception
	GetProfile(userID string) (*UserInfo, *exception.Exception)
	UpdateProfile(userID string, req *UpdateProfileRequest) (*UserInfo, *exception.Exception)
	ChangePassword(userID string, req *ChangePasswordRequest) *exception.Exception
	Disable(userID string, disabled bool) *exception.Exception
	List(page, size int) ([]UserInfo, int64, *exception.Exception)
}

type userServiceImpl struct {
	db        *gorm.DB
	jwtMgr    jwt.Manager
	userRepo  repository.UserRepository
	tokenRepo repository.RefreshTokenRepository
}

var (
	userSvcInstance UserService
	userSvcOnce     sync.Once
)

func GetUserService() UserService {
	userSvcOnce.Do(func() {
		userSvcInstance = &userServiceImpl{
			db:        database.GetDB(),
			jwtMgr:    jwt.GetManager(),
			userRepo:  repository.GetUserRepository(),
			tokenRepo: repository.GetRefreshTokenRepository(),
		}
	})
	return userSvcInstance
}

func (s *userServiceImpl) Register(req *RegisterRequest) (*UserInfo, *exception.Exception) {
	if _, ex := s.userRepo.FindByUsername(s.db, req.Username); ex == nil {
		return nil, exception.NewConflict("username already exists")
	}
	if req.Email != "" {
		if _, ex := s.userRepo.FindByEmail(s.db, req.Email); ex == nil {
			return nil, exception.NewConflict("email already exists")
		}
	}
	if req.Phone != "" {
		if _, ex := s.userRepo.FindByPhone(s.db, req.Phone); ex == nil {
			return nil, exception.NewConflict("phone already exists")
		}
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, exception.NewInternal(err)
	}

	u := &model.User{
		ID:           uuid.NewString(),
		Username:     strings.TrimSpace(req.Username),
		Email:        strings.TrimSpace(req.Email),
		Phone:        strings.TrimSpace(req.Phone),
		Nickname:     strings.TrimSpace(req.Nickname),
		PasswordHash: hash,
	}
	if u.Nickname == "" {
		u.Nickname = u.Username
	}
	if ex := s.userRepo.Create(s.db, u); ex != nil {
		return nil, ex
	}
	return toUserInfo(u), nil
}

func (s *userServiceImpl) Login(req *LoginRequest) (*TokenInfo, *exception.Exception) {
	u, ex := s.userRepo.FindByUsername(s.db, req.Username)
	if ex != nil {
		return nil, exception.NewUnauthorized("invalid username or password")
	}
	if u.Disabled {
		return nil, exception.NewForbidden("user is disabled")
	}
	if !password.Verify(u.PasswordHash, req.Password) {
		return nil, exception.NewUnauthorized("invalid username or password")
	}
	return s.issueTokens(u)
}

func (s *userServiceImpl) Refresh(refreshToken string) (*TokenInfo, *exception.Exception) {
	claims, ex := s.jwtMgr.Parse(refreshToken)
	if ex != nil {
		return nil, ex
	}
	if claims.Type != jwt.TypeRefresh {
		return nil, exception.NewInvalidToken("not a refresh token")
	}
	hash := hashToken(refreshToken)
	rt, ex := s.tokenRepo.FindByHash(s.db, hash)
	if ex != nil {
		return nil, exception.NewInvalidToken("refresh token not recognized")
	}
	if rt.Revoked {
		return nil, exception.NewInvalidToken("refresh token revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, exception.NewExpiredToken("refresh token expired")
	}
	u, ex := s.userRepo.FindByID(s.db, claims.UserID)
	if ex != nil {
		return nil, exception.NewUnauthorized("user no longer exists")
	}
	_ = s.tokenRepo.Revoke(s.db, rt.ID)
	return s.issueTokens(u)
}

func (s *userServiceImpl) Logout(userID, refreshToken string) *exception.Exception {
	if refreshToken != "" {
		hash := hashToken(refreshToken)
		if rt, ex := s.tokenRepo.FindByHash(s.db, hash); ex == nil && rt.UserID == userID {
			return s.tokenRepo.Revoke(s.db, rt.ID)
		}
	}
	return s.tokenRepo.RevokeByUser(s.db, userID)
}

func (s *userServiceImpl) GetProfile(userID string) (*UserInfo, *exception.Exception) {
	u, ex := s.userRepo.FindByID(s.db, userID)
	if ex != nil {
		return nil, ex
	}
	return toUserInfo(u), nil
}

func (s *userServiceImpl) UpdateProfile(userID string, req *UpdateProfileRequest) (*UserInfo, *exception.Exception) {
	u, ex := s.userRepo.FindByID(s.db, userID)
	if ex != nil {
		return nil, ex
	}
	if req.Email != nil {
		u.Email = strings.TrimSpace(*req.Email)
	}
	if req.Phone != nil {
		u.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Nickname != nil {
		u.Nickname = strings.TrimSpace(*req.Nickname)
	}
	if req.Avatar != nil {
		u.Avatar = strings.TrimSpace(*req.Avatar)
	}
	if ex := s.userRepo.Update(s.db, u); ex != nil {
		return nil, ex
	}
	return toUserInfo(u), nil
}

func (s *userServiceImpl) ChangePassword(userID string, req *ChangePasswordRequest) *exception.Exception {
	u, ex := s.userRepo.FindByID(s.db, userID)
	if ex != nil {
		return ex
	}
	if !password.Verify(u.PasswordHash, req.OldPassword) {
		return exception.NewInvalidPassword("old password is incorrect")
	}
	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		return exception.NewInternal(err)
	}
	if ex := s.userRepo.ChangePassword(s.db, u.ID, hash); ex != nil {
		return ex
	}
	_ = s.tokenRepo.RevokeByUser(s.db, u.ID)
	return nil
}

func (s *userServiceImpl) Disable(userID string, disabled bool) *exception.Exception {
	u, ex := s.userRepo.FindByID(s.db, userID)
	if ex != nil {
		return ex
	}
	u.Disabled = disabled
	if ex := s.userRepo.Update(s.db, u); ex != nil {
		return ex
	}
	if disabled {
		_ = s.tokenRepo.RevokeByUser(s.db, u.ID)
	}
	return nil
}

func (s *userServiceImpl) List(page, size int) ([]UserInfo, int64, *exception.Exception) {
	users, total, ex := s.userRepo.List(s.db, page, size)
	if ex != nil {
		return nil, 0, ex
	}
	out := make([]UserInfo, 0, len(users))
	for i := range users {
		out = append(out, *toUserInfo(&users[i]))
	}
	return out, total, nil
}

func (s *userServiceImpl) issueTokens(u *model.User) (*TokenInfo, *exception.Exception) {
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
	if ex := s.tokenRepo.Create(s.db, rt); ex != nil {
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

func toUserInfo(u *model.User) *UserInfo {
	return &UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		Phone:    u.Phone,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
		Disabled: u.Disabled,
	}
}

func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}
