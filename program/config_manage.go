package program

import (
	"fmt"
	"github.com/latifrons/commongo/files"
	"github.com/latifrons/commongo/format"
	"github.com/latifrons/commongo/utilfuncs"
	"github.com/spf13/viper"
	"os"
	"path"
	"path/filepath"
)

func ReadNormalConfig(configFolder string) {
	configPath := path.Join(configFolder, "config.toml")

	if files.FileExists(configPath) {
		MergeLocalConfig(configPath)
	}
}

func ReadEnvConfig(envPrefix string) {
	// env override
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()
}

func ReadPrivate(privateFolder string) {
	configPath := path.Join(privateFolder, "private.toml")
	if files.FileExists(configPath) {
		MergeLocalConfig(configPath)
	}
	configOverridePath := path.Join(privateFolder, "override.toml")
	if files.FileExists(configOverridePath) {
		MergeLocalConfig(configOverridePath)
	}
}

//func writeConfig() {
//	configPath := files.FixPrefixPath(viper.GetString("rootdir"), path.Join(ConfigDir, "config_dump.toml"))
//	err := viper.WriteConfigAs(configPath)
//	utilfuncs.PanicIfError(err, "dump config")
//}

func MergeLocalConfig(configPath string) {
	absPath, err := filepath.Abs(configPath)
	utilfuncs.PanicIfError(err, fmt.Sprintf("Error on parsing config file path: %s", absPath))

	file, err := os.Open(absPath)
	utilfuncs.PanicIfError(err, fmt.Sprintf("Error on opening config file: %s", absPath))
	defer file.Close()

	viper.SetConfigType("toml")
	err = viper.MergeConfig(file)
	utilfuncs.PanicIfError(err, fmt.Sprintf("Error on reading config file: %s", absPath))
	return
}

func DumpConfig() {
	// print running config in console.
	b, err := format.PrettyJson(viper.AllSettings())
	utilfuncs.PanicIfError(err, "dump json")
	fmt.Println(b)
}

func LoadConfigs(folderConfig FolderConfig, envPrefix string) (folderConfigActual FolderConfig) {
	// init logger first.
	folderConfigActual = EnsureFolders(folderConfig)

	ReadNormalConfig(folderConfigActual.Config)
	ReadPrivate(folderConfigActual.Private)
	ReadEnvConfig(envPrefix)
	DumpConfig()
	return
}
