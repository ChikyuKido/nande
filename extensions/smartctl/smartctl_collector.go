package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ChikyuKido/nande/exporter/extension"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type SmartctlData struct {
	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`
	ModelFamily  string `json:"model_family"`
	ModelName    string `json:"model_name"`
	SerialNumber string `json:"serial_number"`
	UserCapacity struct {
		Blocks int64 `json:"blocks"`
		Bytes  int64 `json:"bytes"`
	} `json:"user_capacity"`
	LogicalBlockSize int `json:"logical_block_size"`
	SmartStatus      struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	AtaSmartAttributes struct {
		Revision int `json:"revision"`
		Table    []struct {
			Id     int64  `json:"id"`
			Name   string `json:"name"`
			Value  int64  `json:"value"`
			Worst  int64  `json:"worst"`
			Thresh int64  `json:"thresh"`
			Flags  struct {
				Prefailure bool `json:"prefailure"`
			} `json:"flags"`
			Raw struct {
				Value  int64  `json:"value"`
				String string `json:"string"`
			} `json:"raw"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
	PowerOnTime struct {
		Hours   int64 `json:"hours"`
		Minutes int64 `json:"minutes"`
	} `json:"power_on_time"`
	PowerCycleCount int64 `json:"power_cycle_count"`
	Temperature     struct {
		Current int64 `json:"current"`
	} `json:"temperature"`
}

type DeviceData struct {
	Blockdevices []struct {
		Name       string  `json:"name"`
		Mountpoint *string `json:"mountpoint"`
		Children   []struct {
			Name       string  `json:"name"`
			Mountpoint *string `json:"mountpoint"`
		} `json:"children,omitempty"`
	} `json:"blockdevices"`
}

type Device struct {
	Name       string
	Serial     string
	Price      float64
	Attributes map[string]int64
}

var oldData = make(map[string]Device)

