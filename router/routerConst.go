package router

type HTTPMethod string

const (
	GET     HTTPMethod = "GET"
	POST    HTTPMethod = "POST"
	PUT     HTTPMethod = "PUT"
	DELETE  HTTPMethod = "DELETE"
	PATCH   HTTPMethod = "PATCH"
	OPTIONS HTTPMethod = "OPTIONS"
	HEAD    HTTPMethod = "HEAD"
)

type middlewareType string

const (
	MethodMiddleware middlewareType = "method"
	AuthMiddleware   middlewareType = "auth"
	RoleMiddleware   middlewareType = "role"
)
