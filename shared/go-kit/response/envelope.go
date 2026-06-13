package response

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Data any       `json:"data"`
	Meta MetaBlock `json:"meta"`
}

type ErrorResponse struct {
	Error ErrorBlock `json:"error"`
	Meta  MetaBlock  `json:"meta"`
}

type MetaBlock struct {
	RequestID  string `json:"request_id"`
	TraceID    string `json:"trace_id"`
	NextCursor string `json:"next_cursor,omitempty"`
}

type ErrorBlock struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []FieldDetail `json:"details,omitempty"`
}

type FieldDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, status int, data any, meta MetaBlock) {
	JSON(w, status, SuccessResponse{Data: data, Meta: meta})
}

func Err(w http.ResponseWriter, status int, code, message string, meta MetaBlock) {
	JSON(w, status, ErrorResponse{
		Error: ErrorBlock{Code: code, Message: message},
		Meta:  meta,
	})
}
