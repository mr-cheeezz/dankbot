package static

import (
	"net/http"
	"os"
	"path/filepath"
)

// IndexHandler always serves the built SPA index.html, but can set a status
// code (useful for /403, /404, or dashboard access gating without redirects).
type IndexHandler struct {
	rootDir string
	status  func(r *http.Request) int
}

func NewIndex(rootDir string, status func(r *http.Request) int) http.Handler {
	return &IndexHandler{rootDir: rootDir, status: status}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	indexPath := filepath.Join(h.rootDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		http.Error(w, "frontend build not found; run `npm run build` in web/", http.StatusServiceUnavailable)
		return
	}

	statusCode := http.StatusOK
	if h.status != nil {
		if next := h.status(r); next > 0 {
			statusCode = next
		}
	}
	if statusCode != http.StatusOK {
		payload, err := os.ReadFile(indexPath)
		if err != nil {
			http.Error(w, "frontend build not found; run `npm run build` in web/", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(statusCode)
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(payload)
		return
	}

	http.ServeFile(w, r, indexPath)
}
