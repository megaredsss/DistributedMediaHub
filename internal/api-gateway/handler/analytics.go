package handler

import (
	"fmt"
	"net/http"

	analyticspb "github.com/yourusername/DistributedMediaHub/api/proto/analytics"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
	"github.com/go-chi/chi/v5"
)

// AnalyticsHandler обрабатывает запросы к Analytics Service
type AnalyticsHandler struct {
	analyticsClient analyticspb.AnalyticsServiceClient
	logger          *logger.Logger
}

// NewAnalyticsHandler создает новый AnalyticsHandler
func NewAnalyticsHandler(analyticsClient analyticspb.AnalyticsServiceClient, log *logger.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsClient: analyticsClient,
		logger:          log,
	}
}

// GetAnalytics обрабатывает получение аналитики по видео
// GET /api/analytics/video/{id}
func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	if videoID == "" {
		WriteError(w, fmt.Errorf("video_id is required"), http.StatusBadRequest)
		return
	}

	req := &analyticspb.GetAnalyticsRequest{
		VideoId: videoID,
	}

	// Вызываем Analytics Service
	resp, err := h.analyticsClient.GetAnalytics(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get analytics")
		WriteError(w, fmt.Errorf("analytics not found"), http.StatusNotFound)
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}
