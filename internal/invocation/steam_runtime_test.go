package invocation

import "testing"

func TestSteamRuntimeTargetIndex(t *testing.T) {
	// This mirrors an Endless Sky launch through Steam Linux Runtime.
	args := []string{
		"/home/user/.steam/steamrt64/steam-launch-wrapper",
		"--",
		"/home/user/.steam/steamrt64/reaper",
		"SteamLaunch",
		"AppId=404410",
		"--",
		"/games/SteamLinuxRuntime_sniper/_v2-entry-point",
		"--verb=waitforexitandrun",
		"--",
		"/games/Endless Sky/endless-sky",
	}

	if got := gameTargetIndex(args); got != 9 {
		t.Fatalf("target index = %d, want 9", got)
	}
}

func TestSteamRuntimeTargetIndexSkipsScoutCompatibilityEntryPoint(t *testing.T) {
	args := []string{
		"/games/SteamLinuxRuntime_soldier/_v2-entry-point",
		"--verb=waitforexitandrun",
		"--",
		"/games/SteamLinuxRuntime/scout-on-soldier-entry-point-v2",
		"--",
		"/games/Terraria/Terraria",
	}

	if got := gameTargetIndex(args); got != 5 {
		t.Fatalf("target index = %d, want 5", got)
	}
}

func TestSteamReaperTargetIndexWithoutContainerRuntime(t *testing.T) {
	args := []string{
		"/home/user/.steam/steamrt64/reaper",
		"SteamLaunch",
		"AppId=123",
		"--",
		"/games/native-game",
	}

	if got := gameTargetIndex(args); got != 4 {
		t.Fatalf("target index = %d, want 4", got)
	}
}
