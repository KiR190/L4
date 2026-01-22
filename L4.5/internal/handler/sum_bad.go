package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type SumResponse struct {
	Result int `json:"result"`
}

func SumHandlerBad(w http.ResponseWriter, r *http.Request) {
	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	a, _ := strconv.Atoi(aStr)
	b, _ := strconv.Atoi(bStr)

	sum := a + b

	resp := SumResponse{Result: sum}

	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
