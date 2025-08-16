// Package httputil provides utility functions for handling common web-related tasks,
// such as writing standardized JSON responses and handling errors.
package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hoyci/bookday/pkg/fault"
)

func RespondWithError(w http.ResponseWriter, err error) {
	var f *fault.Error
	if errors.As(err, &f) {
		RespondWithJSON(w, f.Code, map[string]string{"error": f.Message})
		return
	}
	RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "an unexpected error occurred"})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Error marshaling JSON response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
