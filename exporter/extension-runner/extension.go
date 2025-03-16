package extension_runner

import "time"

type Extension struct {
	Name           string
	Interval       int32
	LastSync       time.Time
	ProcessingTime int32
}
