package utils

import (
	"fmt"
	"ftp/internal/models"
	"math"
	"os"
	"path/filepath"
	"strings"
)

func ScanFiles(path string) ([]models.File, error) {
	var fileList []models.File
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

		fileList = append(fileList, models.File{
			Filename: filename,
			Size:     info.Size(),
			Date:     info.ModTime().Format("2006-01-02 15:04"),
			IsDir:    entry.IsDir(),
		})
	}
	return fileList, nil
}

func FormatSize(size int64) string {
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
