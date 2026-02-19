package delivery

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes a JSON-encoded response with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
