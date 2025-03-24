package commands

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func CreateDashboard(args []string) {
	if len(args) != 2 {
		logrus.Fatalf("Can't create dashboard. Create dashboard needs the extensions name")
	}
	if args[1] != "all" {
		createDashboardFor(args[1])
	} else {
		dir, err := os.ReadDir(os.Getenv("EXTENSION_FOLDER"))
		if err != nil {
			return
		}
		for _, file := range dir {
			if file.IsDir() {
				fmt.Println("Creating dashboard for", file.Name())
				createDashboardFor(file.Name())
			}
		}
	}
}
func createDashboardFor(name string) {
	dir := os.Getenv("EXTENSION_FOLDER") + "/" + name
	_, err := os.ReadDir(dir)
	if err != nil {
		logrus.Fatalf("Extension '%s' does not exists", name)
	}
	data := getConfigForExtension(dir, name)
	err = sendConfigToGrafana(data)
	if err != nil {
		logrus.Fatalf("Could not create dashboard on grafana: %v", err)
	}
}
func sendConfigToGrafana(data []byte) error {
	url := os.Getenv("GRAFANA_URL") + "/api/dashboards/import"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GRAFANA_TOKEN"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Dashboard created successfully!")
	} else {
		return errors.New(string(body))
	}
	return nil
}

func getConfigForExtension(dir string, name string) []byte {
	var out bytes.Buffer
	cmd := exec.Command("./run", "grafana")
	cmd.Dir = dir
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logrus.Fatalf("Can't load grafana extension for '%s' extension: %v", name, err)
	}
	data := out.String()
	data = strings.TrimSpace(data)
	if strings.HasPrefix(data, "1") {
		logrus.Fatalf("Failed to get extension data for %s: %s", name, data[1:])
	}
	return reformatConfigForGrafana(out.Bytes())
}

func reformatConfigForGrafana(data []byte) []byte {
	jsonData := strings.Trim(string(data), " ")
	jsonData = strings.Trim(jsonData, "\n")

	formattedJSON := fmt.Sprintf("{\"dashboard\": %s}", jsonData)

	grafanaDatasourceID := os.Getenv("GRAFANA_INFLUX_DATASOURCE_ID")
	finalJSON := strings.ReplaceAll(formattedJSON, "${DS_INFLUXDB}", grafanaDatasourceID)
	return []byte(finalJSON)
}
