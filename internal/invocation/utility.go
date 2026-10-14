package invocation

import (
	"errors"
	"fmt"
	"strings"
)

const (
	winecfg = "winecfg"
	regedit = "regedit"
)

func UtilityInvocation(args []string, name string) ([]string, string, error) {
	var path string
	switch name {
	case "":
		return args, "", nil
	case "winecfg":
		path = winecfg
	case "regedit":
		path = regedit
	default:
		return nil, "", fmt.Errorf("unsupported utility %q", name)
	}

	if len(args) == 0 {
		return []string{path}, "", nil
	}

	protonArgs, replaced, err := replaceProtonTarget(args, name)
	if err != nil {
		return nil, "", fmt.Errorf("--%s: %w", name, err)
	}

	return protonArgs, replaced, nil
}

func replaceProtonTarget(args []string, target string) ([]string, string, error) {
	i := protonTargetIndex(args)
	if i < 0 {
		return nil, "", errors.New("expected a Proton %command% after the option")
	}

	replaced := ExecutableName(args[i:])
	result := append([]string(nil), args[:i]...)
	return append(result, target), replaced, nil
}

// UtilityEnvironment removes game-only settings and explicitly disables MangoHud.
func UtilityEnvironment(environment []string) []string {
	filtered := environmentWithout(
		environment,
		"GW_GAMEMODE",
		"GW_MANGOHUD",
		"GW_VRR",
		"GW_VRR_OUTPUT",
		"MANGOHUD",
		"LD_PRELOAD",
		"VK_INSTANCE_LAYERS",
		"VK_LOADER_LAYERS_ENABLE",
	)

	// Remove current and future MangoHud-specific configuration variables.
	withoutMangoHudOptions := filtered[:0]
	for _, entry := range filtered {
		if !strings.HasPrefix(entry, "MANGOHUD_") {
			withoutMangoHudOptions = append(withoutMangoHudOptions, entry)
		}
	}

	return append(withoutMangoHudOptions, "MANGOHUD=0")
}

func environmentWithout(environment []string, names ...string) []string {
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		remove := false
		for _, name := range names {
			if strings.HasPrefix(entry, name+"=") {
				remove = true
				break
			}
		}

		if !remove {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
