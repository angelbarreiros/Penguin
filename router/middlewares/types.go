package middlewares

import (
	"net/http"
)

type middlewareFunc func(handleFunc) handleFunc
type handleFunc = func(http.ResponseWriter, *http.Request)
