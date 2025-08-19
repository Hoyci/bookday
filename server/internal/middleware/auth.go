package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/hoyci/bookday/pkg/fault"
	"github.com/hoyci/bookday/pkg/httputil"
	"github.com/hoyci/bookday/pkg/jwt"
)

// Definimos tipos personalizados para as chaves do contexto para evitar colisões.
type contextKey string

const (
	UserIDKey contextKey = "userID"
	RolesKey  contextKey = "roles"
)

// Authenticator é um middleware que depende do serviço JWT para validar tokens.
type Authenticator struct {
	jwtSvc jwt.Service
}

// NewAuthenticator cria uma nova instância do middleware de autenticação.
func NewAuthenticator(jwtSvc jwt.Service) *Authenticator {
	return &Authenticator{jwtSvc: jwtSvc}
}

// AuthMiddleware verifica a presença e validade de um Access Token JWT.
func (a *Authenticator) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extrair o cabeçalho de autorização.
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := fault.New("cabeçalho de autorização em falta",
				fault.WithKind(fault.KindUnauthenticated),
				fault.WithHTTPCode(http.StatusUnauthorized))
			httputil.RespondWithError(w, err)
			return
		}

		// 2. Verificar se o formato é "Bearer <token>".
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
			err := fault.New("formato de autorização inválido",
				fault.WithKind(fault.KindUnauthenticated),
				fault.WithHTTPCode(http.StatusUnauthorized))
			httputil.RespondWithError(w, err)
			return
		}

		tokenString := headerParts[1]

		// 3. Validar o Access Token.
		claims, err := a.jwtSvc.ValidateAccessToken(tokenString)
		if err != nil {
			// O erro de validação já vem formatado do serviço JWT.
			httputil.RespondWithError(w, err)
			return
		}

		// 4. Se o token for válido, adicionar os dados do utilizador ao contexto da requisição.
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RolesKey, claims.Roles)

		// Passar a requisição (com o novo contexto) para o próximo handler na cadeia.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole é um middleware que verifica se o utilizador autenticado possui um dos papéis permitidos.
// Esta função retorna um middleware, permitindo a sua configuração com os papéis desejados.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Recuperar os papéis do utilizador a partir do contexto.
			roles, ok := r.Context().Value(RolesKey).([]string)
			if !ok || len(roles) == 0 {
				err := fault.New("acesso negado: papéis não encontrados no contexto",
					fault.WithKind(fault.KindForbidden),
					fault.WithHTTPCode(http.StatusForbidden))
				httputil.RespondWithError(w, err)
				return
			}

			// Criar um mapa para uma verificação mais eficiente.
			allowedRolesMap := make(map[string]struct{})
			for _, role := range allowedRoles {
				allowedRolesMap[role] = struct{}{}
			}

			// 2. Verificar se o utilizador possui pelo menos um dos papéis permitidos.
			isAuthorized := false
			for _, userRole := range roles {
				if _, found := allowedRolesMap[userRole]; found {
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

			// 3. Se autorizado, continuar para o próximo handler.
			next.ServeHTTP(w, r)
		})
	}
}
