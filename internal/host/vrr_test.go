package host

import (
	"reflect"
	"strings"
	"testing"

	"gamewrap/internal/config"
)

func TestSelectVRROutputRejectsAmbiguousDetection(t *testing.T) {
	never := 0
	outputs := []displayOutput{
		{Name: "DP-3", Connected: true, Enabled: true, VRRPolicy: &never},
		{Name: "DP-4", Connected: true, Enabled: true, VRRPolicy: &never},
	}

	_, err := selectVRROutput(outputs, config.VRROutputDetect)
	if err == nil || !strings.Contains(err.Error(), "DP-3, DP-4") {
		t.Fatalf("error = %v, want ambiguous outputs", err)
	}
}

func TestEnableVRRUsesResolvedAutomaticOutput(t *testing.T) {
	runner := &fakeCommandRunner{Results: map[string]fakeCommandResult{
		commandKey(kscreenDoctor, "--json"): {
			Output: []byte(`{"outputs":[{"name":"DP-3","connected":true,"enabled":true,"vrrPolicy":0},{"name":"HDMI-A-1","connected":true,"enabled":false}]}`),
		},
	}}
	s := &Session{
		log:      discardLogger(),
		settings: config.Settings{VRR: "automatic", VRROutput: config.VRROutputDetect},
		commands: runner,
	}

	s.enableVRR()

	if !s.vrrChanged || s.vrrOutput != "DP-3" || s.vrrOriginal != "never" {
		t.Fatalf("VRR session state = %+v", s)
	}

	wantCalls := []commandCall{
		{Name: kscreenDoctor, Args: []string{"--json"}},
		{Name: kscreenDoctor, Args: []string{"output.DP-3.vrrpolicy.automatic"}},
	}

	if !reflect.DeepEqual(runner.Calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runner.Calls, wantCalls)
	}
}
