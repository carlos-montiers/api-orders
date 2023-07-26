package utils

import (
	"encoding/json"
	"net/http"
)

// JSON encodes data to json and writes it to the http response.
func WriteJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		WriteJSON(w, map[string]string{"error": err.Error()})
		return
	}

	w.Write(b)
}
