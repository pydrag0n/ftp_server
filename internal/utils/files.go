package utils

import (
	"fmt"
	"ftp/config"
	"ftp/internal/models"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"crypto/rand"
    "encoding/hex"
    "sync"
    "time"

)

type Session struct {
    Data      map[string]interface{}
    ExpiresAt time.Time
}

type SessionStore struct {
    Sessions map[string]Session
    MU       sync.Mutex
}

var Store = &SessionStore{
    Sessions: make(map[string]Session),
}


func GenerateSessionID() string {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        panic(err)
    }
    return hex.EncodeToString(b)
}

func ScanFiles(path string, basePath string) ([]models.File, error) {
    var fileList []models.File

    entries, err := os.ReadDir(path)
    if err != nil {
        return nil, fmt.Errorf("scan error for %s: %w", path, err)
    }

    // Добавляем ссылку на родительскую директорию
    if path != config.RootPath {
        fileList = append(fileList, models.File{
            Filename: "..",
            Path:     filepath.Dir(basePath),
            IsDir:    true,
        })
    }

    for _, entry := range entries {
        info, err := entry.Info()
        if err != nil {
            log.Printf("Skipping problematic entry %s: %v", entry.Name(), err)
            continue
        }

        relPath := filepath.Join(basePath, entry.Name())
        if basePath == "/" {
            relPath = "/" + entry.Name()
        }

        fileList = append(fileList, models.File{
            Filename: entry.Name(),
            Path:     relPath,
            Size:     info.Size(),
            Date:     info.ModTime().Format("2006-01-02 15:04"),
            IsDir:    entry.IsDir(),
        })
    }

    return fileList, nil
}

func HasInvalidChars(filename string) bool {
    return strings.ContainsAny(filename, "\\/:*?<>|")
}
func FormatSize(size int64) string {
    if size == 0 {
        return "0"
    }
    suffixes := []string{"", "K", "M", "GB", "TB", "PB"}
    order := math.Log2(float64(size)) / 10
    if order > 5.0 {
        order = 5.0
    }
    value := float64(size) / math.Pow(1024, math.Floor(order))
	if order > 1 {
		return fmt.Sprintf("%.1f%s", value, suffixes[int(order)])
	} else {
		return fmt.Sprintf("%.0f%s", value, suffixes[int(order)])
	}
}


func GetIconForExtension(filename string) string {
    ext := strings.ToLower(filepath.Ext(filename))
    switch ext {
    case ".zip", ".rar", ".7z", ".tar", ".gz", ".xz":
        return "archive.png"
    case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
        return "image.png"
    case ".txt", ".md", ".csv":
        return "text.png"
    default:
        return "unknown.png"
    }
}
