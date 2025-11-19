package handler

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse представляет стандартный формат ошибки
type ErrorResponse struct {
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Code      int                    `json:"code"`
	Details   map[string]interface{} `json:"details,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

// SuccessResponse представляет стандартный формат успешного ответа
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WriteJSON отправляет JSON ответ
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteError отправляет JSON ответ с ошибкой
func WriteError(w http.ResponseWriter, err error, statusCode int) error {
	resp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: err.Error(),
		Code:    statusCode,
	}
	return WriteJSON(w, statusCode, resp)
}

// WriteSuccess отправляет JSON ответ с успехом
func WriteSuccess(w http.ResponseWriter, data interface{}, message string) error {
	resp := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	return WriteJSON(w, http.StatusOK, resp)
}

// DecodeJSON декодирует JSON из тела запроса
func DecodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
