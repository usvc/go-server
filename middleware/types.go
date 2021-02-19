package middleware

import "net/http"

type Middlewares []Middleware

type Middleware func(http.Handler) http.Handler

type Getter func(interface{}) Middleware
