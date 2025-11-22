package router

import (
	"net/http"
	"path/filepath"

	"url-shortener-wb/internal/http-server/handler"
)

type Handler struct {
	UrlH       *handler.URLHandler
	AnalyticsH *handler.AnalyticsHandler
}

func SetupRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", h.UrlH.CreateShortURL)
	mux.HandleFunc("GET /s/{alias}", h.UrlH.RedirectToOriginal)
	mux.HandleFunc("GET /analytics/{alias}", h.AnalyticsH.GetAnalytics)

	staticDir := "./static"
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	return mux
}
