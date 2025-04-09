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


	SFTPWrapper := handlers.NewServerFTPWrappper(cfg.Template.Path, cfg)
	MainMux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static/"))
	MainMux.Handle("/static/", http.StripPrefix("/static/", fs))


	MainMux.HandleFunc("/", SFTPWrapper.ListFilesHandler)


	log.Printf("Server starting on http://%s:%d...",cfg.Server.Host, cfg.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d",cfg.Server.Host, cfg.Server.Port), MainMux))
}
