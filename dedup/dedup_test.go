package dedup

import (
	. "encoding/binary"
	"net"
	"testing"
	"time"

	. "github.com/google/gopacket/layers"
)

func m(mac string) net.HardwareAddr {
	m, _ := net.ParseMAC(mac)
	return m
}

func key2mac(key uint32) string {
	m := [6]byte{}
	BigEndian.PutUint32(m[2:], key)
	return net.HardwareAddr(m[:]).String()
}

func buildStubPacket(da, sa string, ethType EthernetType, payload uint32) []byte {
	packet := [128]byte{}
	copy(packet[:], m(da))
	copy(packet[6:], m(sa))
	BigEndian.PutUint16(packet[12:], uint16(ethType))
	BigEndian.PutUint32(packet[56:], payload)
	return packet[:]
}

func buildStubVlanTaggedPacket(da, sa string, vid uint16, ethType EthernetType, payload uint32) []byte {
	packet := [128]byte{}
	copy(packet[:], m(da))
	copy(packet[6:], m(sa))
	BigEndian.PutUint16(packet[12:], uint16(EthernetTypeDot1Q))
	BigEndian.PutUint16(packet[14:], vid&0xFFF)
	BigEndian.PutUint32(packet[60:], payload)
	return packet[:]
}

func TestMatched(t *testing.T) {
	da := "00:00:00:35:02:b0"
	sa := "00:00:00:fc:a4:0b"
	packet := buildStubPacket(da, sa, EthernetTypeIPv4, 1)

	if Lookup(packet, 0) {
		t.Error("Should not match")
	}

	if !Lookup(packet, 0) {
		t.Error("Should match")
	}

	if Lookup(packet, 0) {
		t.Error("Should not match")
	}
}

func TestMultipleFlowsOnSameDirection(t *testing.T) {
	da := "00:00:00:fc:a4:0b"
	sa := "00:00:00:25:3f:63"
	packet := buildStubPacket(da, sa, EthernetTypeIPv4, 1)
	packet2 := buildStubPacket(da, sa, EthernetTypeARP, 2)

	Lookup(packet, 0)
	Lookup(packet2, 0)

	if !Lookup(packet, 0) {
		t.Error("Should not match")
	}

	if !Lookup(packet2, 0) {
		t.Error("Should not match")
	}
}

func TestPacketLoss(t *testing.T) {
	da := "00:00:00:fc:a4:0b"
	sa := "00:00:00:25:3f:63"
	packet := buildStubPacket(da, sa, EthernetTypeIPv4, 1)
	packet2 := buildStubPacket(da, sa, EthernetTypeIPv4, 2)

	Lookup(packet, 0)
	Lookup(packet2, 0)

	if !Lookup(packet2, 0) {
		t.Error("Should not match")
	}
}

func TestHashCollision(t *testing.T) {
	da := "00:00:00:fc:a4:0b"
	sa := "00:00:00:25:3f:63"
	packet1 := buildStubPacket(da, sa, EthernetTypeIPv4, 0)
	BigEndian.PutUint32(packet1[48:], 1)

	packet2 := buildStubPacket(da, sa, EthernetTypeIPv4, 0)
	BigEndian.PutUint32(packet1[80:], 1790366114)

	if Lookup(packet1, 0) {
		t.Error("Should not hit")
	}
	if Lookup(packet2, 0) {
		t.Error("Should not hit")
	}
}

func TestVlanTagged(t *testing.T) {
	da := "00:00:00:fc:a4:0b"
	sa := "00:00:00:25:3f:63"
	packet1 := buildStubVlanTaggedPacket(da, sa, 1, EthernetTypeIPv4, 1)
	packet2 := buildStubVlanTaggedPacket(da, sa, 2, EthernetTypeIPv4, 1)
	if Lookup(packet1, 0) {
		t.Error("Should not hit")
	}
	if !Lookup(packet2, 0) {
		t.Error("Should hit")
	}
}

func TestChecksum(t *testing.T) {
	da := "00:00:00:fc:a4:0b"
	sa := "00:00:00:25:3f:63"
	packet := buildStubPacket(da, sa, EthernetTypeIPv4, 1)
	packet[14] = 5 // ihl
	packet[23] = byte(IPProtocolUDP)
	BigEndian.PutUint16(packet[40:], 0x0101)
	Lookup(packet, 0)
	BigEndian.PutUint16(packet[40:], 0x1010)
	if !Lookup(packet, 0) {
		t.Error("Should hit")
	}
}

func TestTimeout(t *testing.T) {
	da := "00:00:00:fc:a4:0b"
	sa := "00:00:00:25:3f:63"
	packet := buildStubPacket(da, sa, EthernetTypeIPv4, 1)
	Lookup(packet, 0)

	if Lookup(packet, 110*time.Millisecond) {
		t.Error("Should not hit")
	}
}

func TestOverLimit(t *testing.T) {
	da := "00:00:00:fc:a4:0b"
	sa := "00:00:00:25:3f:63"
	for i := 0; i <= ELEMENTS_LIMIT+1; i++ {
		Lookup(buildStubPacket(da, sa, EthernetTypeIPv4, uint32(i)), 0)
	}

	first := buildStubPacket(da, sa, EthernetTypeIPv4, 0)
	if Lookup(first, 0) {
		t.Error("Should hit")
	}

	middle := buildStubPacket(da, sa, EthernetTypeIPv4, 500)
	if !Lookup(middle, 0) {
		t.Error("Should hit")
	}
}
