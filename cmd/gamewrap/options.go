package main

import (
	"errors"
	"fmt"
	"strings"
)

type invocationOptions struct {
	hideSteam   *bool
	utility     string
	environment map[string]string
}

func parseInvocationOptions(args []string) (invocationOptions, []string, error) {
	options := invocationOptions{environment: make(map[string]string)}

	for len(args) > 0 {
		switch args[0] {
		case "--":
			args = args[1:]
			if len(args) == 0 {
				return options, nil, errors.New("no game command received after --")
			}

			return options, args, nil
		case "--hide-steam", "--no-hide-steam":
			if options.hideSteam != nil {
				return options, nil, errors.New("Steam visibility option specified more than once")
			}

			hideSteam := args[0] == "--hide-steam"
			options.hideSteam = &hideSteam
		case "--winecfg", "--regedit":
			utility := args[0][2:]
			if options.utility != "" {
				return options, nil, fmt.Errorf("cannot combine --%s with --%s", options.utility, utility)
			}

			options.utility = utility
		case "--env":
			if len(args) < 2 {
				return options, nil, errors.New("--env requires NAME=VALUE")
			}

			name, value, err := parseEnvironmentAssignment(args[1])
			if err != nil {
				return options, nil, err
			}

			options.environment[name] = value
			args = args[2:]
			continue
		default:
			if strings.HasPrefix(args[0], "-") {
				return options, nil, fmt.Errorf("unknown option %q", args[0])
			}

			return options, nil, errors.New("game command must follow --")
		}

		args = args[1:]
	}

	if options.utility == "" {
		return options, nil, errors.New("no game command received")
	}

	return options, args, nil
}

func parseEnvironmentAssignment(assignment string) (string, string, error) {
	name, value, found := strings.Cut(assignment, "=")
	if !found {
		return "", "", errors.New("--env requires NAME=VALUE")
	}

	if name == "" || strings.ContainsRune(name, '\x00') {
		return "", "", errors.New("--env has an invalid variable name")
	}

	if strings.ContainsRune(value, '\x00') {
		return "", "", fmt.Errorf("--env value for %s contains a null byte", name)
	}

	return name, value, nil
}
