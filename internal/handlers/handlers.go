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
        "splitPath": func(path string) []string {
            return strings.Split(strings.Trim(path, "/"), "/")
        },
    }).ParseFiles(templatePath))

    return &FileHandler{tmpl: tmpl}
}


func (h *FileHandler) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
    currentPath := strings.TrimPrefix(r.URL.Path, "/files")
    if currentPath == "" {
        currentPath = "/"
    }

    fullPath := filepath.Join(config.RootPath, currentPath)

    // Проверка существования пути
    fi, err := os.Stat(fullPath)
    if err != nil {
        http.Error(w, "Not Found", http.StatusNotFound)
        log.Printf("Path not found: %s, error: %v", fullPath, err)
        return
    }

    // Если это файл - отдаем его
    if !fi.IsDir() {
        http.ServeFile(w, r, fullPath)
        return
    }

    // Получаем список файлов
    files, err := utils.ScanFiles(fullPath, currentPath)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        log.Printf("Error scanning directory %s: %v", fullPath, err)
        return
    }

    data := struct {
        CurrentPath string
        Files       []models.File
    }{
        CurrentPath: currentPath,
        Files:       files,
    }

    if err := h.tmpl.Execute(w, data); err != nil {
        log.Printf("Template execution error: %v, path: %s", err, currentPath)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}


func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    currentPath := r.URL.Query().Get("path")
    fullPath := filepath.Join(config.RootPath, currentPath)

    err := r.ParseMultipartForm(10 << 20) // 10 MB
    if err != nil {
        log.Printf("Upload parse error: %v, path: %s", err, currentPath)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    file, handler, err := r.FormFile("file")
    if err != nil {
        log.Printf("File retrieve error: %v, path: %s", err, currentPath)
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Проверка имени файла
    if utils.HasInvalidChars(handler.Filename) {
        http.Error(w, "Invalid filename", http.StatusBadRequest)
        return
    }

    dstPath := filepath.Join(fullPath, handler.Filename)
    if _, err := os.Stat(dstPath); err == nil {
        http.Error(w, "File exists", http.StatusConflict)
        return
    }

    dst, err := os.Create(dstPath)
    if err != nil {
        log.Printf("File create error: %v, path: %s", err, dstPath)
        http.Error(w, "Internal Error", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    if _, err = io.Copy(dst, file); err != nil {
        log.Printf("File copy error: %v, path: %s", err, dstPath)
        http.Error(w, "Internal Error", http.StatusInternalServerError)
        return
    }

    log.Printf("File uploaded successfully: %s", dstPath)
    http.Redirect(w, r, "/files/"+currentPath, http.StatusSeeOther)
}
