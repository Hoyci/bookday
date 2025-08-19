package jwt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hoyci/bookday/pkg/fault"
)

// AccessClaims contém os dados para autorização e identificação do usuário.
// Usado no token de curta duração.
type AccessClaims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// RefreshClaims contém apenas o necessário para renovar uma sessão.
// Usado no token de longa duração.
type RefreshClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Service define a interface para nosso serviço JWT.
type Service interface {
	GenerateAccessToken(userID string, roles []string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(tokenString string) (*AccessClaims, error)
	ValidateRefreshToken(tokenString string) (*RefreshClaims, error)
}

type jwtService struct {
	accessSecret      string
	refreshSecret     string
	accessExpiration  time.Duration
	refreshExpiration time.Duration
	issuer            string
}

func NewService(accessSecret, refreshSecret, issuer string, accessExpMinutes, refreshExpHours int) Service {
	return &jwtService{
		accessSecret:      accessSecret,
		refreshSecret:     refreshSecret,
		issuer:            issuer,
		accessExpiration:  time.Minute * time.Duration(accessExpMinutes),
		refreshExpiration: time.Hour * time.Duration(refreshExpHours),
	}
}

func (s *jwtService) GenerateAccessToken(userID string, roles []string) (string, error) {
	expirationTime := time.Now().Add(s.accessExpiration)

	claims := &AccessClaims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.accessSecret))
	if err != nil {
		return "", fault.New(
			"could not sign access token",
			fault.WithError(err),
			fault.WithHTTPCode(http.StatusInternalServerError),
			fault.WithKind(fault.KindUnexpected),
		)
	}

	return signedToken, nil
}

func (s *jwtService) GenerateRefreshToken(userID string) (string, error) {
	expirationTime := time.Now().Add(s.refreshExpiration)

	claims := &RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return "", fault.New(
			"could not sign refresh token",
			fault.WithError(err),
			fault.WithHTTPCode(http.StatusInternalServerError),
			fault.WithKind(fault.KindUnexpected),
		)
	}

	return signedToken, nil
}

func (s *jwtService) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.accessSecret), nil
	})
	if err != nil {
		return nil, fault.New("invalid access token",
			fault.WithKind(fault.KindUnauthenticated),
			fault.WithHTTPCode(http.StatusUnauthorized),
			fault.WithError(err))
	}

	if !token.Valid {
		return nil, fault.New("access token is not valid",
			fault.WithKind(fault.KindUnauthenticated),
			fault.WithHTTPCode(http.StatusUnauthorized))
	}

	return claims, nil
}

func (s *jwtService) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.refreshSecret), nil
	})
	if err != nil {
		return nil, fault.New("invalid access token",
			fault.WithKind(fault.KindUnauthenticated),
			fault.WithHTTPCode(http.StatusUnauthorized),
			fault.WithError(err))
	}

	if !token.Valid {
		return nil, fault.New("access token is not valid",
			fault.WithKind(fault.KindUnauthenticated),
			fault.WithHTTPCode(http.StatusUnauthorized))
	}

	return claims, nil
}
