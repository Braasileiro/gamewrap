package config

import "testing"

func TestResolveSettingsUsesActiveOutputByDefault(t *testing.T) {
	settings, resolveErrors := ResolveSettings(ProfileConfig{}, nil)
	if len(resolveErrors) != 0 {
		t.Fatalf("errors = %v", resolveErrors)
	}

	if settings.VRROutput != VRROutputActive {
		t.Fatalf("VRROutput = %q, want %q", settings.VRROutput, VRROutputActive)
	}
}
