package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// sendCommand is based on https://github.com/jaedle/golang-tplink-hs100
func sendCommand(address string, port int, command []byte, timeoutSeconds int) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", address+":"+strconv.Itoa(port), time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	_, err = writer.Write(encryptWithHeader(command))
	if err != nil {
		return nil, err
	}
	writer.Flush()

	response, err := readHeader(conn)
	if err != nil {
		return nil, err
	}

	payload, err := readPayload(conn, payloadLength(response))
	if err != nil {
		return nil, err
	}

	return decrypt(payload), nil
}

const headerLength = 4

func readHeader(conn net.Conn) ([]byte, error) {
	headerReader := io.LimitReader(conn, int64(headerLength))
	var response = make([]byte, headerLength)
	_, err := headerReader.Read(response)
	return response, err
}

func readPayload(conn net.Conn, length uint32) ([]byte, error) {
	payloadReader := io.LimitReader(conn, int64(length))
	var payload = make([]byte, length)
	_, err := payloadReader.Read(payload)
	return payload, err
}

func payloadLength(header []byte) uint32 {
	payloadLength := binary.BigEndian.Uint32(header)
	return payloadLength
}

const lengthHeader = 4

func encrypt(input []byte) []byte {
	s := string(input)

	key := byte(0xAB)
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = s[i] ^ key
		key = b[i]
	}
	return b
}

func encryptWithHeader(input []byte) []byte {
	s := string(input)

	lengthPayload := len(s)
	b := make([]byte, lengthHeader+lengthPayload)
	copy(b[:lengthHeader], header(lengthPayload))
	copy(b[lengthHeader:], encrypt(input))
	return b
}

func header(lengthPayload int) []byte {
	h := make([]byte, lengthHeader)
	binary.BigEndian.PutUint32(h, uint32(lengthPayload))
	return h
}

func decrypt(b []byte) []byte {
	k := byte(0xAB)
	var newKey byte
	for i := 0; i < len(b); i++ {
		newKey = b[i]
		b[i] = b[i] ^ k
		k = newKey
	}

	return b
}

func decryptWithHeader(b []byte) []byte {
	return decrypt(payload(b))
}

func payload(b []byte) []byte {
	return b[lengthHeader:]
}

func discovered(r chan Device, addr *net.UDPAddr, rlen int, buff []byte) {
	r <- Device{
		Addr: addr,
		Data: decrypt(buff[:rlen]),
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
		if d.Info != nil && d.Info.System != nil && d.Info.EMeter != nil && !strings.HasPrefix(d.Info.System.Info.Alias, "TP-LINK") {
			measurement.SmartPlugs = append(measurement.SmartPlugs, BigQuerySmartPlug{
				Name:                  d.Info.System.Info.Alias,
				CurrentPowerUsageWatt: float64(d.Info.EMeter.RealTime.PowerMilliWatt / 1000),
				TotalWattSecond:       float64(d.Info.EMeter.RealTime.TotalWattHour * 3600),
			})
		}
	}

	if len(measurement.SmartPlugs) == 0 {
		return nil
	}

	return &measurement
}
