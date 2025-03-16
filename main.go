package main

import (
	"github.com/ChikyuKido/nande/exporter"
	"github.com/ChikyuKido/nande/util"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{}
var grafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "For the grafana dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&util.CustomFormatter{Group: "Nande-Grafana"})
		util.CheckEnvForRun()

	},
}
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the program",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.SetFormatter(&util.CustomFormatter{Group: "Nande-Server"})
		util.CheckEnvForRun()
		exporter.Run()
	},
}

func main() {
	logrus.SetFormatter(&util.CustomFormatter{Group: "Nande"})
	logrus.SetLevel(logrus.DebugLevel)
	err := godotenv.Load()

	if err != nil {
		logrus.Warnf("No .env file found. This is fine if using docker and providing the environment variables with it")
	}
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(grafanaCmd)

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}
