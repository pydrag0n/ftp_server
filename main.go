package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)


const (
	rootPath      	= "./files"
	iconPath		= "./icon"
	serverPort    	= ":2121"
	templatePath  	= "index.html"
)

type File struct {
    Filename 	string
	Description string
    Size     	int64
    Date     	string
    IsDir    	bool
}

var (
	tmpl *template.Template
)

func main() {
	// Инициализация шаблона один раз при старте
	if err := os.MkdirAll(rootPath, 0755); err != nil {
        log.Fatalf("Failed to create root directory: %v", err)
    }
	var err error
	tmpl, err = template.New(templatePath).Funcs(template.FuncMap{
		"formatSize": formatSize,
		"iconForExt": getIconForExtension,
	}).ParseFiles(templatePath)

	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("icon"))
	mux.Handle("/icon/", http.StripPrefix("/icon/", fs))

	mux.HandleFunc("/files/", loggingMiddleware(listFilesHandler))
	mux.HandleFunc("/upload", loggingMiddleware(uploadFileHandler))

	log.Printf("Server starting on %s...", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, mux))
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
    size       int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
    size, err := lrw.ResponseWriter.Write(b)
    lrw.size += size
    return size, err
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lrw := &loggingResponseWriter{w, http.StatusOK, 0}
        next(lrw, r)

        log.Printf("[%s] %s %s %d %dbytes %v",
            r.RemoteAddr,
            r.Method,
            r.URL.Path,
            lrw.statusCode,
            lrw.size + int(r.ContentLength),
            time.Since(start),
        )
    }
}

func scanFiles(path string) ([]File, error) {
	var fileList []File
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Пропускаем проблемные файлы
		}
		filename := entry.Name()

		fileList = append(fileList, File{
			Filename: filename,
			Size:     info.Size(),
			Date:     info.ModTime().Format("2006-01-02 15:04"),
			IsDir:    entry.IsDir(),
		})
	}
	return fileList, nil
}

func getIconForExtension(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".zip", ".rar", ".7z", ".tar", ".gz":
        return "archive.png"
    case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
        return "image.png"
    case ".txt", ".md", ".csv":
        return "text.png"
    default:
        return "unknown.png"
    }
}


func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Обработка запроса файла
	filePath := filepath.Join(rootPath, strings.TrimPrefix(r.URL.Path, "/files/"))
	if stat, err := os.Stat(filePath); err == nil && !stat.IsDir() {
		http.ServeFile(w, r, filePath)
		return
	}

	// Обработка директории
	files, err := scanFiles(rootPath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error scanning files: %v", err)
		return
	}
	data := struct{
		Files []File
		}{
			Files: files,
		}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template Error", http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
	}
}
func formatSize(size int64) string {
    if size == 0 {
        return "0 B"
    }
    suffixes := []string{"B", "KB", "MB", "GB", "TB", "PB"}
    order := math.Log2(float64(size)) / 10
    if order > 5 {
        order = 5
    }
    value := float64(size) / math.Pow(1024, math.Floor(order))
    return fmt.Sprintf("%.1f %s", value, suffixes[int(order)])
}


func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse multipart form with reasonable memory limits
    err := r.ParseMultipartForm(10 << 20) // 10 MB limit
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

    dstPath := filepath.Join(rootPath, handler.Filename)

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



// File funcs
func (f *File) SetFilename(name string) {
    f.Filename = name
}

func (f *File) SetDescription(description string) {
	f.Description = description
}

func (f *File) SetSize(size int64) {
	f.Size = size
}

func (f *File) SetDate(date string) {
	f.Date = date
}

func (f *File) SetIsDir(isDir bool) {
	f.IsDir = isDir
}
