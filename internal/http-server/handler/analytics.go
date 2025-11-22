package handler

import (
	"encoding/json"
	"net/http"
	"time"
	"url-shortener-wb/internal/http-server/handler/dto"

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
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	report, err := h.usecase.GetAnalytics(r.Context(), alias)
	if err != nil {
		http.Error(w, "failed to get analytics", http.StatusInternalServerError)
		h.logger.Error().Err(err).Str("alias", alias).Msg("get analytics failed")
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
	json.NewEncoder(w).Encode(resp)
}
