package handler

import (
	"fmt"
	"net/http"

	uploadpb "github.com/yourusername/DistributedMediaHub/api/proto/upload"
	"github.com/yourusername/DistributedMediaHub/internal/logger"
)

// UploadHandler обрабатывает запросы к Upload Service
type UploadHandler struct {
	uploadClient uploadpb.UploadServiceClient
	logger       *logger.Logger
}

// NewUploadHandler создает новый UploadHandler
func NewUploadHandler(uploadClient uploadpb.UploadServiceClient, log *logger.Logger) *UploadHandler {
	return &UploadHandler{
		uploadClient: uploadClient,
		logger:       log,
	}
}

// UploadFile обрабатывает загрузку файла
// POST /api/upload
func (h *UploadHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	var req uploadpb.UploadFileRequest
	if err := DecodeJSON(r, &req); err != nil {
		h.logger.WithError(err).Error("Failed to decode upload request")
		WriteError(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	// Валидация
	if req.FilePath == "" {
		WriteError(w, fmt.Errorf("file_path is required"), http.StatusBadRequest)
		return
	}

	// Вызываем Upload Service
	resp, err := h.uploadClient.UploadFile(r.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Upload failed")
		WriteError(w, fmt.Errorf("upload failed"), http.StatusInternalServerError)
		return
	}

	h.logger.WithField("file_path", resp.FilePath).Info("File uploaded successfully")
	WriteJSON(w, http.StatusCreated, resp)
}
