package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	CORSAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	CORSAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	CORSAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	CORSAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	CORSAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	CORSAccessControlMaxAge           = "Access-Control-Max-Age"
	CORSOrigin                        = "Origin"
	CORSAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	CORSAccessControlRequestMethod    = "Access-Control-Request-Method"
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

func NewCORS(config interface{}) Middleware {
	conf := config.(CORSConfiguration)
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
			requestHeaders := strings.Split(r.Header.Get(CORSAccessControlRequestHeaders), ",")
			if len(requestHeaders) > 0 {
				w.Header().Add("Vary", CORSAccessControlRequestHeaders)
			}
			requestMethod := r.Method
			if r.Header.Get(CORSAccessControlRequestMethod) != "" {
				requestMethod = r.Header.Get(CORSAccessControlRequestMethod)
				w.Header().Add("Vary", CORSAccessControlRequestMethod)
			}
			requestOrigin := r.Header.Get(CORSOrigin)
			if len(requestOrigin) > 0 {
				w.Header().Add("Vary", CORSOrigin)
			}
			success := true

			_, originAllowed := allowedOrigins[requestOrigin]
			if originAllowed {
				w.Header().Add(CORSAccessControlAllowOrigin, requestOrigin)
			}
			if len(requestOrigin) > 0 {
				success = success && originAllowed
			}

			_, methodAllowed := allowedMethods[requestMethod]
			if methodAllowed {
				w.Header().Add(CORSAccessControlAllowMethods, requestMethod)
			}
			if len(requestMethod) > 0 {
				success = success && methodAllowed
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
				w.Header().Add(CORSAccessControlAllowHeaders, strings.Join(allowHeaders, ","))
			}
			success = success && allHeadersFound

			if conf.AllowCredentials {
				w.Header().Add(CORSAccessControlAllowCredentials, "true")
			}

			if r.Method == http.MethodOptions && r.Header.Get(CORSAccessControlRequestMethod) != "" {
				if conf.MaxAge > 0 {
					w.Header().Add(CORSAccessControlMaxAge, strconv.FormatUint(uint64(conf.MaxAge.Seconds()), 10))
				}

				if !conf.EnablePassthrough {
					if success {
						w.WriteHeader(http.StatusNoContent)
					} else {
						w.WriteHeader(http.StatusBadRequest)
					}
					return
				}
			}

			if conf.ExposeHeaders != nil && len(conf.ExposeHeaders) > 0 {
				w.Header().Add(CORSAccessControlExposeHeaders, strings.Join(conf.ExposeHeaders, ","))
			}

			next.ServeHTTP(w, r)
		})
	}
}
