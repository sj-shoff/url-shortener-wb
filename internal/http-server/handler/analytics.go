package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"url-shortener-wb/internal/http-server/handler/dto"
	"url-shortener-wb/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
)

type AnalyticsHandler struct {
	usecase AnalyticsUsecase
	logger  *zlog.Zerolog
}

func NewAnalyticsHandler(
	usecase AnalyticsUsecase,
	logger *zlog.Zerolog,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		usecase: usecase,
		logger:  logger,
	}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	alias := chi.URLParam(r, "alias")
	if alias == "" {
		h.sendJSONError(w, "alias is required", http.StatusBadRequest)
		return
	}

	report, err := h.usecase.GetAnalytics(r.Context(), alias)
	if err != nil {
		if errors.Is(err, usecase.ErrNotFound) || errors.Is(err, usecase.ErrInvalidAlias) {
			h.sendJSONError(w, "url not found", http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Str("alias", alias).Msg("get analytics failed")
		h.sendJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := dto.AnalyticsResponse{
		TotalClicks:    report.TotalClicks,
		DailyStats:     report.DailyStats,
		MonthlyStats:   report.MonthlyStats,
		UserAgentStats: report.UserAgentStats,
		Clicks:         make([]dto.ClickAnalytics, len(report.Clicks)),
	}

	for i, click := range report.Clicks {
		resp.Clicks[i] = dto.ClickAnalytics{
			UserAgent: click.UserAgent,
			IPAddress: click.IPAddress,
			ClickedAt: click.ClickedAt.Format(time.RFC3339),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode analytics response")
	}
}

func (h *AnalyticsHandler) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{"error": message}
	json.NewEncoder(w).Encode(errorResponse)
}
