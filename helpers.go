package main

import (
	"net"
	"time"
)

func encrypt(request []byte, key uint8) []byte {
	result := make([]byte, 4+len(request))
	result[0] = 0x0
	result[1] = 0x0
	result[2] = 0x0
	result[3] = 0x0
	for i, c := range request {
		var a = key ^ uint8(c)
		key = uint8(a)
		result[i+4] = a
	}
	return result
}

func decrypt(request []byte, key uint8) []byte {
	result := make([]byte, len(request))
	for i, c := range request {
		var a = key ^ uint8(c)
		key = uint8(c)
		result[i] = a
	}
	return result
}

func discovered(r chan Device, addr *net.UDPAddr, rlen int, buff []byte) {
	r <- Device{
		Addr: addr,
		Data: decrypt(buff[:rlen], 171),
	}
}

// mapDevicesToBigQueryMeasurement converts device information into a bigquery measurement and returns nil if the devices don't have sufficient information
func mapDevicesToBigQueryMeasurement(devices []Device) *BigQueryMeasurement {
	if len(devices) == 0 {
		return nil
	}

	measurement := BigQueryMeasurement{
		InsertedAt: time.Now().UTC(),
	}

	for _, d := range devices {
		if d.Info != nil && d.Info.System != nil && d.Info.EMeter != nil {
			measurement.SmartPlugs = append(measurement.SmartPlugs, BigQuerySmartPlug{
				Name:                  d.Info.System.Info.Alias,
				CurrentPowerUsageWatt: float64(d.Info.EMeter.RealTime.PowerMilliWatt / 1000),
			})
		}
	}

	if len(measurement.SmartPlugs) == 0 {
		return nil
	}

	return &measurement
}
