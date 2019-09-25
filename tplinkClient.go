package main

import (
	"encoding/json"
	"net"
	"time"
)

// TPLinkClient is the interface for querying tp-link smart home devices
type TPLinkClient interface {
	DiscoverDevices(timeout int) ([]Device, error)
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

	BroadcastAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:9999")
	if err != nil {
		return
	}

	FromAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:8755")
	if err != nil {
		return
	}

	sock, err := net.ListenUDP("udp", FromAddr)
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

	var encrypted = encrypt(eJSON, 171)[4:]
	_, err = sock.WriteToUDP(encrypted, BroadcastAddr)
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
