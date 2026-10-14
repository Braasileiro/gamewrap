package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	homeToken      = "@home"
	targetDirToken = "@target-dir"
)

// EnvironmentUsesTargetDir reports whether a configured value needs the game directory.
func EnvironmentUsesTargetDir(values map[string]string) bool {
	for _, value := range values {
		if value == targetDirToken || strings.HasPrefix(value, targetDirToken+"/") {
			return true
		}
	}

	return false
}

// ExpandEnvironmentTokens expands gamewrap path tokens in configured environment values.
func ExpandEnvironmentTokens(values map[string]string, targetDir string) (map[string]string, error) {
	expanded := make(map[string]string, len(values))
	for name, value := range values {
		if name == "GW_EXECUTABLE" && hasEnvironmentToken(value) {
			return nil, fmt.Errorf("environment variable %s does not support path tokens", name)
		}

		resolved, err := expandEnvironmentValue(value, targetDir)
		if err != nil {
			return nil, fmt.Errorf("environment variable %s: %w", name, err)
		}

		expanded[name] = resolved
	}

	return expanded, nil
}

func hasEnvironmentToken(value string) bool {
	_, home := tokenSuffix(value, homeToken)
	_, target := tokenSuffix(value, targetDirToken)
	return home || target
}

func expandEnvironmentValue(value, targetDir string) (string, error) {
	if suffix, matched := tokenSuffix(value, homeToken); matched {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}

		return joinTokenPath(home, suffix), nil
	}

	if suffix, matched := tokenSuffix(value, targetDirToken); matched {
		if targetDir == "" {
			return "", fmt.Errorf("%s requires a game target", targetDirToken)
		}

		return joinTokenPath(targetDir, suffix), nil
	}

	return value, nil
}

func tokenSuffix(value, token string) (string, bool) {
	if value == token {
		return "", true
	}

	if strings.HasPrefix(value, token+"/") {
		return value[len(token)+1:], true
	}

	return "", false
}

func joinTokenPath(base, suffix string) string {
	if suffix == "" {
		return base
	}

	// Keep Wine and Proton target paths in their original path style.
	if strings.Contains(base, `\`) && !strings.Contains(base, "/") {
		return strings.TrimRight(base, `\`) + `\` + strings.ReplaceAll(suffix, "/", `\`)
	}

	return filepath.Join(base, suffix)
}

func envWithout(environment []string, names ...string) []string {
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

func environmentLookup(environment []string, name string) (string, bool) {
	for _, entry := range environment {
		key, value, found := strings.Cut(entry, "=")
		if found && key == name {
			return value, true
		}
	}

	return "", false
}

// EnvironmentWithOverrides replaces inherited values and appends new ones deterministically.
func EnvironmentWithOverrides(environment []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return append([]string(nil), environment...)
	}

	result := make([]string, 0, len(environment)+len(overrides))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if _, replaced := overrides[name]; !replaced {
			result = append(result, entry)
		}
	}

	names := make([]string, 0, len(overrides))
	for name := range overrides {
		names = append(names, name)
	}

	sort.Strings(names)
	for _, name := range names {
		result = append(result, name+"="+overrides[name])
	}

	return result
}
