package config

import (
	"reflect"
	"testing"
)

func TestExpandEnvironmentTokens(t *testing.T) {
	t.Setenv("HOME", "/home/tester")
	values := map[string]string{
		"CONFIG":  "@home/.config/gamewrap",
		"DATA":    "@target-dir/data",
		"LITERAL": "prefix/@home/path",
	}

	got, err := ExpandEnvironmentTokens(values, "/games/example")
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"CONFIG":  "/home/tester/.config/gamewrap",
		"DATA":    "/games/example/data",
		"LITERAL": "prefix/@home/path",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("environment = %v, want %v", got, want)
	}
}

func TestExpandEnvironmentTokensPreservesWinePathSeparators(t *testing.T) {
	got, err := ExpandEnvironmentTokens(map[string]string{"DATA": "@target-dir/data/cache"}, `C:\Games\Example`)
	if err != nil {
		t.Fatal(err)
	}

	if got["DATA"] != `C:\Games\Example\data\cache` {
		t.Fatalf("DATA = %q", got["DATA"])
	}
}

func TestExecutableOverrideRejectsPathTokens(t *testing.T) {
	if _, err := ExpandEnvironmentTokens(map[string]string{"GW_EXECUTABLE": "@target-dir/game"}, "/games/example"); err == nil {
		t.Fatal("expected GW_EXECUTABLE token error")
	}
}
