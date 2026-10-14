package main

import (
	"reflect"
	"strings"
	"testing"

	"gamewrap/internal/config"
	"gamewrap/internal/invocation"
	gameprocess "gamewrap/internal/process"
)

func TestBuildProtonCommandFromProfile(t *testing.T) {
	mangoHud := config.MangoHudWrapper
	gameMode := true
	vrr := config.VRRAutomatic
	hideSteam := true
	profile := config.GameProfile{
		Config: config.ProfileConfig{
			GameMode:  &gameMode,
			MangoHud:  &mangoHud,
			VRR:       &vrr,
			HideSteam: &hideSteam,
		},
		Env: map[string]string{"PROFILE_VALUE": "enabled"},
	}

	protonArgs := []string{
		"/runtime/_v2-entry-point",
		"/compat/Proton/proton",
		"waitforexitandrun",
		`C:\Games\game.exe`,
	}

	runtime, resolveErrors := config.ResolveRuntimeConfig(
		profile,
		[]string{"GW_VRR=automatic"},
		map[string]string{"GW_VRR": "never"},
		nil,
	)
	if len(resolveErrors) != 0 {
		t.Fatalf("unexpected resolution errors: %v", resolveErrors)
	}

	hiddenArgs, target, err := invocation.HideSteamInvocation(protonArgs, runtime.HideSteam)
	if err != nil {
		t.Fatal(err)
	}

	command := gameprocess.GameCommand(hiddenArgs, runtime.GameMode, runtime.MangoHudMode)
	wantPrefix := []string{"gamemoderun", "mangohud", "/runtime/_v2-entry-point", "/usr/bin/env"}

	if runtime.Settings.VRR != "never" || target != "game.exe" {
		t.Fatalf("VRR = %q, target = %q", runtime.Settings.VRR, target)
	}

	if value, found := testEnvironmentLookup(runtime.Environment, "PROFILE_VALUE"); !found || value != "enabled" {
		t.Fatalf("runtime environment = %v", runtime.Environment)
	}

	if len(command) < len(wantPrefix) || !reflect.DeepEqual(command[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("command prefix = %v, want %v", command, wantPrefix)
	}
}

func testEnvironmentLookup(environment []string, name string) (string, bool) {
	for _, entry := range environment {
		key, value, found := strings.Cut(entry, "=")
		if found && key == name {
			return value, true
		}
	}

	return "", false
}
