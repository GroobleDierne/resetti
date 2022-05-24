// Package cfg provides the various configuration types used by resetti,
// along with functionality for reading and writing resetti's configuration
// file.
package cfg

import (
	"os"
	"resetti/x11"

	"gopkg.in/yaml.v2"
)

const (
	KeyReset int = 0
	KeyFocus int = 1
)

// Config contains all of the configuration for resetti.
type Config struct {
	OBS   ObsSettings   `yaml:"obs"`   // The settings to use for resetti's OBS integration.
	Keys  ConfigKeys    `yaml:"keys"`  // The hotkeys to use for resetti's actions.
	Reset ResetSettings `yaml:"reset"` // Reset settings
}

// ConfigKeys contains the user's keybindings.
type ConfigKeys struct {
	Reset x11.Key `yaml:"reset"`
	Focus x11.Key `yaml:"focus"`
}

// McSettings contains the user's preferred Minecraft settings for
// automatically adjusting them when resetting.
type McSettings struct {
	Fov         uint8 `yaml:"fov"`
	Render      uint8 `yaml:"rd"`
	Sensitivity uint8 `yaml:"sensitivity"`
}

// ObsSettings contains the user's OBS settings.
type ObsSettings struct {
	Enabled  bool   `yaml:"enabled"`
	Port     uint16 `yaml:"port"`
	Password string `yaml:"password"` // If empty, no authentication will be used.
}

// ResetSettings contains the user's settings for resetting instances.
type ResetSettings struct {
	Mc          McSettings `yaml:"mc"`           // The Minecraft settings to use.
	SetSettings bool       `yaml:"set-settings"` // Whether or not Minecraft settings should be reset automatically.
	Delay       uint16     `yaml:"delay"`        // Delay (in milliseconds) between menu switches.
}

var DefaultConfig = Config{
	ObsSettings{
		Enabled:  false,
		Port:     4440,
		Password: "password",
	},
	ConfigKeys{},
	ResetSettings{
		McSettings{
			70,
			16,
			100,
		},
		false,
		50,
	},
}

// GetConfig attempts to read the user's configuration file and return it
// in its parsed form.
func GetConfig() (*Config, error) {
	cfgPath, err := GetPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cfgPath); err != nil {
		return nil, err
	}

	// If the configuration file exists, read it.
	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Function GetPath returns the path to the user's configuration file.
func GetPath() (string, error) {
	// Get configuration path.
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		cfgDir = home + "/.config"
	}
	cfgPath := cfgDir + "/resetti.yml"
	return cfgPath, nil
}