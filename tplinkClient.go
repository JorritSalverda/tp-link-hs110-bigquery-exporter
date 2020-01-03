package main

import (
	"encoding/json"
	"net"
	"sync"
	"time"
)

// TPLinkClient is the interface for querying tp-link smart home devices
type TPLinkClient interface {
	DiscoverDevices(timeout int) ([]Device, error)
	GetUsageForDevice(device Device, timeout int) (Device, error)
	GetUsageForAllDevices(devices []Device, timeout int) ([]Device, error)
}

type tplinkClient struct {
}

// NewTPLinkClient returns new TPLinkClient
func NewTPLinkClient() (TPLinkClient, error) {
	return &tplinkClient{}, nil
}

// DiscoverDevices retrieves all tp-link smart home devices
func (tpc *tplinkClient) DiscoverDevices(timeout int) (devices []Device, err error) {

	request := DeviceInfoRequest{}

	broadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:9999")
	if err != nil {
		return
	}

	fromAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8755")
	if err != nil {
		return
	}

	sock, err := net.ListenUDP("udp", fromAddr)
	defer sock.Close()
	if err != nil {
		return
	}
	sock.SetReadBuffer(2048)

	r := make(chan Device)

	go func(s *net.UDPConn, request interface{}) {
		for {
			buff := make([]byte, 2048)
			rlen, addr, err := s.ReadFromUDP(buff)
			if err != nil {
				break
			}
			go discovered(r, addr, rlen, buff)
		}
	}(sock, request)

	eJSON, err := json.Marshal(&request)
	if err != nil {
		return
	}

	var encrypted = encrypt(eJSON)
	_, err = sock.WriteToUDP(encrypted, broadcastAddr)
	if err != nil {
		return
	}
	started := time.Now()
Q:
	for {
		select {
		case x := <-r:

			info := DeviceInfoResponse{}
			json.Unmarshal(x.Data, &info)

			x.Info = &info

			devices = append(devices, x)
		default:
			if now := time.Now(); now.Sub(started) >= time.Duration(timeout)*time.Second {
				break Q
			}
		}
	}
	return devices, nil
}

func (tpc *tplinkClient) GetUsageForDevice(device Device, timeout int) (Device, error) {

	request := DeviceInfoRequest{}

	eJSON, err := json.Marshal(&request)
	if err != nil {
		return device, err
	}

	response, err := sendCommand(device.Addr.IP.String(), device.Addr.Port, eJSON, timeout)
	if err != nil {
		return device, err
	}

	err = json.Unmarshal([]byte(response), &device.Info)
	if err != nil {
		return device, err
	}

	return device, nil
}

func (tpc *tplinkClient) GetUsageForAllDevices(devices []Device, timeout int) (updatedDevices []Device, err error) {

	var wg sync.WaitGroup
	wg.Add(len(devices))

	devicesChannel := make(chan Device, len(devices))
	errors := make(chan error, len(devices))

	for _, d := range devices {
		go func(d Device) {
			defer wg.Done()
			d, err := tpc.GetUsageForDevice(d, timeout)
			if err != nil {
				errors <- err
				return
			}
			devicesChannel <- d
		}(d)
	}

	wg.Wait()

	close(errors)
	for e := range errors {
		return nil, e
	}

	close(devicesChannel)
	for d := range devicesChannel {
		updatedDevices = append(updatedDevices, d)
	}

	return updatedDevices, nil
}
