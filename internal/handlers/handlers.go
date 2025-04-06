package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"ftp/config"
	"ftp/internal/models"
	"ftp/internal/utils"
)

type FileHandler struct {
	tmpl *template.Template
}

func NewFileHandler(templatePath string) *FileHandler {
    tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
        "iconForExt": utils.GetIconForExtension,
        "formatSize": utils.FormatSize,
    }).ParseFiles(templatePath))

    return &FileHandler{
        tmpl: tmpl,
    }
}

func (h *FileHandler) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join(config.RootPath, strings.TrimPrefix(r.URL.Path, "/files/"))
	if stat, err := os.Stat(filePath); err == nil && !stat.IsDir() {
		http.ServeFile(w, r, filePath)
		return
	}

	files, err := utils.ScanFiles(config.RootPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error scanning files: %v", err)
		return
	}

	data := struct {
		Files []models.File
	}{
		Files: files,
	}

	if err := h.tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
	}
}



func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse multipart form with reasonable memory limits
    err := r.ParseMultipartForm(10*8*1024) // 10 MB limit
    if err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }

    file, handler, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Error retrieving file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create destination path safely
	if strings.ContainsAny(handler.Filename, "\\/:*?<>|") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

    dstPath := filepath.Join(config.RootPath, handler.Filename)

	if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
		http.Error(w, "File already exists", http.StatusConflict)
		return
	}

    dst, err := os.Create(dstPath)

    if err != nil {
        http.Error(w, "Error creating file", http.StatusInternalServerError)
        return
    }

    defer dst.Close()

    // Copy file content
    if _, err = io.Copy(dst, file); err != nil {
        http.Error(w, "Error saving file", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/files/", http.StatusSeeOther)
}
