package main

import (
	"reflect"
	"testing"
)

func TestParseInvocationOptions(t *testing.T) {
	options, args, err := parseInvocationOptions([]string{
		"--hide-steam",
		"--env", "FIRST=one",
		"--env", "FIRST=two=parts",
		"--",
		"game",
		"--game-option",
	})
	if err != nil {
		t.Fatal(err)
	}

	if options.hideSteam == nil || !*options.hideSteam {
		t.Fatal("hide Steam should be enabled")
	}

	if options.environment["FIRST"] != "two=parts" {
		t.Fatalf("environment = %v", options.environment)
	}

	if len(args) != 2 || args[0] != "game" || args[1] != "--game-option" {
		t.Fatalf("remaining arguments = %v", args)
	}
}

func TestParseInvocationOptionsRequiresSeparator(t *testing.T) {
	if _, _, err := parseInvocationOptions([]string{"game"}); err == nil {
		t.Fatal("expected missing separator error")
	}
}

func TestResolvedUtilityEnvironmentPrecedence(t *testing.T) {
	inherited := []string{"INHERITED=kept", "SHARED=inherited"}
	profile := map[string]string{"PROFILE_ONLY": "enabled", "SHARED": "profile"}
	command := map[string]string{"COMMAND_ONLY": "enabled", "SHARED": "command"}

	got := resolvedUtilityEnvironment(inherited, profile, command)
	want := []string{
		"INHERITED=kept",
		"PROFILE_ONLY=enabled",
		"COMMAND_ONLY=enabled",
		"SHARED=command",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("utility environment = %v, want %v", got, want)
	}
}
