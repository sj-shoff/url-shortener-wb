package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"

	"url-shortener-wb/internal/http-server/handler/dto"
)

type URLHandler struct {
	urlUsecase URLUsecase
	logger     *zlog.Zerolog
}

func NewURLHandler(
	urlUsecase URLUsecase,
	logger *zlog.Zerolog,
) *URLHandler {
	return &URLHandler{
		urlUsecase: urlUsecase,
		logger:     logger,
	}
}

func (h *URLHandler) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Post("/shorten", h.CreateShortURL)
	r.Get("/s/{alias}", h.RedirectToOriginal)
	r.Get("/analytics/{alias}", h.GetAnalytics)
	return r
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	alias, err := h.urlUsecase.CreateShortURL(r.Context(), req.URL, req.Custom)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.logger.Error().Err(err).Msg("create short url failed")
		return
	}

	shortURL := r.Host + "/s/" + alias
	resp := dto.CreateShortURLResponse{
		ShortURL: shortURL,
		Alias:    alias,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *URLHandler) RedirectToOriginal(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")
	if alias == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	userAgent := r.UserAgent()
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.urlUsecase.RecordClick(ctx, alias, userAgent, ip); err != nil {
		h.logger.Warn().Err(err).Str("alias", alias).Msg("failed to record click")
	}

	originalURL, err := h.urlUsecase.GetOriginalURL(r.Context(), alias)
	if err != nil {
		http.Error(w, "url not found", http.StatusNotFound)
		h.logger.Error().Err(err).Str("alias", alias).Msg("url not found")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
