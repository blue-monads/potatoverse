package funnel

import (
	"encoding/binary"
	"io"
	"net"

	nanoid "github.com/jaevor/go-nanoid"
)

type PacketType = uint8

const (
	PTypeSendHeader PacketType = iota
	PtypeSendBody   PacketType = iota
	PtypeEndBody    PacketType = iota
	PtypeReSendBody PacketType = iota
)

type Packet struct {
	PType  PacketType
	Offset int32 // current offset
	Total  int32 // total body size
	Data   []byte
}

const FragmentSize = 1024 * 512

// WritePacket writes a packet to a net.Conn
func WritePacket(conn net.Conn, packet *Packet) error {
	// write packet type
	_, err := conn.Write([]byte{packet.PType})
	if err != nil {
		return err
	}

	// length, offset, total
	intBytes := make([]byte, 4)

	// write length
	binary.BigEndian.PutUint32(intBytes, uint32(len(packet.Data)))
	_, err = conn.Write(intBytes)
	if err != nil {
		return err
	}

	// write offset
	binary.BigEndian.PutUint32(intBytes, uint32(packet.Offset))
	_, err = conn.Write(intBytes)
	if err != nil {
		return err
	}

	// write total
	binary.BigEndian.PutUint32(intBytes, uint32(packet.Total))
	_, err = conn.Write(intBytes)
	if err != nil {
		return err
	}

	// write data
	totalWritten := 0
	for {
		written, err := conn.Write(packet.Data[totalWritten:])
		if err != nil {
			return err
		}
		totalWritten += written
		if totalWritten >= len(packet.Data) {
			break
		}
	}

	return nil
}

// ReadPacket reads a packet from a net.Conn
func ReadPacket(conn net.Conn) (*Packet, error) {
	packet := &Packet{}
	intBytes := make([]byte, 4)

	// read packet type
	_, err := io.ReadFull(conn, intBytes[:1])
	if err != nil {
		return nil, err
	}

	ptype := uint8(intBytes[0])

	// read length
	_, err = io.ReadFull(conn, intBytes)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(intBytes)

	// read offset
	_, err = io.ReadFull(conn, intBytes)
	if err != nil {
		return nil, err
	}
	offset := binary.BigEndian.Uint32(intBytes)

	// read total
	_, err = io.ReadFull(conn, intBytes)
	if err != nil {
		return nil, err
	}
	total := binary.BigEndian.Uint32(intBytes)

	// read data
	dataBytes := make([]byte, length)
	_, err = io.ReadFull(conn, dataBytes)
	if err != nil {
		return nil, err
	}

	packet.PType = ptype
	packet.Offset = int32(offset)
	packet.Total = int32(total)
	packet.Data = dataBytes

	return packet, nil
}

var idgen, _ = nanoid.ASCII(16)

func GetRequestId() string {
	id := idgen()
	return string(id)
}
