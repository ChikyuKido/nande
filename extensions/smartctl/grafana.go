package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func CreateGrafanaConf() (string, error) {
	grafanaDashboardData, err := readFile("grafana.json")
	if err != nil {
		return "", err
	}
	var grafanaDashboard map[string]interface{}
	err = json.Unmarshal(grafanaDashboardData, &grafanaDashboard)
	if err != nil {
		return "", errors.New("failed to parse grafana dashbord file: " + err.Error())
	}
	driveRow, err := readFile("drive_row.json")
	if err != nil {
		return "", err
	}
	drives := strings.Split(os.Getenv("SCAN_HDDS"), ",")
	if len(drives) == 0 {
		return "", errors.New("no Drive found to add")
	}
	for i, driveWithPrice := range drives {
		if strings.TrimSpace(driveWithPrice) == "" {
			return "", errors.New("drive is empty")
		}
		args := strings.Split(driveWithPrice, ":")
		drive := args[0]
		serial, err := getSerialForDrive(drive)
		if err != nil {
			return "", errors.New("failed to get drive serial: " + err.Error())
		}
		model, err := getModelForDrive(drive)
		if err != nil {
			return "", errors.New("failed to get drive model: " + err.Error())
		}
		rowJson, err := adjustDriveRow(driveRow, i, serial, model)
		if err != nil {
			return "", errors.New("failed to adjust the row for the drive " + drive + ": " + err.Error())
		}
		err = addRowToDashboard(grafanaDashboard, rowJson)
		if err != nil {
			return "", errors.New("failed to add the dashboard to the grafana dashboard")
		}
	}
	data, err := json.Marshal(grafanaDashboard)
	if err != nil {
		return "", err
	}

	return string(data), nil

}

func addRowToDashboard(dashboard, row map[string]interface{}) error {
	dashboardPanels, ok := dashboard["panels"].([]interface{})
	if !ok {
		return errors.New("dashboard does not contain panels")
	}
	rowPanels, ok := row["panels"].([]interface{})
	if ok {
		dashboardPanels = append(dashboardPanels, rowPanels...)
		dashboard["panels"] = dashboardPanels
	}
	return nil
}

func getSerialForDrive(drive string) (string, error) {
	buffer := new(bytes.Buffer)
	cmd := exec.Command("sudo", "hdparm", "-I", drive)
	cmd.Stdout = buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`Serial Number:\s+(\S+)`)
	matches := re.FindStringSubmatch(buffer.String())
	if len(matches) < 2 {
		return "", fmt.Errorf("serial number not found in output")
	}
	return strings.TrimSpace(matches[1]), nil
}
func getModelForDrive(drive string) (string, error) {
	buffer := new(bytes.Buffer)
	cmd := exec.Command("sudo", "hdparm", "-I", drive)
	cmd.Stdout = buffer
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`Model Number:\s+(\S+)`)
	matches := re.FindStringSubmatch(buffer.String())
	if len(matches) < 2 {
		return "", fmt.Errorf("model number not found in output")
	}
	return strings.TrimSpace(matches[1]), nil
}

func adjustDriveRow(data []byte, count int, serial string, model string) (map[string]interface{}, error) {
	dataString := string(data)
	dataString = strings.ReplaceAll(dataString, "{SERIAL_DRIVE_ID}", serial)
	dataString = strings.ReplaceAll(dataString, "Drive {Name}", model+" "+serial)
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
