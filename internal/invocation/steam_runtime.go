package invocation

import (
	"path/filepath"
	"slices"
)

const (
	steamLaunchWrapper       = "steam-launch-wrapper"
	steamReaper              = "reaper"
	steamRuntimeEntryPoint   = "_v2-entry-point"
	scoutOnSoldierEntryPoint = "scout-on-soldier-entry-point-v2"
)

func gameTargetIndex(args []string) int {
	if index := protonTargetIndex(args); index >= 0 {
		return index
	}

	if index := steamRuntimeTargetIndex(args); index >= 0 {
		return index
	}

	if index := steamReaperTargetIndex(args); index >= 0 {
		return index
	}

	if index := steamLaunchWrapperTargetIndex(args); index >= 0 {
		return index
	}

	if len(args) > 0 {
		return 0
	}

	return -1
}

func steamRuntimeTargetIndex(args []string) int {
	for index, argument := range args {
		if filepath.Base(argument) != steamRuntimeEntryPoint {
			continue
		}

		separator := separatorIndex(args, index+1)
		if separator < 0 || !containsArgument(args[index+1:separator], "--verb=waitforexitandrun") {
			continue
		}

		target := separator + 1
		if target >= len(args) {
			return -1
		}

		// The scout compatibility entry point adds one more wrapper layer.
		if filepath.Base(args[target]) == scoutOnSoldierEntryPoint {
			separator = separatorIndex(args, target+1)
			if separator < 0 || separator+1 >= len(args) {
				return -1
			}

			target = separator + 1
		}

		return target
	}

	return -1
}

func steamReaperTargetIndex(args []string) int {
	for index, argument := range args {
		if filepath.Base(argument) != steamReaper || index+1 >= len(args) || args[index+1] != "SteamLaunch" {
			continue
		}

		separator := separatorIndex(args, index+2)
		if separator >= 0 && separator+1 < len(args) {
			return separator + 1
		}
	}

	return -1
}

func steamLaunchWrapperTargetIndex(args []string) int {
	for index, argument := range args {
		if filepath.Base(argument) != steamLaunchWrapper {
			continue
		}

		separator := separatorIndex(args, index+1)
		if separator >= 0 && separator+1 < len(args) {
			return separator + 1
		}
	}

	return -1
}

func separatorIndex(args []string, start int) int {
	for index := start; index < len(args); index++ {
		if args[index] == "--" {
			return index
		}
	}

	return -1
}

func containsArgument(args []string, expected string) bool {
	return slices.Contains(args, expected)
}
