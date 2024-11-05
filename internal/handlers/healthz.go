package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *ApiConfig) handlerHealthz(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Status string `json:"status"`
	}

	res := response{
		Status: "ok",
	}

	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", res)
		log.Printf("Error: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
