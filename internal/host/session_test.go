package host

import (
	"os"
	"reflect"
	"testing"
)

func TestSessionRestoreUsesResolvedOutputAndCommandRunner(t *testing.T) {
	lock, err := os.CreateTemp(t.TempDir(), "session-lock")
	if err != nil {
		t.Fatal(err)
	}

	runner := &fakeCommandRunner{}
	s := &Session{
		log:         discardLogger(),
		commands:    runner,
		lock:        lock,
		vrrOutput:   "DP-3",
		vrrOriginal: "never",
		vrrChanged:  true,
	}

	s.Restore()

	wantCalls := []commandCall{{Name: kscreenCtl, Args: []string{"DP-3", "set-vrr-policy", "never"}}}
	if !reflect.DeepEqual(runner.Calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", runner.Calls, wantCalls)
	}
}
