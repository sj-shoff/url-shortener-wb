package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"

	"url-shortener-wb/internal/http-server/handler/dto"
)

type URLHandler struct {
	usecase     URLUsecase
	analyticsUC AnalyticsUsecase
	logger      *zlog.Zerolog
}

func NewURLHandler(
	usecase URLUsecase,
	analyticsUC AnalyticsUsecase,
	logger *zlog.Zerolog,
) *URLHandler {
	return &URLHandler{
		usecase:     usecase,
		analyticsUC: analyticsUC,
		logger:      logger,
	}
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	alias, err := h.usecase.CreateShortURL(r.Context(), req.URL, req.Custom)
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

	originalURL, err := h.usecase.GetOriginalURL(r.Context(), alias)
	if err != nil {
		http.Error(w, "url not found", http.StatusNotFound)
		h.logger.Error().Err(err).Str("alias", alias).Msg("url not found")
		return
	}

	go func() {
		ctx := context.Background()
		if err := h.analyticsUC.RecordClick(ctx, alias, userAgent, ip); err != nil {
			h.logger.Warn().Err(err).Str("alias", alias).Msg("failed to record click")
		}
	}()

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
