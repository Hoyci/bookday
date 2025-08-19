package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	repo   Repository
	log    *log.Logger
	jwtSvc jwt.Service
}

// NewService cria uma nova instância do serviço de autenticação.
func NewService(repo Repository, logger *log.Logger, jwtSvc jwt.Service) Service {
	return &service{
		repo:   repo,
		log:    logger,
		jwtSvc: jwtSvc,
	}
}

// Register lida com a lógica de registro de um novo usuário.
func (s *service) Register(ctx context.Context, dto RegisterUserDTO) (*UserResponseDTO, error) {
	s.log.Info("starting user registration process", "email", dto.Email)

	if err := dto.Validate(); err != nil {
		return nil, fault.New("invalid registration data", fault.WithKind(fault.KindValidation), fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err))
	}

	// 1. Verifica se o e-mail já existe. Um erro "NotFound" aqui é o caminho feliz.
	_, err := s.repo.FindUserByEmail(ctx, dto.Email)
	if err == nil {
		// Se err é nil, o usuário foi encontrado, o que é um conflito.
		s.log.Warn("registration attempt with existing email", "email", dto.Email)
		return nil, fault.New("email already registered", fault.WithKind(fault.KindConflict), fault.WithHTTPCode(http.StatusConflict))
	}
	var f *fault.Error
	if errors.As(err, &f) && f.Kind != fault.KindNotFound {
		// Se o erro não for 'NotFound', é um erro inesperado do banco.
		s.log.Error("failed to check for existing user", "error", err)
		return nil, err
	}

	// 2. Gera o hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", "error", err)
		return nil, fault.New("could not process registration", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	// 3. Cria a entidade de domínio User
	// Papéis (roles) e CreatedAt serão definidos pelo repositório/banco
	newUser, err := NewUser(uuid.NewString(), dto.Name, dto.Email, string(hashedPassword), nil, time.Now())
	if err != nil {
		s.log.Error("failed to create new user entity", "error", err)
		return nil, fault.New("invalid user entity data", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	// 4. Salva o usuário no banco
	if err := s.repo.CreateUser(ctx, newUser); err != nil {
		s.log.Error("failed to save user to repository", "error", err)
		return nil, fault.New("could not save user", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	s.log.Info("user registered successfully", "user_id", newUser.ID(), "email", dto.Email)
	return toUserResponseDTO(newUser), nil
}

// Login lida com a lógica de autenticação e retorna um par de tokens (access e refresh).
func (s *service) Login(ctx context.Context, dto LoginDTO) (*LoginResponseDTO, error) {
	s.log.Info("user login attempt", "email", dto.Email)

	if err := dto.Validate(); err != nil {
		return nil, fault.New("invalid login data", fault.WithKind(fault.KindValidation), fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err))
	}

	// 1. Busca o usuário pelo e-mail
	user, err := s.repo.FindUserByEmail(ctx, dto.Email)
	if err != nil {
		s.log.Warn("login failed: user not found", "email", dto.Email, "error", err)
		// Retorna um erro genérico para não informar ao atacante se o email existe ou não.
		return nil, fault.New("invalid credentials", fault.WithKind(fault.KindUnauthenticated), fault.WithHTTPCode(http.StatusUnauthorized))
	}

	// 2. Compara a senha fornecida com o hash armazenado
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(dto.Password))
	if err != nil {
		s.log.Warn("login failed: invalid password", "email", dto.Email)
		// Erro genérico novamente por segurança.
		return nil, fault.New("invalid credentials", fault.WithKind(fault.KindUnauthenticated), fault.WithHTTPCode(http.StatusUnauthorized))
	}

	// 3. Gera o Access Token e o Refresh Token
	accessToken, err := s.jwtSvc.GenerateAccessToken(user.ID(), user.Roles())
	if err != nil {
		s.log.Error("failed to generate access token for user", "user_id", user.ID(), "error", err)
		return nil, fault.New("could not process login", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	refreshToken, err := s.jwtSvc.GenerateRefreshToken(user.ID())
	if err != nil {
		s.log.Error("failed to generate refresh token for user", "user_id", user.ID(), "error", err)
		return nil, fault.New("could not process login", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	s.log.Info("user logged in successfully", "user_id", user.ID())

	response := &LoginResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return response, nil
}

// toUserResponseDTO converte a entidade User para o DTO de resposta.
func toUserResponseDTO(user *User) *UserResponseDTO {
	return &UserResponseDTO{
		ID:        user.ID(),
		Name:      user.Name(),
		Email:     user.Email(),
		Roles:     user.Roles(), // O repositório já atribui o papel 'CUSTOMER'
		CreatedAt: user.CreatedAt(),
	}
}
