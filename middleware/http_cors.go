package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CORSConfiguration struct {
	// AllowCredentials sets the Access-Control-Allow-Credentials response header
	// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-allow-credentials
	AllowCredentials bool
	// AllowHeaders sets the Access-Control-Allow-Headers response header
	// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-allow-origin
	// request ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-request-headers
	AllowHeaders []string
	// AllowMethods sets the Access-Control-Allow-Methods response header
	// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-allow-methods
	// request ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-request-method
	AllowMethods []string
	// AllowOrigins sets the Access-Control-Allow-Origin response header
	// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-allow-origin
	// request ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#origin
	AllowOrigins []string
	// EnablePassthrough enables the preflight request to hit the actual endpoint
	EnablePassthrough bool
	// ExposeHeaders sets the Access-Control-Expose-Header response header
	// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-expose-headers
	ExposeHeaders []string
	// MaxAge sets the Access-Control-Max-Age response header
	// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS#access-control-max-age
	MaxAge time.Duration
}

func (cc CORSConfiguration) Get() interface{} {
	return cc
}

func NewCORS(config Configuration) Middleware {
	conf := config.Get().(CORSConfiguration)
	allowedHeaders := map[string]bool{}
	if conf.AllowHeaders != nil && len(conf.AllowHeaders) > 0 {
		for _, allowedHeader := range conf.AllowHeaders {
			allowedHeaders[http.CanonicalHeaderKey(allowedHeader)] = true
		}
	}
	allowedMethods := map[string]bool{}
	if conf.AllowMethods != nil && len(conf.AllowMethods) > 0 {
		for _, allowedMethod := range conf.AllowMethods {
			allowedMethods[allowedMethod] = true
		}
	}
	allowedOrigins := map[string]bool{}
	if conf.AllowOrigins != nil && len(conf.AllowOrigins) > 0 {
		for _, allowedOrigin := range conf.AllowOrigins {
			allowedOrigins[allowedOrigin] = true
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestHeaders := strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",")
			requestMethod := r.Header.Get("Access-Control-Request-Method")
			requestOrigin := r.Header.Get("Origin")
			w.Header().Add("Vary", "Origin")
			w.Header().Add("Vary", "Access-Control-Request-Headers")
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Origin")

			if _, allowed := allowedOrigins[requestOrigin]; allowed {
				w.Header().Add("Access-Control-Allow-Origin", requestOrigin)
			}

			if _, allowed := allowedMethods[requestMethod]; allowed {
				w.Header().Add("Access-Control-Allow-Methods", requestMethod)
			}

			for i := 0; i < len(requestHeaders); i++ {
				requestHeaders[i] = http.CanonicalHeaderKey(strings.Trim(requestHeaders[i], " "))
			}
			allowHeaders := []string{}
			allHeadersFound := true
			for _, key := range requestHeaders {
				_, allowed := allowedHeaders[key]
				allHeadersFound = allHeadersFound && allowed
				if allowed {
					allowHeaders = append(allowHeaders, key)
				}
			}
			if allHeadersFound {
				w.Header().Add("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
			}

			if conf.AllowCredentials {
				w.Header().Add("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				if conf.MaxAge > 0 {
					w.Header().Add("Access-Control-Max-Age", strconv.FormatUint(uint64(conf.MaxAge.Seconds()), 10))
				}

				if !conf.EnablePassthrough {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			if conf.ExposeHeaders != nil && len(conf.ExposeHeaders) > 0 {
				w.Header().Add("Access-Control-Expose-Headers", strings.Join(conf.ExposeHeaders, ", "))
			}

			next.ServeHTTP(w, r)
		})
	}
}
