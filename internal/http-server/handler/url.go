package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"url-shortener-wb/internal/http-server/handler/dto"
	"url-shortener-wb/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
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
		h.sendJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		h.sendJSONError(w, "invalid url format", http.StatusBadRequest)
		return
	}

	alias, err := h.usecase.CreateShortURL(r.Context(), req.URL, req.Custom)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidURL) || errors.Is(err, usecase.ErrInvalidAlias) {
			h.sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, usecase.ErrAliasExists) {
			h.sendJSONError(w, "alias already exists", http.StatusConflict)
			return
		}

		h.logger.Error().Err(err).Msg("create short url failed")
		h.sendJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	shortURL := fmt.Sprintf("%s://%s/s/%s", scheme, r.Host, alias)

	resp := dto.CreateShortURLResponse{
		ShortURL: shortURL,
		Alias:    alias,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode response")
	}
}

func (h *URLHandler) RedirectToOriginal(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")
	if alias == "" {
		h.sendJSONError(w, "invalid url", http.StatusBadRequest)
		return
	}

	userAgent := r.UserAgent()
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	originalURL, err := h.usecase.GetOriginalURL(r.Context(), alias)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) || errors.Is(err, usecase.ErrInvalidAlias) {
			h.sendJSONError(w, "url not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Str("alias", alias).Msg("get original url failed")
		h.sendJSONError(w, "internal server error", http.StatusInternalServerError)
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

func (h *URLHandler) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{"error": message}
	json.NewEncoder(w).Encode(errorResponse)
}
