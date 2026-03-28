package packetwire

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"io"
	"net"

	nanoid "github.com/jaevor/go-nanoid"
)

var crcTable = crc64.MakeTable(crc64.ISO)

type PacketType = uint8

const (
	PTypeSendHeader        PacketType = iota
	PtypeSendBody          PacketType = iota
	PtypeEndBody           PacketType = iota
	PtypeReSendBody        PacketType = iota
	PtypeWebSocketBinData  PacketType = iota
	PtypeWebSocketTextData PacketType = iota
	PtypeWebSocketPing     PacketType = iota
	PtypeWebSocketPong     PacketType = iota
	PtypeEndSocket         PacketType = iota
	PtypeKcpUpgrade        PacketType = iota
)

type KCPUpgradePacket struct {
	Port       int32  `json:"port"`
	Token      string `json:"token"`
	DirectHost string `json:"direct_host"`
}

func (p *KCPUpgradePacket) Encode() []byte {
	res, _ := json.Marshal(p)
	return res
}

func DecodeKCPUpgradePacket(data []byte) (*KCPUpgradePacket, error) {
	var p KCPUpgradePacket
	err := json.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type Packet struct {
	PType  PacketType
	Offset int32 // current offset
	Total  int32 // total body size
	Data   []byte
}

const FragmentSize = 1024 * 512

func WritePacketFull(conn net.Conn, packet *Packet, reqId string) error {

	_, err := conn.Write([]byte(reqId))
	if err != nil {
		return err
	}

	return WritePacket(conn, packet)
}

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

	// write checksum
	checksumBytes := make([]byte, 8)
	checksum := crc64.Checksum(packet.Data, crcTable)
	binary.BigEndian.PutUint64(checksumBytes, checksum)
	_, err = conn.Write(checksumBytes)
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

	// read checksum
	checksumBytes := make([]byte, 8)
	_, err = io.ReadFull(conn, checksumBytes)
	if err != nil {
		return nil, err
	}
	expectedChecksum := binary.BigEndian.Uint64(checksumBytes)

	// read data
	dataBytes := make([]byte, length)
	_, err = io.ReadFull(conn, dataBytes)
	if err != nil {
		return nil, err
	}

	actualChecksum := crc64.Checksum(dataBytes, crcTable)
	if actualChecksum != expectedChecksum {
		return nil, fmt.Errorf("data corruption detected: expected checksum %016x, got %016x", expectedChecksum, actualChecksum)
	}

	packet.PType = ptype
	packet.Offset = int32(offset)
	packet.Total = int32(total)
	packet.Data = dataBytes

	return packet, nil
}

var idgen, _ = nanoid.Standard(16)

func GetRequestId() string {
	id := idgen()
	return string(id)
}
