package utils

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

)

type Session struct {
	Data map[string]interface{}
	ExpiresAt time.Time
}

type SessionStore struct {
	Sessions map[string]Session
	MU sync.Mutex
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
