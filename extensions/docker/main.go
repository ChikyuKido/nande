package main

import (
	"github.com/ChikyuKido/nande/exporter/extension"
	"os"
)

func main() {
	extension.Start(os.Args, DockerCollector, CreateGrafanaConf)
}
