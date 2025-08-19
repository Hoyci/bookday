package auth

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type RegisterUserDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (dto RegisterUserDTO) Validate() error {
	return v.ValidateStruct(&dto,
		v.Field(&dto.Name, v.Required.Error("name is required"), v.Length(3, 100)),
		v.Field(&dto.Email, v.Required.Error("email is required"), is.Email),
		v.Field(&dto.Password, v.Required.Error("password is required"), v.Length(8, 100)),
	)
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (dto LoginDTO) Validate() error {
	return v.ValidateStruct(&dto,
		v.Field(&dto.Email, v.Required.Error("email is required"), is.Email),
		v.Field(&dto.Password, v.Required.Error("password is required")),
	)
}

type LoginResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponseDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
}
