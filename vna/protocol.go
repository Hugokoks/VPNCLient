package vna

type PacketType byte

const (
	PacketIPRequest PacketType = 1
	PacketHandshake PacketType = 2
	PacketData      PacketType = 3
)
