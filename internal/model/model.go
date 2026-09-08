package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Username     string         `gorm:"uniqueIndex;type:varchar(64);not null" json:"username"`
	Email        string         `gorm:"uniqueIndex;type:varchar(128)" json:"email,omitempty"`
	Phone        string         `gorm:"index;type:varchar(32)" json:"phone,omitempty"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string         `gorm:"type:varchar(64)" json:"nickname,omitempty"`
	Avatar       string         `gorm:"type:varchar(255)" json:"avatar,omitempty"`
	Disabled     bool           `gorm:"default:false" json:"disabled"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "uc_users" }

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index;type:varchar(36);not null" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;type:varchar(255);not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"default:false" json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

func (RefreshToken) TableName() string { return "uc_refresh_tokens" }

type OAuthClient struct {
	ClientID     string    `gorm:"primaryKey;type:varchar(64)" json:"client_id"`
	ClientSecret string    `gorm:"type:varchar(255);not null" json:"-"`
	Name         string    `gorm:"type:varchar(128)" json:"name"`
	RedirectURIs string    `gorm:"type:text" json:"redirect_uris"`
	Scopes       string    `gorm:"type:varchar(255)" json:"scopes"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (OAuthClient) TableName() string { return "uc_oauth_clients" }

type AuthorizationCode struct {
	Code        string    `gorm:"primaryKey;type:varchar(128)" json:"code"`
	ClientID    string    `gorm:"index;type:varchar(64);not null" json:"client_id"`
	UserID      string    `gorm:"index;type:varchar(36);not null" json:"user_id"`
	RedirectURI string    `gorm:"type:varchar(512)" json:"redirect_uri"`
	Scope       string    `gorm:"type:varchar(255)" json:"scope"`
	ExpiresAt   time.Time `gorm:"not null" json:"expires_at"`
	Used        bool      `gorm:"default:false" json:"used"`
	CreatedAt   time.Time `json:"created_at"`
}

func (AuthorizationCode) TableName() string { return "uc_authorization_codes" }
