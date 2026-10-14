package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadGameProfilePrecedence(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "global.toml", `[config]
mangohud = "disabled"
gamemode = true
vrr = "never"
vrr_output = "DP-3"

[env]
SHARED = "global"
GLOBAL_ONLY = true
`)
	writeProfile(t, dir, "game.exe.toml", `[config]
mangohud = "wrapper"
gamemode = false

[env]
SHARED = "executable"
`)
	writeProfile(t, dir, "3259600.toml", `[config]
executable = "game-alt.exe"
vrr = "automatic"

[env]
SHARED = "app-id"
`)

	profile, loaded, profileErrors := LoadGameProfileFrom(dir, "game.exe", "3259600")
	if len(profileErrors) != 0 {
		t.Fatalf("unexpected profile errors: %v", profileErrors)
	}

	wantLoaded := []string{
		filepath.Join(dir, "global.toml"),
		filepath.Join(dir, "game.exe.toml"),
		filepath.Join(dir, "3259600.toml"),
	}

	if !reflect.DeepEqual(loaded, wantLoaded) {
		t.Fatalf("loaded profiles = %v, want %v", loaded, wantLoaded)
	}

	if profile.Config.MangoHud == nil || *profile.Config.MangoHud != MangoHudWrapper ||
		profile.Config.GameMode == nil || *profile.Config.GameMode ||
		profile.Config.VRR == nil || *profile.Config.VRR != VRRAutomatic ||
		profile.Config.VRROutput == nil || *profile.Config.VRROutput != "DP-3" ||
		profile.Config.Executable == nil || *profile.Config.Executable != "game-alt.exe" {
		t.Fatalf("merged config = %+v", profile.Config)
	}

	if profile.Env["SHARED"] != "app-id" || profile.Env["GLOBAL_ONLY"] != "true" {
		t.Fatalf("merged environment = %v", profile.Env)
	}
}

func TestReadProfileRejectsUnknownConfig(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "invalid.toml", "[config]\nunknown = true\n")

	if _, err := readProfile(filepath.Join(dir, "invalid.toml")); err == nil {
		t.Fatal("expected unknown key error")
	}
}

func writeProfile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
}
