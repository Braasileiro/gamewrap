package invocation

import (
	"reflect"
	"testing"
)

func TestOverrideExecutableResolvesProtonTargetNextToOriginal(t *testing.T) {
	args := []string{
		"/runtime/_v2-entry-point",
		"/compat/Proton/proton",
		"waitforexitandrun",
		`C:\Games\Original.exe`,
		"--original-argument",
	}

	got, replaced, err := OverrideExecutable(args, "Game.exe")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"/runtime/_v2-entry-point",
		"/compat/Proton/proton",
		"waitforexitandrun",
		`C:\Games\Game.exe`,
	}

	if !reflect.DeepEqual(got, want) || replaced != "Original.exe" {
		t.Fatalf("override = %v, replaced = %q", got, replaced)
	}
}

func TestOverrideExecutablePreservesSteamLinuxRuntime(t *testing.T) {
	args := []string{
		"/home/user/.steam/steamrt64/steam-launch-wrapper",
		"--",
		"/games/SteamLinuxRuntime_sniper/_v2-entry-point",
		"--verb=waitforexitandrun",
		"--",
		"/games/Original/original",
		"--original-argument",
	}

	got, replaced, err := OverrideExecutable(args, "replacement")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"/home/user/.steam/steamrt64/steam-launch-wrapper",
		"--",
		"/games/SteamLinuxRuntime_sniper/_v2-entry-point",
		"--verb=waitforexitandrun",
		"--",
		"/games/Original/replacement",
	}

	if !reflect.DeepEqual(got, want) || replaced != "original" {
		t.Fatalf("override = %v, replaced = %q", got, replaced)
	}
}

func TestOverrideExecutableReplacesNativeCommand(t *testing.T) {
	got, replaced, err := OverrideExecutable([]string{"/games/original", "--argument"}, "game")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, []string{"/games/game"}) || replaced != "original" {
		t.Fatalf("override = %v, replaced = %q", got, replaced)
	}
}
