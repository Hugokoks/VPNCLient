package vna

type PacketType byte

const (
	PacketIPRequest  PacketType = 1
	PacketIPResponse PacketType = 2
	PacketHandshake  PacketType = 3
	PacketData       PacketType = 4
)

func buildPacket(t PacketType, payload []byte) []byte {
	pkt := make([]byte, 1+len(payload))
	pkt[0] = byte(t)
	copy(pkt[1:], payload)
	return pkt
}
