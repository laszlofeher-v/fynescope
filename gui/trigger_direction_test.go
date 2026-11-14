package gui

import (
	"testing"

	"fynescope/genericps"
)

// TestTriggerDirectionOptionAliased checks that a saved Falling/Rising direction
// is restored when the driver aliases Enter/Exit to Rising/Falling, as the
// PS2000A API does (PS2000A_ENTER == PS2000A_RISING, PS2000A_EXIT == PS2000A_FALLING).
func TestTriggerDirectionOptionAliased(t *testing.T) {
	saved := [...]genericps.ThresholdDirection{
		genericps.TriggerRising, genericps.TriggerFalling, genericps.TriggerEnter,
		genericps.TriggerExit, genericps.TriggerEnterOrExit,
	}
	defer func() {
		genericps.TriggerRising, genericps.TriggerFalling, genericps.TriggerEnter,
			genericps.TriggerExit, genericps.TriggerEnterOrExit =
			saved[0], saved[1], saved[2], saved[3], saved[4]
		initMaps()
	}()

	genericps.TriggerRising = 2
	genericps.TriggerFalling = 3
	genericps.TriggerEnter = 2
	genericps.TriggerExit = 3
	genericps.TriggerEnterOrExit = 4
	initMaps()

	cases := []struct {
		dir  genericps.ThresholdDirection
		opts []string
		want string
	}{
		{genericps.TriggerFalling, triggerDirectionOptions, failing},
		{genericps.TriggerRising, triggerDirectionOptions, rising},
		{genericps.TriggerFalling, triggerWindowDirectionOptions, exit},
		{genericps.TriggerRising, triggerWindowDirectionOptions, enter},
	}
	for _, c := range cases {
		got, ok := triggerDirectionOption(c.dir, c.opts)
		if !ok || got != c.want {
			t.Errorf("triggerDirectionOption(%d, %v) = %q, %v; want %q", c.dir, c.opts, got, ok, c.want)
		}
	}
	if _, ok := triggerDirectionOption(genericps.TriggerEnterOrExit, triggerDirectionOptions); ok {
		t.Errorf("EnterOrExit should not map to a Rising/Falling option")
	}
}

