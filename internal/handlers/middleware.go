package handlers

import (
	"log"
	"net"
	"net/http"
	"time"
)

type IPStore struct {
	bannedIPs map[string]bool
}


type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size int
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

func NewIPStore() *IPStore {
	return &IPStore{
		bannedIPs: make(map[string]bool),
	}
}


func (store *IPStore) BannedIPMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get client IP address
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // fallback if port parsing fails
		}

		if isBanned, exists := store.bannedIPs[ip]; exists && isBanned {
			http.Redirect(w, r, "/banned", 200)
			return
		}
		next(w, r)
	}
}


func (store *IPStore) BanIP(ip string) {
	store.bannedIPs[ip] = true
}

func (store *IPStore) UnbanIP(ip string) {
	delete(store.bannedIPs, ip)
}
