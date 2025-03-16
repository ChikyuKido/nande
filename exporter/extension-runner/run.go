package extension_runner

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

func RunExtensions(folder string) bool {
	dir, err := os.ReadDir(folder)
	if err != nil {
		return false
	}

	for _, entry := range dir {
		if entry.IsDir() {
			extensionDir := fmt.Sprintf("%s/%s", folder, entry.Name())
			if !IsExtensionActive(extensionDir) {
				logrus.Infof("Extension in %s not active. Skipping it.\n", extensionDir)
				continue
			}
			cmd := exec.Command("./run")
			cmd.Dir = extensionDir
			cmd.Env = append(os.Environ(), "URL=http://localhost:"+os.Getenv("WEB_PORT")+"/metrics/send")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			go func() {
				logrus.Infof("Running extension command: %s", cmd.String())
				err := cmd.Run()
				if err != nil {
					logrus.Errorf("failed to run extension runner: %v", err)
				}
			}()
		}
	}
	return true
}

func IsExtensionActive(folder string) bool {
	envMap, err := godotenv.Read(folder + "/.env")
	if err != nil {
		logrus.Errorf("Error reading .env file: %v", err)
		return false
	}
	return envMap["ACTIVE"] == "true"
}
