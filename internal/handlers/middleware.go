package handlers

import (
	"log"
	"net/http"
	"time"
)


type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &LoggingResponseWriter{w, http.StatusOK, 0}
		next(lrw, r)

		log.Printf("[%s] %s %s %d %dbytes %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			lrw.size+int(r.ContentLength),
			time.Since(start),
		)
	}
}
