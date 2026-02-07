package funnel

import (
	"io"
	"net"

	"github.com/blue-monads/potatoverse/backend/services/buddyhub-poc/packetwire"
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

	// If total is 0 and we haven't received anything, we need to read the EndBody packet
	// and return EOF immediately
	if r.total == 0 && r.received == 0 {
		packet, err := packetwire.ReadPacket(r.conn)
		if err != nil {
			return 0, err
		}
		if packet.PType != packetwire.PtypeEndBody {
			return 0, io.ErrUnexpectedEOF
		}
		// Consumed the EndBody packet, return EOF
		return 0, io.EOF
	}

	// Read next packet
	packet, err := packetwire.ReadPacket(r.conn)
	if err != nil {
		return 0, err
	}

	if packet.PType != packetwire.PtypeSendBody && packet.PType != packetwire.PtypeEndBody {
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
	if packet.PType == packetwire.PtypeEndBody || (r.total > 0 && r.received >= r.total) {
		if len(r.buffer) == 0 {
			return n, io.EOF
		}
	}

	return n, nil
}

func (r *responseReader) Close() error {
	return nil
}
