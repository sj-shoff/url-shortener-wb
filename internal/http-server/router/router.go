package router

import (
	"net/http"
	"path/filepath"
	"strings"

	"url-shortener-wb/internal/http-server/handler"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	UrlH       *handler.URLHandler
	AnalyticsH *handler.AnalyticsHandler
}

func SetupRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Post("/shorten", h.UrlH.CreateShortURL)
	r.Get("/s/{alias}", h.UrlH.RedirectToOriginal)
	r.Get("/analytics/{alias}", h.AnalyticsH.GetAnalytics)

	staticDir := "./static"
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/s/") ||
			strings.HasPrefix(r.URL.Path, "/analytics/") ||
			strings.HasPrefix(r.URL.Path, "/static/") {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	return r
}
