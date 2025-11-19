package handler

import (
	"fmt"
	"net/http"

	streamingpb "github.com/yourusername/DistributedMediaHub/api/proto/streaming"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
	"github.com/go-chi/chi/v5"
)

// StreamingHandler обрабатывает запросы к Streaming Service
type StreamingHandler struct {
	streamingClient streamingpb.StreamingServiceClient
	logger          *logger.Logger
}

// NewStreamingHandler создает новый StreamingHandler
func NewStreamingHandler(streamingClient streamingpb.StreamingServiceClient, log *logger.Logger) *StreamingHandler {
	return &StreamingHandler{
		streamingClient: streamingClient,
		logger:          log,
	}
}

// GetStream обрабатывает получение потока для видео
// GET /api/stream/{id}
func (h *StreamingHandler) GetStream(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	if videoID == "" {
		WriteError(w, fmt.Errorf("video_id is required"), http.StatusBadRequest)
		return
	}

	req := &streamingpb.GetStreamRequest{
		VideoId: videoID,
	}

	// Вызываем Streaming Service
	resp, err := h.streamingClient.GetStream(r.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get stream")
		WriteError(w, fmt.Errorf("stream not found"), http.StatusNotFound)
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}
