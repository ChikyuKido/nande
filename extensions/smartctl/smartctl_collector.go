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
	Attributes map[string]interface{}
}

func SmartCtlCollector() extension.Data {
	drives := strings.Split(os.Getenv("SCAN_HDDS"), ",")
	//devices := make([]Device, 0)
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

		driveName := smartData.Device.Name
		model := smartData.ModelName
		serial := smartData.SerialNumber
		capacityBytes := smartData.UserCapacity.Bytes
		lbaSize := smartData.LogicalBlockSize
		healthPassed := smartData.SmartStatus.Passed
		powerOnHours := smartData.PowerOnTime.Hours
		powerCycles := smartData.PowerCycleCount
		temperature := smartData.Temperature.Current
		prefailureCount := 0

		var totalLBAsWritten, totalLBAsRead int64
		for _, attr := range smartData.AtaSmartAttributes.Table {
			switch attr.Name {
			case "Total_LBAs_Written":

				totalLBAsWritten = attr.Raw.Value
			case "Total_LBAs_Read":
				totalLBAsRead = attr.Raw.Value
			}
			if attr.Flags.Prefailure {
				prefailureCount++
			}
		}
		totalBytesWritten := totalLBAsWritten * int64(lbaSize)
		totalBytesRead := totalLBAsRead * int64(lbaSize)
		spaceUsed, err := getSpaceUsedForDrive(drive)
		if err != nil {
			logrus.Errorf("Failed to get data for %s: %v", drive, err)
		}
		fmt.Printf("Drive: %s (%s) - Serial: %s\n", driveName, model, serial)
		fmt.Printf("Capacity: %d GB | LBA Size: %d Bytes\n", capacityBytes/1e9, lbaSize)
		fmt.Printf("SMART Health: %t | Prefailure Attributes: %d\n", healthPassed, prefailureCount)
		fmt.Printf("Power-On Hours: %d | Power Cycles: %d\n", powerOnHours, powerCycles)
		fmt.Printf("Temperature: %d°C\n", temperature)
		fmt.Printf("Total used space: %.2f GB\n", float64(spaceUsed)/1e9)
		fmt.Printf("Total Data Written: %d GB | Total Data Read: %d GB\n", totalBytesWritten/1e9, totalBytesRead/1e9)
		fmt.Println("------------------------------------------------------")

	}
	return extension.Data{}
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
