package main

import (
	"github.com/ChikyuKido/nande/exporter/extension"
	"os"
)

func main() {

	if len(os.Args) == 1 {
		extension.Run(DockerCollector)
	} else if os.Args[1] == "grafana" {

	}
}
