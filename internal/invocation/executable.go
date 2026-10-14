package invocation

import (
	"errors"
	"strings"
)

// OverrideExecutable replaces the original target and drops its arguments.
// A bare filename is resolved next to the original target.
func OverrideExecutable(args []string, executable string) ([]string, string, error) {
	if executable == "" {
		return args, "", nil
	}

	if len(args) == 0 {
		return nil, "", errors.New("cannot override executable: no game command received")
	}

	targetIndex := gameTargetIndex(args)
	replaced := executableBase(args[targetIndex])
	executable = siblingExecutable(args[targetIndex], executable)
	result := append([]string(nil), args[:targetIndex]...)
	return append(result, executable), replaced, nil
}

func siblingExecutable(original, executable string) string {
	if strings.ContainsAny(executable, `/\`) {
		return executable
	}

	directory, separator, found := executablePathDirectory(original)
	if !found {
		return executable
	}

	if strings.HasSuffix(directory, separator) {
		return directory + executable
	}

	return directory + separator + executable
}

// executablePathDirectory splits native and Wine target paths without host-specific conversion.
func executablePathDirectory(path string) (string, string, bool) {
	separatorIndex := strings.LastIndexAny(path, `/\`)
	if separatorIndex < 0 {
		return "", "", false
	}

	directoryEnd := separatorIndex
	if separatorIndex == 0 || (separatorIndex == 2 && len(path) >= 3 && path[1] == ':') {
		directoryEnd++
	}

	return path[:directoryEnd], path[separatorIndex : separatorIndex+1], true
}

func isWineDrivePath(path string) bool {
	return len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/')
}
