package routes

import "net/http"

func SendMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
