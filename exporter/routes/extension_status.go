package routes

import (
	"encoding/json"
	extension_runner "github.com/ChikyuKido/nande/exporter/extension-runner"
	"net/http"
)

func ExtensionStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		extensions := extension_runner.GetExtensionStatuses()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(extensions); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
