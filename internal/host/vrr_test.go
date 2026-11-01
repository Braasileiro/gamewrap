package host

import (
	"reflect"
	"strings"
	"testing"

	"gamewrap/internal/config"
)

func TestEnableVRRUsesResolvedAutomaticOutput(t *testing.T) {
	runner := &fakeCommandRunner{Results: map[string]fakeCommandResult{
		commandKey(kscreenCtl, config.VRROutputActive, "--json"): {
			Output: []byte(`{"name":"DP-3","capabilities":["vrr","hdr"],"vrrPolicy":"never"}`),
		},
	}}
	s := &Session{
		log:      discardLogger(),
		settings: config.Settings{VRR: "automatic", VRROutput: config.VRROutputActive},
		commands: runner,
	}

	s.enableVRR()

	if !s.vrrChanged || s.vrrOutput != "DP-3" || s.vrrOriginal != "never" {
		t.Fatalf("VRR session state = %+v", s)
	}

	wantCalls := []commandCall{
		{Name: kscreenCtl, Args: []string{config.VRROutputActive, "--json"}},
		{Name: kscreenCtl, Args: []string{"DP-3", "set-vrr-policy", "automatic"}},
	}

	if !reflect.DeepEqual(runner.Calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runner.Calls, wantCalls)
	}
}

func TestGetVRRRequiresVRRCapability(t *testing.T) {
	runner := &fakeCommandRunner{Results: map[string]fakeCommandResult{
		commandKey(kscreenCtl, "HDMI-A-1", "--json"): {
			Output: []byte(`{"name":"HDMI-A-1","capabilities":["icc"],"vrrPolicy":"never"}`),
		},
	}}

	_, _, err := getVRR(runner, "HDMI-A-1")
	if err == nil || !strings.Contains(err.Error(), "does not support VRR") {
		t.Fatalf("error = %v, want unsupported VRR error", err)
	}
}
