package main

import (
	"encoding/json"
	"fmt"
	"github.com/ChikyuKido/nande/exporter/extension"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strconv"
	"strings"
)

type DockerStats struct {
	ID                  string
	Name                string
	MemPercent          float64
	MemUsed             float64
	CPUPercent          float64
	TotalIORead         float64
	TotalIOWrite        float64
	IncrementalIORead   float64
	IncrementalIOWrite  float64
	TotalNetRead        float64
	TotalNetWrite       float64
	IncrementalNetRead  float64
	IncrementalNetWrite float64
}

var previousStats = make(map[string]DockerStats)

func DockerCollector() extension.Data {
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		logrus.Error(err)
	}
	outputString := string(output)
	allStats := make([]DockerStats, 0)
	for _, line := range strings.Split(outputString, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var stats map[string]interface{}
		err = json.Unmarshal([]byte(line), &stats)
		if err != nil {
			logrus.Fatalf("Failed to parse json: %v", err)
		}
		var oldStats DockerStats
		if previous, ok := previousStats[stats["ID"].(string)]; ok {
			oldStats = previous
		}
		var result DockerStats
		result.ID = stats["ID"].(string)
		result.Name = stats["Name"].(string)
		blockIO := strings.Split(stats["BlockIO"].(string), "/")
		result.TotalIORead = parseIO(blockIO[0])
		result.TotalIOWrite = parseIO(blockIO[1])
		netIO := strings.Split(stats["NetIO"].(string), "/")
		result.TotalNetRead = parseIO(netIO[0])
		result.TotalNetWrite = parseIO(netIO[1])
		mem := strings.Split(stats["MemUsage"].(string), "/")
		result.MemUsed = parseMem(mem[0])
		result.CPUPercent = parsePercent(stats["CPUPerc"].(string))
		result.MemPercent = parsePercent(stats["MemPerc"].(string))

		if oldStats.ID != "" {
			result.IncrementalNetRead = result.TotalNetRead - oldStats.TotalNetRead
			result.IncrementalNetWrite = result.TotalNetWrite - oldStats.TotalNetWrite
			result.IncrementalIORead = result.TotalIORead - oldStats.TotalIORead
			result.IncrementalIOWrite = result.TotalIOWrite - oldStats.TotalIOWrite
		}
		previousStats[result.ID] = result
		allStats = append(allStats, result)
	}

	data := extension.Data{}
	for _, stat := range allStats {
		insertString := fmt.Sprintf("docker,id=%s,name=%s mem_percent=%.2f,mem_used=%.2f,cpu_percent=%.2f,total_io_read=%.2f,total_io_write=%.2f,incremental_io_read=%.2f,incremental_io_write=%.2f,total_net_read=%.2f,total_net_write=%.2f,incremental_net_read=%.2f,incremental_net_write=%.2f",
			stat.ID,
			stat.Name,
			stat.MemPercent,
			stat.MemUsed,
			stat.CPUPercent,
			stat.TotalIORead,
			stat.TotalIOWrite,
			stat.IncrementalIORead,
			stat.IncrementalIOWrite,
			stat.TotalNetRead,
			stat.TotalNetWrite,
			stat.IncrementalNetRead,
			stat.IncrementalNetWrite)

		data.Metrics = append(data.Metrics, insertString)
	}
	return data
}

func parsePercent(value string) float64 {
	value = strings.Trim(value, " ")
	value = strings.ReplaceAll(value, "%", "")
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logrus.Errorf("Failed to parse float in parsePercent: %v", err)
	}
	return floatValue
}
func parseMem(value string) float64 {
	value = strings.Trim(value, " ")
	unit := value[len(value)-3:]

	value = value[:len(value)-3]
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logrus.Errorf("Failed to parse float in parseMem: %v", err)
	}

	switch unit {
	case "kiB":
		floatValue *= 1024
	case "MiB":
		floatValue *= 1024 * 1024
	case "GiB":
		floatValue *= 1024 * 1024 * 1024
	}

	return floatValue
}
func parseIO(value string) float64 {
	value = strings.Trim(value, " ")
	unit := value[len(value)-2:]

	value = value[:len(value)-2]

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		logrus.Errorf("Failed to parse float in parseIO: %v", err)
	}

	switch unit {
	case "kB":
		floatValue *= 1000
	case "MB":
		floatValue *= 1000 * 1000
	case "GB":
		floatValue *= 1000 * 1000 * 1000
	}

	return floatValue
}
