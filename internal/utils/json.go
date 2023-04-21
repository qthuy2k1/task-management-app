package utils

import (
	"encoding/json"
	"net/http"
)

func RenderJson(w http.ResponseWriter, data interface{}) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}
