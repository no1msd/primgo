package settings

import "os"

func localConfigDir() string {
	return os.Getenv("AppData")
}
