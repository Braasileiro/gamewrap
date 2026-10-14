package invocation

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecutableName returns the game target, skipping Steam Runtime and Proton wrappers.
func ExecutableName(args []string) string {
	targetIndex := gameTargetIndex(args)
	if targetIndex < 0 {
		return ""
	}

	return executableBase(args[targetIndex])
}

// ExecutableDirectory returns the effective game target directory in its native path style.
func ExecutableDirectory(args []string) (string, error) {
	targetIndex := gameTargetIndex(args)
	if targetIndex < 0 {
		return "", fmt.Errorf("resolve target directory: no game command received")
	}

	target := args[targetIndex]
	if strings.Contains(target, `\`) || isWineDrivePath(target) {
		directory, _, found := executablePathDirectory(target)
		if !found {
			return "", fmt.Errorf("resolve target directory for %q", target)
		}

		return directory, nil
	}

	if !strings.Contains(target, "/") {
		resolved, err := exec.LookPath(target)
		if err != nil {
			return "", fmt.Errorf("resolve target directory for %q: %w", target, err)
		}

		target = resolved
	}

	absolute, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve target directory for %q: %w", target, err)
	}

	return filepath.Dir(absolute), nil
}

func SteamAppID(environment []string) string {
	for _, name := range []string{"SteamAppId", "SteamGameId", "STEAM_COMPAT_APP_ID"} {
		if value, found := environmentLookup(environment, name); found && onlyDigits(strings.TrimSpace(value)) {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func executableBase(path string) string {
	return filepath.Base(strings.ReplaceAll(path, `\`, "/"))
}

func environmentLookup(environment []string, name string) (string, bool) {
	for _, entry := range environment {
		key, value, found := strings.Cut(entry, "=")
		if found && key == name {
			return value, true
		}
	}

	return "", false
}

func onlyDigits(value string) bool {
	if value == "" {
		return false
	}

	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}
