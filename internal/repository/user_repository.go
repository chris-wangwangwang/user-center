package repository

import (
	"errors"
	"sync"

	"user-center/internal/exception"
	"user-center/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(db *gorm.DB, u *model.User) *exception.Exception
	Update(db *gorm.DB, u *model.User) *exception.Exception
	Delete(db *gorm.DB, id string) *exception.Exception
	FindByID(db *gorm.DB, id string) (*model.User, *exception.Exception)
	FindByUsername(db *gorm.DB, username string) (*model.User, *exception.Exception)
	FindByEmail(db *gorm.DB, email string) (*model.User, *exception.Exception)
	FindByPhone(db *gorm.DB, phone string) (*model.User, *exception.Exception)
	List(db *gorm.DB, page, size int) ([]model.User, int64, *exception.Exception)
	ChangePassword(db *gorm.DB, id, newHash string) *exception.Exception
}

type userRepoImpl struct{}

var (
	userRepoInstance UserRepository
	userRepoOnce     sync.Once
)

func GetUserRepository() UserRepository {
	userRepoOnce.Do(func() {
		userRepoInstance = &userRepoImpl{}
	})
	return userRepoInstance
}

func (r *userRepoImpl) Create(db *gorm.DB, u *model.User) *exception.Exception {
	if err := db.Create(u).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *userRepoImpl) Update(db *gorm.DB, u *model.User) *exception.Exception {
	if err := db.Save(u).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *userRepoImpl) Delete(db *gorm.DB, id string) *exception.Exception {
	if err := db.Delete(&model.User{}, "id = ?", id).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}

func (r *userRepoImpl) FindByID(db *gorm.DB, id string) (*model.User, *exception.Exception) {
	var u model.User
	if err := db.First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.NewNotFound("user not found")
		}
		return nil, exception.NewDatabase(err)
	}
	return &u, nil
}

func (r *userRepoImpl) FindByUsername(db *gorm.DB, username string) (*model.User, *exception.Exception) {
	var u model.User
	if err := db.First(&u, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.NewNotFound("user not found")
		}
		return nil, exception.NewDatabase(err)
	}
	return &u, nil
}

func (r *userRepoImpl) FindByEmail(db *gorm.DB, email string) (*model.User, *exception.Exception) {
	if email == "" {
		return nil, exception.NewNotFound("user not found")
	}
	var u model.User
	if err := db.First(&u, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.NewNotFound("user not found")
		}
		return nil, exception.NewDatabase(err)
	}
	return &u, nil
}

func (r *userRepoImpl) FindByPhone(db *gorm.DB, phone string) (*model.User, *exception.Exception) {
	if phone == "" {
		return nil, exception.NewNotFound("user not found")
	}
	var u model.User
	if err := db.First(&u, "phone = ?", phone).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.NewNotFound("user not found")
		}
		return nil, exception.NewDatabase(err)
	}
	return &u, nil
}

func (r *userRepoImpl) List(db *gorm.DB, page, size int) ([]model.User, int64, *exception.Exception) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 200 {
		size = 20
	}
	var (
		list []model.User
		total int64
	)
	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, exception.NewDatabase(err)
	}
	if err := db.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, exception.NewDatabase(err)
	}
	return list, total, nil
}

func (r *userRepoImpl) ChangePassword(db *gorm.DB, id, newHash string) *exception.Exception {
	if err := db.Model(&model.User{}).Where("id = ?", id).Update("password_hash", newHash).Error; err != nil {
		return exception.NewDatabase(err)
	}
	return nil
}
