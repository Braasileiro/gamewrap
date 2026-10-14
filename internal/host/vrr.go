package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"gamewrap/internal/config"
)

const kscreenDoctor = "kscreen-doctor"

// KScreen exposes VRR policies as numeric values in its JSON output.
const (
	kscreenVRRNever     = 0
	kscreenVRRAlways    = 1
	kscreenVRRAutomatic = 2
)

type displayOutput struct {
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Enabled   bool   `json:"enabled"`
	VRRPolicy *int   `json:"vrrPolicy"`
}

func (s *Session) enableVRR() {
	outputName, original, err := getVRR(s.commandRunner(), s.settings.VRROutput)
	if err != nil {
		s.log.Printf("VRR: failed to select or read output %s: %v", s.settings.VRROutput, err)
		return
	}

	if original == s.settings.VRR {
		s.log.Printf("VRR: %s already set on %s", s.settings.VRR, outputName)
		return
	}

	if err := setVRR(s.commandRunner(), outputName, s.settings.VRR); err != nil {
		s.log.Printf("VRR: failed to set %s on %s: %v", s.settings.VRR, outputName, err)
		return
	}

	s.vrrOutput = outputName
	s.vrrOriginal = original
	s.vrrChanged = true
	s.log.Printf("VRR: changed %s to %s on %s", original, s.settings.VRR, outputName)
}

func getVRR(commands commandRunner, requestedOutput string) (string, config.VRRMode, error) {
	outputs, err := getDisplayOutputs(commands)
	if err != nil {
		return "", "", err
	}

	output, err := selectVRROutput(outputs, requestedOutput)
	if err != nil {
		return "", "", err
	}

	policy, err := vrrPolicyName(output.VRRPolicy)
	if err != nil {
		return "", "", fmt.Errorf("output %s: %w", output.Name, err)
	}

	return output.Name, policy, nil
}

func getDisplayOutputs(commands commandRunner) ([]displayOutput, error) {
	var displayConfig struct {
		Outputs []displayOutput `json:"outputs"`
	}

	output, err := commands.Run(kscreenDoctor, "--json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(output, &displayConfig); err != nil {
		return nil, err
	}

	return displayConfig.Outputs, nil
}

// selectVRROutput detects an output only when exactly one active output supports VRR.
func selectVRROutput(outputs []displayOutput, requestedOutput string) (displayOutput, error) {
	if requestedOutput != config.VRROutputDetect {
		for _, output := range outputs {
			if output.Name == requestedOutput {
				if output.VRRPolicy == nil {
					return displayOutput{}, errors.New("output does not support VRR")
				}

				return output, nil
			}
		}

		return displayOutput{}, fmt.Errorf("output %q not found", requestedOutput)
	}

	var candidates []displayOutput
	for _, output := range outputs {
		if output.Connected && output.Enabled && output.VRRPolicy != nil {
			candidates = append(candidates, output)
		}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	if len(candidates) == 0 {
		return displayOutput{}, errors.New("no active output with VRR support found")
	}

	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.Name)
	}

	sort.Strings(names)
	return displayOutput{}, fmt.Errorf("multiple active outputs support VRR (%s); set vrr_output explicitly", strings.Join(names, ", "))
}

func vrrPolicyName(policy *int) (config.VRRMode, error) {
	if policy == nil {
		return "", errors.New("output does not support VRR")
	}

	switch *policy {
	case kscreenVRRNever:
		return config.VRRNever, nil
	case kscreenVRRAlways:
		return config.VRRAlways, nil
	case kscreenVRRAutomatic:
		return config.VRRAutomatic, nil
	default:
		return "", fmt.Errorf("unknown VRR policy %d", *policy)
	}
}

func setVRR(commands commandRunner, output string, policy config.VRRMode) error {
	_, err := commands.Run(kscreenDoctor, fmt.Sprintf("output.%s.vrrpolicy.%s", output, policy))
	return err
}
