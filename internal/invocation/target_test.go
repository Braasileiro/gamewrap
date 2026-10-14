package invocation

import "testing"

func TestExecutableNameFindsGameTarget(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "Proton",
			args: []string{"/runtime/_v2-entry-point", "/compat/Proton/proton", "waitforexitandrun", `C:\Games\game.exe`},
			want: "game.exe",
		},
		{
			name: "native",
			args: []string{"/games/native-game", "argument.exe"},
			want: "native-game",
		},
		{
			name: "Steam Linux Runtime",
			args: []string{
				"/home/user/.steam/steamrt64/steam-launch-wrapper",
				"--",
				"/games/SteamLinuxRuntime_sniper/_v2-entry-point",
				"--verb=waitforexitandrun",
				"--",
				"/games/Endless Sky/endless-sky",
			},
			want: "endless-sky",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ExecutableName(test.args); got != test.want {
				t.Fatalf("executable name = %q, want %q", got, test.want)
			}
		})
	}
}

func TestExecutableDirectoryPreservesTargetPathStyle(t *testing.T) {
	tests := []struct {
		target string
		want   string
	}{
		{target: "/games/Example/game", want: "/games/Example"},
		{target: `C:\Games\Example\Game.exe`, want: `C:\Games\Example`},
		{target: `C:/Games/Example/Game.exe`, want: `C:/Games/Example`},
	}

	for _, test := range tests {
		got, err := ExecutableDirectory([]string{test.target})
		if err != nil {
			t.Fatal(err)
		}

		if got != test.want {
			t.Fatalf("directory for %q = %q, want %q", test.target, got, test.want)
		}
	}
}
