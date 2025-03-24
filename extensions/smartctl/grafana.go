package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func CreateGrafanaConf() (string, error) {
	grafanaDashboardData, err := readFile("grafana.json")
	if err != nil {
		return "", err
	}
	var grafanaDasboard map[string]interface{}
	err = json.Unmarshal(grafanaDashboardData, &grafanaDasboard)
	if err != nil {
		return "", errors.New("failed to parse grafana dashbord file: " + err.Error())
	}
	driveRow, err := readFile("drive_row.json")
	if err != nil {
		return "", err
	}
	drives := strings.Split(os.Getenv("SCAN_HDDS"), ",")
	if len(drives) == 0 {
		logrus.Error("No drive found to add")
		return "", errors.New("no Drive found to add")
	}
	for i, driveWithPrice := range drives {
		if strings.TrimSpace(driveWithPrice) == "" {
			logrus.Error("Drive is empty")
			continue
		}
		args := strings.Split(driveWithPrice, ":")
		drive := args[0]
		serial, err := getSerialForDrive(drive)
		if err != nil {
			logrus.Errorf("Failed to get serial for drive %s. The dashboard is not complete : %v", drive, err)
			continue
		}
		rowJson, err := adjustDriveRow(driveRow, i, serial)
		if err != nil {
			logrus.Errorf("Failed to adjust the row for the drive %s. The dashboard is not complete : %v", drive, err)
			continue
		}
		err = addRowToDashboard(grafanaDasboard, rowJson)
		if err != nil {
			logrus.Errorf("Failed to add the row to the grafana dashboard. The dashboard is not complete : %v", err)
			continue
		}
	}
	data, err := json.Marshal(grafanaDasboard)
	if err != nil {
		return "", err
	}

	return string(data), nil

}

func addRowToDashboard(dashboard, row map[string]interface{}) error {
	if dashboardPanels, ok := dashboard["panels"].([]interface{}); ok {
		if rowPanels, ok := row["panels"].([]interface{}); ok {
			for _, rowPanel := range rowPanels {
				dashboardPanels = append(dashboardPanels, rowPanel)
			}
		}
	} else {
		return errors.New("dashboard does not contain panels")
	}
	return nil
}

func getSerialForDrive(drive string) (string, error) {
	buffer := new(bytes.Buffer)
	cmd := exec.Command("hdparm", "-I", drive, "|", "grep", "Serial Number")
	cmd.Stdout = buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	serial := strings.Split(buffer.String(), ":")[1]
	serial = strings.TrimSpace(serial)
	return serial, nil
}

func adjustDriveRow(data []byte, count int, serial string) (map[string]interface{}, error) {

	dataString := string(data)
	dataString = strings.ReplaceAll(dataString, "{SERIAL_DRIVE_ID}", serial)
	newData := []byte(dataString)

	var rowJson map[string]interface{}
	err := json.Unmarshal(newData, &rowJson)
	if err != nil {
		return nil, errors.New("failed to parse grafana dashbord file: " + err.Error())
	}

	if panels, ok := rowJson["panels"].([]interface{}); ok {
		for _, panel := range panels {
			if panelMap, ok := panel.(map[string]interface{}); ok {
				if gridPos, ok := panelMap["gridPos"].(map[string]interface{}); ok {
					if y, ok := gridPos["y"].(float64); ok {
						gridPos["y"] = y + float64(count)*27
					}
				}
			}
		}
	} else {
		return nil, errors.New("grafana dashboard does not contain panels")
	}
	return rowJson, nil
}

func readFile(path string) ([]byte, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("grafana conf does not exist")
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("failed to read grafana.json file: " + err.Error())
	}
	return bytes, nil
}
