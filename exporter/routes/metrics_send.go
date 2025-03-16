package routes

import (
	"context"
	"github.com/ChikyuKido/nande/exporter/database"
	"github.com/ChikyuKido/nande/exporter/extension"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func SendMetrics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		data, err := extension.DeserializeData(bytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		failed := false
		for _, line := range data.Metrics {
			err := database.DB.WriteApi.WriteRecord(context.Background(), line)
			if err != nil {
				logrus.Errorf("Write error: %v", err)
				failed = true
			}
		}
		if failed {
			http.Error(w, "Failed to insert all metrics", http.StatusBadRequest)
			return
		}
		logrus.Infof("Inserted %d metrics", len(data.Metrics))
		w.WriteHeader(http.StatusOK)
	}
}
