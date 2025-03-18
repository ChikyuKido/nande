package main

import (
	"errors"
	"os"
)

func CreateGrafanaConf() (string, error) {
	file := "grafana.json"
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return "", errors.New("grafana conf does not exist")
	}
	bytes, err := os.ReadFile(file)
	if err != nil {
		return "", errors.New("failed to read grafana.json file: " + err.Error())
	}
	return string(bytes), nil
}
