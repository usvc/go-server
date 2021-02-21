package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/usvc/go-server/types"
)

const (
	ProbeResponseOK          = "\"ok\""
	ProbeResponseCodeSuccess = http.StatusOK
	ProbeResponseCodeError   = http.StatusInternalServerError
)

func GetHTTPLivenessProbe(handlers types.HTTPProbeHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if errs := handlers.Do(); errs != nil {
			errsAsJSON, marshalError := json.Marshal(errs)
			w.WriteHeader(http.StatusInternalServerError)
			if marshalError != nil {
				w.Write([]byte(fmt.Sprintf("\"%s\"", marshalError.Error())))
				return
			}
			w.Write(errsAsJSON)
			return
		}
		w.WriteHeader(ProbeResponseCodeSuccess)
		w.Write([]byte(ProbeResponseOK))
	}
}

func GetHTTPMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	}
}

func GetHTTPReadinessProbe(handlers types.HTTPProbeHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if errs := handlers.Do(); errs != nil {
			errsAsJSON, marshalError := json.Marshal(errs)
			w.WriteHeader(http.StatusInternalServerError)
			if marshalError != nil {
				w.Write([]byte(fmt.Sprintf("\"%s\"", marshalError.Error())))
				return
			}
			w.Write(errsAsJSON)
			return
		}
		w.WriteHeader(ProbeResponseCodeSuccess)
		w.Write([]byte(ProbeResponseOK))
	}
}

func GetHTTPVersion(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(version))
	}
}
