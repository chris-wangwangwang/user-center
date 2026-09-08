package repository

import (
	"errors"
	"sync"

	"github.com/chris/usercenter/internal/exception"
	"github.com/chris/usercenter/internal/model"
	"gorm.io/gorm"
)

type OAuthClientRepository interface {
	Create(db *gorm.DB, c *model.OAuthClient) *exception.Exception
	FindByClientID(db *gorm.DB, clientID string) (*model.OAuthClient, *exception.Exception)
}

type oauthClientRepoImpl struct{}

var (
	ocInstance OAuthClientRepository
	ocOnce     sync.Once
)

func GetOAuthClientRepository() OAuthClientRepository {
	ocOnce.Do(func() {
		ocInstance = &oauthClientRepoImpl{}
	})
	return ocInstance
}

func (r *oauthClientRepoImpl) Create(db *gorm.DB, c *model.OAuthClient) *exception.Exception {
	if err := db.Create(c).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *oauthClientRepoImpl) FindByClientID(db *gorm.DB, clientID string) (*model.OAuthClient, *exception.Exception) {
	var c model.OAuthClient
	if err := db.First(&c, "client_id = ?", clientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.NewNotFound("oauth client not found")
		}
		return nil, exception.NewDatabase(err)
	}
	return &c, nil
}
