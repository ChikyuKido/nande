package grafana

import "github.com/ChikyuKido/nande/grafana/commands"

func Run(args []string) {
	if args[0] == "create" {
		commands.CreateDashboard(args)
	}
}
