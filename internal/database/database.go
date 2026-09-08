package database

import (
	"log"

	"github.com/chris/usercenter/internal/config"
	"github.com/chris/usercenter/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

func Init() *gorm.DB {
	if db != nil {
		return db
	}
	cfg := config.Get()
	conn, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err := conn.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.OAuthClient{},
		&model.AuthorizationCode{},
	); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}
	db = conn
	return db
}

func GetDB() *gorm.DB { return db }
