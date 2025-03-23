package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ChikyuKido/nande/exporter/extension"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
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
			Id    int64  `json:"id"`
			Name  string `json:"name"`
			Value int64  `json:"value"`
			Flags struct {
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
	Attributes map[string]int64
}

var oldData map[string]Device = make(map[string]Device)

func SmartCtlCollector() extension.Data {
	drives := strings.Split(os.Getenv("SCAN_HDDS"), ",")
	devices := make([]Device, 0)
	for _, drive := range drives {
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
		lbaSize := smartData.LogicalBlockSize
		var prefailureCount int64 = 0
		for _, attr := range smartData.AtaSmartAttributes.Table {
			if attr.Flags.Prefailure {
				prefailureCount++
			}
		}
		spaceUsed, err := getSpaceUsedForDrive(drive)
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
		fluxInsert := fmt.Sprintf("smartctl,serial=%s,model=%s "+
			"capacity=%d,health=%d,power_on_hours=%d,power_cycles=%d,temperature=%d,prefailure_count=%d,"+
			"total_bytes_written=%d,total_bytes_read=%d,space_used=%d,reallocated_sector_count=%d,"+
			"current_pending_sector_count=%d,uncorrectable_sector_count=%d,raw_read_error_rate=%d,"+
			"seek_error_rate=%d,end_to_end_error=%d,reported_uncorrect=%d,command_timeout=%d,"+
			"load_cycle_count=%d,head_flying_hours=%d,bytes_written_since_period=%d,bytes_read_since_period=%d",
			device.Serial,
			device.Name,
			device.Attributes["capacity"],
			device.Attributes["health"],
			device.Attributes["power_on_hours"],
			device.Attributes["power_cycles"],
			device.Attributes["temperature"],
			device.Attributes["prefailure_count"],
			device.Attributes["total_bytes_written"],
			device.Attributes["total_bytes_read"],
			device.Attributes["space_used"],
			device.Attributes["reallocated_sector_count"],
			device.Attributes["current_pending_sector_count"],
			device.Attributes["uncorrectable_sector_count"],
			device.Attributes["raw_read_error_rate"],
			device.Attributes["seek_error_rate"],
			device.Attributes["end_to_end_error"],
			device.Attributes["reported_uncorrect"],
			device.Attributes["command_timeout"],
			device.Attributes["load_cycle_count"],
			device.Attributes["head_flying_hours"],
			device.Attributes["bytes_written_since_period"],
			device.Attributes["bytes_read_since_period"])
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

func getSpaceUsedForDrive(drive string) (int64, error) {
	partitions, err := getAllPartitions(drive)
	if err != nil {
		return 0, err
	}
	size, err := getSpaceUsedForPartitions(partitions)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func getSpaceUsedForPartitions(partitions []string) (int64, error) {
	var usedSpace int64
	for _, partition := range partitions {
		cmd := exec.Command("df", "-B1", "--output=used", partition)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return 0, fmt.Errorf("failed to get used space for %s: %v", partition, err)
		}
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(lines) < 2 {
			return 0, fmt.Errorf("failed to parse used space for %s", partition)
		}
		usedSpaceStr := strings.TrimSpace(lines[1])
		currentUsedSpace, err := strconv.ParseInt(usedSpaceStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse used space for %s: %v", partition, err)
		}
		usedSpace += currentUsedSpace
	}
	return usedSpace, nil
}

func getAllPartitions(disk string) ([]string, error) {
	cmd := exec.Command("lsblk", "-o", "NAME,MOUNTPOINT", "-J")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get partitions for %s: %v", disk, err)
	}

	var data DeviceData
	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %v", err)
	}
	partitions := make([]string, 0)
	for _, block := range data.Blockdevices {
		if block.Name == filepath.Base(disk) {
			for _, child := range block.Children {
				if child.Mountpoint != nil {
					partitions = append(partitions, fmt.Sprintf("/dev/%s", child.Name))
				}
			}
		}
	}

	return partitions, nil
}
