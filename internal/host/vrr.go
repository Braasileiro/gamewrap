package host

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"gamewrap/internal/config"
)

const kscreenCtl = "kscreenctl"

type displayOutput struct {
	Name         string         `json:"name"`
	Capabilities []string       `json:"capabilities"`
	VRRPolicy    config.VRRMode `json:"vrrPolicy"`
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
	output, err := getDisplayOutput(commands, requestedOutput)
	if err != nil {
		return "", "", err
	}

	if !outputSupportsVRR(output) {
		return "", "", fmt.Errorf("output %s does not support VRR", output.Name)
	}

	// Return the resolved connector name so restoration cannot follow a changed active output.
	switch output.VRRPolicy {
	case config.VRRNever, config.VRRAlways, config.VRRAutomatic:
		return output.Name, output.VRRPolicy, nil
	default:
		return "", "", fmt.Errorf("output %s has unknown VRR policy %q", output.Name, output.VRRPolicy)
	}
}

func getDisplayOutput(commands commandRunner, requestedOutput string) (displayOutput, error) {
	output, err := commands.Run(kscreenCtl, requestedOutput, "--json")
	if err != nil {
		return displayOutput{}, err
	}

	var display displayOutput
	if err := json.Unmarshal(output, &display); err != nil {
		return displayOutput{}, err
	}

	if display.Name == "" {
		return displayOutput{}, errors.New("output information is missing its name")
	}

	return display, nil
}

func outputSupportsVRR(output displayOutput) bool {
	// kscreenctl reports a VRR policy even for outputs without the VRR capability.
	return slices.Contains(output.Capabilities, "vrr")
}

func setVRR(commands commandRunner, output string, policy config.VRRMode) error {
	_, err := commands.Run(kscreenCtl, output, "set-vrr-policy", string(policy))
	return err
}
