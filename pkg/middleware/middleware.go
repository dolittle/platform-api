package middleware

import (
	"fmt"
	"mime"
	"net/http"

	"github.com/dolittle/platform-api/pkg/utils"
)

func LogTenantUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerID := r.Header.Get("Tenant-ID")
		userID := r.Header.Get("User-ID")

		fmt.Println(customerID, userID)
		next.ServeHTTP(w, r)
	})
}

func RestrictHandlerWithSharedSecretAndIDS(secret string, name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xSecret := r.Header.Get(name)
			if xSecret != secret {
				utils.RespondWithError(w, http.StatusForbidden, "Shared secret is wrong")
				return
			}

			customerID := r.Header.Get("Tenant-ID")
			userID := r.Header.Get("User-ID")

			if customerID == "" || userID == "" {
				utils.RespondWithError(w, http.StatusForbidden, "Tenant-ID or User-ID is missing")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RestrictHandlerWithSharedSecret(secret string, name string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xSecret := r.Header.Get(name)
			if xSecret != secret {
				utils.RespondWithError(w, http.StatusForbidden, "Shared secret is missing")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
