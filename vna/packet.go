package vna

func buildPacket(t PacketType, payload []byte) []byte {
	pkt := make([]byte, 1+len(payload))
	pkt[0] = byte(t)
	copy(pkt[1:], payload)
	return pkt
}
