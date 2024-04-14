package config

import (
	"os"
	"path"
	"time"

	"github.com/adrg/xdg"
)

const (
	StdDirName        = "anchor"
	StdConfigPath     = "anchor/config.yaml"
	StdStorageKey     = "storage"
	StdSyncModeKey    = "sync"
	StdHttpTimeout    = 3 * time.Second
	StdSyncMsg        = "Sync bookmarks"
	StdFileMode       = os.FileMode(0o666)
	StdLabel          = "root"
	StdLabelSeparator = "."
)

func SettingsFilePath() string {
	config, err := xdg.ConfigFile(StdConfigPath)
	if err != nil {
		panic("Cannot open config path")
	}

	return config
}

func DataDirPath() string {
	return path.Join(xdg.DataHome, StdDirName)
}
