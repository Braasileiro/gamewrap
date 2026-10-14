package host

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type commandRunner interface {
	Run(name string, args ...string) ([]byte, error)
}

type systemCommandRunner struct{}

func (systemCommandRunner) Run(name string, args ...string) ([]byte, error) {
	output, err := hostCommand(name, args...).CombinedOutput()
	if err != nil {
		return output, commandError(err, output)
	}

	return output, nil
}

func hostCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	// Host-side helpers should not inherit game-specific preload libraries.
	// They may have the wrong architecture and are not needed by these commands.
	cmd.Env = environmentWithout(os.Environ(), "LD_PRELOAD")
	return cmd
}

func environmentWithout(environment []string, name string) []string {
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		if !strings.HasPrefix(entry, name+"=") {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func commandError(err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return err
	}

	return fmt.Errorf("%w: %s", err, detail)
}
