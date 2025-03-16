package extension

import (
	"bytes"
	"encoding/binary"
)

type Data struct {
	ExtensionName  string
	Interval       int32
	ProcessingTime int32
	Metrics        []string
}

func SerializeData(data Data) ([]byte, error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int32(len(data.ExtensionName)))
	buf.WriteString(data.ExtensionName)
	binary.Write(&buf, binary.LittleEndian, data.Interval)
	binary.Write(&buf, binary.LittleEndian, data.ProcessingTime)

	binary.Write(&buf, binary.LittleEndian, int32(len(data.Metrics)))
	for _, metric := range data.Metrics {
		binary.Write(&buf, binary.LittleEndian, int32(len(metric)))
		buf.WriteString(metric)
	}

	return buf.Bytes(), nil
}

func DeserializeData(data []byte) (Data, error) {
	var buf bytes.Buffer
	buf.Write(data)

	var result Data

	var extensionNameLength int32
	binary.Read(&buf, binary.LittleEndian, &extensionNameLength)
	extensionNameBytes := make([]byte, extensionNameLength)
	buf.Read(extensionNameBytes)
	result.ExtensionName = string(extensionNameBytes)
	var interval int32
	binary.Read(&buf, binary.LittleEndian, &interval)
	result.Interval = interval
	var processingTime int32
	binary.Read(&buf, binary.LittleEndian, &processingTime)
	result.ProcessingTime = processingTime

	var metricsLength int32
	binary.Read(&buf, binary.LittleEndian, &metricsLength)

	for i := int32(0); i < metricsLength; i++ {
		var metric string

		var metricLength int32
		binary.Read(&buf, binary.LittleEndian, &metricLength)
		metricBytes := make([]byte, metricLength)
		buf.Read(metricBytes)
		metric = string(metricBytes)

		result.Metrics = append(result.Metrics, metric)
	}

	return result, nil
}
