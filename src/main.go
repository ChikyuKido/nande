package main

import (
	"github.com/ChikyuKido/nande/exporter"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the program",
	Run: func(cmd *cobra.Command, args []string) {
		CheckEnvForRun()
		exporter.Run()
	},
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}
}
