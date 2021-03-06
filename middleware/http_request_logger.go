package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/usvc/go-server/types"
)

type RequestLoggerConfiguration struct {
	Log types.Logger
}

func NewRequestLogger(config interface{}) Middleware {
	log := config.(RequestLoggerConfiguration).Log
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestStart := time.Now()
			responseWriterInstance := useResponseWriter(w)
			next.ServeHTTP(responseWriterInstance, r)
			requestDuration := time.Now().Sub(requestStart)

			message := fmt.Sprintf("%s - %s [%s] \"%s %s %s\" %v %v \"%s\" \"%s\" rt=%v id=%s",
				formatLog(r.RemoteAddr),
				formatLog(r.URL.User.Username()),
				formatLog(time.Now().UTC().Format("2/Jan/2006:15:04:05 -0700")),
				formatLog(r.Method),
				formatLog(r.RequestURI),
				formatLog(r.Proto),
				responseWriterInstance.GetStatusCode(),
				responseWriterInstance.GetContentLength(),
				formatLog(r.Referer()),
				formatLog(r.UserAgent()),
				float64(float64(requestDuration.Microseconds())/1000),
				formatInterface(r.Context().Value(RequestContextID)),
			)
			log(message)
		})
	}
}

func formatInterface(entry interface{}) string {
	if entry == nil {
		return "-"
	}
	switch entry.(type) {
	case string:
		return fmt.Sprintf("%s", entry)
	}
	return fmt.Sprintf("%v", entry)
}

func formatLog(entry string) string {
	if len(entry) == 0 {
		return "-"
	}
	return entry
}

func useResponseWriter(w http.ResponseWriter) responseWriter {
	return responseWriter{w, new(int), new(int)}
}

type responseWriter struct {
	instance      http.ResponseWriter
	statusCode    *int
	contentLength *int
}

func (rw responseWriter) GetContentLength() int {
	return *rw.contentLength
}

func (rw responseWriter) GetStatusCode() int {
	return *rw.statusCode
}

func (rw responseWriter) Header() http.Header {
	return rw.instance.Header()
}

func (rw responseWriter) Write(content []byte) (int, error) {
	size, err := rw.instance.Write(content)
	*rw.contentLength += size
	return size, err
}

func (rw responseWriter) WriteHeader(statusCode int) {
	*rw.statusCode = statusCode
	rw.instance.WriteHeader(statusCode)
}
