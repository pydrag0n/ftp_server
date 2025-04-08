package api

import "net/http"

type Route struct {
	Method string
	Path string
	Handler http.HandlerFunc
	Middlewares []MiddlewareFunc
}

type MiddlewareFunc func(http.Handler) http.Handler

func NewRouteFunc(method, path string,
handler http.HandlerFunc,
middlewares ...MiddlewareFunc) Route {
	return Route{
		Method: method,
		Path: path,
		Handler: chainMiddlewares(handler, middlewares...),
		Middlewares: middlewares,
	}
}


// chainMiddlewares applies the middlewares to the handler in order
func chainMiddlewares(handler http.HandlerFunc, middlewares ...MiddlewareFunc) http.HandlerFunc {
	finalHandler := http.Handler(handler)
	for i := len(middlewares) - 1; i >= 0; i-- {
		finalHandler = middlewares[i](finalHandler)
	}
	return finalHandler.ServeHTTP
}
