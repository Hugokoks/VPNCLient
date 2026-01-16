package vna

type PacketType byte

const (
	PacketIPRequest    PacketType = 1
	PacketIPResponse   PacketType = 2
	PacketHandshakeReq PacketType = 3
	PacketHandshakeRes PacketType = 4
	PacketData         PacketType = 5
)

func buildPacket(t PacketType, payload []byte) []byte {
	pkt := make([]byte, 1+len(payload))
	pkt[0] = byte(t)
	copy(pkt[1:], payload)
	return pkt
}
