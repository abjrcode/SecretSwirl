package app

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func IsWailsRunningAppToGenerateBindings(osArgs []string) bool {
	if len(osArgs) == 0 {
		return false
	}

	binaryName := filepath.Base(osArgs[0])

	return binaryName == "wailsbindings"
}

func GetAppDataDir(pwd string, isDebugBuild bool) (string, error) {
	if isDebugBuild {
		if strings.Contains(strings.ToLower(pwd), "swervo.app") {
			pwd = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(pwd))))
		} else {
			pwd = filepath.Dir(pwd)
		}
		return filepath.Join(pwd, "swervo_data"), nil

	} else {
		if userHomeDir, err := os.UserHomeDir(); err != nil {
			return "", err
		} else {
			return filepath.Join(userHomeDir, ".swervo_data"), nil
		}
	}
}

// InitializeAppDataDir creates the app data directory if it does not exist
// and exits the program if it cannot be created
func InitializeAppDataDir(appDataDir string) {
	if err := os.MkdirAll(appDataDir, 0700); err != nil {
		if err != nil {
			log.Fatal("Could not create app data directory")
		}
	}
}
