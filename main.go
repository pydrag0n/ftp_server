package main

import (
	"fmt"
	"ftp/config"
	"ftp/internal/handlers"

	"log"
	"net/http"
	"os"
)

func main() {
	cfg, err := config.Load("config/config.json")
	if err != nil {
		fmt.Printf("[ERROR] Load config %v", err)
	}

	if err := os.MkdirAll(cfg.FTP.RootPath, 0755); err != nil {
		log.Fatalf("Failed to create root directory: %v", err)
	}


	ipStore := handlers.NewIPStore()
	SFTPWrapper := handlers.NewServerFTPWrappper(cfg.Template.Path, cfg)
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/banned", ipStore.BannedIPMiddleware(ipStore.BannedPage))
	mux.HandleFunc("/", handlers.LoggingMiddleware(ipStore.BannedIPMiddleware(SFTPWrapper.ListFilesHandler)))
	mux.HandleFunc("/set-theme", handlers.SetThemeHandler)
	mux.HandleFunc("/upload", handlers.LoggingMiddleware(SFTPWrapper.UploadFileHandler))
	mux.HandleFunc("/createdir", handlers.LoggingMiddleware(SFTPWrapper.CreateDirHandler))

	log.Printf("Server starting on http://%s:%d...",cfg.Server.Host, cfg.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d",cfg.Server.Host, cfg.Server.Port), mux))
}
