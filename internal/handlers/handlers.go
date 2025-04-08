package handlers

import (
	"fmt"
	"io"
	"log"
	"net"
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
	theme := GetTheme(r)
    data := struct {
        CurrentPath string
		Theme 		string
        Files       []models.File
    }{
        CurrentPath: currentPath,
		Theme: 		theme,
        Files:       files,
    }

    if err := h.tmpl.Execute(w, data); err != nil {
        log.Printf("Template execution error: %v, path: %s", err, currentPath)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func CreateDirHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Получаем текущую директорию пользователя из query-параметра
    currentDir := r.URL.Query().Get("path")

    // Парсим форму
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    // Извлекаем имя директории из формы
    dirname := r.FormValue("dirname")
    if dirname == "" {
        http.Error(w, "Missing 'dirname' parameter", http.StatusBadRequest)
        return
    }

    // Собираем полный путь до новой директории
    fullPath := filepath.Join(config.RootPath, currentDir, dirname)

    // Нормализуем пути для безопасности
    cleanedRoot := filepath.Clean(config.RootPath)
    cleanedFullPath := filepath.Clean(fullPath)

    // Проверяем, что конечный путь находится внутри корневой директории
    if !strings.HasPrefix(cleanedFullPath, cleanedRoot) {
        http.Error(w, "Invalid directory path", http.StatusBadRequest)
        return
    }

    if err := os.Mkdir(fullPath, 0755); err != nil {
        if os.IsExist(err) {
            http.Error(w, "Directory already exists", http.StatusConflict)
        } else if os.IsPermission(err) {
            http.Error(w, "Permission denied", http.StatusForbidden)
        } else {
            http.Error(w, fmt.Sprintf("Error creating directory: %v", err), http.StatusInternalServerError)
        }
        return
    }
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
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

func (store *IPStore) BannedPage(w http.ResponseWriter, r *http.Request) {
    // Parse the template and handle any errors
    tmpl, err := template.ParseFiles("templates/uBannedPage.html")
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Extract IP from RemoteAddr
    ip, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        ip = r.RemoteAddr // Fallback if SplitHostPort fails
    }

    // Check if the IP is banned
    if isBanned, exists := store.bannedIPs[ip]; exists && isBanned {
        // Execute template and handle errors
        if err := tmpl.Execute(w, struct{ IP string }{ip}); err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        }
    } else {
        http.Error(w, "Forbidden", http.StatusForbidden)
    }
}


func GetTheme(r *http.Request) string {
    cookie, err := r.Cookie("theme")
    if err != nil || (cookie.Value != "light" && cookie.Value != "dark") {
        return "light"
    }
    return cookie.Value
}

func SetThemeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		theme := r.FormValue("theme")
		if theme != "light" && theme != "dark" {
            theme = "light"
        }
		http.SetCookie(w, &http.Cookie{
            Name:  "theme",
            Value: theme,
            MaxAge: 86400 * 30, // 30 дней
        })

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}
