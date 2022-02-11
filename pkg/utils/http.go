package utils

import (
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
)

type HTTPMessageResponse struct {
	Message string `json:"message"`
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, HTTPMessageResponse{
		Message: message,
	})
}

func RespondWithStatusError(w http.ResponseWriter, err *errors.StatusError) {
	RespondWithError(w, int(err.ErrStatus.Code), err.Error())
}

func RespondNoContent(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
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
