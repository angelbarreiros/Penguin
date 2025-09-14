package middlewares

import "net/http"

type middlewareFunc func(http.HandlerFunc) http.HandlerFunc
