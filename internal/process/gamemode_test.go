package process

import (
	"reflect"
	"testing"
)

func TestEnvironmentWithGameModePreload(t *testing.T) {
	environment := []string{"PATH=/usr/bin", "LD_PRELOAD=/steam/overlay.so"}
	want := []string{"PATH=/usr/bin", "LD_PRELOAD=/usr/$LIB/libgamemode.so.0:/steam/overlay.so"}

	got, enabled := environmentWithGameModePreload(environment, true)
	if !enabled || !reflect.DeepEqual(got, want) {
		t.Fatalf("environment = %v, enabled = %t", got, enabled)
	}
}
