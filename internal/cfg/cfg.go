package cfg

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/woofdoggo/resetti/internal/x11"
)

//go:embed default.toml
var defaultConfig string

type Profile struct {
	General struct {
		ResetType   string `toml:"type"`
		CountResets bool   `toml:"count_resets"`
		CountPath   string `toml:"resets_file"`
	} `toml:"general"`
	Hooks struct {
		WallReset string `toml:"wall_reset"`
		Reset     string `toml:"reset"`
		Lock      string `toml:"lock"`
		Unlock    string `toml:"unlock"`
	} `toml:"hooks"`
	Obs struct {
		Enabled  bool   `toml:"enabled"`
		Port     uint16 `toml:"port"`
		Password string `toml:"password"`
	} `toml:"obs"`
	Reset struct {
		Delay        int  `toml:"delay"`
		UnpauseFocus bool `toml:"unpause_on_focus"`
		ClickFocus   bool `toml:"click_on_focus"`
	} `toml:"reset"`
	Keys struct {
		Focus           x11.Key    `toml:"focus"`
		Reset           x11.Key    `toml:"reset"`
		WallReset       x11.Keymod `toml:"wall_reset"`
		WallResetOthers x11.Keymod `toml:"wall_reset_others"`
		WallPlay        x11.Keymod `toml:"wall_play"`
		WallLock        x11.Keymod `toml:"wall_lock"`
	} `toml:"keybinds"`
	Wall struct {
		HideGui         bool   `toml:"hide_gui"`
		StretchWindows  bool   `toml:"stretch_windows"`
		StretchWidth    uint32 `toml:"stretch_width"`
		StretchHeight   uint32 `toml:"stretch_height"`
		UnstretchWidth  uint32 `toml:"unstretch_width"`
		UnstretchHeight uint32 `toml:"unstretch_height"`
		UseMouse        bool   `toml:"use_mouse"`
		GoToLocked      bool   `toml:"goto_locked"`
		SleepBgLock     bool   `toml:"sleepbg_lock"`
		SleepBgLockPath string `toml:"sleepbg_lock_path"`
	} `toml:"wall"`
	AdvancedWall struct {
		Affinity     bool `toml:"affinity"`
		CpusIdle     int  `toml:"affinity_idle"`
		CpusLow      int  `toml:"affinity_low"`
		CpusMid      int  `toml:"affinity_mid"`
		CpusHigh     int  `toml:"affinity_high"`
		CpusActive   int  `toml:"affinity_active"`
		LowThreshold int  `toml:"low_threshold"`
		Freeze       bool `toml:"freeze_idle"`
		FreezeDelay  int  `toml:"freeze_delay"`
		ConcResets   int  `toml:"max_concurrent_resets"`
	} `toml:"advanced_wall"`
}

// GetFolder returns the path to the user's configuration folder.
func GetFolder() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cfgDir = home + "/.config"
	}
	cfgPath := cfgDir + "/resetti/"
	return cfgPath, nil
}

// GetProfile returns a specific configuration profile.
func GetProfile(name string) (Profile, error) {
	dir, err := GetFolder()
	if err != nil {
		return Profile{}, err
	}
	file, err := os.ReadFile(dir + name + ".toml")
	if err != nil {
		return Profile{}, err
	}
	conf := Profile{}
	err = toml.Unmarshal(file, &conf)
	if err != nil {
		return Profile{}, err
	}

	// Validate configuration.
	{
		cpus := runtime.NumCPU()
		if conf.AdvancedWall.Affinity {
			idle := cpus < conf.AdvancedWall.CpusIdle
			low := cpus < conf.AdvancedWall.CpusLow
			mid := cpus < conf.AdvancedWall.CpusMid
			active := cpus < conf.AdvancedWall.CpusActive
			if idle || low || mid || active {
				return Profile{}, errors.New("too many CPUs set in advanced affinity")
			}
		}
		if conf.AdvancedWall.ConcResets > 0 && !conf.AdvancedWall.Freeze {
			return Profile{}, errors.New("instance freezing must be enabled for maximum concurrent resets")
		}
		if conf.Keys.Focus == conf.Keys.Reset {
			return Profile{}, errors.New("keybinds cannot be the same")
		}
		a := conf.Keys.WallReset
		b := conf.Keys.WallResetOthers
		c := conf.Keys.WallPlay
		d := conf.Keys.WallLock
		if a == b || a == c || a == d || b == c || b == d || c == d {
			return Profile{}, errors.New("keybinds cannot be the same")
		}
		mode := conf.General.ResetType
		if mode != "standard" && mode != "wall" {
			return Profile{}, errors.New("invalid reset type")
		}
		if mode != "standard" && !conf.Obs.Enabled {
			return Profile{}, errors.New("obs must be enabled for this reset mode")
		}
	}

	// Set SleepBackground lock path.
	if conf.Wall.SleepBgLock && conf.Wall.SleepBgLockPath == "" {
		userDir, err := os.UserHomeDir()
		if err != nil {
			return Profile{}, fmt.Errorf("failed to get user dir for sleepbg lock: %s", err)
		}
		conf.Wall.SleepBgLockPath = userDir
	}
	conf.Wall.SleepBgLockPath += "/sleepbg.lock"

	return conf, nil
}

// GetProfileList returns a list of all available configuration profiles.
func GetProfileList() ([]string, error) {
	dir, err := GetFolder()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	profiles := make([]string, 0)
	for _, v := range entries {
		if v.IsDir() || v.Name()[0] == '.' {
			continue
		}
		profiles = append(profiles, v.Name())
	}
	return profiles, nil
}

// MakeProfile makes a new profile with the default configuration.
func MakeProfile(name string) error {
	dir, err := GetFolder()
	if err != nil {
		return err
	}
	if stat, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(dir, 0644)
			if err != nil {
				return fmt.Errorf("failed to create config dir: %s", err)
			}
		} else if !stat.IsDir() {
			return errors.New("config path is not a directory")
		}
	}
	return os.WriteFile(
		dir+name+".toml",
		[]byte(defaultConfig),
		0644,
	)
}
