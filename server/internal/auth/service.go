package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	models "github.com/hoyci/bookday/internal/infra/database/model"
	"github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	repo   Repository
	log    *log.Logger
	jwtSvc jwt.Service
}

func NewService(repo Repository, logger *log.Logger, jwtSvc jwt.Service) Service {
	return &service{
		repo:   repo,
		log:    logger,
		jwtSvc: jwtSvc,
	}
}

func (s *service) Register(ctx context.Context, dto RegisterUserDTO) (*UserResponseDTO, error) {
	dto.Role = ""
	return s.createUser(ctx, dto, string(models.RoleCustomer))
}

func (s *service) CreateUserByAdmin(ctx context.Context, dto RegisterUserDTO) (*UserResponseDTO, error) {
	s.log.Info("admin attempting to create a new user", "email", dto.Email, "role", dto.Role)
	return s.createUser(ctx, dto, dto.Role)
}

func (s *service) createUser(ctx context.Context, dto RegisterUserDTO, roleToCreate string) (*UserResponseDTO, error) {
	s.log.Info("starting user creation process", "email", dto.Email, "role", roleToCreate)

	if err := dto.Validate(); err != nil {
		return nil, fault.New("invalid registration data", fault.WithKind(fault.KindValidation), fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err))
	}

	_, err := s.repo.FindUserByEmail(ctx, dto.Email)
	if err == nil {
		s.log.Warn("creation attempt with existing email", "email", dto.Email)
		return nil, fault.New("email already registered", fault.WithKind(fault.KindConflict), fault.WithHTTPCode(http.StatusConflict))
	}
	var f *fault.Error
	if errors.As(err, &f) && f.Kind != fault.KindNotFound {
		s.log.Error("failed to check for existing user", "error", err)
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", "error", err)
		return nil, fault.New("could not process registration", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	newUser, err := NewUser(uuid.NewString(), dto.Name, dto.Email, string(hashedPassword), []string{roleToCreate}, time.Now())
	if err != nil {
		s.log.Error("failed to create new user entity", "error", err)
		return nil, fault.New("invalid user entity data", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	if err := s.repo.CreateUser(ctx, newUser, roleToCreate); err != nil {
		s.log.Error("failed to save user to repository", "error", err)
		return nil, fault.New("could not save user", fault.WithHTTPCode(http.StatusInternalServerError), fault.WithError(err))
	}

	s.log.Info("user created successfully", "user_id", newUser.ID(), "email", dto.Email, "role", roleToCreate)
	return toUserResponseDTO(newUser), nil
}

func (s *service) Login(ctx context.Context, dto LoginDTO) (*LoginResponseDTO, error) {
	s.log.Info("user login attempt", "email", dto.Email)

	if err := dto.Validate(); err != nil {
		return nil, fault.New("invalid login data", fault.WithKind(fault.KindValidation), fault.WithHTTPCode(http.StatusBadRequest), fault.WithError(err))
	}

	user, err := s.repo.FindUserByEmail(ctx, dto.Email)
	if err != nil {
		s.log.Warn("login failed: user not found", "email", dto.Email, "error", err)
		return nil, fault.New("invalid credentials", fault.WithKind(fault.KindUnauthenticated), fault.WithHTTPCode(http.StatusUnauthorized))
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash()), []byte(dto.Password))
	if err != nil {
		s.log.Warn("login failed: invalid password", "email", dto.Email)
		return nil, fault.New("invalid credentials", fault.WithKind(fault.KindUnauthenticated), fault.WithHTTPCode(http.StatusUnauthorized))
	}

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

func toUserResponseDTO(user *User) *UserResponseDTO {
	return &UserResponseDTO{
		ID:        user.ID(),
		Name:      user.Name(),
		Email:     user.Email(),
		Roles:     user.Roles(),
		CreatedAt: user.CreatedAt(),
	}
}
