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
type GrafanaConfigFunc func() (string, error)

var (
	INTERVAL       int32
	EXTENSION_NAME string
	URL            string
)

func Start(args []string, metricFunc MetricFunc, grafanaConfigFunc GrafanaConfigFunc) {
	initEnv()
	logrus.SetFormatter(&util.CustomFormatter{Group: EXTENSION_NAME})
	success := checkEnv()
	if !success {
		logrus.Fatalf("Failed to start %s", EXTENSION_NAME)
	}
	if len(args) == 1 {
		logrus.Fatalf("Could not start extension because no args were provided")
	}
	if args[1] == "run" {
		Run(metricFunc)
	} else if args[1] == "grafana" {
		RunGrafana(grafanaConfigFunc)
	}

}

func RunGrafana(grafanaConfigFunc GrafanaConfigFunc) {
	data, err := grafanaConfigFunc()
	if err != nil {
		fmt.Printf("1%s", err.Error())
		return
	}
	fmt.Println(data)
}

func Run(metricFunc MetricFunc) {
	err := sendData(Data{
		ExtensionName: EXTENSION_NAME,
		Interval:      INTERVAL,
	})
	if err != nil {
		logrus.Fatalf("Failed to send init commit: %v", err)
	}
	for {
		startTime := time.Now()
		data := metricFunc()
		diff := time.Since(startTime).Milliseconds()
		data.Interval = INTERVAL
		data.ExtensionName = EXTENSION_NAME
		data.ProcessingTime = int32(diff)
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
	if os.Getenv("EXTENSION_NAME") == "" {
		logrus.Error("EXTENSION_NAME environment variable not set")
		return false
	} else {
		EXTENSION_NAME = os.Getenv("EXTENSION_NAME")
	}
	return true
}
