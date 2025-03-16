package extension_runner

import (
	"fmt"
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
			cmd := exec.Command("./run")
			cmd.Dir = fmt.Sprintf("%s/%s", folder, entry.Name())
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
