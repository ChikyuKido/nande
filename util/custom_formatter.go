package util

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type CustomFormatter struct {
	Group string
}

// Format defines how the log entry will be formatted
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s\n",
		f.Group,
		entry.Time.Format(time.RFC3339),
		entry.Level.String(),
		entry.Message,
	)
	return []byte(logLine), nil
}
