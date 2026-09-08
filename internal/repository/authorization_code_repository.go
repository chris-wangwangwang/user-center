package repository

import (
	"errors"
	"sync"
	"time"

	"github.com/chris/usercenter/internal/exception"
	"github.com/chris/usercenter/internal/model"
	"gorm.io/gorm"
)

type AuthorizationCodeRepository interface {
	Create(db *gorm.DB, c *model.AuthorizationCode) *exception.Exception
	FindAndConsume(db *gorm.DB, code string) (*model.AuthorizationCode, *exception.Exception)
	PurgeExpired(db *gorm.DB) *exception.Exception
}

type authCodeRepoImpl struct{}

var (
	acInstance AuthorizationCodeRepository
	acOnce     sync.Once
)

func GetAuthorizationCodeRepository() AuthorizationCodeRepository {
	acOnce.Do(func() {
		acInstance = &authCodeRepoImpl{}
	})
	return acInstance
}

func (r *authCodeRepoImpl) Create(db *gorm.DB, c *model.AuthorizationCode) *exception.Exception {
	if err := db.Create(c).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *authCodeRepoImpl) FindAndConsume(db *gorm.DB, code string) (*model.AuthorizationCode, *exception.Exception) {
	var c model.AuthorizationCode
	tx := db.Begin()
	if tx.Error != nil {
		return nil, exception.NewDatabase(tx.Error)
	}
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&c, "code = ?", code).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.NewNotFound("authorization code not found")
		}
		return nil, exception.NewDatabase(err)
	}
	if c.Used || time.Now().After(c.ExpiresAt) {
		tx.Rollback()
		return nil, exception.NewBadRequest("authorization code invalid or expired")
	}
	if err := tx.Model(&c).Update("used", true).Error; err != nil {
		tx.Rollback()
		return nil, exception.NewDatabase(err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, exception.NewDatabase(err)
	}
	return &c, nil
}

func (r *authCodeRepoImpl) PurgeExpired(db *gorm.DB) *exception.Exception {
	if err := db.Where("expires_at < ?", time.Now()).Delete(&model.AuthorizationCode{}).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}
