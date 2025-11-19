package handler

import (
	"fmt"
	"net/http"

	videopb "github.com/yourusername/DistributedMediaHub/api/proto/video"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
	"github.com/go-chi/chi/v5"
)

// VideoHandler обрабатывает запросы к Video Service
type VideoHandler struct {
	videoClient videopb.VideoServiceClient
	logger      *logger.Logger
}

// NewVideoHandler создает новый VideoHandler
func NewVideoHandler(videoClient videopb.VideoServiceClient, log *logger.Logger) *VideoHandler {
	return &VideoHandler{
		videoClient: videoClient,
		logger:      log,
	}
}

// GetVideo обрабатывает получение видео по ID
// GET /api/videos/{id}
func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	if videoID == "" {
		WriteError(w, fmt.Errorf("video_id is required"), http.StatusBadRequest)
		return
	}

	req := &videopb.GetVideoRequest{
		VideoId: videoID,
	}

	// Вызываем Video Service
	resp, err := h.videoClient.GetVideo(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get video")
		WriteError(w, fmt.Errorf("video not found"), http.StatusNotFound)
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}
