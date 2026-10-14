package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gamewrap/internal/config"
	"gamewrap/internal/host"
	"gamewrap/internal/invocation"
	gameprocess "gamewrap/internal/process"
)

const version = "2.0.1"

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) == 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		return 0
	}

	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("gamewrap %s\n", version)
		return 0
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "gamewrap: no game command received")
		return 1
	}

	logger, closeLogger := newLogger()
	defer closeLogger()
	options, args, err := parseInvocationOptions(os.Args[1:])
	if err != nil {
		logger.Printf("option error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	inheritedEnvironment := os.Environ()
	if options.utility != "" {
		return runUtility(options, args, inheritedEnvironment, logger)
	}

	return runGame(options, args, inheritedEnvironment, logger)
}

func runUtility(options invocationOptions, args, inheritedEnvironment []string, logger *log.Logger) int {
	expandedEnvironment, err := config.ExpandEnvironmentTokens(options.environment, "")
	if err != nil {
		logger.Printf("environment error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	options.environment = expandedEnvironment

	hideSteam := options.hideSteam != nil && *options.hideSteam
	args, _, err = invocation.HideSteamInvocation(args, hideSteam)
	if err != nil {
		logger.Printf("hide Steam error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	args, replaced, err := invocation.UtilityInvocation(args, options.utility)
	if err != nil {
		logger.Printf("utility error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	platform := invocation.ProtonLogFields(args)
	_, commandVars := environmentOverrideCounts(nil, options.environment)
	if commandVars > 0 {
		logger.Printf("environment resolved: profile_vars=0 command_vars=%d", commandVars)
	}

	if replaced == "" {
		logger.Printf("starting utility: name=%s%s hide_steam=%t", options.utility, platform, hideSteam)
	} else {
		logger.Printf("starting utility: name=%s replaced=%s%s hide_steam=%t", options.utility, replaced, platform, hideSteam)
	}

	started := time.Now()
	utilityEnvironment := config.EnvironmentWithOverrides(inheritedEnvironment, options.environment)
	code := gameprocess.Run(args, invocation.UtilityEnvironment(utilityEnvironment), logger)
	logger.Printf("utility exited: name=%s status=%d duration=%s", options.utility, code, elapsedTime(started))
	return code
}

func runGame(options invocationOptions, args, inheritedEnvironment []string, logger *log.Logger) int {
	targetName := invocation.ExecutableName(args)
	appID := invocation.SteamAppID(inheritedEnvironment)
	profile, loadedProfiles, profileErrors := config.LoadGameProfile(targetName, appID)
	if len(loadedProfiles) > 0 {
		profileNames := make([]string, 0, len(loadedProfiles))
		for _, path := range loadedProfiles {
			profileNames = append(profileNames, filepath.Base(path))
		}

		logger.Printf("profiles applied: %s", strings.Join(profileNames, ", "))
	}

	for _, err := range profileErrors {
		logger.Printf("profile: %v", err)
	}

	profileVars, commandVars := environmentOverrideCounts(profile.Env, options.environment)
	if profileVars+commandVars > 0 {
		logger.Printf("environment resolved: profile_vars=%d command_vars=%d", profileVars, commandVars)
	}

	// Resolve the executable first because @target-dir refers to the final target.
	executable := config.ResolveExecutable(profile, inheritedEnvironment, options.environment)
	args, replacedTarget, err := invocation.OverrideExecutable(args, executable)
	if err != nil {
		logger.Printf("executable override error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	targetDir := ""
	if config.EnvironmentUsesTargetDir(profile.Env) || config.EnvironmentUsesTargetDir(options.environment) {
		targetDir, err = invocation.ExecutableDirectory(args)
		if err != nil {
			logger.Printf("environment error: %v", err)
			fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
			return 1
		}
	}

	profile.Env, err = config.ExpandEnvironmentTokens(profile.Env, targetDir)
	if err != nil {
		logger.Printf("profile environment error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	options.environment, err = config.ExpandEnvironmentTokens(options.environment, targetDir)
	if err != nil {
		logger.Printf("command environment error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	runtime, settingsErrors := config.ResolveRuntimeConfig(profile, inheritedEnvironment, options.environment, options.hideSteam)
	for _, err := range settingsErrors {
		logger.Printf("settings: %v", err)
	}

	args, hiddenTarget, err := invocation.HideSteamInvocation(args, runtime.HideSteam)
	if err != nil {
		logger.Printf("hide Steam error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	if runtime.GameMode && invocation.IsProtonCommand(args) {
		var enabled bool
		runtime.Environment, enabled = gameprocess.EnvironmentWithGameModePreload(runtime.Environment)
		if enabled {
			logger.Printf("GameMode library preload enabled")
		}
	}

	if err := gameprocess.ValidateGameDependencies(runtime.GameMode, runtime.MangoHudMode); err != nil {
		logger.Printf("dependency error: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: %v\n", err)
		return 1
	}

	s := host.NewSession(logger, runtime.Settings)
	s.Prepare()
	defer s.Restore()

	gameName := invocation.ExecutableName(args)
	if runtime.HideSteam {
		gameName = hiddenTarget
	}

	replacedField := ""
	if replacedTarget != "" {
		replacedField = " replaced=" + replacedTarget
	}

	logger.Printf("starting game: name=%s%s%s hide_steam=%t", gameName, replacedField, invocation.ProtonLogFields(args), runtime.HideSteam)
	started := time.Now()
	code := gameprocess.RunGame(args, runtime.Environment, runtime.GameMode, runtime.MangoHudMode, logger)
	logger.Printf("game exited: name=%s status=%d duration=%s", gameName, code, elapsedTime(started))
	return code
}

func elapsedTime(started time.Time) time.Duration {
	return time.Since(started).Round(time.Millisecond)
}

func environmentOverrideCounts(profile, command map[string]string) (int, int) {
	profileCount := 0
	for name := range profile {
		if _, overridden := command[name]; !overridden {
			profileCount++
		}
	}

	return profileCount, len(command)
}
