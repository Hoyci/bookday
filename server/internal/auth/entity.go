package auth

import (
	"time"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// User representa a entidade de domínio para um usuário no sistema.
// Os campos são privados para garantir a encapsulação.
type User struct {
	id           string
	name         string
	email        string
	passwordHash string
	roles        []string
	createdAt    time.Time
}

// NewUser é o construtor para a entidade User.
// Ele é usado tanto para criar novos usuários quanto para reconstruir a partir do banco de dados.
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

// validate executa as regras de validação da entidade.
func (u *User) validate() error {
	return v.ValidateStruct(u,
		v.Field(&u.name, v.Required.Error("name is required"), v.Length(3, 100)),
		v.Field(&u.email, v.Required.Error("email is required"), is.Email),
		v.Field(&u.passwordHash, v.Required.Error("password hash is required")),
	)
}

// Getters para acessar os campos privados da entidade
func (u *User) ID() string           { return u.id }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) PasswordHash() string { return u.passwordHash }
func (u *User) Roles() []string      { return u.roles }
func (u *User) CreatedAt() time.Time { return u.createdAt }

// HasRole é um método auxiliar para verificar se um usuário possui um determinado papel.
// Será muito útil no middleware de autorização.
func (u *User) HasRole(roleName string) bool {
	for _, r := range u.roles {
		if r == roleName {
			return true
		}
	}
	return false
}
