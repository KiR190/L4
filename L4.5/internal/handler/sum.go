package handler

import (
	"net/http"
	"strconv"
)

func SumHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	aStr := q.Get("a")
	bStr := q.Get("b")

	a, err := strconv.Atoi(aStr)
	if err != nil {
		http.Error(w, "bad a", http.StatusBadRequest)
		return
	}

	b, err := strconv.Atoi(bStr)
	if err != nil {
		http.Error(w, "bad b", http.StatusBadRequest)
		return
	}

	sum := a + b

	w.Write(strconv.AppendInt(nil, int64(sum), 10))
}
