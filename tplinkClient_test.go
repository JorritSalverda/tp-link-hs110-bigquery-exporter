package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverDevices(t *testing.T) {

	t.Run("ReturnsAllSmartHomeDevices", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client, _ := NewTPLinkClient()

		// act
		devices, err := client.DiscoverDevices(2)

		if assert.Nil(t, err) {
			assert.Equal(t, 1, len(devices))
			assert.Equal(t, "Espresso", devices[0].Info.System.Info.Alias)
			assert.Equal(t, float32(315), devices[0].Info.EMeter.RealTime.PowerMilliWatt)
			assert.Equal(t, float64(0.315), float64(devices[0].Info.EMeter.RealTime.PowerMilliWatt/1000))
		}
	})
}
