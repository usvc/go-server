package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	DefaultRequestIdentifierHeaderKey = "X-Request-ID"
	RequestContextID                  = "request_context_id"
)

type RequestIdentifierConfiguration struct {
	HeaderKey string
}

func NewRequestIdentifier(config interface{}) Middleware {
	headerKey := config.(RequestIdentifierConfiguration).HeaderKey
	if len(headerKey) == 0 {
		headerKey = DefaultRequestIdentifierHeaderKey
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New().String()
			w.Header().Add(headerKey, id)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), RequestContextID, id)))
		})
	}
}