func SmartCtlCollector() extension.Data {
	drives := strings.Split(os.Getenv("SCAN_HDDS"), ",")
	devices := make([]Device, 0)
	if len(drives) == 0 {
		return extension.Data{}
	}
	for _, fullDriveStr := range drives {
		if strings.TrimSpace(fullDriveStr) == "" {
			continue
		}
		args := strings.Split(fullDriveStr, ":")
		drive := args[0]
		mount := args[1]
		priceStr := args[2]
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			fmt.Errorf("Please enter a valid price tag to the drive. You can use 0.0 if you want to disable it: %v", err)
			price = 0.0
		}
		var buffer bytes.Buffer
		cmd := exec.Command("smartctl", "-a", "-j", drive)
		cmd.Stdout = &buffer
		if err := cmd.Run(); err != nil {
			logrus.Errorf("Failed to get data for %s: %v", drive, err)
			continue
		}
		var smartData SmartctlData
		if err := json.Unmarshal(buffer.Bytes(), &smartData); err != nil {
			logrus.Errorf("Failed to unmarshal data for %s: %v", drive, err)
		}
		var device Device
		device.Name = smartData.ModelName
		device.Serial = smartData.SerialNumber
		device.Price = price
		lbaSize := smartData.LogicalBlockSize
		var prefailureCount int64 = 0
		var prefailureWorstCount int64 = 0
		for _, attr := range smartData.AtaSmartAttributes.Table {
			if attr.Flags.Prefailure {
				if attr.Value <= attr.Thresh {
					prefailureCount++
				}
				if attr.Worst <= attr.Thresh {
					prefailureWorstCount++
				}
			}
		}
		spaceUsed, err := getSpaceUsedForDrive(mount)
		if err != nil {
			logrus.Errorf("Failed to get data for %s: %v", drive, err)
		}
		device.Attributes = make(map[string]int64)
		device.Attributes["capacity"] = smartData.UserCapacity.Bytes
		if smartData.SmartStatus.Passed {
			device.Attributes["health"] = 1
		} else {
			device.Attributes["health"] = 0
		}
		device.Attributes["power_on_hours"] = smartData.PowerOnTime.Hours
		device.Attributes["power_cycles"] = smartData.PowerCycleCount
		device.Attributes["temperature"] = smartData.Temperature.Current
		device.Attributes["prefailure_count"] = prefailureCount
		device.Attributes["prefailure_worst_count"] = prefailureWorstCount
		device.Attributes["total_bytes_written"] = getRawValueForId(smartData, 241) * int64(lbaSize)
		device.Attributes["total_bytes_read"] = getRawValueForId(smartData, 242) * int64(lbaSize)
		device.Attributes["space_used"] = spaceUsed
		device.Attributes["reallocated_sector_count"] = getRawValueForId(smartData, 5)
		device.Attributes["current_pending_sector_count"] = getRawValueForId(smartData, 197)
		device.Attributes["uncorrectable_sector_count"] = getRawValueForId(smartData, 198)
		device.Attributes["raw_read_error_rate"] = getRawValueForId(smartData, 1)
		device.Attributes["seek_error_rate"] = getRawValueForId(smartData, 7)
		device.Attributes["end_to_end_error"] = getRawValueForId(smartData, 184)
		device.Attributes["reported_uncorrect"] = getRawValueForId(smartData, 187)
		device.Attributes["command_timeout"] = getRawValueForId(smartData, 188)
		device.Attributes["load_cycle_count"] = getRawValueForId(smartData, 193)
		flyingHoursString := getRawValueStringForId(smartData, 240)
		hoursPart := strings.Split(flyingHoursString, "h")[0]
		flyingHours, err := strconv.ParseInt(hoursPart, 10, 64)
		if err != nil {
			logrus.Errorf("Failed to get data for %s: %v", drive, err)
		} else {
			device.Attributes["head_flying_hours"] = flyingHours
		}

		if val, ok := oldData[device.Serial]; ok {
			device.Attributes["bytes_written_since_period"] = device.Attributes["total_bytes_written"] - val.Attributes["total_bytes_written"]
			device.Attributes["bytes_read_since_period"] = device.Attributes["total_bytes_read"] - val.Attributes["total_bytes_read"]
		} else {
			device.Attributes["bytes_written_since_period"] = 0
			device.Attributes["bytes_read_since_period"] = 0
		}

		oldData[device.Serial] = device
		devices = append(devices, device)
	}

	data := extension.Data{}
	for _, device := range devices {
		fluxInsert := fmt.Sprintf("smartctl,serial=%s,model=%s ", device.Serial, strings.ReplaceAll(device.Name, " ", "_"))

		for key, value := range device.Attributes {
			if key == "price" {
				continue
			}
			// don't add data that was not available
			if value < 0 {
				continue
			}
			fluxInsert += fmt.Sprintf("%s=%d,", key, value)
		}
		pricePerGig := device.Price / float64(device.Attributes["capacity"]/1000/1000/1000)
		fluxInsert += fmt.Sprintf("price_per_gig=%f", pricePerGig)
		data.Metrics = append(data.Metrics, fluxInsert)
	}
	return data
}

func getRawValueForId(data SmartctlData, id int64) int64 {
	for _, attr := range data.AtaSmartAttributes.Table {
		if attr.Id == id {
			return attr.Raw.Value
		}
	}
	return -1
}
func getRawValueStringForId(data SmartctlData, id int64) string {
	for _, attr := range data.AtaSmartAttributes.Table {
		if attr.Id == id {
			return attr.Raw.String
		}
	}
	return ""
}

func getSpaceUsedForDrive(mount string) (int64, error) {
	size, err := getSpaceUsedForPartitions(mount)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func getSpaceUsedForPartitions(mount string) (int64, error) {
	cmd := exec.Command("df", "-P", mount)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to get used space for %s: %v", mount, err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("failed to parse used space for %s", mount)
	}
	fields := strings.Fields(lines[1])
	usedSpaceStr := strings.TrimSpace(fields[2])
	usedSpace, err := strconv.ParseInt(usedSpaceStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse used space for %s: %v", mount, err)
	}
	return usedSpace * 1000, nil
}
