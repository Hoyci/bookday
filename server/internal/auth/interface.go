package auth

import (
	"context"

	models "github.com/hoyci/bookday/internal/infra/database/model"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, user *User, role string) error
	FindUsersByRole(ctx context.Context, role models.RolesType) ([]*User, error)
}

type Service interface {
	Register(ctx context.Context, dto RegisterUserDTO) (*UserResponseDTO, error)
	Login(ctx context.Context, dto LoginDTO) (*LoginResponseDTO, error)
	CreateUserByAdmin(ctx context.Context, dto RegisterUserDTO) (*UserResponseDTO, error)
}
