package exporter

import (
	"github.com/ChikyuKido/nande/exporter/routes"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func Run() {
	http.HandleFunc("/extensions/status", routes.ExtensionStatus())
	http.HandleFunc("/metrics/send", routes.SendMetrics())
	err := http.ListenAndServe(os.Getenv("PORT"), nil)
	if err != nil {
		logrus.Fatal(err)
	}
}
