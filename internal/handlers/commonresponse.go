package handlers

import (
	"encoding/json"
	"net/http"
)

func makeResponse(w http.ResponseWriter, request *http.Request, content interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(content); err != nil {
		panic(err)
	}
}
