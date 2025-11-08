package mcphub

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type PotatoTransport struct {
	r *bufio.Reader
	w io.Writer
}

func NewPotatoTransport(r io.Reader, w io.Writer) *PotatoTransport {
	return &PotatoTransport{
		r: bufio.NewReader(r),
		w: w,
	}
}

func (t *PotatoTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return &potatoConn{
		r: t.r,
		w: t.w,
	}, nil
}

// Connection for PotatoTransport
type potatoConn struct {
	r *bufio.Reader
	w io.Writer
}

func (t *potatoConn) Read(context.Context) (jsonrpc.Message, error) {
	data, err := t.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	return jsonrpc.DecodeMessage(data[:len(data)-1])
}

func (t *potatoConn) Write(_ context.Context, msg jsonrpc.Message) error {
	data, err := jsonrpc.EncodeMessage(msg)
	if err != nil {
		return err
	}

	_, err1 := t.w.Write(data)
	_, err2 := t.w.Write([]byte{'\n'})
	return errors.Join(err1, err2)
}

func (t *potatoConn) Close() error {
	return nil
}

func (t *potatoConn) SessionID() string {
	return ""
}
