// Package mc provides facilities for representing Minecraft instances
// and their state, as well as managing and resetting them.
package mc

import (
	"fmt"
	"os"
	"resetti/x11"
	"sort"
	"strconv"
	"strings"

	"github.com/jezek/xgb/xproto"
)

// InstanceState represents the state of a given instance.
type InstanceState int

const (
	StateUnknown    InstanceState = 0 // The instance's state is unknown; no actions have been performed yet.
	StatePaused     InstanceState = 1 // The instance is currently idle and paused following world generation.
	StateIngame     InstanceState = 2 // The instance is currently being played on.
	StateGenerating InstanceState = 3 // The instance is currently generating a world.
	StatePreview    InstanceState = 4 // The instance is currently on the world preview.
)

func (i InstanceState) String() string {
	// NOTE: Update switch case with the InstanceState enum.
	switch i {
	case StateUnknown:
		return "???"
	case StatePaused:
		return "Paused"
	case StateIngame:
		return "Ingame"
	case StateGenerating:
		return "Generating"
	case StatePreview:
		return "Preview"
	}
	panic("unreachable")
}

// Version represents the Minecraft version of an instance.
type Version int

const (
	VersionUnknown Version = 0  // The instance's version is not supported.
	Version1_7     Version = 7  // 1.7.x
	Version1_8     Version = 8  // 1.8.x
	Version1_15    Version = 15 // 1.15.x
	Version1_16    Version = 16 // 1.16.x
)

func (v Version) String() string {
	// NOTE: Update switch case with the Version enum.
	switch v {
	case VersionUnknown:
		return "???"
	case Version1_7:
		return "1.7"
	case Version1_8:
		return "1.8"
	case Version1_15:
		return "1.15"
	case Version1_16:
		return "1.16"
	}
	panic("unreachable")
}

// Instance contains the state and metadata of a Minecraft instance.
type Instance struct {
	Id      int // The identifier/number of the instance.
	Window  xproto.Window
	Dir     string // The instance's `.minecraft` directory.
	Pid     uint32
	State   InstanceState
	Version Version
}

// GetInstances returns a list of running Minecraft instances.
func GetInstances(x *x11.Client) ([]Instance, error) {
	windows, err := x.GetWindowList(x.Root)
	if err != nil {
		return nil, err
	}

	instances := []Instance{}

	for _, win := range windows {
		// Check if the window is a Minecraft window.
		attrs, err := x.GetWindowAttributes(win)
		if err != nil {
			continue
		}

		if !strings.Contains(attrs.Class[0], "Minecraft") {
			continue
		}

		// TODO: This could be made better. MultiMC and its forks omit
		// the --gameDir argument (I believe the vanilla launcher uses
		// it, perhaps more do?)
		//
		// It is also possible to parse the file `/proc/$pid/environ`
		// for INST_DIR, INST_MC_DIR, e.t.c. I would have to
		// investigate vanilla launcher behavior to determine the best
		// method for getting the game directory (although nobody should
		// be using the vanilla launcher, it's pretty bad...)

		// Get the path to the instance.
		argbytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", attrs.Pid))
		if err != nil {
			continue
		}

		dir := ""
		args := strings.Split(string(argbytes), "\x00")
		for _, arg := range args {
			if !strings.Contains(arg, "-Djava.library.path") {
				continue
			}

			dirsplit := strings.Split(arg, "=")
			dir = strings.ReplaceAll(dirsplit[1], "natives", ".minecraft")
			break
		}

		if dir == "" {
			continue
		}

		// Get the instance ID/number.
		var id int

		numbytes, err := os.ReadFile(fmt.Sprintf("%s/instance_num", dir))
		if err == nil {
			id, err = strconv.Atoi(strings.Trim(string(numbytes), "\n"))
			if err != nil {
				continue
			}
		} else {
			id = -1
		}

		// Get the instance version.
		verstr := strings.Split(attrs.Class[0], " ")[1]
		verstr = strings.Split(verstr, ".")[1]
		var version Version

		switch verstr {
		case "7":
			version = Version1_7
		case "8":
			version = Version1_8
		case "15":
			version = Version1_15
		case "16":
			version = Version1_16
		default:
			version = VersionUnknown
		}

		instance := Instance{
			Id:      id,
			Window:  win,
			Dir:     dir,
			Pid:     attrs.Pid,
			State:   StateUnknown,
			Version: version,
		}

		instances = append(instances, instance)
	}

	sort.Slice(instances, func(i, j int) bool {
		return instances[i].Id < instances[j].Id
	})

	return instances, nil
}
