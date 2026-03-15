package static

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Handler struct {
	rootDir string
}

func New(rootDir string) http.Handler {
	return &Handler{rootDir: rootDir}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	indexPath := filepath.Join(h.rootDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		http.Error(w, "frontend build not found; run `npm run build` in web/", http.StatusServiceUnavailable)
		return
	}

	cleanPath := path.Clean("/" + r.URL.Path)
	relativePath := strings.TrimPrefix(cleanPath, "/")
	if relativePath == "" {
		http.ServeFile(w, r, indexPath)
		return
	}

	targetPath := filepath.Join(h.rootDir, filepath.FromSlash(relativePath))
	if info, err := os.Stat(targetPath); err == nil {
		if info.IsDir() {
			dirIndex := filepath.Join(targetPath, "index.html")
			if _, dirErr := os.Stat(dirIndex); dirErr == nil {
				http.ServeFile(w, r, dirIndex)
				return
			}
		} else {
			http.ServeFile(w, r, targetPath)
			return
		}
	}

	if strings.Contains(filepath.Base(relativePath), ".") {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, indexPath)
}
