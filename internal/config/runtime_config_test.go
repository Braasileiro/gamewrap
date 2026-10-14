package config

import "testing"

func TestResolveRuntimeConfigPrecedence(t *testing.T) {
	gameMode := false
	mangoHud := MangoHudWrapper
	vrr := VRRAutomatic
	hideSteam := true
	executable := "profile-game"
	profile := GameProfile{
		Config: ProfileConfig{
			GameMode:   &gameMode,
			MangoHud:   &mangoHud,
			VRR:        &vrr,
			HideSteam:  &hideSteam,
			Executable: &executable,
		},
		Env: map[string]string{"PROFILE_VALUE": "enabled"},
	}

	hideSteamOverride := false

	resolved, resolveErrors := ResolveRuntimeConfig(
		profile,
		[]string{"GW_GAMEMODE=1", "GW_MANGOHUD=env", "GW_VRR=always", "GW_EXECUTABLE=inherited-game"},
		map[string]string{"GW_GAMEMODE": "1", "GW_MANGOHUD": "disabled", "GW_VRR": "never", "GW_EXECUTABLE": "command-game"},
		&hideSteamOverride,
	)
	if len(resolveErrors) != 0 {
		t.Fatalf("unexpected resolution errors: %v", resolveErrors)
	}

	if !resolved.GameMode || resolved.Settings.VRR != VRRNever || resolved.HideSteam ||
		resolved.Executable != "command-game" || resolved.MangoHudMode != MangoHudDisabled {
		t.Fatalf("resolved config = %+v", resolved)
	}

	if value, found := environmentLookup(resolved.Environment, "MANGOHUD"); !found || value != "0" {
		t.Fatalf("resolved environment = %v", resolved.Environment)
	}

	if value, found := environmentLookup(resolved.Environment, "PROFILE_VALUE"); !found || value != "enabled" {
		t.Fatalf("resolved environment = %v", resolved.Environment)
	}
}

func TestUnchangedVRRModeOverridesInheritedPolicy(t *testing.T) {
	vrr := VRRUnchanged
	profile := GameProfile{Config: ProfileConfig{VRR: &vrr}}

	resolved, resolveErrors := ResolveRuntimeConfig(profile, []string{"GW_VRR=always"}, nil, nil)
	if len(resolveErrors) != 0 || resolved.Settings.VRR != VRRUnchanged {
		t.Fatalf("VRR = %q, errors = %v", resolved.Settings.VRR, resolveErrors)
	}
}

func TestInvalidGameModeKeepsDefaultEnabled(t *testing.T) {
	resolved, resolveErrors := ResolveRuntimeConfig(GameProfile{}, []string{"GW_GAMEMODE=invalid"}, nil, nil)
	if len(resolveErrors) != 1 || !resolved.GameMode {
		t.Fatalf("GameMode = %t, errors = %v", resolved.GameMode, resolveErrors)
	}
}

func TestMangoHudWrapperReplacesDirectVariable(t *testing.T) {
	resolved, resolveErrors := ResolveRuntimeConfig(
		GameProfile{},
		[]string{"MANGOHUD=1"},
		map[string]string{"GW_MANGOHUD": "wrapper", "MANGOHUD": "0"},
		nil,
	)

	if resolved.MangoHudMode != MangoHudWrapper {
		t.Fatalf("MangoHud mode = %q", resolved.MangoHudMode)
	}

	if _, found := environmentLookup(resolved.Environment, "MANGOHUD"); found {
		t.Fatalf("resolved environment = %v", resolved.Environment)
	}

	if len(resolveErrors) != 1 {
		t.Fatalf("resolution errors = %v", resolveErrors)
	}
}
