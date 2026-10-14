package invocation

import (
	"reflect"
	"testing"
)

func TestUtilityEnvironmentDisablesManagementAndMangoHud(t *testing.T) {
	environment := []string{
		"PATH=/usr/bin",
		"GW_GAMEMODE=1",
		"GW_MANGOHUD=wrapper",
		"GW_VRR=automatic",
		"MANGOHUD=1",
		"MANGOHUD_CONFIG=fps_limit=60",
		"LD_PRELOAD=libMangoHud.so",
		"VK_INSTANCE_LAYERS=VK_LAYER_MANGOHUD_overlay",
	}

	got := UtilityEnvironment(environment)
	want := []string{"PATH=/usr/bin", "MANGOHUD=0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("utility environment = %v, want %v", got, want)
	}
}
