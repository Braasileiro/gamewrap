package invocation

import "errors"

const environmentCommand = "/usr/bin/env"

var hiddenSteamVariables = []string{
	"SteamAppId",
	"SteamGameId",
	"SteamOverlayGameId",
	"SteamClientLaunch",
	"SteamEnv",
	"SteamPath",
	"SteamUser",
	"SteamAppUser",
}

// HideSteamInvocation unsets Steam variables immediately before Proton starts.
// The preceding Steam Runtime command keeps its original environment.
func HideSteamInvocation(args []string, enabled bool) ([]string, string, error) {
	if !enabled {
		return args, "", nil
	}

	if len(args) == 0 {
		return nil, "", errors.New("cannot hide Steam: expected a Proton %command%")
	}

	protonIndex := protonCommandIndex(args)
	if protonIndex < 0 {
		return nil, "", errors.New("cannot hide Steam: expected a Proton %command%")
	}

	target := ExecutableName(args[protonIndex+2:])
	result := make([]string, 0, len(args)+len(hiddenSteamVariables)*2+2)
	result = append(result, args[:protonIndex]...)
	result = append(result, environmentCommand)
	for _, name := range hiddenSteamVariables {
		result = append(result, "-u", name)
	}

	result = append(result, "--")
	result = append(result, args[protonIndex:]...)
	return result, target, nil
}
