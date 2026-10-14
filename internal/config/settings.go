package config

import (
	"fmt"
	"strings"
)

type VRRMode string

const (
	VRRUnchanged    VRRMode = ""
	VRRAutomatic    VRRMode = "automatic"
	VRRNever        VRRMode = "never"
	VRRAlways       VRRMode = "always"
	VRROutputDetect         = "detect"
)

type Settings struct {
	VRR       VRRMode
	VRROutput string
}

func ResolveSettings(config ProfileConfig, environment []string) (Settings, []error) {
	vrrOutput, outputConfigured := environmentLookup(environment, "GW_VRR_OUTPUT")
	if !outputConfigured && config.VRROutput != nil {
		vrrOutput = *config.VRROutput
	}

	settings := Settings{VRROutput: strings.TrimSpace(vrrOutput)}
	if settings.VRROutput == "" {
		settings.VRROutput = VRROutputDetect
	}

	var errors []error
	vrrValue, vrrConfigured := environmentLookup(environment, "GW_VRR")
	if !vrrConfigured && config.VRR != nil {
		vrrValue = string(*config.VRR)
	}

	vrr, err := vrrPolicy(vrrValue)
	if err != nil {
		errors = append(errors, err)
	} else {
		settings.VRR = vrr
	}

	return settings, errors
}

func vrrPolicy(value string) (VRRMode, error) {
	if value == "" {
		return VRRUnchanged, nil
	}

	mode, err := parseVRRMode(value)
	if err != nil {
		return "", fmt.Errorf("invalid GW_VRR value %q", value)
	}

	return mode, nil
}

func parseVRRMode(value string) (VRRMode, error) {
	switch value {
	case "automatic":
		return VRRAutomatic, nil
	case "never":
		return VRRNever, nil
	case "always":
		return VRRAlways, nil
	case "unchanged":
		return VRRUnchanged, nil
	default:
		return "", fmt.Errorf("config.vrr must be \"automatic\", \"never\", \"always\", or \"unchanged\"")
	}
}
