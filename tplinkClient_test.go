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
		devices, err := client.DiscoverDevices(10)

		if assert.Nil(t, err) {
			if assert.Equal(t, 16, len(devices)) {
				assert.Equal(t, "Oven", devices[0].Info.System.Info.Alias)
				assert.GreaterOrEqual(t, float32(315), devices[0].Info.EMeter.RealTime.PowerMilliWatt)
				assert.GreaterOrEqual(t, float64(0.315), float64(devices[0].Info.EMeter.RealTime.PowerMilliWatt/1000))
			}
		}
	})
}

func TestGetUsageForDevice(t *testing.T) {

	t.Run("ReturnsMeteringDataForSingleDevice", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client, _ := NewTPLinkClient()
		devices, err := client.DiscoverDevices(2)

		// act
		device, err := client.GetUsageForDevice(devices[0], 2)

		if assert.Nil(t, err) {
			assert.Equal(t, "Espresso", device.Info.System.Info.Alias)
			assert.Equal(t, float32(315), device.Info.EMeter.RealTime.PowerMilliWatt)
			assert.Equal(t, float64(0.315), float64(device.Info.EMeter.RealTime.PowerMilliWatt/1000))
		}
	})
}

func TestGetUsageForAllDevices(t *testing.T) {

	t.Run("ReturnsAllSmartHomeDevicesWithMeteringData", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		client, _ := NewTPLinkClient()
		devices, err := client.DiscoverDevices(10)

		// act
		devices, err = client.GetUsageForAllDevices(devices, 2)

		if assert.Nil(t, err) {
			assert.Equal(t, 16, len(devices))
			assert.Equal(t, "Espresso", devices[0].Info.System.Info.Alias)
			assert.Equal(t, float32(315), devices[0].Info.EMeter.RealTime.PowerMilliWatt)
			assert.Equal(t, float64(0.315), float64(devices[0].Info.EMeter.RealTime.PowerMilliWatt/1000))
		}
	})
}
