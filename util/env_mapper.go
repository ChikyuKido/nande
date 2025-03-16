package util

import "os"

func CheckEnvForRun() string {
	if os.Getenv("PORT") == "" {
		_ = os.Setenv("PORT", "6643")
	}
	return ""
}
