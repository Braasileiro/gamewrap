package config

import "fmt"

type MangoHudMode string

const (
	MangoHudDisabled MangoHudMode = "disabled"
	MangoHudEnv      MangoHudMode = "env"
	MangoHudWrapper  MangoHudMode = "wrapper"
)

func parseMangoHudMode(value string) (MangoHudMode, error) {
	switch value {
	case "disabled":
		return MangoHudDisabled, nil
	case "env":
		return MangoHudEnv, nil
	case "wrapper":
		return MangoHudWrapper, nil
	default:
		return "", fmt.Errorf("config.mangohud must be \"disabled\", \"env\", or \"wrapper\"")
	}
}
