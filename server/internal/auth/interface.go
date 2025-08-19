package auth

import (
	"context"
)

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
}

type Service interface {
	Register(ctx context.Context, dto RegisterUserDTO) (*UserResponseDTO, error)
	Login(ctx context.Context, dto LoginDTO) (*LoginResponseDTO, error)
}
