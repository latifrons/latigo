package program

import (
	"os"
	"path"
)

type FolderConfig struct {
	Root    string
	Log     string
	Data    string
	Config  string
	Private string
}

func mkDirPermIfNotExists(path string, perm os.FileMode) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(path, perm)
}

func ensureFolder(folder string, perm os.FileMode) {
	err := mkDirPermIfNotExists(folder, perm)
	if err != nil {
		panic(err)
	}
	return
}

func defaultPath(givenPath string, defaultRoot string, suffix string) string {
	if givenPath == "" {
		return path.Join(defaultRoot, suffix)
	}
	if path.IsAbs(givenPath) {
		return givenPath
	}
	return path.Join(defaultRoot, givenPath)
}

func EnsureFolders(config FolderConfig) FolderConfig {
	config = FolderConfig{
		Root:    config.Root,
		Log:     defaultPath(config.Log, config.Root, "log"),
		Data:    defaultPath(config.Data, config.Root, "data"),
		Config:  defaultPath(config.Config, config.Root, "config"),
		Private: defaultPath(config.Private, config.Root, "private"),
	}
	ensureFolder(config.Root, 0755)
	ensureFolder(config.Log, 0755)
	ensureFolder(config.Data, 0755)
	ensureFolder(config.Config, 0755)
	ensureFolder(config.Private, 0700)
	return config

}
