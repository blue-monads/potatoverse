package funnel

import (
	"io"
	"net"
)

// responseReader reads response body from packets
type responseReader struct {
	conn     net.Conn
	total    int64
	received int64
	buffer   []byte
}

func (r *responseReader) Read(p []byte) (int, error) {
	// If we have buffered data, return it first
	if len(r.buffer) > 0 {
		n := copy(p, r.buffer)
		r.buffer = r.buffer[n:]
		r.received += int64(n)
		return n, nil
	}

	// Read next packet
	packet, err := ReadPacket(r.conn)
	if err != nil {
		return 0, err
	}

	if packet.PType != PtypeSendBody && packet.PType != PtypeEndBody {
		return 0, io.ErrUnexpectedEOF
	}

	// Copy data to buffer
	n := copy(p, packet.Data)
	r.received += int64(n)

	// If there's remaining data, buffer it
	if n < len(packet.Data) {
		r.buffer = packet.Data[n:]
	}

	// Check if we're done
	if packet.PType == PtypeEndBody || (r.total > 0 && r.received >= r.total) {
		if len(r.buffer) == 0 {
			return n, io.EOF
		}
	}

	return n, nil
}

func (r *responseReader) Close() error {
	return nil
}
