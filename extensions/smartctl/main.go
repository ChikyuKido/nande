package main

import (
	"github.com/ChikyuKido/nande/exporter/extension"
)

func main() {
	extension.Run(SmartCtlCollector)
}
