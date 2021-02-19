package middleware

import (
	"fmt"
	"net/http"
)

type ExampleConfiguration struct {
	ToPrint string
}

func NewExample(config interface{}) Middleware {
	conf := config.(ExampleConfiguration)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(conf.ToPrint)
			next.ServeHTTP(w, r)
		})
	}
}
