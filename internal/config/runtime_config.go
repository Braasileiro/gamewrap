package config

import (
	"fmt"
)

type RuntimeConfig struct {
	Settings     Settings
	Environment  []string
	GameMode     bool
	MangoHudMode MangoHudMode
	HideSteam    bool
	Executable   string
}

// ResolveRuntimeConfig applies profiles over inherited values, followed by command overrides.
func ResolveRuntimeConfig(profile GameProfile, environment []string, environmentOverrides map[string]string, hideSteamOverride *bool) (RuntimeConfig, []error) {
	resolvedSettings, settingErrors := settingsForGame(profile, environment, environmentOverrides)
	gameMode, gameModeError := resolveGameMode(profile, environment, environmentOverrides)
	if gameModeError != nil {
		settingErrors = append(settingErrors, gameModeError)
	}

	mangoHudMode, mangoHudError := resolveMangoHudMode(profile, environment, environmentOverrides)
	if mangoHudError != nil {
		settingErrors = append(settingErrors, mangoHudError)
	}

	if mangoHudConfiguredDirectly(profile, environment, environmentOverrides) {
		settingErrors = append(settingErrors, fmt.Errorf("MANGOHUD is managed by gamewrap, use GW_MANGOHUD instead"))
	}

	return RuntimeConfig{
		Settings:     resolvedSettings,
		Environment:  environmentForGame(profile, environment, environmentOverrides, mangoHudMode),
		GameMode:     gameMode,
		MangoHudMode: mangoHudMode,
		HideSteam:    resolveHideSteam(profile, hideSteamOverride),
		Executable:   ResolveExecutable(profile, environment, environmentOverrides),
	}, settingErrors
}

func resolveGameMode(profile GameProfile, inherited []string, overrides map[string]string) (bool, error) {
	value, configured := environmentLookup(inherited, "GW_GAMEMODE")
	if profile.Config.GameMode != nil {
		value = "0"
		if *profile.Config.GameMode {
			value = "1"
		}

		configured = true
	}

	if profileValue, found := profile.Env["GW_GAMEMODE"]; found {
		value = profileValue
		configured = true
	}

	if override, found := overrides["GW_GAMEMODE"]; found {
		value = override
		configured = true
	}

	if !configured {
		return true, nil
	}

	switch value {
	case "1":
		return true, nil
	case "0":
		return false, nil
	default:
		// Invalid input must not unexpectedly disable the default behavior.
		return true, fmt.Errorf("invalid GW_GAMEMODE value %q", value)
	}
}

// ResolveExecutable applies the executable override precedence without resolving other settings.
func ResolveExecutable(profile GameProfile, inherited []string, overrides map[string]string) string {
	executable, _ := environmentLookup(inherited, "GW_EXECUTABLE")
	if profile.Config.Executable != nil {
		executable = *profile.Config.Executable
	}

	if value, configured := profile.Env["GW_EXECUTABLE"]; configured {
		executable = value
	}

	if value, configured := overrides["GW_EXECUTABLE"]; configured {
		executable = value
	}

	return executable
}

func environmentForGame(profile GameProfile, inherited []string, overrides map[string]string, mangoHudMode MangoHudMode) []string {
	environment := EnvironmentWithOverrides(inherited, profile.Env)
	environment = EnvironmentWithOverrides(environment, overrides)

	// The resolved mode owns MangoHud activation so it cannot conflict with its wrapper strategy.
	environment = envWithout(environment, "MANGOHUD")
	switch mangoHudMode {
	case MangoHudDisabled:
		return append(environment, "MANGOHUD=0")
	case MangoHudEnv:
		return append(environment, "MANGOHUD=1")
	default:
		return environment
	}
}

func settingsForGame(profile GameProfile, inherited []string, overrides map[string]string) (Settings, []error) {
	profileSettings := make(map[string]string, 2)
	if profile.Config.VRR != nil {
		profileSettings["GW_VRR"] = string(*profile.Config.VRR)
	}

	if profile.Config.VRROutput != nil {
		profileSettings["GW_VRR_OUTPUT"] = *profile.Config.VRROutput
	}

	environment := EnvironmentWithOverrides(inherited, profileSettings)
	environment = EnvironmentWithOverrides(environment, profile.Env)
	environment = EnvironmentWithOverrides(environment, overrides)
	return ResolveSettings(ProfileConfig{}, environment)
}

func resolveMangoHudMode(profile GameProfile, inherited []string, overrides map[string]string) (MangoHudMode, error) {
	value := string(MangoHudDisabled)
	if inheritedValue, configured := environmentLookup(inherited, "GW_MANGOHUD"); configured {
		value = inheritedValue
	}

	if profile.Config.MangoHud != nil {
		value = string(*profile.Config.MangoHud)
	}

	if profileValue, configured := profile.Env["GW_MANGOHUD"]; configured {
		value = profileValue
	}

	if override, configured := overrides["GW_MANGOHUD"]; configured {
		value = override
	}

	mode, err := parseMangoHudMode(value)
	if err != nil {
		return MangoHudDisabled, fmt.Errorf("invalid GW_MANGOHUD value %q", value)
	}

	return mode, nil
}

func mangoHudConfiguredDirectly(profile GameProfile, inherited []string, overrides map[string]string) bool {
	if _, configured := environmentLookup(inherited, "MANGOHUD"); configured {
		return true
	}

	if _, configured := profile.Env["MANGOHUD"]; configured {
		return true
	}

	_, configured := overrides["MANGOHUD"]
	return configured
}

func resolveHideSteam(profile GameProfile, override *bool) bool {
	if override != nil {
		return *override
	}

	return profile.Config.HideSteam != nil && *profile.Config.HideSteam
}
