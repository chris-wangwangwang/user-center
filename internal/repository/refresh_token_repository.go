package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/chris/usercenter/internal/exception"
	"github.com/chris/usercenter/internal/model"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(db *gorm.DB, t *model.RefreshToken) *exception.Exception
	FindByHash(db *gorm.DB, hash string) (*model.RefreshToken, *exception.Exception)
	Revoke(db *gorm.DB, id uint) *exception.Exception
	RevokeByUser(db *gorm.DB, userID string) *exception.Exception
	PurgeExpired(db *gorm.DB) *exception.Exception
}

type refreshTokenRepoImpl struct{}

var (
	rtInstance RefreshTokenRepository
	rtOnce     sync.Once
)

func GetRefreshTokenRepository() RefreshTokenRepository {
	rtOnce.Do(func() {
		rtInstance = &refreshTokenRepoImpl{}
	})
	return rtInstance
}

func (r *refreshTokenRepoImpl) Create(db *gorm.DB, t *model.RefreshToken) *exception.Exception {
	if err := db.Create(t).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *refreshTokenRepoImpl) FindByHash(db *gorm.DB, hash string) (*model.RefreshToken, *exception.Exception) {
	var t model.RefreshToken
	if err := db.First(&t, "token_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.NewNotFound("refresh token not found")
		}
		return nil, exception.NewDatabase(err)
	}
	return &t, nil
}

func (r *refreshTokenRepoImpl) Revoke(db *gorm.DB, id uint) *exception.Exception {
	if err := db.Model(&model.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *refreshTokenRepoImpl) RevokeByUser(db *gorm.DB, userID string) *exception.Exception {
	if err := db.Model(&model.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *refreshTokenRepoImpl) PurgeExpired(db *gorm.DB) *exception.Exception {
	if err := db.Where("expires_at < ?", time.Now()).Delete(&model.RefreshToken{}).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}
