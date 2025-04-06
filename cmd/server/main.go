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

	fileHandler := handlers.NewFileHandler(config.TemplatePath)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/files/", handlers.LoggingMiddleware(fileHandler.ListFilesHandler))
	mux.HandleFunc("/upload", handlers.LoggingMiddleware(handlers.UploadFileHandler))

	log.Printf("Server starting on http://localhost%s...", config.ServerPort)
	log.Fatal(http.ListenAndServe(config.ServerPort, mux))
}
