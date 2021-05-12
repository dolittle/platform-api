package middleware

import (
	"mime"
	"net/http"

	"github.com/dolittle-entropy/platform-api/pkg/utils"
)

func RestrictHandlerWithHeaderName(secret string, name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xSecret := r.Header.Get(name)
			if xSecret != secret {
				utils.RespondWithError(w, http.StatusUnauthorized, "You are not authorized")
				return
			}

			// TODO confim secret1
			next.ServeHTTP(w, r)
		})
	}
}

func RestrictHandler(secret string) func(next http.Handler) http.Handler {
	return RestrictHandlerWithHeaderName(secret, "x-secret")
}

func EnforceJSONHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if contentType != "" {
			mt, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Malformed Content-Type header")
				return
			}

			if mt != "application/json" {
				utils.RespondWithError(w, http.StatusUnsupportedMediaType, "Content-Type header must be application/json")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
