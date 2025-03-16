package extension_runner

import (
	"github.com/ChikyuKido/nande/exporter/extension"
	"time"
)

var extensionsStatus = make(map[string]Extension)

func UpdateExtensionStats(data extension.Data) {
	if val, ok := extensionsStatus[data.ExtensionName]; !ok {
		extensionsStatus[data.ExtensionName] = Extension{
			Name:     data.ExtensionName,
			Interval: data.Interval,
			LastSync: time.Now(),
		}
	} else {
		val.ProcessingTime = data.ProcessingTime
		val.LastSync = time.Now()
		extensionsStatus[data.ExtensionName] = val
	}
}

func GetExtensionStatuses() []Extension {
	values := make([]Extension, 0, len(extensionsStatus))
	for _, value := range extensionsStatus {
		values = append(values, value)
	}
	return values
}
