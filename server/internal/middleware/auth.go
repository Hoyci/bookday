package middleware

import (
	"context"
	"net/http"
	"strings"

	models "github.com/hoyci/bookday/internal/infra/database/model"
	"github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/httputil"
	"github.com/hoyci/bookday/pkg/jwt"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RolesKey  contextKey = "roles"
)

type Authenticator struct {
	jwtSvc jwt.Service
}

func NewAuthenticator(jwtSvc jwt.Service) *Authenticator {
	return &Authenticator{jwtSvc: jwtSvc}
}

func (a *Authenticator) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := fault.New("missing authentication header",
				fault.WithKind(fault.KindUnauthenticated),
				fault.WithHTTPCode(http.StatusUnauthorized))
			httputil.RespondWithError(w, err)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
			err := fault.New("invalid authentication header format",
				fault.WithKind(fault.KindUnauthenticated),
				fault.WithHTTPCode(http.StatusUnauthorized))
			httputil.RespondWithError(w, err)
			return
		}

		tokenString := headerParts[1]

		claims, err := a.jwtSvc.ValidateAccessToken(tokenString)
		if err != nil {
			httputil.RespondWithError(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RolesKey, claims.Roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(allowedRoles ...models.RolesType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, ok := r.Context().Value(RolesKey).([]string)
			if !ok || len(roles) == 0 {
				err := fault.New("acesso negado: papéis não encontrados no contexto",
					fault.WithKind(fault.KindForbidden),
					fault.WithHTTPCode(http.StatusForbidden))
				httputil.RespondWithError(w, err)
				return
			}

			allowedRolesMap := make(map[models.RolesType]struct{})
			for _, role := range allowedRoles {
				allowedRolesMap[role] = struct{}{}
			}

			isAuthorized := false
			for _, userRole := range roles {
				if _, found := allowedRolesMap[models.RolesType(userRole)]; found {
					isAuthorized = true
					break
				}
			}

			if !isAuthorized {
				err := fault.New("acesso negado: permissões insuficientes",
					fault.WithKind(fault.KindForbidden),
					fault.WithHTTPCode(http.StatusForbidden))
				httputil.RespondWithError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
