package main

import (
	"ftp/config"
	"ftp/internal/handlers"

	"log"
	"net/http"
	"os"
)

func main() {

	if err := os.MkdirAll(config.RootPath, 0755); err != nil {
		log.Fatalf("Failed to create root directory: %v", err)
	}
	ipStore := handlers.NewIPStore()
	fileHandler := handlers.NewFileHandler(config.TemplatePath)
	mux := http.NewServeMux()

	// ipStore.BanIP("26.250.92.105")

	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/banned", ipStore.BannedIPMiddleware(ipStore.BannedPage))
	mux.HandleFunc("/", handlers.LoggingMiddleware(ipStore.BannedIPMiddleware(fileHandler.ListFilesHandler)))
	mux.HandleFunc("/set-theme", handlers.SetThemeHandler)
	mux.HandleFunc("/upload", handlers.LoggingMiddleware(handlers.UploadFileHandler))
	mux.HandleFunc("/createdir", handlers.LoggingMiddleware(handlers.CreateDirHandler))

	log.Printf("Server starting on http://localhost%s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, mux))
}
