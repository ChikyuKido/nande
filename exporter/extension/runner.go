package extension

import (
	"bytes"
	"fmt"
	"github.com/ChikyuKido/nande/util"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"time"
)

type MetricFunc func() Data

var (
	INTERVAL int32
	URL      string
)

func Run(metricFunc MetricFunc, extensionName string) {
	initEnv()
	logrus.SetFormatter(&util.CustomFormatter{Group: extensionName})
	success := checkEnv()
	if !success {
		logrus.Fatalf("Failed to start %s", extensionName)
	}
	for {
		data := metricFunc()
		data.Interval = INTERVAL
		data.ExtensionName = extensionName
		err := sendData(data)
		if err != nil {
			logrus.Errorf("Failed to send metric data: %v", err)
		}
		time.Sleep(time.Duration(INTERVAL) * time.Second)
	}
}

func initEnv() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}
}

func sendData(data Data) error {
	dataBytes, err := SerializeData(data)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(dataBytes)
	post, err := http.Post(URL, "application/octet-stream", reader)
	if err != nil {
		return err
	}
	if post.StatusCode != 200 {
		return fmt.Errorf("failed to send metric data: %d", post.StatusCode)
	}
	return nil
}

func checkEnv() bool {
	if os.Getenv("INTERVAL") == "" {
		logrus.Error("INTERVAL environment variable not set")
		return false
	} else {
		value, err := strconv.ParseInt(os.Getenv("INTERVAL"), 10, 32)
		if err != nil {
			logrus.Errorf("INTERVAL %s is not a int", os.Getenv("INTERVAL"))
			return false
		}
		INTERVAL = int32(value)
	}
	if os.Getenv("URL") == "" {
		logrus.Error("URL environment variable not set")
		return false
	} else {
		URL = os.Getenv("URL")
	}
	return true
}
