package process

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"gamewrap/internal/config"
)

const gamemoderun = "gamemoderun"
const mangoHud = "mangohud"

func ValidateGameDependencies(gameMode bool, mangoHudMode config.MangoHudMode) error {
	var commands []string
	if gameMode {
		commands = append(commands, gamemoderun)
	}

	if mangoHudMode == config.MangoHudWrapper {
		commands = append(commands, mangoHud)
	}

	for _, name := range commands {
		if _, err := exec.LookPath(name); err != nil {
			return fmt.Errorf("required command %q was not found in PATH", name)
		}
	}

	return nil
}

func RunGame(args, environment []string, gameMode bool, mangoHudMode config.MangoHudMode, logger *log.Logger) int {
	return Run(GameCommand(args, gameMode, mangoHudMode), environment, logger)
}

func GameCommand(args []string, gameMode bool, mangoHudMode config.MangoHudMode) []string {
	commandArgs := make([]string, 0, len(args)+2)
	if gameMode {
		commandArgs = append(commandArgs, gamemoderun)
	}

	if mangoHudMode == config.MangoHudWrapper {
		commandArgs = append(commandArgs, mangoHud)
	}

	return append(commandArgs, args...)
}

// Run attaches the child to the current terminal and forwards termination signals.
func Run(args, environment []string, logger *log.Logger) int {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if environment != nil {
		cmd.Env = environment
	}

	if err := cmd.Start(); err != nil {
		logger.Printf("failed to start command: %v", err)
		fmt.Fprintf(os.Stderr, "gamewrap: failed to start command: %v\n", err)
		return 1
	}

	logger.Printf("process started: pid=%d", cmd.Process.Pid)

	signals := make(chan os.Signal, 1)
	done := make(chan error, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(signals)
	go func() { done <- cmd.Wait() }()

	for {
		select {
		case sig := <-signals:
			if err := cmd.Process.Signal(sig); err != nil {
				logger.Printf("failed to forward %s: %v", sig, err)
			}
		case err := <-done:
			return exitCode(err)
		}
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return 1
	}

	if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return 128 + int(status.Signal())
	}

	return exitErr.ExitCode()
}
