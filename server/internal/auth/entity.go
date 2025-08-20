package auth

import (
	"slices"
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type User struct {
	id           string
	name         string
	email        string
	passwordHash string
	roles        []string
	createdAt    time.Time
}

func NewUser(id, name, email, passwordHash string, roles []string, createdAt time.Time) (*User, error) {
	u := &User{
		id:           id,
		name:         name,
		email:        email,
		passwordHash: passwordHash,
		roles:        roles,
		createdAt:    createdAt,
	}

	if err := u.validate(); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) validate() error {
	return v.ValidateStruct(u,
		v.Field(&u.name, v.Required.Error("name is required"), v.Length(3, 100)),
		v.Field(&u.email, v.Required.Error("email is required"), is.Email),
		v.Field(&u.passwordHash, v.Required.Error("password hash is required")),
	)
}

func (u *User) ID() string           { return u.id }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) PasswordHash() string { return u.passwordHash }
func (u *User) Roles() []string      { return u.roles }
func (u *User) CreatedAt() time.Time { return u.createdAt }

func (u *User) HasRole(roleName string) bool {
	return slices.Contains(u.roles, roleName)
}
