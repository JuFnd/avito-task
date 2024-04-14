package middleware

import (
	"avito-track/pkg/util"
	"avito-track/pkg/variables"
	"context"
	"log/slog"
	"net/http"
)

type ICore interface {
	GetUserId(ctx context.Context, sid string) (int64, error)
	GetUserRole(ctx context.Context, id int64) (string, error)
}

func MethodMiddleware(next http.Handler, method string, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			util.SendResponse(w, r, http.StatusMethodNotAllowed, nil, variables.StatusMethodNotAllowedError, nil, logger)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthorizationMiddleware(next http.Handler, core ICore, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie(variables.SessionCookieName)
		if err != nil {
			util.SendResponse(w, r, http.StatusUnauthorized, nil, variables.StatusUnauthorizedError, nil, logger)
			return
		}

		userId, err := core.GetUserId(r.Context(), session.Value)
		if err != nil || userId == 0 {
			util.SendResponse(w, r, http.StatusUnauthorized, nil, variables.StatusUnauthorizedError, nil, logger)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), variables.UserIDKey, userId))
		next.ServeHTTP(w, r)
	})
}

func PermissionsMiddleware(next http.Handler, core ICore, role string, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId, isAuth := r.Context().Value(variables.UserIDKey).(int64)
		if !isAuth {
			util.SendResponse(w, r, http.StatusUnauthorized, nil, variables.StatusUnauthorizedError, nil, logger)
			return
		}

		userRole, err := core.GetUserRole(r.Context(), userId)
		if err != nil {
			util.SendResponse(w, r, http.StatusInternalServerError, nil, variables.StatusInternalServerError, err, logger)
			return
		}
		if userRole != role {
			util.SendResponse(w, r, http.StatusForbidden, nil, variables.StatusForbiddenError, nil, logger)
			return
		}

		next.ServeHTTP(w, r)
	})
}
