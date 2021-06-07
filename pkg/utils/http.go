package utils

import (
	"encoding/json"
	"net/http"
)

type HTTPMessageResponse struct {
	Message string `json:"message"`
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"message": message})
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondWithYAML(w http.ResponseWriter, code int, payload []byte) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(code)
	w.Write(payload)
}
