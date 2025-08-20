package auth

import (
	"context"
	"errors"

	models "github.com/hoyci/bookday/internal/infra/database/model"
	"github.com/hoyci/bookday/pkg/fault"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	var userModel models.UserModel
	result := r.db.WithContext(ctx).
		Preload("Roles").
		First(&userModel, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fault.New("user not found", fault.WithKind(fault.KindNotFound))
		}
		return nil, fault.New("failed to find user by email", fault.WithError(result.Error), fault.WithHTTPCode(500))
	}

	return toUserEntity(&userModel), nil
}

func (r *gormRepository) FindUserByID(ctx context.Context, id string) (*User, error) {
	var userModel models.UserModel
	result := r.db.WithContext(ctx).
		Preload("Roles").
		First(&userModel, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fault.New("user not found", fault.WithKind(fault.KindNotFound))
		}
		return nil, fault.New("failed to find user by id", fault.WithError(result.Error), fault.WithHTTPCode(500))
	}

	return toUserEntity(&userModel), nil
}

func (r *gormRepository) CreateUser(ctx context.Context, user *User, role string) error {
	userModel := models.UserModel{
		ID:           user.ID(),
		Name:         user.Name(),
		Email:        user.Email(),
		PasswordHash: user.PasswordHash(),
		CreatedAt:    user.CreatedAt(),
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var roleToAssign models.RoleModel
		if err := tx.First(&roleToAssign, "name = ?", role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fault.New("The specified role does not exist", fault.WithKind(fault.KindUnexpected), fault.WithError(err))
			}
			return fault.New("failed to query role", fault.WithError(err))
		}

		userModel.Roles = []models.RoleModel{roleToAssign}

		if err := tx.Create(&userModel).Error; err != nil {
			return fault.New("failed to create user in database", fault.WithError(err), fault.WithHTTPCode(500))
		}

		return nil
	})
}

func toUserEntity(model *models.UserModel) *User {
	roleNames := make([]string, len(model.Roles))
	for i, role := range model.Roles {
		roleNames[i] = string(role.Name)
	}

	user, _ := NewUser(
		model.ID,
		model.Name,
		model.Email,
		model.PasswordHash,
		roleNames,
		model.CreatedAt,
	)
	return user
}
