package exporter

import (
	"github.com/ChikyuKido/nande/exporter/database"
	extension_runner "github.com/ChikyuKido/nande/exporter/extension-runner"
	"github.com/ChikyuKido/nande/exporter/routes"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func Run() {
	s := database.InitDB(os.Getenv("INFLUX_URL"), os.Getenv("INFLUX_TOKEN"), os.Getenv("INFLUX_ORG"), os.Getenv("INFLUX_BUCKET"))
	if !s {
		logrus.Fatalf("Could not connect to InfluxDB")
	}
	s = extension_runner.RunExtensions(os.Getenv("EXTENSION_FOLDER"))
	if !s {
		logrus.Fatalf("Could not start extensions")
	}
	http.HandleFunc("/extensions/status", routes.ExtensionStatus())
	http.HandleFunc("/metrics/send", routes.SendMetrics())
	err := http.ListenAndServe(":"+os.Getenv("WEB_PORT"), nil)
	if err != nil {
		logrus.Fatal(err)
	}
}
