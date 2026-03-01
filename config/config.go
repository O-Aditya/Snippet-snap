package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	appName       = "snippet-snap"
	configDirName = "snippet-snap"
	dbFileName    = "snippets.db"
)

// Load reads configuration from ~/.config/snippet-snap/config.yaml.
// It sets reasonable defaults and creates the config directory if missing.
func Load(cfgFile string) error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		configDir, err := defaultConfigDir()
		if err != nil {
			return fmt.Errorf("config dir: %w", err)
		}

		// Ensure the config directory exists
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Defaults
	dbDir, _ := defaultConfigDir()
	viper.SetDefault("db_path", filepath.Join(dbDir, dbFileName))
	viper.SetDefault("editor", editorDefault())

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is okay — we use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config: %w", err)
		}
	}

	return nil
}

// DBPath returns the path to the SQLite database file.
func DBPath() string {
	return viper.GetString("db_path")
}

// Editor returns the preferred editor command.
func Editor() string {
	return viper.GetString("editor")
}

func defaultConfigDir() (string, error) {
	home, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName), nil
}

func editorDefault() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}
	return "notepad"
}
