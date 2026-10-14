package process

import (
	"os"
	"strings"
)

const gameModeLibraryPreload = "/usr/$LIB/libgamemode.so.0"

var gameModeLibraries = []string{
	"/usr/lib/libgamemode.so.0",
	"/usr/lib64/libgamemode.so.0",
}

// EnvironmentWithGameModePreload exposes the GameMode library to Steam Runtime containers.
func EnvironmentWithGameModePreload(environment []string) ([]string, bool) {
	// Preload only when both architectures are available to Steam Runtime.
	for _, path := range gameModeLibraries {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			return append([]string(nil), environment...), false
		}
	}

	return environmentWithGameModePreload(environment, true)
}

func environmentWithGameModePreload(environment []string, available bool) ([]string, bool) {
	result := append([]string(nil), environment...)
	if !available {
		return result, false
	}

	for index, entry := range result {
		name, value, found := strings.Cut(entry, "=")
		if !found || name != "LD_PRELOAD" {
			continue
		}

		if preloadContains(value, gameModeLibraryPreload) {
			return result, true
		}

		if value == "" {
			result[index] = "LD_PRELOAD=" + gameModeLibraryPreload
		} else {
			result[index] = "LD_PRELOAD=" + gameModeLibraryPreload + ":" + value
		}

		return result, true
	}

	return append(result, "LD_PRELOAD="+gameModeLibraryPreload), true
}

func preloadContains(value, library string) bool {
	for _, entry := range strings.FieldsFunc(value, func(character rune) bool {
		return character == ':' || character == ' '
	}) {
		if entry == library {
			return true
		}
	}

	return false
}
