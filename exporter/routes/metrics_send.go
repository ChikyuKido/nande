package routes

import (
	"context"
	"fmt"
	"github.com/ChikyuKido/nande/exporter/database"
	"github.com/ChikyuKido/nande/exporter/extension"
	extension_runner "github.com/ChikyuKido/nande/exporter/extension-runner"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
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
		timestamp := time.Now().UnixNano()
		for _, line := range data.Metrics {
			line = line + fmt.Sprintf("%d", timestamp)
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
		logrus.Infof("Inserted %d metrics from %s", len(data.Metrics), data.ExtensionName)
		extension_runner.UpdateExtensionStats(data)
		w.WriteHeader(http.StatusOK)
	}
}
