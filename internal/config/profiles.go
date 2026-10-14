package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type ProfileConfig struct {
	Executable *string
	GameMode   *bool
	MangoHud   *MangoHudMode
	VRR        *VRRMode
	VRROutput  *string
	HideSteam  *bool
}

type GameProfile struct {
	Config ProfileConfig
	Env    map[string]string
}

type profileDocument struct {
	Config struct {
		Executable *string `toml:"executable"`
		GameMode   *bool   `toml:"gamemode"`
		MangoHud   *string `toml:"mangohud"`
		VRR        *string `toml:"vrr"`
		VRROutput  *string `toml:"vrr_output"`
		HideSteam  *bool   `toml:"hide_steam"`
	} `toml:"config"`
	Env map[string]any `toml:"env"`
}

// LoadGameProfile loads profiles next to the resolved gamewrap binary.
// Values are merged from global, executable, and Steam App ID profiles, in that order.
func LoadGameProfile(targetName, appID string) (GameProfile, []string, []error) {
	executable, err := os.Executable()
	if err != nil {
		return GameProfile{}, nil, []error{fmt.Errorf("find executable: %w", err)}
	}

	if resolved, resolveErr := filepath.EvalSymlinks(executable); resolveErr == nil {
		executable = resolved
	}

	dir := filepath.Join(filepath.Dir(executable), "profiles")
	return LoadGameProfileFrom(dir, targetName, appID)
}

// LoadGameProfileFrom loads and merges profiles from an explicit directory.
func LoadGameProfileFrom(dir, targetName, appID string) (GameProfile, []string, []error) {
	profile := GameProfile{Env: make(map[string]string)}
	names := []string{"global.toml"}
	if targetName != "" {
		names = append(names, targetName+".toml")
	}

	if appID != "" {
		names = append(names, appID+".toml")
	}

	var loaded []string
	var profileErrors []error
	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			continue
		}

		seen[name] = true

		path := filepath.Join(dir, name)
		loadedProfile, err := readProfile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				profileErrors = append(profileErrors, fmt.Errorf("%s: %w", path, err))
			}

			continue
		}

		mergeProfile(&profile, loadedProfile)
		loaded = append(loaded, path)
	}

	return profile, loaded, profileErrors
}

func readProfile(path string) (GameProfile, error) {
	var document profileDocument
	metadata, err := toml.DecodeFile(path, &document)
	if err != nil {
		return GameProfile{}, err
	}

	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return GameProfile{}, fmt.Errorf("unknown key %q", undecoded[0].String())
	}

	profile := GameProfile{
		Config: ProfileConfig{
			Executable: document.Config.Executable,
			GameMode:   document.Config.GameMode,
			VRROutput:  document.Config.VRROutput,
			HideSteam:  document.Config.HideSteam,
		},
		Env: make(map[string]string, len(document.Env)),
	}
	if profile.Config.Executable != nil {
		executable := *profile.Config.Executable
		if strings.ContainsRune(executable, '\x00') {
			return GameProfile{}, fmt.Errorf("executable contains a null byte")
		}

		profile.Config.Executable = &executable
	}

	if document.Config.MangoHud != nil {
		mode, err := parseMangoHudMode(*document.Config.MangoHud)
		if err != nil {
			return GameProfile{}, err
		}

		profile.Config.MangoHud = &mode
	}

	if document.Config.VRR != nil {
		mode, err := parseVRRMode(*document.Config.VRR)
		if err != nil {
			return GameProfile{}, err
		}

		profile.Config.VRR = &mode
	}

	for name, value := range document.Env {
		if name == "" || strings.ContainsAny(name, "=\x00") {
			return GameProfile{}, fmt.Errorf("invalid environment variable name %q", name)
		}

		converted, err := environmentValue(value)
		if err != nil {
			return GameProfile{}, fmt.Errorf("environment variable %s: %w", name, err)
		}

		profile.Env[name] = converted
	}

	return profile, nil
}

func environmentValue(value any) (string, error) {
	var converted string
	switch value := value.(type) {
	case string:
		converted = value
	case int64:
		converted = strconv.FormatInt(value, 10)
	case float64:
		converted = strconv.FormatFloat(value, 'g', -1, 64)
	case bool:
		converted = strconv.FormatBool(value)
	default:
		return "", fmt.Errorf("unsupported TOML value type %T", value)
	}

	if strings.ContainsRune(converted, '\x00') {
		return "", fmt.Errorf("value contains a null byte")
	}

	return converted, nil
}

func mergeProfile(target *GameProfile, source GameProfile) {
	if source.Config.Executable != nil {
		target.Config.Executable = source.Config.Executable
	}

	if source.Config.GameMode != nil {
		target.Config.GameMode = source.Config.GameMode
	}

	if source.Config.MangoHud != nil {
		target.Config.MangoHud = source.Config.MangoHud
	}

	if source.Config.VRR != nil {
		target.Config.VRR = source.Config.VRR
	}

	if source.Config.VRROutput != nil {
		target.Config.VRROutput = source.Config.VRROutput
	}

	if source.Config.HideSteam != nil {
		target.Config.HideSteam = source.Config.HideSteam
	}

	maps.Copy(target.Env, source.Env)
}
