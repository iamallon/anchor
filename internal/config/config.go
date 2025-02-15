package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
)

const (
	StdDirName        = "anchor"
	StdStorageKey     = "storage"
	StdSyncModeKey    = "sync"
	StdHttpTimeout    = 3 * time.Second
	StdSyncMsg        = "Sync bookmarks"
	StdFileMode       = os.FileMode(0o666)
	StdLabel          = "root"
	StdLabelSeparator = "."
)

func SettingsFilePath() string {
	config, err := xdg.ConfigFile(filepath.Join(StdDirName, "config.yaml"))
	if err != nil {
		panic("Cannot open config path")
	}

	return config
}

func DataDirPath() string {
	return filepath.Join(xdg.DataHome, StdDirName, "data")
}

func ArchiveDirPath() string {
	return filepath.Join(xdg.DataHome, StdDirName, "archive")
}
