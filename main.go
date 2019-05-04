package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func main() {

	if len(os.Args) != 2 {
		log.Fatal(errors.New("please give the machine name"))
	}

	content, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err)
	}

	cfg := config{}
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	mac := ""
	for _, v := range cfg.Machines {
		if v.Name == os.Args[1] {
			mac = v.MAC
		}
	}

	if mac == "" {
		log.Fatal(errors.New("machine not found in config"))
	}

	mp, err := newMagicPacket(mac)
	if err != nil {
		log.Fatal(err)
	}

	err = sendUDPPacket(mp, cfg.Broadcast)
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	Broadcast string    `json:"broadcast"`
	Machines  []machine `json:"machines"`
}

type machine struct {
	Name string `json:"name"`
	MAC  string `json:"mac"`
}

type magicPacket [102]byte

func newMagicPacket(macAddr string) (packet magicPacket, err error) {
	mac, err := net.ParseMAC(macAddr)
	if err != nil {
		return packet, err
	}

	if len(mac) != 6 {
		return packet, errors.New("invalid EUI-48 MAC address")
	}

	// write magic bytes to packet
	copy(packet[0:], []byte{255, 255, 255, 255, 255, 255})
	offset := 6

	for i := 0; i < 16; i++ {
		copy(packet[offset:], mac)
		offset += 6
	}

	return packet, nil
}

func sendUDPPacket(mp magicPacket, addr string) (err error) {
	conn, err := net.Dial("udp", addr+":9")
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(mp[:])
	return err
}
