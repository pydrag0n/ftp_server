package api

import "net/http"

type Route struct {
    Method      string          // HTTP-метод: GET, POST, PUT и т.д.
    Path        string          // URL-путь, например: "/users/{id}"
    Handler     http.HandlerFunc     // Функция-обработчик запроса
    Middlewares []MiddlewareFunc // Список middleware (опционально)
}

type MiddlewareFunc func(http.Handler) http.Handler


/*
example:
	route := NewRouteFunc(
		"GET",
		"/hello",
		HelloHandler,
		LoggingMiddleware,
		HeaderMiddleware,
	)
*/
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
