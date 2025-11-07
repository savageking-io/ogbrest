package packet

import (
	"encoding/binary"
	"fmt"
)

const MagicJson = 0x1235
const MagicProtobuf = 0x1236

type Packet struct {
	Magic       uint16 // 2 bytes
	ServiceId   uint16 // 2 bytes
	PayloadSize uint16 // 2 bytes
	UserId      uint64 // 8 bytes
	Header      []byte // 2 bytes
	Payload     []byte
}

func NewPacketJson(serviceId uint16, userId uint64, header []byte, payload []byte) (*Packet, error) {
	return &Packet{
		Magic:       MagicJson,
		ServiceId:   serviceId,
		PayloadSize: uint16(len(payload)),
		UserId:      userId,
		Header:      header,
		Payload:     payload,
	}, nil
}

func NewPacketProtobuf(serviceId uint16, userId uint64, header []byte, payload []byte) (*Packet, error) {
	return &Packet{
		Magic:       MagicProtobuf,
		ServiceId:   serviceId,
		PayloadSize: uint16(len(payload)),
		UserId:      userId,
		Header:      header,
		Payload:     payload,
	}, nil
}

func Marshal(p *Packet, out []byte) error {
	if p == nil {
		return fmt.Errorf("packet is nil")
	}

	out = make([]byte, 0)
	binary.BigEndian.PutUint16(out[0:2], p.Magic)
	binary.BigEndian.PutUint16(out[2:4], p.ServiceId)
	binary.BigEndian.PutUint16(out[4:6], p.PayloadSize)
	binary.BigEndian.PutUint64(out[6:14], p.UserId)
	copy(out[14:16], p.Header)
	copy(out[16:], p.Payload)

	return nil
}

func Unmarshal(in []byte) (*Packet, error) {
	if len(in) < 18 {
		return nil, fmt.Errorf("packet is too small")
	}

	p := &Packet{}
	p.Magic = binary.BigEndian.Uint16(in[0:2])
	if p.Magic != MagicJson && p.Magic != MagicProtobuf {
		return nil, fmt.Errorf("invalid packet")
	}
	p.ServiceId = binary.BigEndian.Uint16(in[2:4])
	p.PayloadSize = binary.BigEndian.Uint16(in[4:6])
	p.UserId = binary.BigEndian.Uint64(in[6:14])
	p.Header = make([]byte, 2)
	copy(p.Header, in[14:16])
	p.Payload = make([]byte, p.PayloadSize)
	copy(p.Payload, in[16:])

	return p, nil
}
