package util

import (
	"github.com/sirupsen/logrus"
	"os"
)

func CheckEnvForRun() string {
	if os.Getenv("WEB_PORT") == "" {
		_ = os.Setenv("WEB_PORT", "6643")
	}
	if os.Getenv("EXTENSION_FOLDER") == "" {
		_ = os.Setenv("EXTENSION_FOLDER", "./extension-build")
	}
	if os.Getenv("INFLUX_URL") == "" {
		logrus.Fatalf("INFLUX_URL not set")
	}
	if os.Getenv("INFLUX_TOKEN") == "" {
		logrus.Fatalf("INFLUX_TOKEN not set")
	}
	if os.Getenv("INFLUX_ORG") == "" {
		logrus.Fatalf("INFLUX_ORG not set")
	}
	if os.Getenv("INFLUX_BUCKET") == "" {
		logrus.Fatalf("INFLUX_BUCKET not set")
	}
	return ""
}
